package casso

import (
	"errors"
	"math"
)

type tag struct {
	priority Priority
	marker   Symbol
	other    Symbol
}

// Solver implements the Cassowary constraint solving algorithm.
type Solver struct {
	tabs map[Symbol]Constraint
	tags map[Symbol]tag

	infeasible []Symbol

	objective  expr
	artificial expr
}

// NewSolver creates a new constraint solver.
func NewSolver() *Solver {
	return &Solver{
		tabs: make(map[Symbol]Constraint),
		tags: make(map[Symbol]tag),
	}
}

// Val returns the current value of a symbol.
func (s *Solver) Val(id Symbol) float64 {
	row, ok := s.tabs[id]
	if !ok {
		return 0
	}
	return row.expr.constant
}

// Add adds a constraint to the solver with the given priority.
func (s *Solver) Add(priority Priority, cell Constraint) (Symbol, error) {
	t := tag{priority: priority}

	c := cell
	c.expr.terms = make([]Term, 0, len(c.expr.terms))

	for _, term := range cell.expr.terms {
		if eqz(term.coeff) {
			continue
		}
		if term.id.isZero() {
			return zero, errBadTerm
		}
		resolved, exists := s.tabs[term.id]
		if !exists {
			c.expr.addSymbol(term.coeff, term.id)
			continue
		}
		c.expr.addExpr(term.coeff, resolved.expr)
	}

	switch c.op {
	case LTE, GTE:
		coeff := 1.0
		if c.op == GTE {
			coeff = -1.0
		}

		t.marker = next(slack)
		c.expr.addSymbol(coeff, t.marker)

		if priority < required {
			t.other = next(errorSym)
			c.expr.addSymbol(-coeff, t.other)
			s.objective.addSymbol(float64(priority), t.other)
		}
	case EQ:
		if priority < required {
			t.marker = next(errorSym)
			t.other = next(errorSym)

			c.expr.addSymbol(-1.0, t.marker)
			c.expr.addSymbol(1.0, t.other)

			s.objective.addSymbol(float64(priority), t.marker)
			s.objective.addSymbol(float64(priority), t.other)
		} else {
			t.marker = next(dummy)
			c.expr.addSymbol(1.0, t.marker)
		}
	}

	if c.expr.constant < 0.0 {
		c.expr.negate()
	}

	subject, err := s.findSubject(c, t)
	if err != nil {
		return zero, err
	}

	if subject.isZero() {
		if err := s.augmentArtificialVariable(c); err != nil {
			return t.marker, err
		}
	} else {
		c.expr.solveFor(subject)
		s.substitute(subject, c.expr)
		s.tabs[subject] = c
	}

	s.tags[t.marker] = t

	return t.marker, s.optimizeAgainst(&s.objective)
}

const required Priority = 1e9

func (s *Solver) findSubject(cell Constraint, t tag) (Symbol, error) {
	for _, term := range cell.expr.terms {
		if term.id.external() {
			return term.id, nil
		}
	}

	if t.marker.restricted() {
		idx := cell.expr.find(t.marker)
		if idx != -1 && cell.expr.terms[idx].coeff < 0.0 {
			return t.marker, nil
		}
	}

	if t.other.restricted() {
		idx := cell.expr.find(t.other)
		if idx != -1 && cell.expr.terms[idx].coeff < 0.0 {
			return t.other, nil
		}
	}

	for _, term := range cell.expr.terms {
		if !term.id.isDummy() {
			return zero, nil
		}
	}

	if !eqz(cell.expr.constant) {
		return zero, errUnsatisfiable
	}

	return t.marker, nil
}

func (s *Solver) substitute(id Symbol, e expr) {
	for symbol := range s.tabs {
		row := s.tabs[symbol]
		row.expr.substitute(id, e)
		s.tabs[symbol] = row
		if symbol.external() || row.expr.constant >= 0.0 {
			continue
		}
		s.infeasible = append(s.infeasible, symbol)
	}
	s.objective.substitute(id, e)
	s.artificial.substitute(id, e)
}

func (s *Solver) optimizeAgainst(objective *expr) error {
	for {
		entry := zero
		exit := zero

		for _, term := range objective.terms {
			if !term.id.isDummy() && term.coeff < 0.0 {
				entry = term.id
				break
			}
		}
		if entry.isZero() {
			return nil
		}

		ratio := math.MaxFloat64

		for symbol := range s.tabs {
			if symbol.external() {
				continue
			}
			idx := s.tabs[symbol].expr.find(entry)
			if idx == -1 {
				continue
			}
			coeff := s.tabs[symbol].expr.terms[idx].coeff
			if coeff >= 0.0 {
				continue
			}
			r := -s.tabs[symbol].expr.constant / coeff
			if r < ratio {
				ratio, exit = r, symbol
			}
		}

		row := s.tabs[exit]
		delete(s.tabs, exit)

		row.expr.solveForSymbols(exit, entry)

		s.substitute(entry, row.expr)
		s.tabs[entry] = row
	}
}

func (s *Solver) augmentArtificialVariable(row Constraint) error {
	art := next(slack)

	s.tabs[art] = row.clone()
	s.artificial = row.expr.clone()

	if err := s.optimizeAgainst(&s.artificial); err != nil {
		return err
	}

	success := eqz(s.artificial.constant)
	s.artificial = newExpr(0.0)

	artificial, ok := s.tabs[art]
	if ok {
		delete(s.tabs, art)

		if len(artificial.expr.terms) == 0 {
			return nil
		}

		entry := zero
		for _, term := range artificial.expr.terms {
			if term.id.restricted() {
				entry = term.id
				break
			}
		}
		if entry.isZero() {
			return errUnsatisfiable
		}

		artificial.expr.solveForSymbols(art, entry)

		s.substitute(entry, artificial.expr)
		s.tabs[entry] = artificial
	}

	for symbol, row := range s.tabs {
		idx := row.expr.find(art)
		if idx == -1 {
			continue
		}
		row.expr.delete(idx)
		s.tabs[symbol] = row
	}

	idx := s.objective.find(art)
	if idx != -1 {
		s.objective.delete(idx)
	}

	if !success {
		return errUnsatisfiable
	}
	return nil
}

var (
	errUnsatisfiable = errors.New("casso: constraint is unsatisfiable")
	errBadTerm       = errors.New("casso: term references a nil symbol")
)
