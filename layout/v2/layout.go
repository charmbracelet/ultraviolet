package layout

import (
	"fmt"
	"math"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/casso"
)

const floatPrecisionMultiplier float64 = 100.0

const (
	spacerSizeEq     casso.Strength = casso.Required / 10.0
	minSizeGTE       casso.Strength = casso.Strong * 100.0
	maxSizeLTE       casso.Strength = casso.Strong * 100.0
	lengthSizeEq     casso.Strength = casso.Strong * 10.0
	percentageSizeEq casso.Strength = casso.Strong
	ratioSizeEq      casso.Strength = casso.Strong / 10.0
	minSizeEq        casso.Strength = casso.Medium * 10.0
	maxSizeEq        casso.Strength = casso.Medium * 10.0
	grow             casso.Strength = 100.0
	fillGrow         casso.Strength = casso.Medium
	spaceGrow        casso.Strength = casso.Weak * 10.0
	allSegmentGrow   casso.Strength = casso.Weak
)

type Splitted []uv.Rectangle

func (s Splitted) Assign(areas ...*uv.Rectangle) {
	for i := range areas {
		if areas[i] != nil {
			*areas[i] = s[i]
		}
	}
}

type Direction int

const (
	DirectionVertical Direction = iota
	DirectionHorizontal
)

func Vertical(constraints ...Constraint) Layout {
	return New(DirectionVertical, constraints...)
}

func Horizontal(constraints ...Constraint) Layout {
	return New(DirectionHorizontal, constraints...)
}

func New(direction Direction, constraints ...Constraint) Layout {
	return Layout{
		Direction:   direction,
		Constraints: constraints,
		Flex:        FlexLegacy,
		Spacing:     SpacingSpace(0),
	}
}

type Layout struct {
	Direction   Direction
	Constraints []Constraint
	Padding     Padding
	Spacing     Spacing
	Flex        Flex
}

func (l Layout) WithDirection(direction Direction) Layout {
	l.Direction = direction

	return l
}

func (l Layout) WithPadding(padding Padding) Layout {
	l.Padding = padding
	return l
}

func (l Layout) WithFlex(flex Flex) Layout {
	l.Flex = flex
	return l
}

func (l Layout) WithSpacing(spacing Spacing) Layout {
	l.Spacing = spacing
	return l
}

func (l Layout) WithConstraints(constraints ...Constraint) Layout {
	l.Constraints = append(l.Constraints, constraints...)
	return l
}

func (l Layout) SplitWithSpacers(area uv.Rectangle) (segments, spacers Splitted) {
	segments, spacers, err := l.split(area)
	if err != nil {
		panic(err)
	}

	return segments, spacers
}

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

	variableCount := len(l.Constraints)*2 + 2

	variables := make([]casso.Variable, variableCount)
	for i := range variableCount {
		variables[i] = casso.NewVariable()
	}

	spacerElements := newElements(variables)
	segmentElements := newElements(variables[1:])

	var spacing int

	switch s := l.Spacing.(type) {
	case SpacingSpace:
		spacing = int(s)

	case SpacingOverlap:
		spacing = -int(s)
	}

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

	fetched := solver.FetchChanges()

	changes := make(map[casso.Variable]float64, len(fetched))
	for _, c := range fetched {
		changes[c.Variable] = c.Constant
	}

	segments = changesToRects(changes, segmentElements, innerArea, l.Direction)
	spacers = changesToRects(changes, spacerElements, innerArea, l.Direction)

	return segments, spacers, nil
}

func changesToRects(
	changes map[casso.Variable]float64,
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

	if len(validConstraints) < 2 {
		return nil
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

		case Percentage:
			size := area.size().MulConstant(float64(constraint)).DivConstant(100)

			if err := solver.AddConstraint(segment.hasSize(size, percentageSizeEq)); err != nil {
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

	case FlexSpaceAround:
		if len(spacersExceptFirstAndLast) >= 2 {
			for _, indices := range combinations(len(spacersExceptFirstAndLast), 2) {
				i, j := indices[0], indices[1]

				left, right := spacersExceptFirstAndLast[i], spacersExceptFirstAndLast[j]

				if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
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

	case FlexSpaceBetween:
		if len(spacersExceptFirstAndLast) >= 2 {
			for _, indices := range combinations(len(spacersExceptFirstAndLast), 2) {
				i, j := indices[0], indices[1]

				left, right := spacersExceptFirstAndLast[i], spacersExceptFirstAndLast[j]

				if err := solver.AddConstraint(left.hasSize(right.size(), spacerSizeEq)); err != nil {
					return fmt.Errorf("add has size constraint: %w", err)
				}
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
	variables []casso.Variable,
) error {
	variables = variables[1:]

	count := len(variables)

	for i := 0; i < count-count%2; i += 2 {
		left, right := variables[i], variables[i+1]

		constraint := casso.LessThanEqual(casso.Required).VariableLHS(left).VariableRHS(right)

		if err := solver.AddConstraint(constraint); err != nil {
			return fmt.Errorf("add constraint: %w", err)
		}
	}

	return nil
}

func configureVariableInAreaConstraints(
	solver *casso.Solver,
	variables []casso.Variable,
	area element,
) error {
	for _, v := range variables {
		start := casso.GreaterThanEqual(casso.Required).VariableLHS(v).VariableRHS(area.Start)
		end := casso.LessThanEqual(casso.Required).VariableLHS(v).VariableRHS(area.End)

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
	startConstraint := casso.Equal(casso.Required).VariableLHS(area.Start).ConstantRHS(areaStart)
	endConstraint := casso.Equal(casso.Required).VariableLHS(area.End).ConstantRHS(areaEnd)

	if err := solver.AddConstraint(startConstraint); err != nil {
		return fmt.Errorf("add start constraint: %w", err)
	}

	if err := solver.AddConstraint(endConstraint); err != nil {
		return fmt.Errorf("add end constraint: %w", err)
	}

	return nil
}

func newElements(variables []casso.Variable) []element {
	count := len(variables)

	elements := make([]element, 0, count/2+1)

	for i := 0; i < count-count%2; i += 2 {
		start, end := variables[i], variables[i+1]

		elements = append(elements, element{Start: start, End: end})
	}

	return elements
}

type element struct {
	Start, End casso.Variable
}

func (e element) size() casso.Expression {
	return e.End.Sub(e.Start)
}

func (e element) isEmpty() casso.Constraint {
	return casso.
		Equal(casso.Required - 1).
		ExpressionLHS(e.size()).
		ConstantRHS(0)
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
		return data
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
		panic("layout: binomial: bad set size")
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
