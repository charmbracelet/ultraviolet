package layout

import "github.com/charmbracelet/uv"

// Padding adds padding to the given area.
type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// IsZero checks if the padding is zero in all directions.
func (p Padding) IsZero() bool {
	return p.Top == 0 && p.Right == 0 && p.Bottom == 0 && p.Left == 0
}

// Apply applies the padding to the given area and returns a new area.
func (p Padding) Apply(area uv.Rectangle) uv.Rectangle {
	if p.IsZero() {
		return area
	}
	area.Min.X += p.Left
	area.Min.Y += p.Top
	area.Max.X -= p.Right
	area.Max.Y -= p.Bottom
	return area
}
