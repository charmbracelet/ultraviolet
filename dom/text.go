package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/clipperhouse/displaywidth"
)

// text represents a text element that renders a string.
type text struct {
	content  string
	style    uv.Style
	hardWrap bool
}

// Text creates a new text element with the given content.
func Text(content string) Element {
	return &text{content: content, hardWrap: false}
}

// TextHardWrap creates a new text element that hard-wraps at the boundary.
func TextHardWrap(content string) Element {
	return &text{content: content, hardWrap: true}
}

// Styled creates a new text element with the given content and style.
func Styled(content string, style uv.Style) Element {
	return &text{content: content, style: style, hardWrap: false}
}

// StyledHardWrap creates a new text element with style that hard-wraps.
func StyledHardWrap(content string, style uv.Style) Element {
	return &text{content: content, style: style, hardWrap: true}
}

// Render implements the Element interface.
func (t *text) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	if t.hardWrap {
		t.renderHardWrap(scr, area)
	} else {
		t.renderNoWrap(scr, area)
	}
}

// renderNoWrap renders text without hard-wrapping (original behavior).
func (t *text) renderNoWrap(scr uv.Screen, area uv.Rectangle) {
	lines := strings.Split(t.content, "\n")
	y := area.Min.Y

	for _, line := range lines {
		if y >= area.Max.Y {
			break
		}

		x := area.Min.X
		gr := displaywidth.StringGraphemes(line)

		for gr.Next() {
			if x >= area.Max.X {
				break
			}

			grapheme := string(gr.Value())
			width := gr.Width()

			if x+width > area.Max.X {
				break
			}

			cell := &uv.Cell{
				Content: grapheme,
				Width:   width,
				Style:   t.style,
			}
			scr.SetCell(x, y, cell)
			x += width
		}

		y++
	}
}

// renderHardWrap renders text with hard-wrapping at the boundary.
func (t *text) renderHardWrap(scr uv.Screen, area uv.Rectangle) {
	lines := strings.Split(t.content, "\n")
	y := area.Min.Y

	for _, line := range lines {
		if y >= area.Max.Y {
			break
		}

		x := area.Min.X
		gr := displaywidth.StringGraphemes(line)

		for gr.Next() {
			if y >= area.Max.Y {
				return
			}

			grapheme := string(gr.Value())
			width := gr.Width()

			// Hard-wrap to next line if grapheme doesn't fit
			if x+width > area.Max.X {
				y++
				x = area.Min.X
				if y >= area.Max.Y {
					return
				}
			}

			cell := &uv.Cell{
				Content: grapheme,
				Width:   width,
				Style:   t.style,
			}
			scr.SetCell(x, y, cell)
			x += width
		}

		// Move to next line after each input line
		y++
	}
}

// MinSize implements the Element interface.
func (t *text) MinSize(scr uv.Screen) (width, height int) {
	if t.hardWrap {
		// For hard-wrap, we need more calculation, but minimum is 1 char wide
		lines := strings.Split(t.content, "\n")
		height = len(lines) // At least one line per newline
		width = 1            // Minimum width
		for _, line := range lines {
			w := displaywidth.String(line)
			if w > width {
				width = w
			}
		}
	} else {
		lines := strings.Split(t.content, "\n")
		height = len(lines)

		for _, line := range lines {
			w := displaywidth.String(line)
			if w > width {
				width = w
			}
		}
	}

	return width, height
}

// paragraph represents a text element that word-wraps its content.
type paragraph struct {
	content string
	style   uv.Style
}

// Paragraph creates a new paragraph element that word-wraps text.
func Paragraph(content string) Element {
	return &paragraph{content: content}
}

// ParagraphStyled creates a new paragraph element with a style.
func ParagraphStyled(content string, style uv.Style) Element {
	return &paragraph{content: content, style: style}
}

// Render implements the Element interface.
func (p *paragraph) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	words := strings.Fields(p.content)
	if len(words) == 0 {
		return
	}

	x := area.Min.X
	y := area.Min.Y
	maxWidth := area.Dx()

	for _, word := range words {
		wordWidth := displaywidth.String(word)
		spaceWidth := 1 // Space between words

		// Check if we need to wrap
		if x > area.Min.X && x+spaceWidth+wordWidth > area.Min.X+maxWidth {
			y++
			x = area.Min.X
			if y >= area.Max.Y {
				return
			}
		}

		// Add space before word (except at start of line)
		if x > area.Min.X {
			cell := &uv.Cell{
				Content: " ",
				Width:   1,
				Style:   p.style,
			}
			scr.SetCell(x, y, cell)
			x++
		}

		// Render word grapheme by grapheme
		gr := displaywidth.StringGraphemes(word)
		for gr.Next() {
			if x >= area.Min.X+maxWidth {
				break
			}
			if y >= area.Max.Y {
				return
			}

			grapheme := string(gr.Value())
			width := gr.Width()

			cell := &uv.Cell{
				Content: grapheme,
				Width:   width,
				Style:   p.style,
			}
			scr.SetCell(x, y, cell)
			x += width
		}
	}
}

// MinSize implements the Element interface.
func (p *paragraph) MinSize(scr uv.Screen) (width, height int) {
	// For paragraphs, we need at least the width of the longest word
	words := strings.Fields(p.content)
	for _, word := range words {
		w := displaywidth.String(word)
		if w > width {
			width = w
		}
	}
	// Height is flexible based on wrapping, return minimum of 1
	return width, 1
}

// filler represents an element that fills an area with a specific cell.
type filler struct {
	cell *uv.Cell
}

// Filler creates a new filler element that fills the area with the given cell.
func Filler(cell *uv.Cell) Element {
	return &filler{cell: cell}
}

// Render implements the Element interface.
func (f *filler) Render(scr uv.Screen, area uv.Rectangle) {
	screen.FillArea(scr, f.cell, area)
}

// MinSize implements the Element interface.
func (f *filler) MinSize(scr uv.Screen) (width, height int) {
	return 0, 0
}
