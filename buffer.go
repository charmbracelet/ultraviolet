package tv

import (
	"image"
	"strings"
)

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

// Line represents cells in a line.
type Line []*Cell

// Set sets the cell at the given x position.
func (l Line) Set(x int, c *Cell) {
	// maxCellWidth is the maximum width a terminal cell is expected to have.
	const maxCellWidth = 5

	width := len(l)
	if x < 0 || x >= width {
		return
	}

	if c != nil {
		logger.Printf("Set cell at %d: %v", x, c)
	}

	// When a wide cell is partially overwritten, we need
	// to fill the rest of the cell with space cells to
	// avoid rendering issues.
	prev := l.At(x)
	if prev != nil && prev.Width > 1 {
		// Writing to the first wide cell
		for j := 0; j < prev.Width && x+j < width; j++ {
			l[x+j] = prev.Clone().Blank()
		}
	} else if prev != nil && prev.Width == 0 {
		// Writing to wide cell placeholders
		for j := 1; j < maxCellWidth && x-j >= 0; j++ {
			wide := l.At(x - j)
			if wide != nil && wide.Width > 1 && j < wide.Width {
				for k := 0; k < wide.Width; k++ {
					l[x-j+k] = wide.Clone().Blank()
				}
				break
			}
		}
	}

	if c != nil && x+c.Width > width {
		// If the cell is too wide, we write blanks with the same style.
		for i := 0; i < c.Width && x+i < width; i++ {
			l[x+i] = c.Clone().Blank()
		}
	} else {
		l[x] = c

		// Mark wide cells with an empty cell zero width
		// We set the wide cell down below
		if c != nil && c.Width > 1 {
			for j := 1; j < c.Width && x+j < width; j++ {
				var wide Cell
				l[x+j] = &wide
			}
		}
	}
}

// At returns the cell at the given x position.
// If the cell does not exist, it returns nil.
func (l Line) At(x int) *Cell {
	if x < 0 || x >= len(l) {
		return nil
	}

	c := l[x]
	if c == nil {
		newCell := BlankCell
		return &newCell
	}

	return c
}

// String returns the string representation of the line. Any trailing spaces
// are removed.
func (l Line) String() (s string) {
	for _, c := range l {
		if c == nil {
			s += " "
		} else if c.Empty() {
			continue
		} else {
			s += c.String()
		}
	}
	s = strings.TrimRight(s, " ")
	return
}

// Buffer represents a cell buffer that contains the contents of a screen.
type Buffer struct {
	Lines []Line
}

// NewBuffer creates a new buffer with the given width and height.
// This is a convenience function that initializes a new buffer and resizes it.
func NewBuffer(width int, height int) *Buffer {
	b := new(Buffer)
	b.Resize(width, height)
	return b
}

// String returns the string representation of the buffer.
func (b *Buffer) String() (s string) {
	for i, l := range b.Lines {
		s += l.String()
		if i < len(b.Lines)-1 {
			s += "\r\n"
		}
	}
	return
}

// Line returns a pointer to the line at the given y position.
// If the line does not exist, it returns nil.
func (b *Buffer) Line(y int) Line {
	if y < 0 || y >= len(b.Lines) {
		return nil
	}
	return b.Lines[y]
}

// CellAt returns the cell at the given position. It returns nil if the
// position is out of bounds.
func (b *Buffer) CellAt(x int, y int) *Cell {
	if y < 0 || y >= len(b.Lines) || x < 0 || x >= len(b.Lines[y]) {
		return nil
	}
	return b.Lines[y][x]
}

// SetCell sets the cell at the given x, y position.
func (b *Buffer) SetCell(x, y int, c *Cell) {
	if y < 0 || y >= len(b.Lines) {
		logger.Printf("SetCell: y out of bounds: %d", y)
		return
	}

	b.Lines[y].Set(x, c)
}

// Height implements Screen.
func (b *Buffer) Height() int {
	return len(b.Lines)
}

// Width implements Screen.
func (b *Buffer) Width() int {
	if len(b.Lines) == 0 {
		return 0
	}
	return len(b.Lines[0])
}

// Bounds returns the bounds of the buffer.
func (b *Buffer) Bounds() Rectangle {
	return Rect(0, 0, b.Width(), b.Height())
}

// Resize resizes the buffer to the given width and height.
func (b *Buffer) Resize(width int, height int) {
	logger.Printf("Resize: %d %d -> %d %d %q", b.Width(), b.Height(), width, height, b)
	if width == 0 || height == 0 {
		b.Lines = nil
		return
	}

	if width > b.Width() {
		line := make(Line, width-b.Width())
		for i := range b.Lines {
			b.Lines[i] = append(b.Lines[i], line...)
		}
	} else if width < b.Width() {
		for i := range b.Lines {
			b.Lines[i] = b.Lines[i][:width]
		}
	}

	if height > len(b.Lines) {
		for i := len(b.Lines); i < height; i++ {
			b.Lines = append(b.Lines, make(Line, width))
		}
	} else if height < len(b.Lines) {
		b.Lines = b.Lines[:height]
	}
}

// Fill fills the buffer with the given cell and rectangle.
func (b *Buffer) Fill(c *Cell) {
	b.FillArea(c, b.Bounds())
}

// FillArea fills the buffer with the given cell and rectangle.
func (b *Buffer) FillArea(c *Cell, rect Rectangle) {
	cellWidth := 1
	if c != nil && c.Width > 1 {
		cellWidth = c.Width
	}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x += cellWidth {
			b.SetCell(x, y, c)
		}
	}
}

// Clear clears the buffer with space cells and rectangle.
func (b *Buffer) Clear() {
	b.ClearArea(b.Bounds())
}

// ClearArea clears the buffer with space cells within the specified
// rectangles. Only cells within the rectangle's bounds are affected.
func (b *Buffer) ClearArea(rect Rectangle) {
	b.FillArea(nil, rect)
}
