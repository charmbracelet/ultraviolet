package layout

import "fmt"

// Constraint that defines the size of a layout element.
//
// Constraints are the core mechanism for defining how space should be allocated within a
// [Layout]. They can specify fixed sizes ([Len]), proportional sizes
// ([Percentage], [Ratio]), size limits ([Min], [Max]), or proportional fill values for layout elements.
// Relative constraints ([Percentage], [Ratio]) are calculated relative to the entire space being
// divided, rather than the space available after applying more fixed
// constraints ([Min], [Max], [Len]).
//
// Constraints are prioritized in the following order:
//
//   - [Min]
//   - [Max]
//   - [Len]
//   - [Percentage]
//   - [Ratio]
//   - [Fill]
type Constraint interface{ isConstraint() }

type (
	// Min applies a minimum size constraint to the element
	//
	// The element size is set to at least the specified amount.
	//
	// # Examples
	//
	// 	[Percentage(100), Min(20)]
	//
	// 	┌────────────────────────────┐┌──────────────────┐
	// 	│            30 px           ││       20 px      │
	// 	└────────────────────────────┘└──────────────────┘
	//
	// 	[Percentage(100), Min(10)]
	//
	// 	┌──────────────────────────────────────┐┌────────┐
	// 	│                 40 px                ││  10 px │
	// 	└──────────────────────────────────────┘└────────┘
	Min int

	// Max applies a maximum size constraint to the element
	//
	// The element size is set to at most the specified amount.
	//
	// # Examples
	//
	// 	[Percentage(0), Max(20)]
	//
	// 	┌────────────────────────────┐┌──────────────────┐
	// 	│            30 px           ││       20 px      │
	// 	└────────────────────────────┘└──────────────────┘
	//
	// 	[Percentage(0), Max(10)]
	//
	// 	┌──────────────────────────────────────┐┌────────┐
	// 	│                 40 px                ││  10 px │
	// 	└──────────────────────────────────────┘└────────┘
	Max int

	// Len applies a length constraint to the element
	//
	// The element size is set to the specified amount.
	//
	// # Examples
	//
	// 	[Length(20), Length(20)]
	//
	// 	┌──────────────────┐┌──────────────────┐
	// 	│       20 px      ││       20 px      │
	// 	└──────────────────┘└──────────────────┘
	//
	// 	[Length(20), Length(30)]
	//
	// 	┌──────────────────┐┌────────────────────────────┐
	// 	│       20 px      ││            30 px           │
	// 	└──────────────────┘└────────────────────────────┘
	Len int

	// Percentage applies a percentage of the available space to the element
	//
	// Converts the given percentage to a floating-point value and multiplies that with area. This
	// value is rounded back to a integer as part of the layout split calculation.
	//
	// **Note**: As this value only accepts a int, certain percentages that cannot be
	// represented exactly (e.g. 1/3) are not possible. You might want to use
	// [Ratio] or [Fill] in such cases.
	//
	// # Examples
	//
	// 	[Percentage(75), Fill(1)]
	//
	// 	┌────────────────────────────────────┐┌──────────┐
	// 	│                38 px               ││   12 px  │
	// 	└────────────────────────────────────┘└──────────┘
	//
	// 	[Percentage(50), Fill(1)]
	//
	// 	┌───────────────────────┐┌───────────────────────┐
	// 	│         25 px         ││         25 px         │
	// 	└───────────────────────┘└───────────────────────┘
	Percentage int

	// Ratio applies a ratio of the available space to the element
	//
	// Converts the given ratio to a floating-point value and multiplies that with area.
	// This value is rounded back to a integer as part of the layout split calculation.
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

	// Fill applies the scaling factor proportional to all other [Fill] elements
	// to fill excess space
	//
	// The element will only expand or fill into excess available space, proportionally matching
	// other [Fill] elements while satisfying all other constraints.
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
	// 	[Fill(1), Percentage(50), Fill(1)]
	//
	// 	┌───────────┐┌───────────────────────┐┌──────────┐
	// 	│   13 px   ││         25 px         ││   12 px  │
	// 	└───────────┘└───────────────────────┘└──────────┘
	Fill int
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
