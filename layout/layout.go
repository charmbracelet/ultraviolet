// Package layout provides a comprehensive set of types and traits for working with layout and
// positioning in terminal applications. It implements a flexible layout system that allows you to
// divide the terminal screen into different areas using constraints, manage positioning and
// sizing, and handle complex UI arrangements.
//
// The layout system is based on the [Cassowary constraint solver algorithm].
// This allows for sophisticated constraint-based layouts where
// multiple requirements can be satisfied simultaneously, with priorities determining which
// constraints take precedence when conflicts arise.
//
// # Layout Fundamentals
//
// Layouts form the structural foundation of your terminal UI. The [Layout] struct divides
// available screen space into rectangular areas using a constraint-based approach. You define
// multiple constraints for how space should be allocated, and the Cassowary solver determines
// the optimal layout that satisfies as many constraints as possible. These areas can then be
// used to render widgets or nested layouts.
//
// Note that the [Layout] struct is not required to create layouts - you can also manually
// calculate and create [uv.Rectangle] areas using simple mathematics to divide up the terminal space
// if you prefer direct control over positioning and sizing.
//
// # Acknowledgements
//
// This implementation is heavily based on [Ratatui] source code and
// is roughly 1:1 translation from Rust with some minor API adjustments.
//
// [Cassowary constraint solver algorithm]: https://en.wikipedia.org/wiki/Cassowary_(software)
// [Ratatui]: https://ratatui.rs/
package layout

import (
	"fmt"
	"math"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/casso"
)

// floatPrecisionMultiplier decides floating point precision when rounding.
// The number of zeros in this number is the precision for the rounding in layout calculations.
const floatPrecisionMultiplier float64 = 100.0

const (
	// spacerSizeEq is the strength to apply to Spacers to ensure that their sizes are equal.
	//
	// 	┌     ┐┌───┐┌     ┐┌───┐┌     ┐
	// 	  ==x  │   │  ==x  │   │  ==x
	// 	└     ┘└───┘└     ┘└───┘└     ┘
	spacerSizeEq casso.Strength = casso.Required / 10.0

	// minSizeGTE is the strength to apply to [Min] inequality constraints.
	//
	// 	┌────────┐
	// 	│Min(>=x)│
	// 	└────────┘
	minSizeGTE casso.Strength = casso.Strong * 100.0

	// maxSizeLTE is the strength to apply to [Max] inequality constraints.
	//
	// 	┌────────┐
	// 	│Max(<=x)│
	// 	└────────┘
	maxSizeLTE casso.Strength = casso.Strong * 100.0

	// lengthSizeEq is the strength to apply to [Len] constraints.
	//
	// 	┌────────┐
	// 	│Len(==x)│
	// 	└────────┘
	lengthSizeEq casso.Strength = casso.Strong * 10.0

	// percentSizeEq is the strength to apply to [Percent] constraints.
	//
	// 	┌────────────┐
	// 	│Percent(==x)│
	// 	└────────────┘
	percentSizeEq casso.Strength = casso.Strong

	// ratioSizeEq is the strength to apply to [Ratio] constraints.
	//
	// 	┌────────────┐
	// 	│Ratio(==x,y)│
	// 	└────────────┘
	ratioSizeEq casso.Strength = casso.Strong / 10.0

	// minSizeEq is the strength to apply to [Min] equality constraints.
	//
	// 	┌────────┐
	// 	│Min(==x)│
	// 	└────────┘
	minSizeEq casso.Strength = casso.Medium * 10.0

	// maxSizeEq the strength to apply to [Max] equality constraints.
	//
	// 	┌────────┐
	// 	│Max(==x)│
	// 	└────────┘
	maxSizeEq casso.Strength = casso.Medium * 10.0

	// fillGrow is the strength to apply to [Fill] growing constraints.
	//
	// 	┌─────────────────────┐
	// 	│<=     Fill(x)     =>│
	// 	└─────────────────────┘
	fillGrow casso.Strength = casso.Medium

	// grow is the strength to apply to growing constraints.
	//
	// 	┌────────────┐
	// 	│<= Min(x) =>│
	// 	└────────────┘
	grow casso.Strength = 100.0

	// spaceGrow is the strength to apply to Spacer growing constraints.
	//
	// 	┌       ┐
	// 	 <= x =>
	// 	└       ┘
	spaceGrow casso.Strength = casso.Weak * 10.0

	// allSegmentGrow is he strength to apply to growing the size of all segments equally.
	//
	// ┌───────┐
	// │<= x =>│
	// └───────┘
	allSegmentGrow casso.Strength = casso.Weak
)

