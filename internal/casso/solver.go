package casso

import (
	"fmt"
	"math"
)

type PublicChange struct {
	Variable Variable
	Constant float64
}

type _Tag struct {
	marker _Symbol
	other  _Symbol
}

type _EditInfo struct {
	tag        _Tag
	constraint Constraint
	constant   float64
}

type _VariableData struct {
	constant float64
	symbol   _Symbol
	id       uint8
}

type Solver struct {
	cns                map[Constraint]_Tag
	varData            map[Variable]_VariableData
	varForSymbol       map[_Symbol]Variable
	publicChanges      []PublicChange
	changed            map[Variable]struct{}
	shouldClearChanges bool
	rows               map[_Symbol]_Row
	edits              map[Variable]_EditInfo
	infeasibleRows     []_Symbol
	objective          _Row
	artificial         *_Row
	idTick             uint8
}

func NewSolver() Solver {
	return Solver{
		cns:                make(map[Constraint]_Tag),
		varData:            make(map[Variable]_VariableData),
		varForSymbol:       make(map[_Symbol]Variable),
		publicChanges:      nil,
		changed:            make(map[Variable]struct{}),
		shouldClearChanges: false,
		rows:               make(map[_Symbol]_Row),
		edits:              make(map[Variable]_EditInfo),
		infeasibleRows:     nil,
		objective:          newRow(0),
		artificial:         nil,
		idTick:             1,
	}
}

func (s *Solver) AddConstraints(constraints ...Constraint) error {
	for _, c := range constraints {
		if err := s.AddConstraint(c); err != nil {
			return err
		}
	}

	return nil
}

func (s *Solver) AddConstraint(constraint Constraint) error {
	if _, ok := s.cns[constraint]; ok {
		return ErrDuplicateConstraint
	}

	row, tag := s.createRow(constraint)
	subject := chooseSubject(row, tag)

	if subject.Type == SymbolTypeInvalid && allDummies(row) {
		if !nearZero(row.constant) {
			return ErrUnsatisfiableConstraint
		}

		subject = tag.marker
	}

	// If an entering symbol still isn't found, then the row must
	// be added using an artificial variable. If that fails, then
	// the row represents an unsatisfiable constraint.
	if subject.Type == SymbolTypeInvalid {
		ok, err := s.addWithArtificialVariable(row)
		if err != nil {
			return err
		}

		if !ok {
			return ErrUnsatisfiableConstraint
		}
	} else {
		row.SolveForSymbol(subject)
		s.substitute(subject, row)

		if subject.Type == SymbolTypeExternal && row.constant != 0 {
			v := s.varForSymbol[subject]
			s.varChanged(v)
		}

		s.rows[subject] = row
	}

	s.cns[constraint] = tag

	if err := s.optimize(&s.objective); err != nil {
		return err
	}

	return nil
}

// FetchChanges fetches all changes to the values of variables since the last call to this function.
//
// The list of changes returned is not in a specific order. Each change comprises the variable changed and
// the new value of that variable.
func (s *Solver) FetchChanges() []PublicChange {
	if s.shouldClearChanges {
		clear(s.changed)
		s.shouldClearChanges = false
	} else {
		s.shouldClearChanges = true
	}

	clear(s.publicChanges)

	for v := range s.changed {
		if varData, ok := s.varData[v]; ok {
			var newValue float64

			if row, ok := s.rows[varData.symbol]; ok {
				newValue = row.constant
			}

			oldValue := varData.constant

			if oldValue != newValue {
				s.publicChanges = append(s.publicChanges, PublicChange{
					Variable: v,
					Constant: newValue,
				})

				varData.constant = newValue
				s.varData[v] = varData
			}
		}
	}

	return s.publicChanges
}

func (s *Solver) GetValue(v Variable) float64 {
	if data, ok := s.varData[v]; ok {
		if row, ok := s.rows[data.symbol]; ok {
			return row.constant
		}
	}

	return 0
}

