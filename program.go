package tv

import (
	"fmt"
	"io"
)

// Programable is a generic type that represents a program that can be
// displayed on a screen.
type Programable interface {
	Screen
	Displayer
}

// Program is a generic type that represents a program with a [Screen].
type Program[T Programable] struct {
	scr      T
	buf      *Buffer  // The new buffer to be drawn to the screen.
	size     Size     // The last known size of the terminal.
	viewport Viewport // The viewport area to operate on.
	started  bool
}

// NewProgram creates a new [Programable] with the given screen.
func NewProgram[T Programable](scr T) *Program[T] {
	return &Program[T]{scr: scr}
}

// SetViewport sets the viewport area for the program. It defaults to
// [FullViewport] which covers the entire program screen.
func (p *Program[T]) SetViewport(v Viewport) {
	if v == nil {
		p.viewport = FullViewport{}
	} else {
		p.viewport = v
	}
}

// Start starts the program and initializes the screen. Call [Program.Close] to
// close the program and release any resources.
func (p *Program[T]) Start() error {
	w, h, err := p.scr.GetSize()
	if err != nil {
		return fmt.Errorf("error getting screen size: %w", err)
	}

	p.size = Size{Width: w, Height: h}
	p.buf = NewBuffer(p.size.Width, p.size.Height)
	p.viewport = FullViewport{}
	p.started = true
	return nil
}

// Close closes the program and releases any resources.
func (p *Program[T]) Close() error {
	if closer, ok := any(p.scr).(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (p *Program[T]) Display(fn func(f *Frame)) error {
	if !p.started {
		return fmt.Errorf("program not started")
	}

	f := &Frame{
		Buffer:   p.buf,
		Viewport: p.viewport,
		Area:     p.viewport.ComputeArea(p.size),
	}
	fn(f)

	return p.scr.Display(f)
}
