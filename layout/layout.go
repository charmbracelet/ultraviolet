package layout

import (
	"image"

	uv "github.com/charmbracelet/ultraviolet"
)

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
func SplitVertical(area uv.Rectangle, constraint Constraint) (top uv.Rectangle, bottom uv.Rectangle) {
	height := min(constraint.Apply(area.Dy()), area.Dy())
	top = uv.Rectangle{Min: area.Min, Max: uv.Position{X: area.Max.X, Y: area.Min.Y + height}}
	bottom = uv.Rectangle{Min: uv.Position{X: area.Min.X, Y: area.Min.Y + height}, Max: area.Max}
	return
}

// SplitHorizontal splits the area horizontally into two parts based on the
// given [Constraint].
//
// It returns the left and right rectangles.
func SplitHorizontal(area uv.Rectangle, constraint Constraint) (left uv.Rectangle, right uv.Rectangle) {
	width := min(constraint.Apply(area.Dx()), area.Dx())
	left = uv.Rectangle{Min: area.Min, Max: uv.Position{X: area.Min.X + width, Y: area.Max.Y}}
	right = uv.Rectangle{Min: uv.Position{X: area.Min.X + width, Y: area.Min.Y}, Max: area.Max}
	return
}

// CenterRect returns a new [uv.Rectangle] centered within the given area with the
// specified width and height.
func CenterRect(area uv.Rectangle, width, height int) uv.Rectangle {
	centerX := area.Min.X + area.Dx()/2
	centerY := area.Min.Y + area.Dy()/2
	minX := centerX - width/2
	minY := centerY - height/2
	maxX := minX + width
	maxY := minY + height
	return image.Rect(minX, minY, maxX, maxY)
}

// TopLeftRect returns a new [uv.Rectangle] positioned at the top-left corner of the
// given area with the specified width and height.
func TopLeftRect(area uv.Rectangle, width, height int) uv.Rectangle {
	return image.Rect(area.Min.X, area.Min.Y, area.Min.X+width, area.Min.Y+height).Intersect(area)
}

// TopCenterRect returns a new [uv.Rectangle] positioned at the top-center of the
// given area with the specified width and height.
func TopCenterRect(area uv.Rectangle, width, height int) uv.Rectangle {
	centerX := area.Min.X + area.Dx()/2
	minX := centerX - width/2
	return image.Rect(minX, area.Min.Y, minX+width, area.Min.Y+height).Intersect(area)
}

// TopRightRect returns a new [uv.Rectangle] positioned at the top-right corner of
// the given area with the specified width and height.
func TopRightRect(area uv.Rectangle, width, height int) uv.Rectangle {
	return image.Rect(area.Max.X-width, area.Min.Y, area.Max.X, area.Min.Y+height).Intersect(area)
}

// RightCenterRect returns a new [uv.Rectangle] positioned at the right-center of
// the given area with the specified width and height.
func RightCenterRect(area uv.Rectangle, width, height int) uv.Rectangle {
	centerY := area.Min.Y + area.Dy()/2
	minY := centerY - height/2
	return image.Rect(area.Max.X-width, minY, area.Max.X, minY+height).Intersect(area)
}

// LeftCenterRect returns a new [uv.Rectangle] positioned at the left-center of the
// given area with the specified width and height.
func LeftCenterRect(area uv.Rectangle, width, height int) uv.Rectangle {
	centerY := area.Min.Y + area.Dy()/2
	minY := centerY - height/2
	return image.Rect(area.Min.X, minY, area.Min.X+width, minY+height).Intersect(area)
}

// BottomLeftRect returns a new [uv.Rectangle] positioned at the bottom-left corner
// of the given area with the specified width and height.
func BottomLeftRect(area uv.Rectangle, width, height int) uv.Rectangle {
	return image.Rect(area.Min.X, area.Max.Y-height, area.Min.X+width, area.Max.Y).Intersect(area)
}

// BottomCenterRect returns a new [uv.Rectangle] positioned at the bottom-center of
// the given area with the specified width and height.
func BottomCenterRect(area uv.Rectangle, width, height int) uv.Rectangle {
	centerX := area.Min.X + area.Dx()/2
	minX := centerX - width/2
	return image.Rect(minX, area.Max.Y-height, minX+width, area.Max.Y).Intersect(area)
}

// BottomRightRect returns a new [uv.Rectangle] positioned at the bottom-right
// corner of the given area with the specified width and height.
func BottomRightRect(area uv.Rectangle, width, height int) uv.Rectangle {
	return image.Rect(area.Max.X-width, area.Max.Y-height, area.Max.X, area.Max.Y).Intersect(area)
}
