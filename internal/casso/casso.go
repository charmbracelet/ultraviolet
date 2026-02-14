// Package casso implements [Cassowary constraint solving algorithm].
//
// Code is roughly 1:1 translation of [ratatui/kasuari] rust crate.
//
// Cassowary is designed for solving constraints to lay out user interfaces.
// Constraints typically take the form "this button must line up with this text box",
// or "this box should try to be 3 times the size of this other box".
//
// [Cassowary constraint solving algorithm]: https://en.wikipedia.org/wiki/Cassowary_(software)
// [ratatui/kasuari]: https://github.com/ratatui/kasuari
package casso

import (
	"maps"
	"slices"
	"sync/atomic"
)

type Strength float64

const (
	Required Strength = 1_001_001_000
	Strong   Strength = 1_000_000
	Medium   Strength = 1_000
	Weak     Strength = 1
)

type RelationOperator int

const (
	RelationOperatorLessThanEqual RelationOperator = iota + 1
	RelationOperatorEqual
	RelationOperatorGreaterThanEqual
)

var _variableID atomic.Uint64

type Variable uint64

func NewVariable() Variable {
	defer _variableID.Add(1)

	return Variable(_variableID.Load())
}

type Term struct {
	Variable    Variable
	Coefficient float64
}

func NewTerm(variable Variable, coefficient float64) Term {
	return Term{
		Variable:    variable,
		Coefficient: coefficient,
	}
}

func (t Term) Negate() Term {
	t.Coefficient = -t.Coefficient
	return t
}

type Expression struct {
	Terms    []Term
	Constant float64
}

func NewExpressionFromConstant(v float64) Expression {
	return Expression{Constant: v}
}

func NewExpressionFromTerm(term Term) Expression {
	return Expression{Terms: []Term{term}}
}

func NewExpression(constant float64, terms ...Term) Expression {
	return Expression{
		Terms:    terms,
		Constant: constant,
	}
}

func (e Expression) Negate() Expression {
	e.Terms = slices.Clone(e.Terms)
	e.Constant = -e.Constant

	for i := range e.Terms {
		e.Terms[i] = e.Terms[i].Negate()
	}

	return e
}

type ConstraintData struct {
	expression Expression
	strength   Strength
	op         RelationOperator
}

type Constraint *ConstraintData

func NewConstraint(e Expression, op RelationOperator, strength Strength) Constraint {
	data := ConstraintData{
		expression: e,
		strength:   strength,
		op:         op,
	}

	return &data
}

func (cd ConstraintData) Expression() Expression {
	return cd.expression
}

func (cd ConstraintData) Strength() Strength {
	return cd.strength
}

func (cd ConstraintData) Op() RelationOperator {
	return cd.op
}

type WeightedRelation struct {
	Operator RelationOperator
	Strength Strength
}

func (w WeightedRelation) ExpressionLHS(expression Expression) PartialConstraint {
	return PartialConstraint{
		Expression: expression,
		Relation:   w,
	}
}

func (w WeightedRelation) VariableLHS(variable Variable) PartialConstraint {
	return PartialConstraint{
		Expression: NewExpressionFromTerm(NewTerm(variable, 1)),
		Relation:   w,
	}
}

func Equal(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: RelationOperatorEqual, Strength: strength}
}

func LessThanEqual(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: RelationOperatorLessThanEqual, Strength: strength}
}

func GreaterThanEqual(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: RelationOperatorGreaterThanEqual, Strength: strength}
}

type PartialConstraint struct {
	Expression Expression
	Relation   WeightedRelation
}

func (p PartialConstraint) ConstantRHS(v float64) Constraint {
	return NewConstraint(
		p.Expression.SubConstant(v),
		p.Relation.Operator,
		p.Relation.Strength,
	)
}

func (p PartialConstraint) ExpressionRHS(e Expression) Constraint {
	return NewConstraint(
		p.Expression.Sub(e),
		p.Relation.Operator,
		p.Relation.Strength,
	)
}

func (p PartialConstraint) VariableRHS(v Variable) Constraint {
	return NewConstraint(
		p.Expression.SubVariable(v),
		p.Relation.Operator,
		p.Relation.Strength,
	)
}

type SymbolType int

const (
	SymbolTypeInvalid SymbolType = iota + 1
	SymbolTypeExternal
	SymbolTypeSlack
	SymbolTypeError
	SymbolTypeDummy
)

type _Symbol struct {
	Value uint8
	Type  SymbolType
}

func newInvalidSymbol() _Symbol {
	return _Symbol{
		Value: 0,
		Type:  SymbolTypeInvalid,
	}
}

type _Row struct {
	cells    map[_Symbol]float64
	constant float64
}

func newRow(constant float64) _Row {
	return _Row{
		cells:    make(map[_Symbol]float64),
		constant: constant,
	}
}

func (r *_Row) Clone() _Row {
	return _Row{
		cells:    maps.Clone(r.cells),
		constant: r.constant,
	}
}

func (r *_Row) reverseSign() {
	r.constant = -r.constant

	for s := range r.cells {
		r.cells[s] = -r.cells[s]
	}
}

func (r *_Row) Add(v float64) float64 {
	r.constant += v
	return r.constant
}

func (r *_Row) InsertSymbol(s _Symbol, coefficient float64) {
	if value, ok := r.cells[s]; ok {
		value += coefficient

		if nearZero(value) {
			delete(r.cells, s)
		} else {
			r.cells[s] = value
		}

		return
	}

	if nearZero(coefficient) {
		return
	}

	r.cells[s] = coefficient
}

func (r *_Row) InsertRow(other _Row, coefficient float64) bool {
	constantDiff := other.constant * coefficient
	r.constant += constantDiff

	for s, v := range other.cells {
		r.InsertSymbol(s, v*coefficient)
	}

	return constantDiff != 0
}

func (r *_Row) Remove(s _Symbol) {
	delete(r.cells, s)
}

func (r *_Row) SolveForSymbol(s _Symbol) {
	coeff := -1.0 / r.cells[s]
	delete(r.cells, s)

	r.constant *= coeff

	for s := range r.cells {
		r.cells[s] *= coeff
	}
}

func (r *_Row) SolveForSymbols(lhs, rhs _Symbol) {
	r.InsertSymbol(lhs, -1)
	r.SolveForSymbol(rhs)
}

func (r *_Row) CoefficientFor(s _Symbol) float64 {
	return r.cells[s]
}

func (r *_Row) Substitute(s _Symbol, row _Row) bool {
	if coeff, ok := r.cells[s]; ok {
		delete(r.cells, s)
		return r.InsertRow(row, coeff)
	}

	return false
}

func nearZero(value float64) bool {
	const epsilon float64 = 1e-8

	if value < 0.0 {
		return -value < epsilon
	}

	return value < epsilon
}
