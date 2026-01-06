package dom

import (
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
)

// Box is a base container that all elements can use or embed.
// It provides common functionality like borders, padding, scrolling, focus, and selection.
type Box struct {
	// Content is the child element to render inside the box
	Content Element

	// Border settings
	BorderStyle  *BorderStyle
	BorderFg     color.Color
	BorderBg     color.Color
	BorderAttrs  uint8

	// Padding (inner spacing between border and content)
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int

	// Scrolling support
	ScrollOffsetX int
	ScrollOffsetY int

	// Focus state
	Focusable bool
	Focused   bool

	// Selection state
	Selectable    bool
	Selected      bool
	SelectionFg   color.Color
	SelectionBg   color.Color
	SelectionAttr uint8
}

// BorderStyle holds the characters used for each part of the border.
type BorderStyle struct {
	Top         string
	Bottom      string
	Left        string
	Right       string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
}

// BorderStyleNormal returns a standard box-drawing border style.
func BorderStyleNormal() *BorderStyle {
	return &BorderStyle{"─", "─", "│", "│", "┌", "┐", "└", "┘"}
}

// BorderStyleRounded returns a border style with rounded corners.
func BorderStyleRounded() *BorderStyle {
	return &BorderStyle{"─", "─", "│", "│", "╭", "╮", "╰", "╯"}
}

// BorderStyleDouble returns a border style with double lines.
func BorderStyleDouble() *BorderStyle {
	return &BorderStyle{"═", "═", "║", "║", "╔", "╗", "╚", "╝"}
}

// BorderStyleThick returns a border style with thick lines.
func BorderStyleThick() *BorderStyle {
	return &BorderStyle{"━", "━", "┃", "┃", "┏", "┓", "┗", "┛"}
}

// NewBox creates a new box with the given content.
func NewBox(content Element) *Box {
	return &Box{
		Content:       content,
		SelectionAttr: uv.AttrReverse,
	}
}

// WithBorder adds a border to the box.
func (b *Box) WithBorder(style *BorderStyle) *Box {
	b.BorderStyle = style
	return b
}

// WithPadding sets the padding for all sides.
func (b *Box) WithPadding(padding int) *Box {
	b.PaddingTop = padding
	b.PaddingRight = padding
	b.PaddingBottom = padding
	b.PaddingLeft = padding
	return b
}

// WithFocus makes the box focusable and optionally sets its focus state.
func (b *Box) WithFocus(focused bool) *Box {
	b.Focusable = true
	b.Focused = focused
	return b
}

// WithSelection makes the box selectable and optionally sets its selection state.
func (b *Box) WithSelection(selected bool) *Box {
	b.Selectable = true
	b.Selected = selected
	return b
}

// ScrollUp scrolls the content up by the given amount.
func (b *Box) ScrollUp(amount int) {
	b.ScrollOffsetY = max(0, b.ScrollOffsetY-amount)
}

// ScrollDown scrolls the content down by the given amount.
func (b *Box) ScrollDown(amount int) {
	b.ScrollOffsetY += amount
}

// ScrollLeft scrolls the content left by the given amount.
func (b *Box) ScrollLeft(amount int) {
	b.ScrollOffsetX = max(0, b.ScrollOffsetX-amount)
}

// ScrollRight scrolls the content right by the given amount.
func (b *Box) ScrollRight(amount int) {
	b.ScrollOffsetX += amount
}

// Render implements the Element interface.
func (b *Box) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	contentArea := area

	// Draw border if present
	if b.BorderStyle != nil {
		b.renderBorder(scr, area)
		// Shrink content area to account for border
		contentArea = uv.Rect(
			area.Min.X+1,
			area.Min.Y+1,
			max(0, area.Dx()-2),
			max(0, area.Dy()-2),
		)
	}

	// Apply padding
	if b.PaddingTop > 0 || b.PaddingRight > 0 || b.PaddingBottom > 0 || b.PaddingLeft > 0 {
		contentArea = uv.Rect(
			contentArea.Min.X+b.PaddingLeft,
			contentArea.Min.Y+b.PaddingTop,
			max(0, contentArea.Dx()-b.PaddingLeft-b.PaddingRight),
			max(0, contentArea.Dy()-b.PaddingTop-b.PaddingBottom),
		)
	}

	if contentArea.Dx() <= 0 || contentArea.Dy() <= 0 {
		return
	}

	// Render content with scrolling offset
	if b.Content != nil {
		// Create a virtual screen that's offset by the scroll position
		// The content will render thinking it's at contentArea, but we'll
		// adjust the actual rendering position
		offsetArea := uv.Rect(
			contentArea.Min.X-b.ScrollOffsetX,
			contentArea.Min.Y-b.ScrollOffsetY,
			contentArea.Dx(),
			contentArea.Dy(),
		)
		b.Content.Render(scr, offsetArea)
	}

	// Apply selection highlighting if selected
	if b.Selected && b.Selectable {
		b.applySelection(scr, area)
	}

	// Apply focus indicator if focused
	if b.Focused && b.Focusable {
		b.applyFocus(scr, area)
	}
}

