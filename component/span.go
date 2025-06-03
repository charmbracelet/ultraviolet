package component

import (
	"github.com/charmbracelet/uv"
	"github.com/rivo/uniseg"
)

// Span represents a continuous text span with the same style.
type Span struct {
	// Content is the text content of the span.
	Content string
	// Style is the style of the span.
	Style uv.Style
	// Link is the link of the span.
	Link uv.Link
}

// Draw draws the span on the given screen within the specified area.
func (s Span) Draw(scr uv.Screen, area uv.Rectangle) {
	if s.Content == "" {
		return
	}

	wm := scr.WidthMethod()
	spanArea := area
	spanArea.Max.X = spanArea.Min.X + wm.StringWidth(s.Content)
	spanArea.Max.Y = spanArea.Min.Y + 1
	gr := uniseg.NewGraphemes(s.Content)
	x := spanArea.Min.X
	for gr.Next() {
		str := gr.Str()
		if str == "\n" {
			// We only care about the first line of the content.
			break
		}
		width := wm.StringWidth(str)
		if width == 0 {
			continue
		}
		cell := uv.Cell{
			Content: str,
			Width:   width,
			Style:   s.Style,
			Link:    s.Link,
		}
		scr.SetCell(x, spanArea.Min.Y, &cell)
		x += width
	}
}