func (s *Solver) Reset() {
	clear(s.rows)
	clear(s.cns)
	clear(s.varData)
	clear(s.varForSymbol)
	clear(s.changed)

	s.shouldClearChanges = false

	clear(s.edits)
	clear(s.infeasibleRows)

	s.objective = newRow(0)
	s.artificial = nil
	s.idTick = 1
}

func ptr[T any](value T) *T {
	return &value
}

func (s *Solver) addWithArtificialVariable(row _Row) (bool, error) {
	// Create and add the artificial variable to the tableau
	art := _Symbol{Value: s.idTick, Type: SymbolTypeSlack}
	s.idTick++
	s.rows[art] = row.Clone()
	s.artificial = ptr(row.Clone())

	// Optimize the artificial objective. This is successful
	// only if the artificial objective is optimized to zero.
	if err := s.optimize(s.artificial); err != nil {
		return false, fmt.Errorf("optimize: %w", err)
	}

	success := nearZero(s.artificial.constant)
	s.artificial = nil

	if row, ok := s.rows[art]; ok {
		delete(s.rows, art)

		if len(row.cells) == 0 {
			return success, nil
		}

		entering := anyPivotableSymbol(row)
		if entering.Type == SymbolTypeInvalid {
			return false, nil
		}

		row.SolveForSymbols(art, entering)
		s.substitute(entering, row)
		s.rows[entering] = row
	}

	// Remove the artificial row from the tableau
	for symbol, row := range s.rows {
		row.Remove(art)

		s.rows[symbol] = row
	}

	s.objective.Remove(art)

	return success, nil
}

func (s *Solver) optimize(objective *_Row) error {
	for {
		entering := getEnteringSymbol(*objective)
		if entering.Type == SymbolTypeInvalid {
			return nil
		}

		leaving, row, ok := s.getLeavingRow(entering)
		if !ok {
			return InternalSolverError("unbounded objective")
		}

		row.SolveForSymbols(leaving, entering)

		s.substitute(entering, row)

		if entering.Type == SymbolTypeExternal && row.constant != 0 {
			v := s.varForSymbol[entering]
			s.varChanged(v)
		}
		s.rows[entering] = row
	}
}

func (s *Solver) varChanged(v Variable) {
	if s.shouldClearChanges {
		clear(s.changed)
		s.shouldClearChanges = false
	}
	s.changed[v] = struct{}{}
}

func (s *Solver) substitute(symbol _Symbol, row _Row) {
	for otherSymbol, otherRow := range s.rows {
		constantChanged := otherRow.Substitute(symbol, row)
		s.rows[otherSymbol] = otherRow

		if otherSymbol.Type == SymbolTypeExternal && constantChanged {
			v := s.varForSymbol[otherSymbol]

			s.varChanged(v)
		}

		if otherSymbol.Type != SymbolTypeExternal && otherRow.constant < 0 {
			s.infeasibleRows = append(s.infeasibleRows, otherSymbol)
		}
	}

	s.objective.Substitute(symbol, row)

	if s.artificial != nil {
		s.artificial.Substitute(symbol, row)
	}
}

func (s *Solver) getLeavingRow(entering _Symbol) (_Symbol, _Row, bool) {
	ratio := math.Inf(1)

	var (
		found _Symbol
		ok    bool
	)

	for s, r := range s.rows {
		if s.Type != SymbolTypeExternal {
			temp := r.CoefficientFor(entering)

			if temp < 0 {
				tempRatio := -r.constant / temp

				if tempRatio < ratio {
					ratio = tempRatio
					found = s
					ok = true
				}
			}
		}
	}

	if !ok {
		return _Symbol{}, _Row{}, false
	}

	row := s.rows[found]
	delete(s.rows, found)

	return found, row, true
}