// MinSize implements the Element interface.
func (b *Box) MinSize(scr uv.Screen) (width, height int) {
	if b.Content != nil {
		width, height = b.Content.MinSize(scr)
	}

	// Add border size
	if b.BorderStyle != nil {
		width += 2
		height += 2
	}

	// Add padding
	width += b.PaddingLeft + b.PaddingRight
	height += b.PaddingTop + b.PaddingBottom

	return width, height
}

// renderBorder draws the border around the box.
func (b *Box) renderBorder(scr uv.Screen, area uv.Rectangle) {
	style := uv.Style{
		Fg:    b.BorderFg,
		Bg:    b.BorderBg,
		Attrs: b.BorderAttrs,
	}

	// Draw corners
	cell := uv.NewCell(scr.WidthMethod(), b.BorderStyle.TopLeft)
	cell.Style = style
	scr.SetCell(area.Min.X, area.Min.Y, cell)

	cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.TopRight)
	cell.Style = style
	scr.SetCell(area.Max.X-1, area.Min.Y, cell)

	cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.BottomLeft)
	cell.Style = style
	scr.SetCell(area.Min.X, area.Max.Y-1, cell)

	cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.BottomRight)
	cell.Style = style
	scr.SetCell(area.Max.X-1, area.Max.Y-1, cell)

	// Draw top and bottom edges
	for x := area.Min.X + 1; x < area.Max.X-1; x++ {
		cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.Top)
		cell.Style = style
		scr.SetCell(x, area.Min.Y, cell)

		cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.Bottom)
		cell.Style = style
		scr.SetCell(x, area.Max.Y-1, cell)
	}

	// Draw left and right edges
	for y := area.Min.Y + 1; y < area.Max.Y-1; y++ {
		cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.Left)
		cell.Style = style
		scr.SetCell(area.Min.X, y, cell)

		cell = uv.NewCell(scr.WidthMethod(), b.BorderStyle.Right)
		cell.Style = style
		scr.SetCell(area.Max.X-1, y, cell)
	}
}

// applySelection applies selection highlighting to the box area.
func (b *Box) applySelection(scr uv.Screen, area uv.Rectangle) {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			cell := scr.CellAt(x, y)
			if cell != nil {
				// Apply selection style
				cell.Style.Attrs |= b.SelectionAttr
				if b.SelectionFg != nil {
					cell.Style.Fg = b.SelectionFg
				}
				if b.SelectionBg != nil {
					cell.Style.Bg = b.SelectionBg
				}
				scr.SetCell(x, y, cell)
			}
		}
	}
}

// applyFocus applies focus indicator to the box.
func (b *Box) applyFocus(scr uv.Screen, area uv.Rectangle) {
	// Simple focus indicator: reverse the border or add a marker
	// For simplicity, we'll just apply reverse attribute to borders if present
	if b.BorderStyle != nil {
		for x := area.Min.X; x < area.Max.X; x++ {
			cell := scr.CellAt(x, area.Min.Y)
			if cell != nil {
				cell.Style.Attrs |= uv.AttrReverse
				scr.SetCell(x, area.Min.Y, cell)
			}
			cell = scr.CellAt(x, area.Max.Y-1)
			if cell != nil {
				cell.Style.Attrs |= uv.AttrReverse
				scr.SetCell(x, area.Max.Y-1, cell)
			}
		}
		for y := area.Min.Y; y < area.Max.Y; y++ {
			cell := scr.CellAt(area.Min.X, y)
			if cell != nil {
				cell.Style.Attrs |= uv.AttrReverse
				scr.SetCell(area.Min.X, y, cell)
			}
			cell = scr.CellAt(area.Max.X-1, y)
			if cell != nil {
				cell.Style.Attrs |= uv.AttrReverse
				scr.SetCell(area.Max.X-1, y, cell)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
