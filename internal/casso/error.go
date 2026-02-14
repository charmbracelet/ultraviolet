package casso

import "errors"

var (
	ErrDuplicateConstraint     = errors.New("duplicate constraint")
	ErrUnsatisfiableConstraint = errors.New("unsatisfiable constraint")
	ErrUnknownConstraint       = errors.New("unknown constraint")
	ErrDuplicateEditVariable   = errors.New("duplicate edit variable")
	ErrBadRequiredStrength     = errors.New("bad required strength")
	ErrUnknownEditVariable     = errors.New("unknown edit variable")
)

type InternalSolverError string

func (e InternalSolverError) Error() string {
	return string(e)
}
