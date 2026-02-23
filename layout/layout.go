// Package layout partitions terminal screen space into rectangular regions
// using a constraint-based solver.
//
// Under the hood, it relies on the [Cassowary constraint solver algorithm] to
// resolve competing size requirements. Each constraint carries a priority, so
// the solver can pick the best trade-off when not every requirement fits.
//
// # How It Works
//
// A [Layout] takes the available area and a list of constraints ([Len], [Ratio],
// [Percent], [Fill], [Min], [Max]) and produces a set of non-overlapping rectangles.
// The solver tries to honour every constraint; when that is impossible it
// relaxes lower-priority ones first.
//
// You are not required to use [Layout] at all. If you prefer manual control,
// you can compute [uv.Rectangle] values yourself with plain arithmetic.
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
	"hash/fnv"
	"math"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/casso"
)

// floatPrecisionMultiplier scales cell positions into a higher-precision
// floating-point domain before handing them to the constraint solver.
// The number of trailing zeros determines the decimal precision kept
// during rounding.
const floatPrecisionMultiplier float64 = 100.0

const (
	required casso.Priority = 1_001_001_000
	strong   casso.Priority = 1_000_000
	medium   casso.Priority = 1_000
	weak     casso.Priority = 1
)

const (
	// spacerSizeEq enforces equal sizing across spacers.
	//
	// 	┌     ┐┌───┐┌     ┐┌───┐┌     ┐
	// 	  ==x  │   │  ==x  │   │  ==x
	// 	└     ┘└───┘└     ┘└───┘└     ┘
	spacerSizeEq = required / 10.0

	// minSizeGTE enforces the lower-bound inequality for [Min] constraints.
	//
	// 	┌────────┐
	// 	│Min(>=x)│
	// 	└────────┘
	minSizeGTE = strong * 100.0

	// maxSizeLTE enforces the upper-bound inequality for [Max] constraints.
	//
	// 	┌────────┐
	// 	│Max(<=x)│
	// 	└────────┘
	maxSizeLTE = strong * 100.0

	// lengthSizeEq pins the segment to the exact size requested by a [Len] constraint.
	//
	// 	┌────────┐
	// 	│Len(==x)│
	// 	└────────┘
	lengthSizeEq = strong * 10.0

	// percentSizeEq tries to make the segment match its [Percent] target.
	//
	// 	┌────────────┐
	// 	│Percent(==x)│
	// 	└────────────┘
	percentSizeEq = strong

	// ratioSizeEq tries to make the segment match its [Ratio] target.
	//
	// 	┌────────────┐
	// 	│Ratio(==x,y)│
	// 	└────────────┘
	ratioSizeEq = strong / 10.0

	// minSizeEq is an equality companion for the [Min] lower-bound; it
	// nudges the segment toward the minimum value when room is tight.
	//
	// 	┌────────┐
	// 	│Min(==x)│
	// 	└────────┘
	minSizeEq = medium * 10.0

	// maxSizeEq is an equality companion for the [Max] upper-bound;
	// it nudges the segment toward the maximum value.
	//
	// 	┌────────┐
	// 	│Max(==x)│
	// 	└────────┘
	maxSizeEq = medium * 10.0

	// fillGrow lets [Fill] segments expand into available space.
	//
	// 	┌─────────────────────┐
	// 	│<=     Fill(x)     =>│
	// 	└─────────────────────┘
	fillGrow = medium

	// grow is a general expansion priority (used by [Min] in non-legacy flex).
	//
	// 	┌────────────┐
	// 	│<= Min(x) =>│
	// 	└────────────┘
	grow casso.Priority = 100.0

	// spaceGrow allows spacers to expand and absorb remaining room.
	//
	// 	┌       ┐
	// 	 <= x =>
	// 	└       ┘
	spaceGrow = weak * 10.0

	// allSegmentGrow encourages all segments to share the same size.
	//
	// 	┌───────┐
	// 	│<= x =>│
	// 	└───────┘
	allSegmentGrow = weak
)

// Splitted holds the rectangles produced by a [Layout.Split] call.
type Splitted []uv.Rectangle

// Assign stores each resulting rectangle into the corresponding pointer.
//
// Nil pointers are silently skipped.
//
// Panics when len(areas) exceeds the number of rectangles in [Splitted].
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

