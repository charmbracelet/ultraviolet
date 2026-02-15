package casso

import "slices"

func (v Variable) Sub(other Variable) Expression {
	return NewExpression(0, NewTerm(v, 1), NewTerm(other, -1))
}

func (v Variable) Add(other Variable) Expression {
	return NewExpression(0, NewTerm(v, 1), NewTerm(other, 1))
}

func (v Variable) AddConstant(other float64) Expression {
	return NewExpression(other, NewTerm(v, 1))
}

func (e Expression) SubConstant(other float64) Expression {
	e.Constant -= other
	return e
}

func (e Expression) Sub(other Expression) Expression {
	other = other.Negate()

	e.Terms = append(e.Terms, other.Terms...)
	e.Constant += other.Constant

	return e
}

func (e Expression) SubVariable(other Variable) Expression {
	e.Terms = append(e.Terms, NewTerm(other, -1.0))

	return e
}

func (e Expression) MulConstant(other float64) Expression {
	e.Terms = slices.Clone(e.Terms)
	e.Constant *= other

	for i := range e.Terms {
		e.Terms[i].Coefficient *= other
	}

	return e
}

func (e Expression) DivConstant(other float64) Expression {
	e.Terms = slices.Clone(e.Terms)
	e.Constant /= other

	for i := range e.Terms {
		e.Terms[i].Coefficient /= other
	}

	return e
}
