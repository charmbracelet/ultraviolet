package tv

// Drawer is an interface for drawing to a screen.
type Drawer interface {
	// SetCell sets the cell at the given position. It returns whether the cell
	// was set successfully.
	SetCell(x, y int) bool

	// Clear clears the entire screen.
	Clear()

	// SetPosition sets the position of the cursor after drawing.
	SetPosition(x, y int)
}

// Screen represents a screen that can be drawn to.
type Screen interface {
	Drawer

	// Bounds returns the bounds of the screen.
	Bounds() Rectangle

	// CellAt returns the cell at the given position.
	CellAt(x, y int) *Cell

	// Position returns the last known position of the cursor.
	Position() (x, y int)

	// Flush flushes pending changes to the screen.
	Flush() error
}
