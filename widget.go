package tv

// Widget represents a displayable widget on a [Buffer].
type Widget interface {
	// Display displays the widget on the screen within the given area.
	Display(buf *Buffer, area Rectangle) error
}
