package component

import "github.com/charmbracelet/uv"

// Margin adds padding to the given area.
type Margin struct {
	Style uv.Style

	Top    int
	Right  int
	Bottom int
	Left   int
}

// IsZero checks if the padding is zero in all directions.
func (m Margin) IsZero() bool {
	return m.Top == 0 && m.Right == 0 && m.Bottom == 0 && m.Left == 0
}

// InnerArea computes the inner area of the margin, excluding the margin
// itself.
func (m Margin) InnerArea(_ uv.WidthMethod, area uv.Rectangle) uv.Rectangle {
	if m.IsZero() {
		return area
	}
	marginArea := area
	marginArea.Min.X += m.Left
	marginArea.Min.Y += m.Top
	marginArea.Max.X -= m.Right
	marginArea.Max.Y -= m.Bottom
	return area.Intersect(marginArea)
}

// Draw draws the padding on the given screen within the specified area.
func (m Margin) Draw(scr uv.Screen, area uv.Rectangle) {
	if m.IsZero() {
		return
	}

	cell := uv.EmptyCell
	cell.Style = m.Style

	for _, r := range []uv.Rectangle{
		{Min: area.Min, Max: uv.Pos(area.Max.X, area.Min.Y+m.Top)},    // Top padding
		{Min: uv.Pos(area.Min.X, area.Max.Y-m.Bottom), Max: area.Max}, // Bottom padding
		{Min: area.Min, Max: uv.Pos(area.Min.X+m.Left, area.Max.Y)},   // Left padding
		{Min: uv.Pos(area.Max.X-m.Right, area.Min.Y), Max: area.Max},  // Right padding
	} {
		uv.FillArea(scr, &cell, r)
	}
}
