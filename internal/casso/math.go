// Package casso implements a Cassowary constraint solver.
//
// Ported from github.com/lithdew/casso with unused API removed.
package casso

import "sync/atomic"

type symbolKind uint8

const (
	external symbolKind = iota
	slack
	errorSym
	dummy
)

func (s symbolKind) restricted() bool { return s == slack || s == errorSym }

// Symbol is an opaque identifier for a variable in the solver.
type Symbol uint64

var (
	count uint64
	zero  Symbol
)

// New creates a new external symbol.
func New() Symbol { return next(external) }

func next(typ symbolKind) Symbol {
	return Symbol((atomic.AddUint64(&count, 1) & 0x3fffffffffffffff) | (uint64(typ) << 62))
}

func (sym Symbol) kind() symbolKind { return symbolKind(sym >> 62) }
func (sym Symbol) isZero() bool     { return sym == zero }
func (sym Symbol) restricted() bool { return !sym.isZero() && sym.kind().restricted() }
func (sym Symbol) external() bool   { return !sym.isZero() && sym.kind() == external }
func (sym Symbol) isDummy() bool    { return !sym.isZero() && sym.kind() == dummy }

// T creates a Term with the given coefficient.
func (sym Symbol) T(coeff float64) Term { return Term{coeff: coeff, id: sym} }

// Priority represents the strength of a constraint.
type Priority float64

// Op represents a constraint operator.
type Op uint8

const (
	EQ Op = iota
	GTE
	LTE
)

// Constraint represents a linear constraint.
type Constraint struct {
	op   Op
	expr expr
}

// NewConstraint creates a new constraint.
func NewConstraint(op Op, constant float64, terms ...Term) Constraint {
	return Constraint{op: op, expr: newExpr(constant, terms...)}
}

func (c Constraint) clone() Constraint {
	return Constraint{op: c.op, expr: c.expr.clone()}
}

// Term represents a variable with a coefficient.
type Term struct {
	coeff float64
	id    Symbol
}

type expr struct {
	constant float64
	terms    []Term
}

func newExpr(constant float64, terms ...Term) expr {
	return expr{constant: constant, terms: terms}
}

func (c expr) clone() expr {
	res := expr{constant: c.constant, terms: make([]Term, len(c.terms))}
	copy(res.terms, c.terms)
	return res
}

func (c expr) find(id Symbol) int {
	for i := 0; i < len(c.terms); i++ {
		if c.terms[i].id == id {
			return i
		}
	}
	return -1
}

func (c *expr) delete(idx int) {
	copy(c.terms[idx:], c.terms[idx+1:])
	c.terms = c.terms[:len(c.terms)-1]
}

func (c *expr) addSymbol(coeff float64, id Symbol) {
	idx := c.find(id)
	if idx == -1 {
		if !eqz(coeff) {
			c.terms = append(c.terms, Term{coeff: coeff, id: id})
		}
		return
	}
	c.terms[idx].coeff += coeff
	if eqz(c.terms[idx].coeff) {
		c.delete(idx)
	}
}

func (c *expr) addExpr(coeff float64, other expr) {
	c.constant += coeff * other.constant
	for i := 0; i < len(other.terms); i++ {
		c.addSymbol(coeff*other.terms[i].coeff, other.terms[i].id)
	}
}

func (c *expr) negate() {
	c.constant = -c.constant
	for i := 0; i < len(c.terms); i++ {
		c.terms[i].coeff = -c.terms[i].coeff
	}
}

func (c *expr) solveFor(id Symbol) {
	idx := c.find(id)
	if idx == -1 {
		return
	}

	coeff := -1.0 / c.terms[idx].coeff
	c.delete(idx)

	if coeff == 1.0 {
		return
	}

	c.constant *= coeff
	for i := 0; i < len(c.terms); i++ {
		c.terms[i].coeff *= coeff
	}
}

func (c *expr) solveForSymbols(lhs, rhs Symbol) {
	c.addSymbol(-1.0, lhs)
	c.solveFor(rhs)
}

func (c *expr) substitute(id Symbol, other expr) {
	idx := c.find(id)
	if idx == -1 {
		return
	}
	coeff := c.terms[idx].coeff
	c.delete(idx)
	c.addExpr(coeff, other)
}

func eqz(val float64) bool {
	if val < 0 {
		return -val < 1.0e-8
	}
	return val < 1.0e-8
}
