package component

import (
	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/layout"
)

const tail = "…"

// Block represents a component block that can be drawn on a screen. Any
// attributes set on the block will be inherited by the component it contains.
type Block struct {
	// Style is the style of the block.
	Style uv.Style
	// Padding is the padding around the block.
	Padding layout.Padding
	// Border is the border of the block.
	Border Border
	// Titles contains a list of titles to be drawn on the block border.
	Titles []Title
}

// Draw draws the block on the given screen within the specified area.
func (b Block) Draw(scr uv.Screen, area uv.Rectangle) {
	if b.Border.IsZero() {
		// If no border is set, use empty border.
		b.Border = HiddenBorder()
	}

	b.Border.Draw(scr, area)

	// Draw titles on the block border.
	for _, title := range b.Titles {
		switch title.Direction {
		case layout.Horizontal:
			titleArea := title.ComputeArea(scr.WidthMethod(), area)
		case layout.Vertical:
		}
	}
}
