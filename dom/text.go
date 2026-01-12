package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/clipperhouse/displaywidth"
)

// Text represents a text node in the DOM.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Text
type Text struct {
	baseNode
	data string // The text content
}

// NewText creates a new text node with the given data.
func NewText(data string) *Text {
	return &Text{
		baseNode: baseNode{
			nodeType: TextNode,
			nodeName: "#text",
		},
		data: data,
	}
}

// Data returns the text content of the node.
func (t *Text) Data() string {
	return t.data
}

// SetData sets the text content of the node.
func (t *Text) SetData(data string) {
	t.data = data
}

// TextContent returns the text content of the node.
func (t *Text) TextContent() string {
	return t.data
}

// SetTextContent sets the text content of the node.
func (t *Text) SetTextContent(text string) {
	t.data = text
}

// CloneNode creates a copy of the text node.
func (t *Text) CloneNode(deep bool) Node {
	return &Text{
		baseNode: baseNode{
			nodeType: TextNode,
			nodeName: "#text",
		},
		data: t.data,
	}
}

// Render renders the text node to the screen.
func (t *Text) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	// Split text into lines
	lines := strings.Split(t.data, "\n")
	
	y := area.Min.Y
	for _, line := range lines {
		if y >= area.Max.Y {
			break
		}

		x := area.Min.X
		graphemes := displaywidth.StringGraphemes(line)
		
		for graphemes.Next() {
			if x >= area.Max.X {
				break
			}

			gStr := graphemes.Value()
			gWidth := graphemes.Width()
			
			cell := uv.NewCell(scr.WidthMethod(), gStr)
			scr.SetCell(x, y, cell)
			x += gWidth
		}
		
		y++
	}
}

// MinSize returns the minimum size needed to render the text.
func (t *Text) MinSize(scr uv.Screen) (width, height int) {
	lines := strings.Split(t.data, "\n")
	height = len(lines)
	
	for _, line := range lines {
		lineWidth := 0
		graphemes := displaywidth.StringGraphemes(line)
		for graphemes.Next() {
			lineWidth += graphemes.Width()
		}
		if lineWidth > width {
			width = lineWidth
		}
	}
	
	return width, height
}
