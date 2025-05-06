package tv

import (
	"image"
	"io"
	"strings"

	"github.com/charmbracelet/x/ansi"
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

// Render renders the line to a string with all the required attributes and
// styles.
func (l Line) Render() string {
	var buf strings.Builder
	return renderLine(&buf, l)
}

func renderLine(buf interface {
	io.Writer
	io.StringWriter
	String() string
}, l Line,
) string {
	var pen Style
	var link Link
	var pendingLine string
	var pendingWidth int // this ignores space cells until we hit a non-space cell

	writePending := func() {
		// If there's no pending line, we don't need to do anything.
		if len(pendingLine) == 0 {
			return
		}
		buf.WriteString(pendingLine)
		pendingWidth = 0
		pendingLine = ""
	}

	for x := 0; x < len(l); x++ {
		if cell := l.At(x); cell != nil && cell.Width > 0 {
			// Convert the cell's style and link to the given color profile.
			cellStyle := cell.Style
			cellLink := cell.Link
			if cellStyle.Empty() && !pen.Empty() {
				writePending()
				buf.WriteString(ansi.ResetStyle) //nolint:errcheck
				pen.Reset()
			}
			if !cellStyle.Equal(&pen) {
				writePending()
				seq := cellStyle.DiffSequence(pen)
				buf.WriteString(seq) // nolint:errcheck
				pen = cellStyle
			}

			// Write the URL escape sequence
			if cellLink != link && link.URL != "" {
				writePending()
				buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
				link.Reset()
			}
			if cellLink != link {
				writePending()
				buf.WriteString(ansi.SetHyperlink(cellLink.URL, cellLink.Params)) //nolint:errcheck
				link = cellLink
			}

			// We only write the cell content if it's not empty. If it is, we
			// append it to the pending line and width to be evaluated later.
			if cell.Equal(&BlankCell) {
				pendingLine += cell.String()
				pendingWidth += cell.Width
			} else {
				writePending()
				buf.WriteString(cell.String())
			}
		}
	}
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.Empty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}
	return strings.TrimRight(buf.String(), " ") // Trim trailing spaces
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
func (b *Buffer) String() string {
	var buf strings.Builder
	for i, l := range b.Lines {
		buf.WriteString(l.String())
		if i < len(b.Lines)-1 {
			buf.WriteString("\r\n") //nolint:errcheck
		}
	}
	return buf.String()
}

// Render renders the buffer to a string with all the required attributes and
// styles.
func (b *Buffer) Render() string {
	var buf strings.Builder
	for i, l := range b.Lines {
		buf.WriteString(renderLine(&buf, l))
		if i < len(b.Lines)-1 {
			buf.WriteString("\r\n") //nolint:errcheck
		}
	}
	return buf.String()
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
	if y < 0 || y >= len(b.Lines) {
		return nil
	}
	return b.Lines[y].At(x)
}

// SetCell sets the cell at the given x, y position.
func (b *Buffer) SetCell(x, y int, c *Cell) {
	if y < 0 || y >= len(b.Lines) {
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
	w := len(b.Lines[0])
	for _, l := range b.Lines {
		if len(l) > w {
			w = len(l)
		}
	}
	return w
}

// Bounds returns the bounds of the buffer.
func (b *Buffer) Bounds() Rectangle {
	return Rect(0, 0, b.Width(), b.Height())
}

// Resize resizes the buffer to the given width and height.
func (b *Buffer) Resize(width int, height int) {
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
