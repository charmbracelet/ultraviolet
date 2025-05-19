package tv

// Direction represents a direction in a coordinate system.
type Direction uint8

// Direction constants.
const (
	// Vertical direction.
	Vertical Direction = iota
	// Horizontal direction.
	Horizontal
)

// Constraint represents a constraint on a rectangle.
type Constraint interface {
	// Apply applies the constraint to the given rectangle.
	Apply(length int) int
}

// Length is a constraint that splits a rectangle by a fixed length.
type Length int

// Apply applies the length constraint to the given rectangle.
func (l Length) Apply(length int) int {
	if l < 0 {
		return 0
	}
	return int(l)
}

// Percent is a constraint that splits a rectangle by a percentage of its
// length.
type Percent int

// Apply applies the percent constraint to the given rectangle.
func (p Percent) Apply(length int) int {
	if p < 0 {
		return 0
	}
	return int(p) * length / 100
}

// Ratio returns a [Constraint] that splits a rectangle by a ratio of its
// length. This is a convenience function for creating a [Percent] constraint.
func Ratio(n, d int) Constraint {
	if n <= 0 || d <= 0 {
		return Length(0)
	}
	return Percent(n * 100 / d)
}

// Layout is a set of [Constraint]s and a [Direction] that can be applied to a
// [Rectangle] to split it into smaller ones.
type Layout struct {
	// dir is the direction of the layout.
	dir Direction
	// consts is a list of constraints to apply to the rectangle.
	consts []Constraint
	// The layout margins.
	marginTop, marginRight, marginBot, marginLeft int
	// The space between the split rectangles. Positive values indicate
	// additional space between the rectangles. Negative values indicate
	// overlapping rectangles. Zero means no space between the rectangles.
	spacing int
}

// NewLayout creates a new [Layout] with the given direction and constraints.
// The constrains are applied in the order they are received.
func NewLayout(direction Direction, constraints ...Constraint) *Layout {
	return &Layout{
		dir:    direction,
		consts: constraints,
	}
}

// NewVerticalLayout creates a new vertical [Layout] with the given
// constraints.
func NewVerticalLayout(constraints ...Constraint) *Layout {
	return NewLayout(Vertical, constraints...)
}

// NewHorizontalLayout creates a new horizontal [Layout] with the given
// constraints.
func NewHorizontalLayout(constraints ...Constraint) *Layout {
	return NewLayout(Horizontal, constraints...)
}

// Split splits the given rectangle into smaller ones based on the layout
// constraints and direction. The resulting rectangles are returned as a slice.
func (l *Layout) Split(r Rectangle) []Rectangle {
	if l.dir == Horizontal {
		rects, _, _ := l.splitHorizontal(r)
		return rects
	}
	rects, _, _ := l.splitVertical(r)
	return rects
}

// Areas splits the given rectangle into smaller ones based on the layout
// constraints and direction. The resulting rectangles are returned as a two
// slices: the first slice contains the rectangles, the second slice contains
// any margin rectangles, and the third slice contains any spacing rectangles.
func (l *Layout) Areas(r Rectangle) ([]Rectangle, []Rectangle, []Rectangle) {
	if l.dir == Horizontal {
		return l.splitHorizontal(r)
	}
	return l.splitVertical(r)
}

// Spacing sets the spacing between the split rectangles. Positive values
// indicate additional space between the rectangles. Negative values indicate
// overlapping rectangles. Zero means no space between the rectangles.
func (l *Layout) Spacing(s int) *Layout {
	l.spacing = s
	return l
}

// Margin sets the margins for the layout. One value sets all margins to the
// same value, two values set the vertical and horizontal margins, three values
// set the top, horizontal, and bottom margins, and four values set the top,
// right, bottom, and left margins. More than four values are ignored.
func (l *Layout) Margin(m ...int) *Layout {
	switch len(m) {
	case 1:
		return l.MarginTop(m[0]).MarginRight(m[0]).MarginBottom(m[0]).MarginLeft(m[0])
	case 2:
		return l.MarginTop(m[0]).MarginRight(m[1]).MarginBottom(m[0]).MarginLeft(m[1])
	case 3:
		return l.MarginTop(m[0]).MarginRight(m[1]).MarginBottom(m[2]).MarginLeft(m[1])
	case 4:
		return l.MarginTop(m[0]).MarginRight(m[1]).MarginBottom(m[2]).MarginLeft(m[3])
	default:
		return l
	}
}

// MarginTop sets the top margin for the layout.
func (l *Layout) MarginTop(m int) *Layout {
	l.marginTop = m
	return l
}

// MarginRight sets the right margin for the layout.
func (l *Layout) MarginRight(m int) *Layout {
	l.marginRight = m
	return l
}

// MarginBottom sets the bottom margin for the layout.
func (l *Layout) MarginBottom(m int) *Layout {
	l.marginBot = m
	return l
}

// MarginLeft sets the left margin for the layout.
func (l *Layout) MarginLeft(m int) *Layout {
	l.marginLeft = m
	return l
}

