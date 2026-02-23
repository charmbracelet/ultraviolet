package layout

import (
	"fmt"
	"io"
)

// Constraint describes how a single segment of a [Layout] should be sized.
//
// Each constraint type expresses a different kind of sizing rule:
// fixed ([Len]), proportional ([Percent], [Ratio]), bounded ([Min], [Max]),
// or greedy ([Fill]). Proportional constraints are evaluated against the
// full area being split rather than the remaining space after fixed
// constraints have been applied.
//
// When the solver cannot satisfy every constraint, it resolves conflicts
// according to the following priority order (highest first):
//
//   - [Min]
//   - [Max]
//   - [Len]
//   - [Percent]
//   - [Ratio]
//   - [Fill]
type Constraint interface {
	hash(w io.Writer)

	// isConstraint is a private method to prevent users implementing the
	// interface making it a sealed enum.
	isConstraint()
}

type (
	// Min ensures the segment is no smaller than the given number of cells.
	//
	// # Examples
	//
	// 	[Percent(100), Min(20)]
	//
	// 	┌────────────────────────────┐┌──────────────────┐
	// 	│            30 px           ││       20 px      │
	// 	└────────────────────────────┘└──────────────────┘
	//
	// 	[Percent(100), Min(10)]
	//
	// 	┌──────────────────────────────────────┐┌────────┐
	// 	│                 40 px                ││  10 px │
	// 	└──────────────────────────────────────┘└────────┘
	Min int

	// Max caps the segment at the given number of cells.
	//
	// # Examples
	//
	// 	[Percent(0), Max(20)]
	//
	// 	┌────────────────────────────┐┌──────────────────┐
	// 	│            30 px           ││       20 px      │
	// 	└────────────────────────────┘└──────────────────┘
	//
	// 	[Percent(0), Max(10)]
	//
	// 	┌──────────────────────────────────────┐┌────────┐
	// 	│                 40 px                ││  10 px │
	// 	└──────────────────────────────────────┘└────────┘
	Max int

	// Len fixes the segment to exactly the given number of cells.
	//
	// # Examples
	//
	// 	[Len(20), Len(20)]
	//
	// 	┌──────────────────┐┌──────────────────┐
	// 	│       20 px      ││       20 px      │
	// 	└──────────────────┘└──────────────────┘
	//
	// 	[Len(20), Len(30)]
	//
	// 	┌──────────────────┐┌────────────────────────────┐
	// 	│       20 px      ││            30 px           │
	// 	└──────────────────┘└────────────────────────────┘
	Len int

	// Percent sizes the segment as a fraction of the total area.
	//
	// The integer value is treated as a percentage (0-100+) and multiplied
	// by the total area; the result is rounded to the nearest cell.
	//
	// Because only whole integers are accepted, some fractions (e.g. 1/3)
	// cannot be represented exactly. Consider [Ratio] or [Fill] instead.
	//
	// # Examples
	//
	// 	[Percent(75), Fill(1)]
	//
	// 	┌────────────────────────────────────┐┌──────────┐
	// 	│                38 px               ││   12 px  │
	// 	└────────────────────────────────────┘└──────────┘
	//
	// 	[Percent(50), Fill(1)]
	//
	// 	┌───────────────────────┐┌───────────────────────┐
	// 	│         25 px         ││         25 px         │
	// 	└───────────────────────┘└───────────────────────┘
	Percent int

	// Ratio sizes the segment as a numerator/denominator fraction of the total area.
	//
	// The fraction is converted to a float, multiplied by the area, and
	// rounded to the nearest cell.
	//
	// # Examples
	//
	// 	[Ratio(1, 2) ; 2]
	//
	// 	┌───────────────────────┐┌───────────────────────┐
	// 	│         25 px         ││         25 px         │
	// 	└───────────────────────┘└───────────────────────┘
	//
	// 	[Ratio(1, 4) ; 4]
	//
	// 	┌───────────┐┌──────────┐┌───────────┐┌──────────┐
	// 	│   13 px   ││   12 px  ││   13 px   ││   12 px  │
	// 	└───────────┘└──────────┘└───────────┘└──────────┘
	Ratio struct{ Num, Den int }

	// Fill distributes remaining space proportionally among all Fill segments
	// according to their respective weights.
	//
	// A Fill segment only expands into space left over after higher-priority
	// constraints have been satisfied. Multiple Fill segments share that
	// leftover in proportion to their integer values.
	//
	// # Examples
	//
	//
	// 	[Fill(1), Fill(2), Fill(3)]
	//
	// 	┌──────┐┌───────────────┐┌───────────────────────┐
	// 	│ 8 px ││     17 px     ││         25 px         │
	// 	└──────┘└───────────────┘└───────────────────────┘
	//
	// 	[Fill(1), Percent(50), Fill(1)]
	//
	// 	┌───────────┐┌───────────────────────┐┌──────────┐
	// 	│   13 px   ││         25 px         ││   12 px  │
	// 	└───────────┘└───────────────────────┘└──────────┘
	Fill int
)

func (m Min) String() string   { return fmt.Sprintf("Min(%d)", m) }
func (m Min) hash(w io.Writer) { fmt.Fprint(w, "min", m) }
func (Min) isConstraint()      {}

func (m Max) String() string   { return fmt.Sprintf("Max(%d)", m) }
func (m Max) hash(w io.Writer) { fmt.Fprint(w, "max", m) }
func (Max) isConstraint()      {}

func (l Len) String() string   { return fmt.Sprintf("Len(%d)", l) }
func (l Len) hash(w io.Writer) { fmt.Fprint(w, "len", l) }
func (Len) isConstraint()      {}

func (p Percent) String() string   { return fmt.Sprintf("Percent(%d)", p) }
func (p Percent) hash(w io.Writer) { fmt.Fprint(w, "percent", p) }
func (Percent) isConstraint()      {}

func (r Ratio) String() string   { return fmt.Sprintf("Ratio(%d / %d)", r.Num, r.Den) }
func (r Ratio) hash(w io.Writer) { fmt.Fprint(w, "ratio", r.Num, r.Den) }
func (Ratio) isConstraint()      {}

func (f Fill) String() string   { return fmt.Sprintf("Fill(%d)", f) }
func (f Fill) hash(w io.Writer) { fmt.Fprint(w, "fill", f) }
func (Fill) isConstraint()      {}
