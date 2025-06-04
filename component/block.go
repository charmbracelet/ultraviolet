package component

import (
	"github.com/charmbracelet/uv"
)

// Block is a component that represents a rectangular area with a base style,
// margins, paddings, and borders. It can be used to create structured layouts
// and visually separate content on the screen.
type Block struct {
	// Style is the style of the block.
	Style uv.Style
	// Margin is the margin around the block outside the border.
	Margin Margin
	// Padding is the padding around the block inside the border.
	Padding Padding
	// Border is the border around the block.
	Border Border
}

// InnerArea computes the inner area of the block, excluding padding and border.
func (b Block) InnerArea(wm uv.WidthMethod, area uv.Rectangle) uv.Rectangle {
	blockArea := area
	blockArea = b.Margin.InnerArea(wm, blockArea)
	blockArea = b.Border.InnerArea(wm, blockArea)
	blockArea = b.Padding.InnerArea(wm, blockArea)
	return area.Intersect(blockArea)
}

// Draw draws the block on the given screen within the specified area.
func (b Block) Draw(scr uv.Screen, area uv.Rectangle) {
	borderArea := area
	if !b.Margin.IsZero() {
		b.Margin.Draw(scr, area)
		borderArea = b.Margin.InnerArea(scr.WidthMethod(), borderArea)
	}

	if !b.Border.IsZero() {
		b.Border.Draw(scr, borderArea)
		borderArea = b.Border.InnerArea(scr.WidthMethod(), borderArea)
	}

	if !b.Padding.IsZero() {
		b.Padding.Draw(scr, borderArea)
		borderArea = b.Padding.InnerArea(scr.WidthMethod(), borderArea)
	}

	// Set the style for the block content.
	for x := borderArea.Min.X; x < borderArea.Max.X; x++ {
		for y := borderArea.Min.Y; y < borderArea.Max.Y; y++ {
			cell := scr.CellAt(x, y)
			if cell != nil {
				cell = cell.Clone()
				cell.Style = cell.Style.Merge(b.Style)
				scr.SetCell(x, y, cell)
			}
		}
	}
	//
	// // Draw titles on the block border.
	// for _, title := range b.Titles {
	// 	var titleArea uv.Rectangle
	// 	title.Style = title.Style.Merge(b.Style)
	// 	switch title.Direction {
	// 	case layout.Horizontal:
	// 		titleArea = uv.Rectangle{
	// 			Min: uv.Pos(area.Min.X, area.Min.Y),
	// 			Max: uv.Pos(area.Max.X, area.Min.Y+1),
	// 		}
	// 	case layout.Vertical:
	// 		titleArea = uv.Rectangle{
	// 			Min: uv.Pos(area.Min.X, area.Min.Y),
	// 			Max: uv.Pos(area.Min.X+1, area.Max.Y),
	// 		}
	// 	}
	// 	titleArea = title.ComputeArea(wm, titleArea)
	// 	title.Draw(scr, titleArea)
	// }
}
