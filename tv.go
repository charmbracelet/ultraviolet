package tv

import (
	"image/color"
)

// Screen represents a screen that can be drawn to.
type Screen interface {
	// GetSize returns the size of the screen. It errors if the size cannot be
	// determined.
	GetSize() (width, height int, err error)

	// CellAt returns the cell at the given position.
	CellAt(x, y int) *Cell

	// ColorModel returns the color model of the screen.
	ColorModel() color.Model
}

// Displayer is an interface that can display a frame.
type Displayer interface {
	// Display displays the given frame.
	Display(f *Frame) error
}

// Frame is a single frame to be displayed on the screen.
type Frame struct {
	// The screen buffer to be displayed.
	Buffer *Buffer
	// The cursor position on the frame. When nil, the cursor is hidden.
	Position *Position
	// The viewport of the program. Use [Frame.ComputeArea] to compute the area
	// of the frame based on the size of the screen.
	Viewport Viewport
}

// ComputeArea computes the area of the frame based on the given size of the screen.
func (f *Frame) ComputeArea(width, height int) Rectangle {
	vp := f.Viewport
	if vp == nil {
		vp = FullViewport{}
	}
	return vp.ComputeArea(width, height)
}

// Resize resizes the frame to the given width and height based on the viewport.
func (f *Frame) Resize(width, height int) {
	switch v := f.Viewport.(type) {
	case nil, FullViewport, FixedViewport:
		f.Buffer.Resize(width, height)
	case InlineViewport:
		area := v.ComputeArea(width, height)
		f.Buffer.Resize(area.Dx(), area.Dy())
	}
}

// RenderComponent renders the given component on the frame.
func (f *Frame) RenderComponent(w Component, area Rectangle) error {
	if fvp, ok := f.Viewport.(FixedViewport); ok {
		// If the viewport is fixed, we need to ensure the area is within bounds.
		area = area.Intersect(Rectangle(fvp))
	}
	if err := w.RenderComponent(f.Buffer, area); err != nil {
		return err
	}
	return nil
}

// SetPosition sets the cursor position on the frame.
func (f *Frame) SetPosition(x, y int) {
	f.Position = &Position{X: x, Y: y}
}

// SetCell sets the cell at the given position on the frame.
func (f *Frame) SetCell(x, y int, c *Cell) {
	f.Buffer.SetCell(x, y, c)
}

// CellAt returns an existing cell at the given position on the frame.
func (f *Frame) CellAt(x, y int) *Cell {
	return f.Buffer.CellAt(x, y)
}
