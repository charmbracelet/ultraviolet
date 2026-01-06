package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// BoxDisplay defines how a box is displayed (similar to CSS display property).
type BoxDisplay int

const (
	// DisplayBlock creates a block-level box that starts on a new line
	// and takes up full width available. Supports margins, borders, padding.
	DisplayBlock BoxDisplay = iota
	
	// DisplayInline creates an inline-level box that flows with text content.
	// Does not start on a new line. Vertical margins/padding are ignored.
	DisplayInline
	
	// DisplayInlineBlock creates an inline-level box that can have block properties.
	// Flows with content but respects width, height, and all margins/padding.
	DisplayInlineBlock
)

// BoxStyle defines the visual and behavioral styling for a Box.
// It follows a CSS-like box model with borders, padding, margins, etc.
type BoxStyle struct {
	// Display mode
	Display BoxDisplay

	// Border configuration
	Border *Border

	// Padding (inner spacing)
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int

	// Margins (outer spacing)
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int

	// Text/content styling
	Style uv.Style

	// Focus styling
	Focusable  bool
	Focused    bool
	FocusStyle uv.Style

	// Selection styling
	Selectable       bool
	Selection        SelectionRange
	HasSelection     bool
	SelectionStyle   uv.Style

	// Text wrapping and truncation
	Wrap     bool // Hard-wrap text at boundaries
	Truncate bool // Truncate text that doesn't fit

	// Background fill
	Background *uv.Cell

	// Width and height constraints
	Width  int // Fixed width (0 = flexible)
	Height int // Fixed height (0 = flexible)
	MinWidth  int
	MinHeight int
	MaxWidth  int
	MaxHeight int
}

// DefaultBoxStyle returns a BoxStyle with sensible defaults.
func DefaultBoxStyle() BoxStyle {
	return BoxStyle{
		Display:        DisplayBlock,
		SelectionStyle: uv.Style{Attrs: uv.AttrReverse},
		FocusStyle:     uv.Style{Attrs: uv.AttrReverse},
		Wrap:           false,
		Truncate:       true,
	}
}

// DefaultInlineStyle returns a BoxStyle with inline display defaults.
func DefaultInlineStyle() BoxStyle {
	return BoxStyle{
		Display:        DisplayInline,
		SelectionStyle: uv.Style{Attrs: uv.AttrReverse},
		FocusStyle:     uv.Style{Attrs: uv.AttrReverse},
		Wrap:           false,
		Truncate:       true,
	}
}

// DefaultInlineBlockStyle returns a BoxStyle with inline-block display defaults.
func DefaultInlineBlockStyle() BoxStyle {
	return BoxStyle{
		Display:        DisplayInlineBlock,
		SelectionStyle: uv.Style{Attrs: uv.AttrReverse},
		FocusStyle:     uv.Style{Attrs: uv.AttrReverse},
		Wrap:           false,
		Truncate:       true,
	}
}

// Box represents a container that wraps an element with styling capabilities.
// It follows a CSS-like box model with borders, padding, margins, focus, selection, etc.
type Box struct {
	child Element
	style BoxStyle
}

// NewBox creates a new Box with the given child element and style.
func NewBox(child Element, style BoxStyle) *Box {
	return &Box{
		child: child,
		style: style,
	}
}

// BlockBox creates a block-level Box with default styling.
func BlockBox(child Element) *Box {
	return &Box{
		child: child,
		style: DefaultBoxStyle(),
	}
}

// InlineBox creates an inline-level Box with default styling.
func InlineBox(child Element) *Box {
	return &Box{
		child: child,
		style: DefaultInlineStyle(),
	}
}

// InlineBlockBox creates an inline-block Box with default styling.
func InlineBlockBox(child Element) *Box {
	return &Box{
		child: child,
		style: DefaultInlineBlockStyle(),
	}
}

// StyledBox creates a Box with default block styling that can be customized.
// Deprecated: Use BlockBox instead.
func StyledBox(child Element) *Box {
	return BlockBox(child)
}

// WithDisplay sets the display mode for this box.
func (b *Box) WithDisplay(display BoxDisplay) *Box {
	b.style.Display = display
	return b
}

// WithBorder sets the border for this box.
func (b *Box) WithBorder(border *Border) *Box {
	b.style.Border = border
	return b
}

// WithPadding sets uniform padding on all sides.
func (b *Box) WithPadding(padding int) *Box {
	b.style.PaddingTop = padding
	b.style.PaddingRight = padding
	b.style.PaddingBottom = padding
	b.style.PaddingLeft = padding
	return b
}

