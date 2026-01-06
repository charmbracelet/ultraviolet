package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// separator represents a visual separator line.
type separator struct {
	vertical bool
	char     string
	style    uv.Style
}

// Separator creates a horizontal separator line.
func Separator() Element {
	return &separator{
		vertical: false,
		char:     "─",
	}
}

// SeparatorStyled creates a horizontal separator with custom character and style.
func SeparatorStyled(char string, style uv.Style) Element {
	return &separator{
		vertical: false,
		char:     char,
		style:    style,
	}
}

// SeparatorVertical creates a vertical separator line.
func SeparatorVertical() Element {
	return &separator{
		vertical: true,
		char:     "│",
	}
}

// SeparatorVerticalStyled creates a vertical separator with custom character and style.
func SeparatorVerticalStyled(char string, style uv.Style) Element {
	return &separator{
		vertical: true,
		char:     char,
		style:    style,
	}
}

// Render implements the Element interface.
func (s *separator) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	char := s.char
	if char == "" {
		if s.vertical {
			char = "│"
		} else {
			char = "─"
		}
	}

	cell := &uv.Cell{
		Content: char,
		Width:   1,
		Style:   s.style,
	}

	if s.vertical {
		// Draw vertical line
		x := area.Min.X
		for y := area.Min.Y; y < area.Max.Y; y++ {
			scr.SetCell(x, y, cell)
		}
	} else {
		// Draw horizontal line
		y := area.Min.Y
		for x := area.Min.X; x < area.Max.X; x++ {
			scr.SetCell(x, y, cell)
		}
	}
}

// MinSize implements the Element interface.
func (s *separator) MinSize(scr uv.Screen) (width, height int) {
	if s.vertical {
		return 1, 0
	}
	return 0, 1
}

// Deprecated: Use BlockBox(child).WithBorder(NormalBorder()) instead.
// This old decorator will be removed in a future version.
type deprecatedBorder struct {
	child  Element
	style  uv.Style
	border uv.Border
}

// Deprecated: Use BlockBox(child).WithBorder(NormalBorder()) instead.
func BorderDeprecated(child Element) Element {
	return &deprecatedBorder{
		child:  child,
		border: uv.NormalBorder(),
	}
}

// Deprecated: Use BlockBox with WithBorder() instead.
func BorderStyledDeprecated(child Element, borderStyle uv.Border, style uv.Style) Element {
	return &deprecatedBorder{
		child:  child,
		border: borderStyle,
		style:  style,
	}
}

// Render implements the Element interface.
func (b *deprecatedBorder) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() < 2 || area.Dy() < 2 {
		// Not enough space for border
		return
	}

	// Draw the border
	styledBorder := b.border.Style(b.style)
	styledBorder.Draw(scr, area)

	// Render child in the inner area
	innerArea := uv.Rect(
		area.Min.X+1,
		area.Min.Y+1,
		area.Dx()-2,
		area.Dy()-2,
	)

	if b.child != nil && innerArea.Dx() > 0 && innerArea.Dy() > 0 {
		b.child.Render(scr, innerArea)
	}
}

// MinSize implements the Element interface.
func (b *deprecatedBorder) MinSize(scr uv.Screen) (width, height int) {
	if b.child != nil {
		width, height = b.child.MinSize(scr)
	}
	// Add border size (2 for each dimension)
	return width + 2, height + 2
}

// padder represents padding around an element.
type padder struct {
	child  Element
	top    int
	right  int
	bottom int
	left   int
}

// Padding creates padding around the given element.
func Padding(child Element, top, right, bottom, left int) Element {
	return &padder{
		child:  child,
		top:    top,
		right:  right,
		bottom: bottom,
		left:   left,
	}
}

// PaddingAll creates uniform padding on all sides.
func PaddingAll(child Element, padding int) Element {
	return &padder{
		child:  child,
		top:    padding,
		right:  padding,
		bottom: padding,
		left:   padding,
	}
}

// Render implements the Element interface.
func (p *padder) Render(scr uv.Screen, area uv.Rectangle) {
	innerArea := uv.Rect(
		area.Min.X+p.left,
		area.Min.Y+p.top,
		area.Dx()-p.left-p.right,
		area.Dy()-p.top-p.bottom,
	)

	if p.child != nil && innerArea.Dx() > 0 && innerArea.Dy() > 0 {
		p.child.Render(scr, innerArea)
	}
}

// MinSize implements the Element interface.
func (p *padder) MinSize(scr uv.Screen) (width, height int) {
	if p.child != nil {
		width, height = p.child.MinSize(scr)
	}
	return width + p.left + p.right, height + p.top + p.bottom
}

// center represents a centered element within its area.
type center struct {
	child Element
}

// Center creates a centered element within its available area.
func Center(child Element) Element {
	return &center{child: child}
}

// Render implements the Element interface.
func (c *center) Render(scr uv.Screen, area uv.Rectangle) {
	if c.child == nil || area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	minW, minH := c.child.MinSize(scr)

	// Calculate centered position
	childArea := uv.CenterRect(area, minW, minH)
	c.child.Render(scr, childArea)
}

// MinSize implements the Element interface.
func (c *center) MinSize(scr uv.Screen) (width, height int) {
	if c.child != nil {
		return c.child.MinSize(scr)
	}
	return 0, 0
}
