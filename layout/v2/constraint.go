package layout

import "fmt"

type Constraint interface {
	fmt.Stringer

	isConstraint()
}

type (
	Min        int
	Max        int
	Len        int
	Percentage int
	Ratio      struct{ Num, Den int }
	Fill       int
)

func (m Min) String() string { return fmt.Sprintf("Min(%d)", m) }
func (Min) isConstraint()    {}

func (m Max) String() string { return fmt.Sprintf("Max(%d)", m) }
func (Max) isConstraint()    {}

func (l Len) String() string { return fmt.Sprintf("Len(%d)", l) }
func (Len) isConstraint()    {}

func (p Percentage) String() string { return fmt.Sprintf("Percentage(%d)", p) }
func (Percentage) isConstraint()    {}

func (r Ratio) String() string { return fmt.Sprintf("Ratio(%d / %d)", r.Num, r.Den) }
func (Ratio) isConstraint()    {}

func (f Fill) String() string { return fmt.Sprintf("Fill(%d)", f) }
func (Fill) isConstraint()    {}
