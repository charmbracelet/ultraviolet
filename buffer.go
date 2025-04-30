package tv

import "image"

// Position represents a position in a coordinate system.
type Position = image.Point

// Pos is a shorthand for creating a new [Position].
func Pos(x, y int) Position {
	return Position{X: x, Y: y}
}

// Rectangle represents a rectangular area.
type Rectangle = image.Rectangle

// Rect is a shorthand for creating a new [Rectangle].
func Rect(x, y, w, h int) Rectangle {
	return Rectangle{Min: image.Point{X: x, Y: y}, Max: image.Point{X: x + w, Y: y + h}}
}

// Cell represents a cell in the terminal.
type Cell struct{}

// Line represents cells in a line.
type Line []*Cell

// Buffer represents a cell buffer that contains the contents of a screen.
type Buffer struct {
	Lines []Line
}

// CellAt returns the cell at the given position. It returns nil if the
// position is out of bounds.
func (b *Buffer) CellAt(x, y int) *Cell {
	return nil
}
