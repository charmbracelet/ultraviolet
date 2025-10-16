package uv

// Constraint represents a size constraint for layout purposes.
type Constraint interface {
	// Apply applies the constraint to the given size and returns the
	// constrained size.
	Apply(size int) int
}

// Percent is a constraint that represents a percentage of the available size.
type Percent int

// Apply applies the percentage constraint to the given size.
func (p Percent) Apply(size int) int {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return size
	}
	return size * int(p) / 100
}

// Ratio is a constraint that represents a ratio of the available size. It is a
// syntactic sugar for [Percent].
func Ratio(numerator, denominator int) Percent {
	if denominator == 0 {
		return 0
	}
	return Percent(numerator * 100 / denominator)
}

// Fixed is a constraint that represents a fixed size.
type Fixed int

// Apply applies the fixed size constraint to the given size.
func (f Fixed) Apply(size int) int {
	if f < 0 {
		return 0
	}
	if int(f) > size {
		return size
	}
	return int(f)
}

// SplitVertical splits the area vertically into two parts based on the given
// [Constraint].
//
// It returns the top and bottom rectangles.
func SplitVertical(area Rectangle, constraint Constraint) (top Rectangle, bottom Rectangle) {
	height := min(constraint.Apply(area.Dy()), area.Dy())
	top = Rectangle{Min: area.Min, Max: Position{X: area.Max.X, Y: area.Min.Y + height}}
	bottom = Rectangle{Min: Position{X: area.Min.X, Y: area.Min.Y + height}, Max: area.Max}
	return
}

// SplitHorizontal splits the area horizontally into two parts based on the
// given [Constraint].
//
// It returns the left and right rectangles.
func SplitHorizontal(area Rectangle, constraint Constraint) (left Rectangle, right Rectangle) {
	width := min(constraint.Apply(area.Dx()), area.Dx())
	left = Rectangle{Min: area.Min, Max: Position{X: area.Min.X + width, Y: area.Max.Y}}
	right = Rectangle{Min: Position{X: area.Min.X + width, Y: area.Min.Y}, Max: area.Max}
	return
}
