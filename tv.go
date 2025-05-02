package tv

import "image/color"

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
	// The Position of the frame on the screen. When nil, the cursor is hidden.
	Position *Position
	// The viewport of the program.
	Viewport Viewport
	// The viewport area of the frame.
	Area Rectangle
}

// RenderWidget renders the given widget on the frame.
func (f *Frame) RenderWidget(w Widget, area Rectangle) error {
	if err := w.Display(f.Buffer, area); err != nil {
		return err
	}
	return nil
}
