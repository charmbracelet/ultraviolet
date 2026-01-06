package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/clipperhouse/displaywidth"
)

// TextNode is an inline display element that renders text content.
// Like DOM Text nodes and HTML's <span>, it flows inline and only breaks on explicit newlines.
// For block-level text containers, use Box with TextNode as content.
type TextNode struct {
	Content string
	Style   uv.Style
}

// Text creates a new inline text element with the given content.
// This follows DOM's document.createTextNode() pattern.
func Text(content string) *TextNode {
	return &TextNode{Content: content}
}

// WithStyle sets the style for the text element.
func (t *TextNode) WithStyle(style uv.Style) *TextNode {
	t.Style = style
	return t
}

// Styled creates a new text element with the given content and style.
// Deprecated: Use Text(content).WithStyle(style) instead.
func Styled(content string, style uv.Style) Element {
	return &TextNode{Content: content, Style: style}
}

// Render implements the Element interface.
func (t *TextNode) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	lines := strings.Split(t.Content, "\n")
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
				Style:   t.Style,
			}
			scr.SetCell(x, y, cell)
			x += width
		}

		y++
	}
}

// MinSize implements the Element interface.
func (t *TextNode) MinSize(scr uv.Screen) (width, height int) {
	lines := strings.Split(t.Content, "\n")
	height = len(lines)

	for _, line := range lines {
		w := displaywidth.String(line)
		if w > width {
			width = w
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