// WithPaddingDetailed sets padding for each side individually.
func (b *Box) WithPaddingDetailed(top, right, bottom, left int) *Box {
	b.style.PaddingTop = top
	b.style.PaddingRight = right
	b.style.PaddingBottom = bottom
	b.style.PaddingLeft = left
	return b
}

// WithMargin sets uniform margin on all sides.
func (b *Box) WithMargin(margin int) *Box {
	b.style.MarginTop = margin
	b.style.MarginRight = margin
	b.style.MarginBottom = margin
	b.style.MarginLeft = margin
	return b
}

// WithMarginDetailed sets margin for each side individually.
func (b *Box) WithMarginDetailed(top, right, bottom, left int) *Box {
	b.style.MarginTop = top
	b.style.MarginRight = right
	b.style.MarginBottom = bottom
	b.style.MarginLeft = left
	return b
}

// WithStyle sets the content style.
func (b *Box) WithStyle(style uv.Style) *Box {
	b.style.Style = style
	return b
}

// WithFocusable makes the box focusable.
func (b *Box) WithFocusable(focusable bool) *Box {
	b.style.Focusable = focusable
	return b
}

// SetFocused sets the focus state.
func (b *Box) SetFocused(focused bool) {
	b.style.Focused = focused
}

// IsFocused returns whether the box is focused.
func (b *Box) IsFocused() bool {
	return b.style.Focused
}

// WithFocusStyle sets the focus indicator style.
func (b *Box) WithFocusStyle(style uv.Style) *Box {
	b.style.FocusStyle = style
	return b
}

// WithSelectable makes the box selectable.
func (b *Box) WithSelectable(selectable bool) *Box {
	b.style.Selectable = selectable
	return b
}

// SetSelection sets the selection range.
func (b *Box) SetSelection(selection SelectionRange) {
	b.style.Selection = selection
	b.style.HasSelection = !selection.IsEmpty()
}

// GetSelection returns the selection range and whether there is a selection.
func (b *Box) GetSelection() (SelectionRange, bool) {
	return b.style.Selection, b.style.HasSelection
}

// ClearSelection clears the selection.
func (b *Box) ClearSelection() {
	b.style.HasSelection = false
	b.style.Selection = SelectionRange{}
}

// WithSelectionStyle sets the selection highlight style.
func (b *Box) WithSelectionStyle(style uv.Style) *Box {
	b.style.SelectionStyle = style
	return b
}

// WithWrap enables or disables text wrapping.
func (b *Box) WithWrap(wrap bool) *Box {
	b.style.Wrap = wrap
	return b
}

// WithTruncate enables or disables text truncation.
func (b *Box) WithTruncate(truncate bool) *Box {
	b.style.Truncate = truncate
	return b
}

// WithBackground sets a background fill.
func (b *Box) WithBackground(cell *uv.Cell) *Box {
	b.style.Background = cell
	return b
}

// WithWidth sets a fixed width.
func (b *Box) WithWidth(width int) *Box {
	b.style.Width = width
	return b
}

// WithHeight sets a fixed height.
func (b *Box) WithHeight(height int) *Box {
	b.style.Height = height
	return b
}

// WithSize sets fixed width and height.
func (b *Box) WithSize(width, height int) *Box {
	b.style.Width = width
	b.style.Height = height
	return b
}

// WithMinSize sets minimum dimensions.
func (b *Box) WithMinSize(width, height int) *Box {
	b.style.MinWidth = width
	b.style.MinHeight = height
	return b
}

// WithMaxSize sets maximum dimensions.
func (b *Box) WithMaxSize(width, height int) *Box {
	b.style.MaxWidth = width
	b.style.MaxHeight = height
	return b
}