// Splitted represents result of [Layout] splitting
// as a slice of rectangles.
type Splitted []uv.Rectangle

// Assign sets splitted rectangles into pointed values.
//
// Nil pointers are skipped.
//
// Panics if given more areas that [Splitted] length.
//
// # Examples
//
//	var top, bottom uv.Rectangle
//
//	layout.New(layout.Fill(1), layout.Len(1)).
//		Split(area).
//	    Assign(&top, &bottom)
func (s Splitted) Assign(areas ...*uv.Rectangle) {
	for i := range areas {
		if areas[i] != nil {
			*areas[i] = s[i]
		}
	}
}

// Direction of a layout.
//
// This is used with [Layout] to specify whether layout
// segments should be arranged horizontally or vertically.
type Direction int

const (
	// DirectionVertical - layout segments are arranged top to bottom (default).
	DirectionVertical Direction = iota
	// DirectionHorizontal - layout segments are arranged side by side (left to right).
	DirectionHorizontal
)

// New creates a new layout with default values.
func New(direction Direction, constraints ...Constraint) Layout {
	return Layout{
		Direction:   direction,
		Constraints: constraints,
		Flex:        FlexLegacy,
	}
}

// Vertical reates a new vertical layout with default values.
func Vertical(constraints ...Constraint) Layout {
	return New(DirectionVertical, constraints...)
}

// Horizontal reates a new horizontal layout with default values.
func Horizontal(constraints ...Constraint) Layout {
	return New(DirectionHorizontal, constraints...)
}

// TODO: Research layout caching

// Layout engine for dividing terminal space using constraints and direction.
//
// A layout is a set of constraints that can be applied to a given area to split it into smaller
// rectangular areas. This is the core building block for creating structured user interfaces in
// terminal applications.
//
// A layout is composed of:
//   - a direction (horizontal or vertical)
//   - a set of constraints ([Len], [Ratio], [Percent], [Fill], [Min], [Max])
//   - a margin (horizontal and vertical), the space between the edge of the main area and the split areas
//   - a flex option that controls space distribution
//   - a spacing option that controls gaps between segments
//
// The algorithm used to compute the layout is based on the Cassowary solver, a linear constraint
// solver that computes positions and sizes to satisfy as many constraints as possible in order of
// their priorities.
type Layout struct {
	Direction   Direction
	Constraints []Constraint
	Padding     Padding
	// Spacing reprsents spacing between
	// segments as a number of cells.
	//
	// Negative spacing causes overlap between segments.
	Spacing int
	Flex    Flex
}

// WithDirection returns a copy of the layout with the given direction.
func (l Layout) WithDirection(direction Direction) Layout {
	l.Direction = direction

	return l
}

// WithPadding returns a copy of the layout with the given padding.
func (l Layout) WithPadding(padding Padding) Layout {
	l.Padding = padding
	return l
}

// WithFlex returns a copy of the layout with the given flex.
func (l Layout) WithFlex(flex Flex) Layout {
	l.Flex = flex

	return l
}

// WithSpacing returns a copy of the layout with the given spacing.
func (l Layout) WithSpacing(spacing int) Layout {
	l.Spacing = spacing

	return l
}

// WithConstraints returns a copy of the layout with the given constraints.
func (l Layout) WithConstraints(constraints ...Constraint) Layout {
	l.Constraints = append(l.Constraints, constraints...)

	return l
}