// Direction controls whether a [Layout] arranges its segments
// horizontally (left to right) or vertically (top to bottom).
type Direction int

const (
	// DirectionVertical - layout segments are arranged top to bottom (default).
	DirectionVertical Direction = iota
	// DirectionHorizontal - layout segments are arranged side by side (left to right).
	DirectionHorizontal
)

// New returns a [Layout] configured with the given direction and constraints.
func New(direction Direction, constraints ...Constraint) Layout {
	return Layout{
		Direction:   direction,
		Constraints: constraints,
	}
}

// Vertical is shorthand for New(DirectionVertical, constraints...).
func Vertical(constraints ...Constraint) Layout {
	return New(DirectionVertical, constraints...)
}

// Horizontal is shorthand for New(DirectionHorizontal, constraints...).
func Horizontal(constraints ...Constraint) Layout {
	return New(DirectionHorizontal, constraints...)
}

// Layout splits a rectangular area into smaller rectangles using a set of constraints.
// It is the primary building block for structuring terminal user interfaces.
//
// Fields:
//   - Direction: whether segments flow vertically or horizontally.
//   - Constraints: the sizing rules ([Len], [Ratio], [Percent], [Fill], [Min], [Max]).
//   - Padding: inset applied to the outer area before solving.
//   - Flex: strategy for distributing leftover space among segments.
//   - Spacing: gap (or overlap, if negative) between adjacent segments.
//
// Internally, sizes are resolved by a Cassowary linear-constraint solver that
// satisfies as many rules as it can, preferring higher-priority constraints
// when trade-offs are necessary.
type Layout struct {
	Direction   Direction
	Constraints []Constraint
	Padding     Padding
	// Spacing is the gap between adjacent segments, measured in cells.
	// A negative value causes segments to overlap by that many cells.
	Spacing int
	Flex    Flex
}

// WithDirection returns a shallow copy of the layout using the specified direction.
func (l Layout) WithDirection(direction Direction) Layout {
	l.Direction = direction

	return l
}

// WithPadding returns a shallow copy of the layout using the specified padding.
func (l Layout) WithPadding(padding Padding) Layout {
	l.Padding = padding
	return l
}

// WithFlex returns a shallow copy of the layout using the specified flex strategy.
func (l Layout) WithFlex(flex Flex) Layout {
	l.Flex = flex

	return l
}

// WithSpacing returns a shallow copy of the layout using the specified spacing value.
func (l Layout) WithSpacing(spacing int) Layout {
	l.Spacing = spacing

	return l
}

// WithConstraints returns a shallow copy of the layout with the given
// constraints appended to its existing list.
func (l Layout) WithConstraints(constraints ...Constraint) Layout {
	l.Constraints = append(l.Constraints, constraints...)

	return l
}

// SplitWithSpacers divides the given area into content segments and the
// gaps (spacers) between them. It returns both slices; use [Layout.Split]
// if you only need the content rectangles.
func (l Layout) SplitWithSpacers(area uv.Rectangle) (segments, spacers Splitted) {
	segments, spacers, err := l.splitCached(area)
	if err != nil {
		panic(err)
	}

	return segments, spacers
}

// Split partitions the area into content rectangles according to the
// layout's direction and constraints.
//
// Because every constraint is evaluated against the total area, mixing
// relative constraints (Percent, Ratio) with absolute ones (Min, Max, Len)
// can produce ambiguous results. For example, splitting 100 cells as
// [Min(20), Percent(50), Percent(50)] will not necessarily yield [20, 40, 40].
func (l Layout) Split(area uv.Rectangle) Splitted {
	segments, _ := l.SplitWithSpacers(area)

	return segments
}

func (l Layout) splitCached(area uv.Rectangle) (segments, spacers []uv.Rectangle, err error) {
	globalCacheMu.Lock()
	defer globalCacheMu.Unlock()

	key := l.cacheKey(area)

	if v, ok := globalCache.Get(key); ok {
		return v.Segments, v.Spacers, nil
	}

	segments, spacers, err = l.split(area)
	if err != nil {
		return nil, nil, err
	}

	globalCache.Add(key, cacheValue{Segments: segments, Spacers: spacers})

	return segments, spacers, nil
}

