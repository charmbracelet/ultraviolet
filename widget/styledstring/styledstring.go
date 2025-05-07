package styledstring

import (
	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/x/ansi"
)

// StyledString is a styled string widget that can be rendered to a screen.
type StyledString struct{ *tv.StyledString }

// New creates a new [StyledString].
func New(method ansi.Method, str string) StyledString {
	return StyledString{tv.NewStyledString(method, str)}
}

var _ tv.Widget = StyledString{}

// Display implements tv.Widget.
func (s StyledString) Display(buf *tv.Buffer, area tv.Rectangle) error {
	var x, y int
	for y = area.Min.Y; y < area.Max.Y; y++ {
		for x = area.Min.X; x < area.Max.X; {
			cell := s.Buffer.CellAt(x-area.Min.X, y-area.Min.Y)
			if cell == nil {
				// Fill nil cells with blank cells.
				buf.SetCell(x, y, nil)
				x++
				continue
			}
			buf.SetCell(x, y, cell)
			x += cell.Width
		}
	}
	return nil
}