// SplitWithSpacers splits the given area into smaller ones
// based on the preferred widths or heights and the direction, with the ability to include
// spacers between the areas.
//
// This method is similar to [Layout.Split], but it returns two sets of rectangles: one for the areas
// and one for the spacers.
func (l Layout) SplitWithSpacers(area uv.Rectangle) (segments, spacers Splitted) {
	segments, spacers, err := l.split(area)
	if err != nil {
		panic(err)
	}

	return segments, spacers
}

// Split a given area into smaller ones based on the preferred
// widths or heights and the direction.
//
// Note that the constraints are applied to the whole area that is to be split, so using
// percentages and ratios with the other constraints may not have the desired effect of
// splitting the area up. (e.g. splitting 100 into [min 20, 50%, 50%], may not result
// in [20, 40, 40] but rather an indeterminate result between [20, 50, 30] and [20, 30, 50]).
func (l Layout) Split(area uv.Rectangle) Splitted {
	segments, _ := l.SplitWithSpacers(area)

	return segments
}

func (l Layout) split(area uv.Rectangle) (segments, spacers []uv.Rectangle, err error) {
	solver := casso.NewSolver()

	innerArea := l.Padding.apply(area)

	var areaStart, areaEnd float64

	switch l.Direction {
	case DirectionHorizontal:
		areaStart = float64(innerArea.Min.X) * floatPrecisionMultiplier
		areaEnd = float64(innerArea.Max.X) * floatPrecisionMultiplier

	case DirectionVertical:
		areaStart = float64(innerArea.Min.Y) * floatPrecisionMultiplier
		areaEnd = float64(innerArea.Max.Y) * floatPrecisionMultiplier
	}

	// 	<───────────────────────────────────area_size──────────────────────────────────>
	// 	┌─area_start                                                          area_end─┐
	// 	V                                                                              V
	// 	┌────┬───────────────────┬────┬─────variables─────┬────┬───────────────────┬────┐
	// 	│    │                   │    │                   │    │                   │    │
	// 	V    V                   V    V                   V    V                   V    V
	// 	┌   ┐┌──────────────────┐┌   ┐┌──────────────────┐┌   ┐┌──────────────────┐┌   ┐
	// 	     │     Max(20)      │     │      Max(20)     │     │      Max(20)     │
	// 	└   ┘└──────────────────┘└   ┘└──────────────────┘└   ┘└──────────────────┘└   ┘
	// 	^    ^                   ^    ^                   ^    ^                   ^    ^
	// 	│    │                   │    │                   │    │                   │    │
	// 	└─┬──┶━━━━━━━━━┳━━━━━━━━━┵─┬──┶━━━━━━━━━┳━━━━━━━━━┵─┬──┶━━━━━━━━━┳━━━━━━━━━┵─┬──┘
	// 	  │            ┃           │            ┃           │            ┃           │
	// 	  └────────────╂───────────┴────────────╂───────────┴────────────╂──Spacers──┘
	// 	               ┃                        ┃                        ┃
	// 	               ┗━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━Segments━━━━━━━━┛

	variableCount := len(l.Constraints)*2 + 2

	variables := make([]casso.Symbol, variableCount)
	for i := range variableCount {
		variables[i] = casso.New()
	}

	spacerElements := newElements(variables)
	segmentElements := newElements(variables[1:])

	spacing := l.Spacing

	areaSize := element{
		Start: variables[0],
		End:   variables[len(variables)-1],
	}

	if err := configureArea(&solver, areaSize, areaStart, areaEnd); err != nil {
		return nil, nil, fmt.Errorf("configure area: %w", err)
	}

	if err := configureVariableInAreaConstraints(&solver, variables, areaSize); err != nil {
		return nil, nil, fmt.Errorf("configure variable in area constraints: %w", err)
	}

	if err := configureVariableConstraints(&solver, variables); err != nil {
		return nil, nil, fmt.Errorf("configure variable constraints: %w", err)
	}

	if err := configureFlexConstraints(&solver, areaSize, spacerElements, l.Flex, spacing); err != nil {
		return nil, nil, fmt.Errorf("configure flex constraints: %w", err)
	}

	if err := configureConstraints(&solver, areaSize, segmentElements, l.Constraints, l.Flex); err != nil {
		return nil, nil, fmt.Errorf("configure constraints: %w", err)
	}

	if err := configureFillConstraints(&solver, segmentElements, l.Constraints, l.Flex); err != nil {
		return nil, nil, fmt.Errorf("configure fill constraints: %w", err)
	}

	if l.Flex != FlexLegacy {
		for i := 0; i < len(segmentElements)-1; i++ {
			left := segmentElements[i]
			right := segmentElements[i+1]

			if err := solver.AddConstraint(left.hasSize(right.size(), allSegmentGrow)); err != nil {
				return nil, nil, fmt.Errorf("add has size constraint: %w", err)
			}
		}
	}

	changes := make(map[casso.Symbol]float64, variableCount)

	for _, v := range variables {
		changes[v] = solver.Val(v)
	}

	segments = changesToRects(changes, segmentElements, innerArea, l.Direction)
	spacers = changesToRects(changes, spacerElements, innerArea, l.Direction)

	return segments, spacers, nil
}

