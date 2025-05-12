package tv

import (
	"image/color"
	"io"
	"log"
	"os"
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
	// The viewport of the program.
	Viewport Viewport
	// The viewport area of the frame.
	Area Rectangle
}

// RenderComponent renders the given component on the frame.
func (f *Frame) RenderComponent(w Component, area Rectangle) error {
	if err := w.Display(f.Buffer, area); err != nil {
		return err
	}
	return nil
}

// SetPosition sets the cursor position on the frame.
func (f *Frame) SetPosition(x, y int) {
	f.Position = &Position{X: x, Y: y}
}

var logger = log.New(io.Discard, "tv", log.LstdFlags|log.Lshortfile)

func init() {
	debug, ok := os.LookupEnv("TV_DEBUG")
	if ok && len(debug) > 0 {
		f, err := os.OpenFile(debug, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			panic("failed to open debug file: " + err.Error())
		}

		logger.SetFlags(log.LstdFlags | log.Lshortfile)
		logger.SetOutput(f)
	}
}
