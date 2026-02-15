package casso

import (
	"github.com/lithdew/casso"
)

type Solver struct {
	solver *casso.Solver
}

func NewSolver() Solver {
	return Solver{
		solver: casso.NewSolver(),
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
	terms := make([]casso.Term, 0, len(constraint.expression.Terms))

	for _, t := range constraint.expression.Terms {
		terms = append(terms, casso.Symbol(t.Symbol).T(t.Coefficient))
	}

	c := casso.NewConstraint(casso.Op(constraint.op), constraint.expression.Constant, terms...)

	_, err := s.solver.AddConstraintWithPriority(casso.Priority(constraint.strength), c)

	return err
}

func (s *Solver) Val(v Symbol) float64 {
	return s.solver.Val(casso.Symbol(v))
}