func changesToRects(
	changes map[casso.Symbol]float64,
	elements []element,
	area uv.Rectangle,
	direction Direction,
) []uv.Rectangle {
	var rects []uv.Rectangle

	for _, e := range elements {
		start := changes[e.Start]
		end := changes[e.End]

		startRounded := int(math.Round(math.Round(start) / floatPrecisionMultiplier))
		endRounded := int(math.Round(math.Round(end) / floatPrecisionMultiplier))

		size := max(0, endRounded-startRounded)

		switch direction {
		case DirectionHorizontal:
			rect := uv.Rect(startRounded, area.Min.Y, size, area.Dy())

			rects = append(rects, rect)

		case DirectionVertical:
			rect := uv.Rect(area.Min.X, startRounded, area.Dx(), size)

			rects = append(rects, rect)
		}
	}

	return rects
}

// configureFillConstraints makes every [Fill] constraint proportionally equal to each other
// This will make it fill up empty spaces equally
//
//	[Fill(1), Fill(1)]
//	┌──────┐┌──────┐
//	│abcdef││abcdef│
//	└──────┘└──────┘
//
//	[Fill(1), Fill(2)]
//	┌──────┐┌────────────┐
//	│abcdef││abcdefabcdef│
//	└──────┘└────────────┘
//
//	size == base_element * scaling_factor
func configureFillConstraints(
	solver *casso.Solver,
	segments []element,
	constraints []Constraint,
	flex Flex,
) error {
	var (
		validConstraints []Constraint
		validSegments    []element
	)

	for i := 0; i < min(len(constraints), len(segments)); i++ {
		c := constraints[i]
		s := segments[i]

		switch c.(type) {
		case Fill, Min:
			if _, ok := c.(Min); ok && flex == FlexLegacy {
				continue
			}

			validConstraints = append(validConstraints, c)
			validSegments = append(validSegments, s)
		}
	}

	for _, indices := range combinations(len(validConstraints), 2) {
		i, j := indices[0], indices[1]

		leftConstraint := validConstraints[i]
		leftSegment := validSegments[i]

		rightConstraint := validConstraints[j]
		rightSegment := validSegments[j]

		getScalingFactor := func(c Constraint) float64 {
			var scalingFactor float64

			switch c := c.(type) {
			case Fill:
				scale := float64(c)

				scalingFactor = 1e-6
				scalingFactor = max(scalingFactor, scale)

			case Min:
				scalingFactor = 1
			}

			return scalingFactor
		}

		leftScalingFactor := getScalingFactor(leftConstraint)
		rightScalingFactor := getScalingFactor(rightConstraint)

		lhs := leftSegment.size().MulConstant(rightScalingFactor)
		rhs := rightSegment.size().MulConstant(leftScalingFactor)

		constraint := casso.Equal(grow).ExpressionLHS(lhs).ExpressionRHS(rhs)
		if err := solver.AddConstraint(constraint); err != nil {
			return fmt.Errorf("add constraint: %w", err)
		}
	}

	return nil
}

