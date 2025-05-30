package uv

// WidthMethod determines how many columns a grapheme occupies on the screen.
type WidthMethod interface {
	StringWidth(s string) int
}

// Screen represents a screen that can be drawn to.
type Screen interface {
	// Size returns the size of the screen buffer. Most of the times, this
	// would also be the size of the screen window.
	// It returns -1 for width and height if the screen is not initialized or
	// has no size.
	Size() Size

	// CellAt returns the cell at the given position. If the position is out of
	// bounds, it returns nil. Otherwise, it always returns a cell, even if it
	// is empty (i.e., a cell with a space character and a width of 1).
	CellAt(x, y int) *Cell

	// SetCell sets the cell at the given position. A nil cell is treated as an
	// empty cell with a space character and a width of 1.
	SetCell(x, y int, c *Cell)

	// Method returns the width method used by the screen.
	WidthMethod() WidthMethod
}

// Clear clears the screen with empty cells. This is equivalent to filling the
// screen with empty cells.
func Clear(scr Screen) {
	Fill(scr, nil)
}

// ClearArea clears the given area of the screen with empty cells. This is
// equivalent to filling the area with empty cells.
func ClearArea(scr Screen, area Rectangle) {
	FillArea(scr, nil, area)
}

// Fill fills the screen with the given cell. If the cell is nil, it fills the
// screen with empty cells.
func Fill(scr Screen, cell *Cell) {
	FillArea(scr, cell, scr.Size().Bounds())
}

// FillArea fills the given area of the screen with the given cell. If the cell
// is nil, it fills the area with empty cells.
func FillArea(scr Screen, cell *Cell, area Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			scr.SetCell(x, y, cell)
		}
	}
}

// CloneArea clones the given area of the screen and returns a new buffer
// with the same size as the area. The new buffer will contain the same cells
// as the area in the screen.
// Use [Buffer.Draw] to draw the cloned buffer to a screen again.
func CloneArea(scr Screen, area Rectangle) *Buffer {
	buf := NewBuffer(area.Dx(), area.Dy())
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			cell := scr.CellAt(x, y)
			if cell == nil || cell.IsZero() {
				continue
			}
			buf.SetCell(x-area.Min.X, y-area.Min.Y, cell.Clone())
		}
	}
	return buf
}

// Clone creates a new [Buffer] clone of the given screen. The new buffer will
// have the same size as the screen and will contain the same cells.
// Use [Buffer.Draw] to draw the cloned buffer to a screen again.
func Clone(scr Screen) *Buffer {
	return CloneArea(scr, scr.Size().Bounds())
}

// CursorShape represents a terminal cursor shape.
type CursorShape int

// Cursor shapes.
const (
	CursorBlock CursorShape = iota
	CursorUnderline
	CursorBar
)

// Encode returns the encoded value for the cursor shape.
func (s CursorShape) Encode(blink bool) int {
	// We're using the ANSI escape sequence values for cursor styles.
	// We need to map both [style] and [steady] to the correct value.
	s = (s * 2) + 1 //nolint:mnd
	if !blink {
		s++
	}
	return int(s)
}

// Displayer is an interface that can display a frame.
type Displayer interface {
	// Display displays the given frame.
	Display() error
}
