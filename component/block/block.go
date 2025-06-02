package block

import (
	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/layout"
)

// Block represents a component block that can be drawn on a screen. Any
// attributes set on the block will be inherited by the component it contains.
type Block struct {
	// Style is the style of the block.
	uv.Style
	// Link is the link of the block.
	uv.Link
	// Padding is the padding around the block.
	Padding layout.Padding
	// Component is the component to be drawn inside the block.
	Component uv.Drawable
}

// Draw draws the block on the given screen within the specified area.
func (b Block) Draw(scr uv.Screen, area uv.Rectangle) {
	if b.Component == nil {
		return
	}

	pad := uv.EmptyCell
	pad.Style = b.Style
	pad.Link = b.Link
	uv.FillArea(scr, &pad, area)

	// Apply padding to the area.
	area = b.Padding.Apply(area)

	// Draw the component within the padded area.
	b.Component.Draw(scr, area)

	// Apply missing style and link to the component.
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			c := scr.CellAt(x, y).Clone()
			if c != nil {
				if c.Style.IsZero() {
					c.Style = b.Style
				}
				if c.Link.IsZero() {
					c.Link = b.Link
				}
				scr.SetCell(x, y, c)
				x += c.Width - 1
			}
		}
	}
}
