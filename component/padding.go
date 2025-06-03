package component

import "github.com/charmbracelet/uv"

const nbsp = "\u00A0" // Non-breaking space character

// Padding adds padding to the given area.
type Padding struct {
	Style uv.Style

	Top    int
	Right  int
	Bottom int
	Left   int
}

// IsZero checks if the padding is zero in all directions.
func (p Padding) IsZero() bool {
	return p.Top == 0 && p.Right == 0 && p.Bottom == 0 && p.Left == 0
}

// InnerArea computes the inner area of the padding, excluding the padding
// itself.
func (p Padding) InnerArea(_ uv.WidthMethod, area uv.Rectangle) uv.Rectangle {
	if p.IsZero() {
		return area
	}
	paddingArea := area
	paddingArea.Min.X += p.Left
	paddingArea.Min.Y += p.Top
	paddingArea.Max.X -= p.Right
	paddingArea.Max.Y -= p.Bottom
	return area.Intersect(paddingArea)
}

// Draw draws the padding on the given screen within the specified area.
func (p Padding) Draw(scr uv.Screen, area uv.Rectangle) {
	if p.IsZero() {
		return
	}

	cell := uv.EmptyCell
	cell.Content = nbsp
	cell.Style = p.Style

	for _, r := range []uv.Rectangle{
		{Min: area.Min, Max: uv.Pos(area.Max.X, area.Min.Y+p.Top)},    // Top padding
		{Min: uv.Pos(area.Min.X, area.Max.Y-p.Bottom), Max: area.Max}, // Bottom padding
		{Min: area.Min, Max: uv.Pos(area.Min.X+p.Left, area.Max.Y)},   // Left padding
		{Min: uv.Pos(area.Max.X-p.Right, area.Min.Y), Max: area.Max},  // Right padding
	} {
		uv.FillArea(scr, &cell, r)
	}
}
