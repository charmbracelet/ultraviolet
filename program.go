package tv

import (
	"context"
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
	buf      Buffer   // The new buffer to be drawn to the screen.
	size     Size     // The last known size of the terminal.
	viewport Viewport // The viewport area to operate on.
	started  bool
}

// NewProgram creates a new [Programable] with the given screen.
func NewProgram[T Programable](scr T) *Program[T] {
	return &Program[T]{
		scr:      scr,
		viewport: FullViewport{},
	}
}

// SetViewport sets the viewport area for the program. It defaults to
// [FullViewport] which covers the entire program screen.
func (p *Program[T]) SetViewport(v Viewport) {
	if v == nil {
		v = FullViewport{}
	}
	p.viewport = v
}

// Resize resizes the program to the given width and height. It returns an
// error if the resize fails.
func (p *Program[T]) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid size: %dx%d", width, height)
	}

	bufw, bufh := width, height
	p.size = Size{Width: width, Height: height}
	switch vp := p.viewport.(type) {
	case FullViewport:
	case InlineViewport:
		bufh = int(vp)
	}

	p.buf.Resize(bufw, bufh)
	if re, ok := any(p.scr).(Resizer); ok {
		if err := re.Resize(bufw, bufh); err != nil {
			return fmt.Errorf("error resizing screen: %w", err)
		}
	}

	return nil
}

// AutoResize queries the screen for its size and resizes the program to fit the
// screen. It returns an error if the resize fails.
func (p *Program[T]) AutoResize() error {
	w, h, err := p.scr.GetSize()
	if err != nil {
		return fmt.Errorf("error getting screen size: %w", err)
	}
	return p.Resize(w, h)
}

// Start starts the program and initializes the screen. If the screen
// implements the [Starter] interface, it will call the [Start] method on the
// screen.
func (p *Program[T]) Start() error {
	if p.started {
		return fmt.Errorf("program already started")
	}
	if err := p.AutoResize(); err != nil {
		return err
	}
	p.started = true
	if s, ok := any(p.scr).(Starter); ok {
		return s.Start()
	}

	return nil
}

// Shutdown shuts down the program gracefully. It returns an error if the
// shutdown fails. It calls the [Shutdowner] interface if the screen implements
// it.
func (p *Program[T]) Shutdown(ctx context.Context) error {
	if !p.started {
		return fmt.Errorf("program not started")
	}
	if s, ok := any(p.scr).(Shutdowner); ok {
		return s.Shutdown(ctx)
	}
	return nil
}

// Close closes the program and releases any resources. If the screen
// implements the [io.Closer] interface, it will call the [Close] method on the
// screen.
func (p *Program[T]) Close() error {
	if !p.started {
		return fmt.Errorf("program not started")
	}
	if closer, ok := any(p.scr).(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Renderer is a function that renders a frame on the screen. It takes a
// [Frame] and returns an error if the rendering fails.
type Renderer = func(f *Frame) error

// Display displays the program on the screen via the given function. It errors
// if the program is not started or if the function fails.
func (p *Program[T]) Display(fn Renderer) error {
	if !p.started {
		return fmt.Errorf("program not started")
	}

	f := &Frame{
		Buffer:   &p.buf,
		Viewport: p.viewport,
		Area:     p.viewport.ComputeArea(p.size),
	}
	if err := fn(f); err != nil {
		return err
	}

	if err := p.scr.Display(f); err != nil {
		return fmt.Errorf("error displaying frame: %w", err)
	}

	if f, ok := any(p.scr).(Flusher); ok {
		if err := f.Flush(); err != nil {
			return fmt.Errorf("error flushing screen: %w", err)
		}
	}

	return nil
}

// Starter represents types that can be started.
type Starter interface {
	Start() error
}

// Shutdowner represents types that can be shut down gracefully.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// Resizer represents types that can be resized.
type Resizer interface {
	Resize(width, height int) error
}

// Flusher represents types that can be flushed.
type Flusher interface {
	Flush() error
}