func (l Layout) split(area uv.Rectangle) (segments, spacers []uv.Rectangle, err error) {
	s := casso.NewSolver()

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

	areaEl := element{
		start: variables[0],
		end:   variables[len(variables)-1],
	}

	if err := configureArea(s, areaEl, areaStart, areaEnd); err != nil {
		return nil, nil, fmt.Errorf("configure area: %w", err)
	}

	if err := configureVariableInAreaConstraints(s, variables, areaEl); err != nil {
		return nil, nil, fmt.Errorf("configure variable in area constraints: %w", err)
	}

	if err := configureVariableConstraints(s, variables); err != nil {
		return nil, nil, fmt.Errorf("configure variable constraints: %w", err)
	}

	if err := configureFlexConstraints(s, areaEl, spacerElements, l.Flex, spacing); err != nil {
		return nil, nil, fmt.Errorf("configure flex constraints: %w", err)
	}

	if err := configureConstraints(s, areaEl, segmentElements, l.Constraints, l.Flex); err != nil {
		return nil, nil, fmt.Errorf("configure constraints: %w", err)
	}

	if err := configureFillConstraints(s, segmentElements, l.Constraints, l.Flex); err != nil {
		return nil, nil, fmt.Errorf("configure fill constraints: %w", err)
	}

	if l.Flex != FlexLegacy {
		for i := 0; i < len(segmentElements)-1; i++ {
			left := segmentElements[i]
			right := segmentElements[i+1]

			if _, err := s.Add(allSegmentGrow, left.sizeEqSize(right)); err != nil {
				return nil, nil, fmt.Errorf("add has size constraint: %w", err)
			}
		}
	}

	changes := make(map[casso.Symbol]float64, variableCount)

	for _, v := range variables {
		changes[v] = s.Val(v)
	}

	segments = changesToRects(changes, segmentElements, innerArea, l.Direction)
	spacers = changesToRects(changes, spacerElements, innerArea, l.Direction)

	return segments, spacers, nil
}

func (l Layout) cacheKey(area uv.Rectangle) cacheKey {
	h := fnv.New64a()

	for _, c := range l.Constraints {
		c.hash(h)
	}

	return cacheKey{
		Area:            area,
		Direction:       l.Direction,
		ConstraintsHash: h.Sum64(),
		Padding:         l.Padding,
		Spacing:         l.Spacing,
		Flex:            l.Flex,
	}
}

