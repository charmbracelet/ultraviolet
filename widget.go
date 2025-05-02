package tv

// Widget is a base interface for all widgets.
type Widget interface {
	// Display displays the widget on the screen within the given area.
	Display(buf *Buffer, area Rectangle) error
}
