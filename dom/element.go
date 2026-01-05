package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Element represents a renderable UI component in the DOM tree.
// Elements can be composed together to build complex interfaces.
type Element interface {
	// Render draws the element on the given screen within the specified area.
	Render(scr uv.Screen, area uv.Rectangle)

	// MinSize returns the minimum width and height required to render this
	// element. This is used for layout calculations. Returns (0, 0) if the
	// element can adapt to any size.
	MinSize(scr uv.Screen) (width, height int)
}

// ElementFunc is a function adapter that implements the Element interface.
type ElementFunc func(scr uv.Screen, area uv.Rectangle)

// Render implements the Element interface.
func (f ElementFunc) Render(scr uv.Screen, area uv.Rectangle) {
	f(scr, area)
}

// MinSize implements the Element interface.
func (f ElementFunc) MinSize(scr uv.Screen) (width, height int) {
	return 0, 0
}

// Empty returns an element that renders nothing.
func Empty() Element {
	return ElementFunc(func(scr uv.Screen, area uv.Rectangle) {})
}
