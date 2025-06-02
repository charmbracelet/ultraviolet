// Package paragraph provides a component for rendering paragraphs in a
// document.
package paragraph

import (
	"github.com/charmbracelet/uv"
	"github.com/rivo/uniseg"
)

// Paragraph represents a block of text in a document that can have a style,
// link, and can be wrapped or truncated with an optional tail.
type Paragraph struct {
	uv.Style
	uv.Link

	Text string
	// If true, the text will be truncated to fit the available width.
	Truncate bool
	// The tail to append to the text when truncated.
	Tail string
}

// Draw draws the paragraph on the given [uv.Screen] within the specified area.
func (p Paragraph) Draw(scr uv.Screen, area uv.Rectangle) {
	if p.Text == "" {
		return
	}

	var (
		x         = area.Min.X
		y         = area.Min.Y
		tailWidth = scr.WidthMethod().StringWidth(p.Tail)
	)
	grs := uniseg.NewGraphemes(p.Text)
	for grs.Next() {
		str := grs.Str()
		w := scr.WidthMethod().StringWidth(str)
		if w == 0 {
			switch str {
			case "\n":
				y++
			case "\r":
				x = area.Min.X
			}
			continue
		}

		if x+w > area.Max.X && y+1 < area.Max.Y {
			// Wrap the string to the width of the screen
			x = area.Min.X
			y++
		}

		pos := uv.Pos(x, y)
		if pos.In(area) {
			if p.Truncate && y >= area.Max.Y-1 && tailWidth > 0 && x+w > area.Max.X-tailWidth {
				// Truncate the string and append the tail if any.
				tailgrs := uniseg.NewGraphemes(p.Tail)
				for tailgrs.Next() {
					t := uv.NewCell(scr.WidthMethod(), tailgrs.Str())
					t.Style = p.Style
					t.Link = p.Link
					scr.SetCell(x, y, t)
					x += t.Width
				}
				// We're done here, so we can break out of the loop.
				break
			} else {
				// Print the cell to the screen
				c := uv.Cell{
					Content: str,
					Width:   w,
					Style:   p.Style,
					Link:    p.Link,
				}
				scr.SetCell(x, y, &c)
				x += w
			}
		}
	}
}
