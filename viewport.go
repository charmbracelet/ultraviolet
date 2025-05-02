package tv

// Viewport represents a rectangular area of the terminal screen to operate on.
// It is used to define the area of the screen that should be drawn to.
type Viewport interface {
	// ComputeArea calculates the area of the viewport based on the given
	// window size.
	ComputeArea(size Size) Rectangle
}

// FullViewport represents a viewport that covers the entire terminal screen.
type FullViewport struct{}

// ComputeArea calculates the area of the full viewport based on the given
// window size.
func (v FullViewport) ComputeArea(size Size) Rectangle {
	return Rect(0, 0, size.Width, size.Height)
}

// InlineViewport represents a viewport that is inline with the terminal screen.
type InlineViewport int

// ComputeArea calculates the area of the inline viewport based on the given
// window size.
func (v InlineViewport) ComputeArea(size Size) Rectangle {
	line := max(0, size.Height-int(v))
	return Rect(0, line, size.Width, size.Height-line)
}

// FixedViewport represents a fixed viewport with a specific size and position.
type FixedViewport Rectangle

// ComputeArea calculates the area of the fixed viewport based on the given
// window size.
func (v FixedViewport) ComputeArea(size Size) Rectangle {
	area := Rect(0, 0, size.Width, size.Height)
	fixed := Rectangle(v)
	if !fixed.In(area) {
		return area
	}
	return fixed
}
