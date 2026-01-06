package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// Element represents a node in the DOM tree that can be rendered.
// This is analogous to DOM's Element interface, which represents objects in the document.
// All DOM nodes that can be rendered implement this interface.
//
// Similar to DOM's Element interface:
//   - Provides core rendering functionality
//   - Can calculate its minimum size requirements
//   - Can be composed into a tree structure
type Element interface {
	// Render draws the element on the given screen within the specified area.
	// Similar to DOM's rendering process, but explicit rather than implicit.
	Render(scr uv.Screen, area uv.Rectangle)

	// MinSize returns the minimum width and height required to render this
	// element. This is analogous to CSS's min-width and min-height properties.
	// Returns (0, 0) if the element can adapt to any size (like flex items).
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