func (s *Solver) createRow(constraint Constraint) (_Row, _Tag) {
	expr := constraint.expression
	row := newRow(expr.Constant)

	for _, term := range expr.Terms {
		if !nearZero(term.Coefficient) {
			symbol := s.getVarSymbol(term.Variable)

			if otherRow, ok := s.rows[symbol]; ok {
				row.InsertRow(otherRow, term.Coefficient)
			} else {
				row.InsertSymbol(symbol, term.Coefficient)
			}
		}
	}

	var tag _Tag

	switch constraint.op {
	case RelationOperatorGreaterThanEqual, RelationOperatorLessThanEqual:
		coeff := -1.0
		if constraint.op == RelationOperatorLessThanEqual {
			coeff = 1.0
		}

		slack := _Symbol{Value: s.idTick, Type: SymbolTypeSlack}
		s.idTick++

		row.InsertSymbol(slack, coeff)

		if constraint.strength < Required {
			errorSymbol := _Symbol{Value: s.idTick, Type: SymbolTypeError}
			s.idTick++

			row.InsertSymbol(errorSymbol, -coeff)
			s.objective.InsertSymbol(errorSymbol, float64(constraint.strength))

			tag = _Tag{
				marker: slack,
				other:  errorSymbol,
			}
		} else {
			tag = _Tag{
				marker: slack,
				other:  newInvalidSymbol(),
			}
		}
	case RelationOperatorEqual:
		if constraint.strength < Required {
			errPlus := _Symbol{Value: s.idTick, Type: SymbolTypeError}
			s.idTick++

			errMinus := _Symbol{Value: s.idTick, Type: SymbolTypeError}
			s.idTick++

			row.InsertSymbol(errPlus, -1)
			row.InsertSymbol(errMinus, 1)

			s.objective.InsertSymbol(errPlus, float64(constraint.strength))
			s.objective.InsertSymbol(errMinus, float64(constraint.strength))

			tag = _Tag{
				marker: errPlus,
				other:  errMinus,
			}
		} else {
			dummy := _Symbol{Value: s.idTick, Type: SymbolTypeDummy}
			s.idTick++

			row.InsertSymbol(dummy, 1)

			tag = _Tag{
				marker: dummy,
				other:  newInvalidSymbol(),
			}
		}
	default:
		panic(fmt.Sprintf("unexpected casso.RelationOperator: %#v", constraint.op))
	}

	if row.constant < 0 {
		row.reverseSign()
	}

	return row, tag
}

func (s *Solver) getVarSymbol(v Variable) _Symbol {
	data, ok := s.varData[v]
	if !ok {
		symbol := _Symbol{Value: s.idTick, Type: SymbolTypeExternal}
		s.varForSymbol[symbol] = v
		s.idTick++
		data = _VariableData{
			constant: math.NaN(),
			symbol:   symbol,
			id:       0,
		}
	}

	data.id++
	s.varData[v] = data

	return data.symbol
}

func chooseSubject(row _Row, tag _Tag) _Symbol {
	for s := range row.cells {
		if s.Type == SymbolTypeExternal {
			return s
		}
	}

	for _, s := range []_Symbol{tag.marker, tag.other} {
		switch s.Type {
		case SymbolTypeSlack, SymbolTypeError:
			if row.CoefficientFor(s) < 0 {
				return s
			}
		}
	}

	return newInvalidSymbol()
}

func allDummies(row _Row) bool {
	for s := range row.cells {
		if s.Type != SymbolTypeDummy {
			return false
		}
	}

	return true
}

func getEnteringSymbol(objective _Row) _Symbol {
	for s, v := range objective.cells {
		if s.Type != SymbolTypeDummy && v < 0 {
			return s
		}
	}

	return newInvalidSymbol()
}

func anyPivotableSymbol(row _Row) _Symbol {
	for s := range row.cells {
		switch s.Type {
		case SymbolTypeSlack, SymbolTypeError:
			return s
		}
	}

	return newInvalidSymbol()
}
