package tv

// Widget is a base interface for all widgets.
type Widget interface {
	// Display displays the widget on the given screen and area.
	Display(scr Screen, area Rectangle) error
}
