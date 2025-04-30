package tv

// Widget is a base interface for all widgets.
// It provides methods for drawing the widget.
type Widget interface {
	// Draw draws the widget to the screen.
	Draw(drw Drawer, area Rectangle) error
}
