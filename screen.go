package tv

// Screen represents a screen that can be drawn to.
type Screen interface {
	// Bounds returns the bounds of the screen.
	Bounds() Rectangle

	// SetCell sets the cell at the given position. It returns whether the cell
	// was set successfully.
	SetCell(x, y int) bool

	// CellAt returns the cell at the given position.
	CellAt(x, y int) *Cell

	// Clear clears the entire screen.
	Clear()

	// SetPosition sets the position of the cursor after drawing.
	SetPosition(x, y int)

	// Position returns the last known position of the cursor.
	Position() (x, y int)
}