// Render implements the Element interface.
func (b *Box) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	// For inline boxes, ignore vertical margins
	marginTop := b.style.MarginTop
	marginBottom := b.style.MarginBottom
	if b.style.Display == DisplayInline {
		marginTop = 0
		marginBottom = 0
	}

	// Apply margins to reduce the area
	innerArea := uv.Rect(
		area.Min.X+b.style.MarginLeft,
		area.Min.Y+marginTop,
		area.Dx()-b.style.MarginLeft-b.style.MarginRight,
		area.Dy()-marginTop-marginBottom,
	)

	if innerArea.Dx() <= 0 || innerArea.Dy() <= 0 {
		return
	}

	// Fill background if specified
	if b.style.Background != nil {
		for y := innerArea.Min.Y; y < innerArea.Max.Y; y++ {
			for x := innerArea.Min.X; x < innerArea.Max.X; x++ {
				scr.SetCell(x, y, b.style.Background)
			}
		}
	}

	// Draw border if specified
	borderArea := innerArea
	if b.style.Border != nil {
		b.style.Border.Draw(scr, innerArea)
		
		// Reduce area for padding and content
		innerArea = uv.Rect(
			innerArea.Min.X+1,
			innerArea.Min.Y+1,
			innerArea.Dx()-2,
			innerArea.Dy()-2,
		)
	}

	if innerArea.Dx() <= 0 || innerArea.Dy() <= 0 {
		return
	}

	// For inline boxes, ignore vertical padding
	paddingTop := b.style.PaddingTop
	paddingBottom := b.style.PaddingBottom
	if b.style.Display == DisplayInline {
		paddingTop = 0
		paddingBottom = 0
	}

	// Apply padding
	contentArea := uv.Rect(
		innerArea.Min.X+b.style.PaddingLeft,
		innerArea.Min.Y+paddingTop,
		innerArea.Dx()-b.style.PaddingLeft-b.style.PaddingRight,
		innerArea.Dy()-paddingTop-paddingBottom,
	)

	if contentArea.Dx() <= 0 || contentArea.Dy() <= 0 {
		return
	}

	// Render child content
	if b.child != nil {
		b.child.Render(scr, contentArea)
	}

	// Apply selection highlighting
	if b.style.HasSelection && b.style.Selectable {
		norm := b.style.Selection.Normalize()
		
		for y := contentArea.Min.Y; y < contentArea.Max.Y; y++ {
			line := y - contentArea.Min.Y
			for x := contentArea.Min.X; x < contentArea.Max.X; x++ {
				col := x - contentArea.Min.X
				
				if norm.Contains(line, col) {
					if cell := scr.CellAt(x, y); cell != nil {
						newStyle := cell.Style
						newStyle.Attrs |= b.style.SelectionStyle.Attrs
						
						if b.style.SelectionStyle.Fg != nil {
							newStyle.Fg = b.style.SelectionStyle.Fg
						}
						if b.style.SelectionStyle.Bg != nil {
							newStyle.Bg = b.style.SelectionStyle.Bg
						}
						
						cell.Style = newStyle
						scr.SetCell(x, y, cell)
					}
				}
			}
		}
	}

	// Apply focus indicator
	if b.style.Focused && b.style.Focusable {
		// Draw focus indicator on the left edge of the border area
		for y := borderArea.Min.Y; y < borderArea.Max.Y; y++ {
			if cell := scr.CellAt(borderArea.Min.X, y); cell != nil {
				newStyle := cell.Style
				newStyle.Attrs |= b.style.FocusStyle.Attrs
				
				if b.style.FocusStyle.Fg != nil {
					newStyle.Fg = b.style.FocusStyle.Fg
				}
				if b.style.FocusStyle.Bg != nil {
					newStyle.Bg = b.style.FocusStyle.Bg
				}
				
				cell.Style = newStyle
				scr.SetCell(borderArea.Min.X, y, cell)
			}
		}
	}
}

// MinSize implements the Element interface.
func (b *Box) MinSize(scr uv.Screen) (width, height int) {
	// Start with child's min size
	if b.child != nil {
		width, height = b.child.MinSize(scr)
	}

	// Apply constraints
	if b.style.MinWidth > 0 && width < b.style.MinWidth {
		width = b.style.MinWidth
	}
	if b.style.MinHeight > 0 && height < b.style.MinHeight {
		height = b.style.MinHeight
	}

	// Add padding (inline boxes ignore vertical padding for MinSize)
	width += b.style.PaddingLeft + b.style.PaddingRight
	if b.style.Display != DisplayInline {
		height += b.style.PaddingTop + b.style.PaddingBottom
	}

	// Add border
	if b.style.Border != nil {
		width += 2
		if b.style.Display != DisplayInline {
			height += 2
		}
	}

	// Add margins (inline boxes ignore vertical margins for MinSize)
	width += b.style.MarginLeft + b.style.MarginRight
	if b.style.Display != DisplayInline {
		height += b.style.MarginTop + b.style.MarginBottom
	}

	// Apply fixed dimensions if set
	if b.style.Width > 0 {
		width = b.style.Width
	}
	if b.style.Height > 0 {
		height = b.style.Height
	}

	// Apply max constraints
	if b.style.MaxWidth > 0 && width > b.style.MaxWidth {
		width = b.style.MaxWidth
	}
	if b.style.MaxHeight > 0 && height > b.style.MaxHeight {
		height = b.style.MaxHeight
	}

	return width, height
}

// GetStyle returns the box style.
func (b *Box) GetStyle() BoxStyle {
	return b.style
}

// SetStyle sets the box style.
func (b *Box) SetStyle(style BoxStyle) {
	b.style = style
}
