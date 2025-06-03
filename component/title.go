package component

import (
	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/layout"
	"github.com/rivo/uniseg"
)

// Title represents a component title that can be drawn on a screen.
type Title struct {
	// Content is the text content of the title.
	Content string
	// Style is the style of the title.
	Style uv.Style
	// Link is the link of the title.
	Link uv.Link
	// Padding contains the padding around the title.
	PaddingStart, PaddingEnd int
	// Alignment specifies how the title should be aligned within its area.
	Alignment layout.Alignment
	// Direction specifies the direction of the title text.
	Direction layout.Direction
}

// ComputeArea computes the area that the title will occupy based on its
// content and the specified area.
func (t Title) ComputeArea(wm uv.WidthMethod, area uv.Rectangle) uv.Rectangle {
	if t.Content == "" {
		return uv.Rectangle{}
	}

	grCount := uniseg.GraphemeClusterCount(t.Content)
	w := wm.StringWidth(t.Content)
	if w == 0 {
		return uv.Rectangle{}
	}

	titleArea := area
	switch t.Direction {
	case layout.Horizontal:
		switch t.Alignment {
		case layout.Start:
			titleArea.Min.X = titleArea.Min.X
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		case layout.Left:
			titleArea.Min.X = titleArea.Min.X + 1
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		case layout.Center:
			titleArea.Min.X = titleArea.Min.X + (titleArea.Dx()-w)/2
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		case layout.Right:
			titleArea.Min.X = titleArea.Max.X - 1 - w
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		case layout.End:
			titleArea.Min.X = titleArea.Max.X - w
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		default:
			if t.Alignment < 0 {
				t.Alignment = layout.Start
			} else if t.Alignment > 100 {
				t.Alignment = layout.End
			}
			titleArea.Min.X = (t.Alignment * titleArea.Dx()) / 100
			titleArea.Max.X = titleArea.Min.X + w + t.PaddingStart + t.PaddingEnd
		}
	case layout.Vertical:
		switch t.Alignment {
		case layout.Start:
			titleArea.Min.Y = titleArea.Min.Y
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		case layout.Top:
			titleArea.Min.Y = titleArea.Min.Y + 1
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		case layout.Center:
			titleArea.Min.Y = titleArea.Min.Y + (titleArea.Dy()-w)/2
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		case layout.Bottom:
			titleArea.Min.Y = titleArea.Max.Y - 1 - w
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		case layout.End:
			titleArea.Min.Y = titleArea.Max.Y - w
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		default:
			if t.Alignment < 0 {
				t.Alignment = layout.Start
			} else if t.Alignment > 100 {
				t.Alignment = layout.End
			}
			titleArea.Min.Y = (t.Alignment*titleArea.Dy())/100 + t.PaddingStart
			titleArea.Max.Y = titleArea.Min.Y + grCount + t.PaddingStart + t.PaddingEnd
		}
	}

	// Ensure the area is within the bounds of the original area.
	return area.Intersect(titleArea)
}

// Draw draws the title on the given screen within the specified area.
func (t Title) Draw(scr uv.Screen, area uv.Rectangle) {
	if t.Content == "" {
		return
	}

	method := scr.WidthMethod()
	startX := area.Min.X
	startY := area.Min.Y
	w := method.StringWidth(t.Content)
	grCount := uniseg.GraphemeClusterCount(t.Content)

	switch t.Direction {
	case layout.Horizontal:
		switch t.Alignment {
		case layout.Start:
			startX = area.Min.X + t.PaddingStart
		case layout.Left:
			startX = area.Min.X + 1 + t.PaddingStart
		case layout.Center:
			startX = area.Min.X + (area.Dx()-w)/2 + t.PaddingStart
		case layout.Right:
			startX = area.Max.X - 1 - w - t.PaddingEnd
		case layout.End:
			startX = area.Max.X - w - t.PaddingEnd
		default:
			if t.Alignment < 0 {
				t.Alignment = layout.Start
			} else if t.Alignment > 100 {
				t.Alignment = layout.End
			}
			startX = area.Min.X + (t.Alignment*area.Dx())/100 + t.PaddingStart
		}
	case layout.Vertical:
		switch t.Alignment {
		case layout.Start:
			startY = area.Min.Y + t.PaddingStart
		case layout.Top:
			startY = area.Min.Y + 1 + t.PaddingStart
		case layout.Center:
			startY = area.Min.Y + (area.Dy()-grCount)/2 + t.PaddingStart
		case layout.Bottom:
			startY = area.Max.Y - 1 - grCount - t.PaddingEnd
		case layout.End:
			startY = area.Max.Y - grCount - t.PaddingEnd
		default:
			if t.Alignment < 0 {
				t.Alignment = layout.Start
			} else if t.Alignment > 100 {
				t.Alignment = layout.End
			}
			startY = area.Min.Y + (t.Alignment*area.Dy())/100 + t.PaddingStart
		}
	}

	var pad *uv.Cell
	if t.PaddingEnd > 0 || t.PaddingStart > 0 {
		pad = new(uv.Cell)
		*pad = uv.EmptyCell
		pad.Style = t.Style
		pad.Link = t.Link
	}

	if t.PaddingStart > 0 {
		switch t.Direction {
		case layout.Horizontal:
			for x := startX - t.PaddingStart; x < startX && x >= area.Min.X; x++ {
				scr.SetCell(x, startY, pad)
			}
		case layout.Vertical:
			for y := startY - t.PaddingStart; y < startY && y >= area.Min.Y; y++ {
				scr.SetCell(startX, y, pad)
			}
		}
	}

	x, y := startX, startY
	gr := uniseg.NewGraphemes(t.Content)
	for gr.Next() {
		str := gr.Str()

		// We only care about the first line of the title.
		if str == "\n" {
			break
		}

		c := uv.Cell{
			Style:   t.Style,
			Link:    t.Link,
			Content: str,
			Width:   method.StringWidth(str),
		}
		switch t.Direction {
		case layout.Horizontal:
			scr.SetCell(x, startY, &c)
			x += c.Width
		case layout.Vertical:
			scr.SetCell(startX, y, &c)
			y++
		}
	}

	if t.PaddingEnd > 0 {
		switch t.Direction {
		case layout.Horizontal:
			for x := startX + w; x < startX+w+t.PaddingEnd && x < area.Max.X; x++ {
				scr.SetCell(x, startY, pad)
			}
		case layout.Vertical:
			for y := startY + grCount; y < startY+grCount+t.PaddingEnd && y < area.Max.Y; y++ {
				scr.SetCell(startX, y, pad)
			}
		}
	}
}
