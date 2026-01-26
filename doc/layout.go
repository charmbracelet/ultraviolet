package doc

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// LayoutBox represents the calculated layout for a node.
type LayoutBox struct {
	// Rect is the absolute position and size on screen (in cells)
	// This includes the entire box: margin + border + padding + content
	Rect uv.Rectangle

	// ContentRect is the area available for content (after margin, border, padding)
	ContentRect uv.Rectangle
	
	// PaddingRect is the area including padding (after margin, border)
	PaddingRect uv.Rectangle
	
	// BorderRect is the area including border (after margin)
	BorderRect uv.Rectangle

	// ScrollOffset is the scroll position for this container
	ScrollOffset uv.Position

	// Dirty indicates this layout needs recalculation
	Dirty bool
}

// NewLayoutBox creates a new LayoutBox with the given rectangle.
func NewLayoutBox(rect uv.Rectangle) *LayoutBox {
	return &LayoutBox{
		Rect:         rect,
		ContentRect:  rect,
		PaddingRect:  rect,
		BorderRect:   rect,
		ScrollOffset: uv.Pos(0, 0),
		Dirty:        true,
	}
}

// Invalidate marks this layout as dirty (needs recalculation).
func (lb *LayoutBox) Invalidate() {
	lb.Dirty = true
}

// IsVisible returns true if this layout box intersects with the given viewport.
func (lb *LayoutBox) IsVisible(viewport uv.Rectangle) bool {
	return lb.Rect.Overlaps(viewport)
}
