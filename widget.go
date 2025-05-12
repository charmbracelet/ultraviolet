package tv

// Component represents a displayable component on a [Buffer].
type Component interface {
	// Display displays the component on the screen within the given area.
	Display(buf *Buffer, area Rectangle) error
}