func changesToRects(
	changes map[casso.Symbol]float64,
	elements []element,
	area uv.Rectangle,
	direction Direction,
) []uv.Rectangle {
	var rects []uv.Rectangle

	for _, e := range elements {
		startVal := changes[e.start]
		endVal := changes[e.end]

		startRounded := int(math.Round(math.Round(startVal) / floatPrecisionMultiplier))
		endRounded := int(math.Round(math.Round(endVal) / floatPrecisionMultiplier))

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

// configureFillConstraints ensures that every [Fill] (and, outside legacy mode,
// every [Min]) segment grows proportionally to its scaling factor, so that
// remaining space is shared according to the declared weights.
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
	s *casso.Solver,
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
		seg := segments[i]

		switch c.(type) {
		case Fill, Min:
			if _, ok := c.(Min); ok && flex == FlexLegacy {
				continue
			}

			validConstraints = append(validConstraints, c)
			validSegments = append(validSegments, seg)
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

		c := casso.NewConstraint(casso.EQ, 0,
			leftSegment.end.T(rightScalingFactor),
			leftSegment.start.T(-rightScalingFactor),
			rightSegment.end.T(-leftScalingFactor),
			rightSegment.start.T(leftScalingFactor),
		)

		if _, err := s.Add(grow, c); err != nil {
			return fmt.Errorf("add constraint: %w", err)
		}
	}

	return nil
}

func configureConstraints(
	s *casso.Solver,
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

			if _, err := s.Add(maxSizeLTE, segment.sizeLTE(size)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(maxSizeEq, segment.sizeEqConst(size)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

		case Min:
			size := int(constraint)

			if _, err := s.Add(minSizeGTE, segment.sizeGTE(size)); err != nil {
				return fmt.Errorf("add has min size constraint: %w", err)
			}

			if flex == FlexLegacy {
				if _, err := s.Add(minSizeEq, segment.sizeEqConst(size)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			} else {
				if _, err := s.Add(fillGrow, segment.sizeEqSize(area)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

		case Len:
			length := int(constraint)

			if _, err := s.Add(lengthSizeEq, segment.sizeEqConst(length)); err != nil {
				return fmt.Errorf("add has int size constraint: %w", err)
			}

		case Percent:
			f := float64(constraint) / 100

			if _, err := s.Add(percentSizeEq, segment.sizeEqScaledSize(area, f)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}

		case Ratio:
			f := float64(constraint.Num) / float64(max(1, constraint.Den))

			if _, err := s.Add(ratioSizeEq, segment.sizeEqScaledSize(area, f)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}

		case Fill:
			if _, err := s.Add(fillGrow, segment.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}
	}

	return nil
}

func configureFlexConstraints(
	s *casso.Solver,
	area element,
	spacers []element,
	flex Flex,
	spacing int,
) error {
	var spacersExceptFirstAndLast []element

	if len(spacers) > 2 {
		spacersExceptFirstAndLast = spacers[1 : len(spacers)-1]
	}

	switch flex {
	case FlexLegacy:
		for _, sp := range spacersExceptFirstAndLast {
			if _, err := s.Add(spacerSizeEq, sp.sizeEqConst(spacing)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			if _, err := s.Add(required-weak, first.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(required-weak, last.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexSpaceEvenly:
		for _, indices := range combinations(len(spacers), 2) {
			i, j := indices[0], indices[1]

			left, right := spacers[i], spacers[j]

			if _, err := s.Add(spacerSizeEq, left.sizeEqSize(right)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		for _, sp := range spacers {
			if _, err := s.Add(spacerSizeEq, sp.sizeGTE(spacing)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(spaceGrow, sp.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexSpaceAround:
		if len(spacers) <= 2 {
			for _, indices := range combinations(len(spacers), 2) {
				i, j := indices[0], indices[1]

				left, right := spacers[i], spacers[j]

				if _, err := s.Add(spacerSizeEq, left.sizeEqSize(right)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

			for _, sp := range spacers {
				if _, err := s.Add(spacerSizeEq, sp.sizeGTE(spacing)); err != nil {
					return fmt.Errorf("add constraints: %w", err)
				}

				if _, err := s.Add(spaceGrow, sp.sizeEqSize(area)); err != nil {
					return fmt.Errorf("add constraints: %w", err)
				}
			}
		} else {
			first, rest := spacers[0], spacers[1:]
			last, middle := rest[len(rest)-1], rest[:len(rest)-1]

			for _, indices := range combinations(len(middle), 2) {
				i, j := indices[0], indices[1]

				left, right := middle[i], middle[j]

				if _, err := s.Add(spacerSizeEq, left.sizeEqSize(right)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}

			if len(middle) > 0 {
				firstMiddle := middle[0]

				for _, e := range []element{first, last} {
					if _, err := s.Add(spacerSizeEq, firstMiddle.sizeEqDouble(e)); err != nil {
						return fmt.Errorf("add has double size constraint: %w", err)
					}
				}
			}

			for _, sp := range spacers {
				if _, err := s.Add(spacerSizeEq, sp.sizeGTE(spacing)); err != nil {
					return fmt.Errorf("add has min size constraint: %w", err)
				}

				if _, err := s.Add(spaceGrow, sp.sizeEqSize(area)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
			}
		}

	case FlexSpaceBetween:
		for _, indices := range combinations(len(spacersExceptFirstAndLast), 2) {
			i, j := indices[0], indices[1]

			left, right := spacersExceptFirstAndLast[i], spacersExceptFirstAndLast[j]

			if _, err := s.Add(spacerSizeEq, left.sizeEqSize(right)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		for _, sp := range spacersExceptFirstAndLast {
			if _, err := s.Add(spacerSizeEq, sp.sizeGTE(spacing)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(spaceGrow, sp.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			if _, err := s.Add(required-weak, first.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(required-weak, last.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexStart:
		for _, sp := range spacersExceptFirstAndLast {
			if _, err := s.Add(spacerSizeEq, sp.sizeEqConst(spacing)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first := spacers[0]
			last := spacers[len(spacers)-1]

			if _, err := s.Add(required-weak, first.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(grow, last.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexCenter:
		for _, sp := range spacersExceptFirstAndLast {
			if _, err := s.Add(spacerSizeEq, sp.sizeEqConst(spacing)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first, last := spacers[0], spacers[len(spacers)-1]

			if _, err := s.Add(grow, first.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(grow, last.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(spacerSizeEq, first.sizeEqSize(last)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}

	case FlexEnd:
		for _, sp := range spacersExceptFirstAndLast {
			if _, err := s.Add(spacerSizeEq, sp.sizeEqConst(spacing)); err != nil {
				return fmt.Errorf("add has size constraint: %w", err)
			}
		}

		if len(spacers) >= 2 {
			first := spacers[0]
			last := spacers[len(spacers)-1]

			if _, err := s.Add(required-weak, last.empty()); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}

			if _, err := s.Add(grow, first.sizeEqSize(area)); err != nil {
				return fmt.Errorf("add constraints: %w", err)
			}
		}
	}

	return nil
}

func configureVariableConstraints(
	s *casso.Solver,
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

		if _, err := s.Add(required, casso.NewConstraint(casso.LTE, 0, left.T(1), right.T(-1))); err != nil {
			return fmt.Errorf("add constraint: %w", err)
		}
	}

	return nil
}

func configureVariableInAreaConstraints(
	s *casso.Solver,
	variables []casso.Symbol,
	area element,
) error {
	for _, v := range variables {
		if _, err := s.Add(required, casso.NewConstraint(casso.GTE, 0, v.T(1), area.start.T(-1))); err != nil {
			return fmt.Errorf("add start constraint: %w", err)
		}

		if _, err := s.Add(required, casso.NewConstraint(casso.LTE, 0, v.T(1), area.end.T(-1))); err != nil {
			return fmt.Errorf("add end constraint: %w", err)
		}
	}

	return nil
}

func configureArea(
	s *casso.Solver,
	area element,
	areaStart, areaEnd float64,
) error {
	if _, err := s.Add(required, casso.NewConstraint(casso.EQ, -areaStart, area.start.T(1))); err != nil {
		return fmt.Errorf("add start constraint: %w", err)
	}

	if _, err := s.Add(required, casso.NewConstraint(casso.EQ, -areaEnd, area.end.T(1))); err != nil {
		return fmt.Errorf("add end constraint: %w", err)
	}

	return nil
}

func newElements(variables []casso.Symbol) []element {
	count := len(variables)

	elements := make([]element, 0, count/2+1)

	for i := 0; i < count-count%2; i += 2 {
		s, e := variables[i], variables[i+1]

		elements = append(elements, element{start: s, end: e})
	}

	return elements
}

type element struct {
	start, end casso.Symbol
}

func (e element) empty() casso.Constraint {
	return casso.NewConstraint(casso.EQ, 0, e.end.T(1), e.start.T(-1))
}

func (e element) sizeEqConst(size int) casso.Constraint {
	return casso.NewConstraint(casso.EQ, -float64(size)*floatPrecisionMultiplier, e.end.T(1), e.start.T(-1))
}

func (e element) sizeLTE(size int) casso.Constraint {
	return casso.NewConstraint(casso.LTE, -float64(size)*floatPrecisionMultiplier, e.end.T(1), e.start.T(-1))
}

func (e element) sizeGTE(size int) casso.Constraint {
	return casso.NewConstraint(casso.GTE, -float64(size)*floatPrecisionMultiplier, e.end.T(1), e.start.T(-1))
}

func (e element) sizeEqSize(other element) casso.Constraint {
	return casso.NewConstraint(casso.EQ, 0, e.end.T(1), e.start.T(-1), other.end.T(-1), other.start.T(1))
}

func (e element) sizeEqScaledSize(other element, f float64) casso.Constraint {
	return casso.NewConstraint(casso.EQ, 0, e.end.T(1), e.start.T(-1), other.end.T(-f), other.start.T(f))
}

func (e element) sizeEqDouble(other element) casso.Constraint {
	return casso.NewConstraint(casso.EQ, 0, e.end.T(1), e.start.T(-1), other.end.T(-2), other.start.T(2))
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