// splitHorizontal splits a rectangle into smaller ones. The resulting
// rectangles are returned as three slices: the first slice contains the area
// rectangles, the second slice contains the margin rectangles, and the third
// slice contains the spacing rectangles.
// Each constraint is applied to the remaining rectangle in the order they are
// received.
func (l *Layout) splitHorizontal(r Rectangle) ([]Rectangle, []Rectangle, []Rectangle) {
	var areas, margins, spacings []Rectangle

	// Apply margins to the rectangle
	if l.marginTop > 0 {
		margins = append(margins, Rect(r.Min.X, r.Min.Y, r.Dx(), l.marginTop))
	}
	if l.marginRight > 0 {
		margins = append(margins, Rect(r.Max.X-l.marginRight, r.Min.Y+l.marginTop, l.marginRight, r.Dy()-l.marginTop-l.marginBot))
	}
	if l.marginBot > 0 {
		margins = append(margins, Rect(r.Min.X, r.Max.Y-l.marginBot, r.Dx(), l.marginBot))
	}
	if l.marginLeft > 0 {
		margins = append(margins, Rect(r.Min.X, r.Min.Y+l.marginTop, l.marginLeft, r.Dy()-l.marginTop-l.marginBot))
	}

	// Apply the margins to the rectangle
	innerRect := Rect(
		r.Min.X+l.marginLeft,
		r.Min.Y+l.marginTop,
		r.Dx()-l.marginLeft-l.marginRight,
		r.Dy()-l.marginTop-l.marginBot,
	)

	// If there are no constraints, return the rectangle as is
	if len(l.consts) == 0 {
		areas = append(areas, innerRect)
		return areas, margins, spacings
	}

	// Apply the constraints sequentially to the remaining area
	x := innerRect.Min.X
	remainingWidth := innerRect.Dx()

	for i, c := range l.consts {
		width := min(c.Apply(remainingWidth), remainingWidth)

		// Add the area
		areas = append(areas, Rect(x, innerRect.Min.Y, width, innerRect.Dy()))

		// Update position and remaining width
		x += width
		remainingWidth -= width

		// Add spacing if not the last constraint and we have room
		if i < len(l.consts)-1 && l.spacing != 0 && remainingWidth > l.spacing {
			spacings = append(spacings, Rect(x, innerRect.Min.Y, l.spacing, innerRect.Dy()))
			x += l.spacing
			remainingWidth -= l.spacing
		}

		// If we've run out of space, stop processing constraints
		if remainingWidth <= 0 {
			break
		}
	}

	// If there's remaining width after all constraints, add it as an additional area
	if remainingWidth > 0 {
		if l.spacing != 0 && remainingWidth > l.spacing {
			spacings = append(spacings, Rect(x, innerRect.Min.Y, l.spacing, innerRect.Dy()))
			x += l.spacing
			remainingWidth -= l.spacing
		}
		areas = append(areas, Rect(x, innerRect.Min.Y, remainingWidth, innerRect.Dy()))
	}

	return areas, margins, spacings
}

// splitVertical splits a rectangle into smaller ones. The resulting
// rectangles are returned as three slices: the first slice contains the area
// rectangles, the second slice contains the margin rectangles, and the third
// slice contains the spacing rectangles.
// Each constraint is applied to the remaining rectangle in the order they are
// received.
func (l *Layout) splitVertical(r Rectangle) ([]Rectangle, []Rectangle, []Rectangle) {
	var areas, margins, spacings []Rectangle

	// Apply margins to the rectangle
	if l.marginTop > 0 {
		margins = append(margins, Rect(r.Min.X, r.Min.Y, r.Dx(), l.marginTop))
	}
	if l.marginRight > 0 {
		margins = append(margins, Rect(r.Max.X-l.marginRight, r.Min.Y+l.marginTop, l.marginRight, r.Dy()-l.marginTop-l.marginBot))
	}
	if l.marginBot > 0 {
		margins = append(margins, Rect(r.Min.X, r.Max.Y-l.marginBot, r.Dx(), l.marginBot))
	}
	if l.marginLeft > 0 {
		margins = append(margins, Rect(r.Min.X, r.Min.Y+l.marginTop, l.marginLeft, r.Dy()-l.marginTop-l.marginBot))
	}

	// Apply the margins to the rectangle
	innerRect := Rect(
		r.Min.X+l.marginLeft,
		r.Min.Y+l.marginTop,
		r.Dx()-l.marginLeft-l.marginRight,
		r.Dy()-l.marginTop-l.marginBot,
	)

	// If there are no constraints, return the rectangle as is
	if len(l.consts) == 0 {
		areas = append(areas, innerRect)
		return areas, margins, spacings
	}

	// Apply the constraints sequentially to the remaining area
	y := innerRect.Min.Y
	remainingHeight := innerRect.Dy()

	for i, c := range l.consts {
		height := min(c.Apply(remainingHeight), remainingHeight)

		// Add the area
		areas = append(areas, Rect(innerRect.Min.X, y, innerRect.Dx(), height))

		// Update position and remaining height
		y += height
		remainingHeight -= height

		// Add spacing if not the last constraint and we have room
		if i < len(l.consts)-1 && l.spacing != 0 && remainingHeight > l.spacing {
			spacings = append(spacings, Rect(innerRect.Min.X, y, innerRect.Dx(), l.spacing))
			y += l.spacing
			remainingHeight -= l.spacing
		}

		// If we've run out of space, stop processing constraints
		if remainingHeight <= 0 {
			break
		}
	}

	// If there's remaining height after all constraints, add it as an additional area
	if remainingHeight > 0 {
		if l.spacing != 0 && remainingHeight > l.spacing {
			spacings = append(spacings, Rect(innerRect.Min.X, y, innerRect.Dx(), l.spacing))
			y += l.spacing
			remainingHeight -= l.spacing
		}
		areas = append(areas, Rect(innerRect.Min.X, y, innerRect.Dx(), remainingHeight))
	}

	return areas, margins, spacings
}
