package tv

// Viewport represents a rectangular area of the terminal screen to operate on.
// It is used to define the area of the screen that should be drawn to.
type Viewport interface {
	// ComputeArea calculates the area of the viewport based on the given
	// window size.
	ComputeArea(width, height int) Rectangle
}

// FullViewport represents a viewport that covers the entire terminal screen.
type FullViewport struct{}

// ComputeArea calculates the area of the full viewport based on the given
// window size.
func (v FullViewport) ComputeArea(width, height int) Rectangle {
	return Rect(0, 0, width, height)
}

// InlineViewport represents a viewport that is inline with the terminal screen.
type InlineViewport int

// ComputeArea calculates the area of the inline viewport based on the given
// window size.
func (v InlineViewport) ComputeArea(width, _ int) Rectangle {
	return Rect(0, 0, width, max(0, int(v)))
}

// FixedViewport represents a viewport with fixed dimensions.
type FixedViewport Rectangle

// ComputeArea returns the fixed rectangle as the area of the viewport.
func (v FixedViewport) ComputeArea(width, height int) Rectangle {
	return Rect(0, 0, width, height).Bounds().Intersect(Rectangle(v))
}
