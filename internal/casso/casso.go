// Package casso implements Cassowary constraint solving algorithm.
//
// It wraps [lithdew/casso] package with additional expression manipulation
// functions, similar to [ratatui/kasuari] implementation.
//
// [lithdew/casso]: https://github.com/lithdew/casso
// [ratatui/kasuari]: https://github.com/ratatui/kasuari
package casso

import (
	"slices"

	"github.com/lithdew/casso"
)

type Strength float64

const (
	Required Strength = 1_001_001_000
	Strong   Strength = 1_000_000
	Medium   Strength = 1_000
	Weak     Strength = 1
)

type Symbol casso.Symbol

func New() Symbol {
	return Symbol(casso.New())
}

type Term struct {
	Symbol      Symbol
	Coefficient float64
}

func NewTerm(variable Symbol, coefficient float64) Term {
	return Term{
		Symbol:      variable,
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
	op         casso.Op
}

type Constraint *ConstraintData

func NewConstraint(e Expression, op casso.Op, strength Strength) Constraint {
	data := ConstraintData{
		expression: e,
		strength:   strength,
		op:         op,
	}

	return &data
}

type WeightedRelation struct {
	Operator casso.Op
	Strength Strength
}

func (w WeightedRelation) ExpressionLHS(expression Expression) PartialConstraint {
	return PartialConstraint{
		Expression: expression,
		Relation:   w,
	}
}

func (w WeightedRelation) SymbolLHS(variable Symbol) PartialConstraint {
	return PartialConstraint{
		Expression: NewExpressionFromTerm(NewTerm(variable, 1)),
		Relation:   w,
	}
}

func Equal(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: casso.EQ, Strength: strength}
}

func LessThanEqual(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: casso.LTE, Strength: strength}
}

func GreaterThanEqual(strength Strength) WeightedRelation {
	return WeightedRelation{Operator: casso.GTE, Strength: strength}
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

func (p PartialConstraint) SymbolRHS(v Symbol) Constraint {
	return NewConstraint(
		p.Expression.SubSymbol(v),
		p.Relation.Operator,
		p.Relation.Strength,
	)
}