func configureConstraints(
	solver *casso.Solver,
	area element,
	segments []element,
	constraints []Constraint,
	flex Flex,
) error {
	for i := 0; i < min(len(constraints), len(segments)); i++ {
		constraint := constraints[i]
		segment := segments[i]

		switch constraint := constraint.(type) {
		case Max:
			size := int(constraint)

			err := solver.AddConstraints(
				segment.hasMaxSize(size, maxSizeLTE),
				segment.hasIntSize(size, maxSizeEq),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

		case Min:
			size := int(constraint)

			if err := solver.AddConstraint(segment.hasMinSize(size, minSizeGTE)); err != nil {
				return fmt.Errorf("add has min size constraint: %w", err)
			}

			if flex == FlexLegacy {
				if err := solver.AddConstraint(segment.hasIntSize(size, minSizeEq)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			} else {
				if err := solver.AddConstraint(segment.hasSize(area.size(), fillGrow)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

		case Len:
			length := int(constraint)

			if err := solver.AddConstraint(segment.hasIntSize(length, lengthSizeEq)); err != nil {
				return fmt.Errorf("add has int size constraint: %w", err)
			}

		case Percent:
			size := area.size().MulConstant(float64(constraint)).DivConstant(100)

			if err := solver.AddConstraint(segment.hasSize(size, percentSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}

		case Ratio:
			size := area.size().MulConstant(float64(constraint.Num)).DivConstant(float64(max(1, constraint.Den)))

			if err := solver.AddConstraint(segment.hasSize(size, ratioSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}

		case Fill:
			if err := solver.AddConstraint(segment.hasSize(area.size(), fillGrow)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}
	}

	return nil
}

func configureFlexConstraints(
	solver *casso.Solver,
	area element,
	spacers []element,
	flex Flex,
	spacing int,
) error {
	var spacersExceptFirstAndLast []element

	if len(spacers) > 2 {
		spacersExceptFirstAndLast = spacers[1 : len(spacers)-1]
	}

	spacingF := float64(spacing) * floatPrecisionMultiplier

	switch flex {
	case FlexLegacy:
		for _, s := range spacersExceptFirstAndLast {
			if err := solver.AddConstraint(s.hasSize(casso.NewExpressionFromConstant(spacingF), spacerSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			err := solver.AddConstraints(first.isEmpty(), last.isEmpty())
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	// All spacers excluding first and last are the same size and will grow to fill
	// any remaining space after the constraints are satisfied.
	// All spacers excluding first and last are also twice the size of the first and last
	// spacers
	case FlexSpaceEvenly:
		for _, indices := range combinations(len(spacers), 2) {
			i, j := indices[0], indices[1]

			left, right := spacers[i], spacers[j]

			if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		for _, s := range spacers {
			err := solver.AddConstraints(
				s.hasMinSize(spacing, spacerSizeEq),
				s.hasSize(area.size(), spaceGrow),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexSpaceAround:
		// If there are two or less spacers, fallback to [FlexSpaceEvenly].
		if len(spacers) <= 2 {
			for _, indices := range combinations(len(spacers), 2) {
				i, j := indices[0], indices[1]

				left, right := spacers[i], spacers[j]

				if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

			for _, s := range spacers {
				err := solver.AddConstraints(
					s.hasMinSize(spacing, spacerSizeEq),
					s.hasSize(area.size(), spaceGrow),
				)
				if err != nil {
					return fmt.Errorf("add constraints: %w", err)
				}
			}
		} else {
			// Separate the first and last spacer from the middle ones
			first, rest := spacers[0], spacers[1:]
			last, middle := rest[len(rest)-1], rest[:len(rest)-1]

			// All middle spacers should be equal in size
			for _, indices := range combinations(len(middle), 2) {
				i, j := indices[0], indices[1]

				left, right := middle[i], middle[j]

				if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

			// First and last spacers should be half the size of any middle spacer
			if len(middle) > 0 {
				firstMiddle := middle[0]

				for _, e := range []element{first, last} {
					if err := solver.AddConstraint(firstMiddle.hasDoubleSize(e.size(), spacerSizeEq)); err != nil {
						return fmt.Errorf("add has double size constraint: %w", err)
					}
				}
			}

			// Apply minimum size and growth constraints
			for _, s := range spacers {
				if err := solver.AddConstraint(s.hasMinSize(spacing, spacerSizeEq)); err != nil {
					return fmt.Errorf("add has min size constraint: %w", err)
				}

				if err := solver.AddConstraint(s.hasSize(area.size(), spaceGrow)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}
		}

	case FlexSpaceBetween:
		for _, indices := range combinations(len(spacersExceptFirstAndLast), 2) {
			i, j := indices[0], indices[1]

			left, right := spacersExceptFirstAndLast[i], spacersExceptFirstAndLast[j]

			if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		for _, s := range spacersExceptFirstAndLast {
			err := solver.AddConstraints(
				s.hasMinSize(spacing, spacerSizeEq),
				s.hasSize(area.size(), spaceGrow),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			err := solver.AddConstraints(first.isEmpty(), last.isEmpty())
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

		}

	case FlexStart:
		for _, s := range spacersExceptFirstAndLast {
			if err := solver.AddConstraint(s.hasSize(casso.NewExpressionFromConstant(spacingF), spacerSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first := spacers[0]
			last := spacers[len(spacers)-1]

			err := solver.AddConstraints(
				first.isEmpty(),
				last.hasSize(area.size(), grow),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexCenter:
		for _, s := range spacersExceptFirstAndLast {
			constraint := s.hasSize(casso.NewExpressionFromConstant(spacingF), spacerSizeEq)

			if err := solver.AddConstraint(constraint); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			err := solver.AddConstraints(
				first.hasSize(area.size(), grow),
				last.hasSize(area.size(), grow),
				first.hasSize(last.size(), spacerSizeEq),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexEnd:
		for _, s := range spacersExceptFirstAndLast {
			if err := solver.AddConstraint(s.hasSize(casso.NewExpressionFromConstant(spacingF), spacerSizeEq)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first := spacers[0]
			last := spacers[len(spacers)-1]

			err := solver.AddConstraints(
				last.isEmpty(),
				first.hasSize(area.size(), grow),
			)
			if err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}
	}

	return nil
}

func configureVariableConstraints(
	solver *casso.Solver,
	variables []casso.Symbol,
) error {
	// 	┌────┬───────────────────┬────┬─────variables─────┬────┬───────────────────┬────┐
	// 	│    │                   │    │                   │    │                   │    │
	// 	v    v                   v    v                   v    v                   v    v
	// 	┌   ┐┌──────────────────┐┌   ┐┌──────────────────┐┌   ┐┌──────────────────┐┌   ┐
	// 	     │     Max(20)      │     │      Max(20)     │     │      Max(20)     │
	// 	└   ┘└──────────────────┘└   ┘└──────────────────┘└   ┘└──────────────────┘└   ┘
	// 	^    ^                   ^    ^                   ^    ^                   ^    ^
	// 	└v0  └v1                 └v2  └v3                 └v4  └v5                 └v6  └v7

	variables = variables[1:]

	count := len(variables)

	for i := 0; i < count-count%2; i += 2 {
		left, right := variables[i], variables[i+1]

		constraint := casso.LessThanEqual(casso.Required).SymbolLHS(left).SymbolRHS(right)

		if err := solver.AddConstraint(constraint); err != nil {
			return fmt.Errorf("add constraint: %w", err)
		}
	}

	return nil
}

func configureVariableInAreaConstraints(
	solver *casso.Solver,
	variables []casso.Symbol,
	area element,
) error {
	for _, v := range variables {
		start := casso.GreaterThanEqual(casso.Required).SymbolLHS(v).SymbolRHS(area.Start)
		end := casso.LessThanEqual(casso.Required).SymbolLHS(v).SymbolRHS(area.End)

		if err := solver.AddConstraint(start); err != nil {
			return fmt.Errorf("add start constraint: %w", err)
		}

		if err := solver.AddConstraint(end); err != nil {
			return fmt.Errorf("add end constraint: %w", err)
		}
	}

	return nil
}

func configureArea(
	solver *casso.Solver,
	area element,
	areaStart, areaEnd float64,
) error {
	startConstraint := casso.Equal(casso.Required).SymbolLHS(area.Start).ConstantRHS(areaStart)
	endConstraint := casso.Equal(casso.Required).SymbolLHS(area.End).ConstantRHS(areaEnd)

	if err := solver.AddConstraint(startConstraint); err != nil {
		return fmt.Errorf("add start constraint: %w", err)
	}

	if err := solver.AddConstraint(endConstraint); err != nil {
		return fmt.Errorf("add end constraint: %w", err)
	}

	return nil
}

func newElements(variables []casso.Symbol) []element {
	count := len(variables)

	elements := make([]element, 0, count/2+1)

	for i := 0; i < count-count%2; i += 2 {
		start, end := variables[i], variables[i+1]

		elements = append(elements, element{Start: start, End: end})
	}

	return elements
}

type element struct {
	Start, End casso.Symbol
}

func (e element) size() casso.Expression {
	return e.End.Sub(e.Start)
}

func (e element) isEmpty() casso.Constraint {
	return casso.
		Equal(casso.Required - casso.Weak).
		ExpressionLHS(e.size()).
		ConstantRHS(0)
}

func (e element) hasDoubleSize(
	size casso.Expression,
	strength casso.Strength,
) casso.Constraint {
	return casso.
		Equal(strength).
		ExpressionLHS(e.size()).
		ExpressionRHS(size.MulConstant(2))
}

func (e element) hasSize(
	size casso.Expression,
	strength casso.Strength,
) casso.Constraint {
	return casso.
		Equal(strength).
		ExpressionLHS(e.size()).
		ExpressionRHS(size)
}

func (e element) hasMaxSize(
	size int,
	strength casso.Strength,
) casso.Constraint {
	return casso.
		LessThanEqual(strength).
		ExpressionLHS(e.size()).
		ConstantRHS(float64(size) * floatPrecisionMultiplier)
}

func (e element) hasMinSize(
	size int,
	strength casso.Strength,
) casso.Constraint {
	return casso.
		GreaterThanEqual(strength).
		ExpressionLHS(e.size()).
		ConstantRHS(float64(size) * floatPrecisionMultiplier)
}

func (e element) hasIntSize(
	size int,
	strength casso.Strength,
) casso.Constraint {
	return casso.
		Equal(strength).
		ExpressionLHS(e.size()).
		ConstantRHS(float64(size) * floatPrecisionMultiplier)
}

func combinations(n, k int) [][]int {
	combins := binomial(n, k)
	data := make([][]int, combins)
	if len(data) == 0 {
		return nil
	}

	data[0] = make([]int, k)
	for i := range data[0] {
		data[0][i] = i
	}

	for i := 1; i < combins; i++ {
		next := make([]int, k)
		copy(next, data[i-1])
		nextCombination(next, n, k)
		data[i] = next
	}

	return data
}

func nextCombination(s []int, n, k int) {
	for j := k - 1; j >= 0; j-- {
		if s[j] == n+j-k {
			continue
		}

		s[j]++

		for l := j + 1; l < k; l++ {
			s[l] = s[j] + l - j
		}

		break
	}
}

func binomial(n, k int) int {
	if n < 0 || k < 0 {
		panic("layout: binomial: negative input")
	}

	if n < k {
		return 0
	}

	// (n,k) = (n, n-k)
	if k > n/2 {
		k = n - k
	}

	b := 1
	for i := 1; i <= k; i++ {
		b = (n - k + i) * b / i
	}

	return b
}
