package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// SelectionRange represents a character-level selection range.
type SelectionRange struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// Normalize ensures start is before end.
func (s SelectionRange) Normalize() SelectionRange {
	if s.StartLine > s.EndLine || (s.StartLine == s.EndLine && s.StartCol > s.EndCol) {
		return SelectionRange{
			StartLine: s.EndLine,
			StartCol:  s.EndCol,
			EndLine:   s.StartLine,
			EndCol:    s.StartCol,
		}
	}
	return s
}

// IsEmpty returns true if the selection is empty.
func (s SelectionRange) IsEmpty() bool {
	return s.StartLine == s.EndLine && s.StartCol == s.EndCol
}

// Contains checks if a position is within the selection.
func (s SelectionRange) Contains(line, col int) bool {
	norm := s.Normalize()
	if line < norm.StartLine || line > norm.EndLine {
		return false
	}
	if line == norm.StartLine && col < norm.StartCol {
		return false
	}
	if line == norm.EndLine && col >= norm.EndCol {
		return false
	}
	return true
}

// Selectable represents an element that can have character-level selection.
type Selectable interface {
	Element
	SetSelection(selection SelectionRange)
	GetSelection() (SelectionRange, bool)
	ClearSelection()
	SetSelectionStyle(style uv.Style)
}

// Focusable represents an element that can receive focus.
type Focusable interface {
	Element
	SetFocused(focused bool)
	IsFocused() bool
}

// selectableElement wraps an element to make it selectable with character-level selection.
type selectableElement struct {
	child          Element
	hasSelection   bool
	selection      SelectionRange
	selectionStyle uv.Style
}

// MakeSelectable wraps an element to make it selectable.
func MakeSelectable(child Element) Selectable {
	return &selectableElement{
		child:          child,
		hasSelection:   false,
		selectionStyle: uv.Style{Attrs: uv.AttrReverse}, // Default terminal selection style
	}
}

// MakeSelectableStyled wraps an element with custom selection style.
func MakeSelectableStyled(child Element, style uv.Style) Selectable {
	return &selectableElement{
		child:          child,
		hasSelection:   false,
		selectionStyle: style,
	}
}

// SetSelection sets the selection range.
func (s *selectableElement) SetSelection(selection SelectionRange) {
	s.selection = selection
	s.hasSelection = !selection.IsEmpty()
}

// GetSelection returns the selection range and whether there is a selection.
func (s *selectableElement) GetSelection() (SelectionRange, bool) {
	return s.selection, s.hasSelection
}

// ClearSelection clears the selection.
func (s *selectableElement) ClearSelection() {
	s.hasSelection = false
	s.selection = SelectionRange{}
}

// SetSelectionStyle sets the style to use for selected text.
func (s *selectableElement) SetSelectionStyle(style uv.Style) {
	s.selectionStyle = style
}

// Render implements the Element interface.
func (s *selectableElement) Render(scr uv.Screen, area uv.Rectangle) {
	// Render the child element first
	s.child.Render(scr, area)

	// Apply selection highlighting if there is a selection
	if s.hasSelection {
		norm := s.selection.Normalize()
		
		// Iterate through the area and apply selection style to selected cells
		for y := area.Min.Y; y < area.Max.Y; y++ {
			line := y - area.Min.Y
			for x := area.Min.X; x < area.Max.X; x++ {
				col := x - area.Min.X
				
				// Check if this position is in the selection
				if norm.Contains(line, col) {
					if cell := scr.CellAt(x, y); cell != nil {
						// Apply selection style on top of existing style
						// Keep the original colors but add the selection attribute
						newStyle := cell.Style
						newStyle.Attrs |= s.selectionStyle.Attrs
						
						// If selection style has colors, use them (override)
						if s.selectionStyle.Fg != nil {
							newStyle.Fg = s.selectionStyle.Fg
						}
						if s.selectionStyle.Bg != nil {
							newStyle.Bg = s.selectionStyle.Bg
						}
						
						cell.Style = newStyle
						scr.SetCell(x, y, cell)
					}
				}
			}
		}
	}
}

// MinSize implements the Element interface.
func (s *selectableElement) MinSize(scr uv.Screen) (width, height int) {
	return s.child.MinSize(scr)
}

// focusableElement wraps an element to make it focusable.
type focusableElement struct {
	child    Element
	focused  bool
	style    uv.Style
	styleSet bool
}

// MakeFocusable wraps an element to make it focusable.
func MakeFocusable(child Element) Focusable {
	return &focusableElement{
		child:   child,
		focused: false,
	}
}

// MakeFocusableStyled wraps an element with custom focus style.
func MakeFocusableStyled(child Element, style uv.Style) Focusable {
	return &focusableElement{
		child:    child,
		focused:  false,
		style:    style,
		styleSet: true,
	}
}

// SetFocused sets the focus state.
func (f *focusableElement) SetFocused(focused bool) {
	f.focused = focused
}

// IsFocused returns whether the element is focused.
func (f *focusableElement) IsFocused() bool {
	return f.focused
}

// Render implements the Element interface.
func (f *focusableElement) Render(scr uv.Screen, area uv.Rectangle) {
	// Render the child element
	f.child.Render(scr, area)

	// If focused, apply focus styling
	if f.focused {
		// Draw focus indicator on the left edge
		for y := area.Min.Y; y < area.Max.Y; y++ {
			if cell := scr.CellAt(area.Min.X, y); cell != nil {
				if f.styleSet {
					// Use custom focus style
					cell.Style = f.style
				} else {
					// Default: reverse video
					cell.Style.Attrs |= uv.AttrReverse
				}
				scr.SetCell(area.Min.X, y, cell)
			}
		}
	}
}

// MinSize implements the Element interface.
func (f *focusableElement) MinSize(scr uv.Screen) (width, height int) {
	return f.child.MinSize(scr)
}

// selectableAndFocusableElement combines both selectable and focusable.
type selectableAndFocusableElement struct {
	child          Element
	hasSelection   bool
	selection      SelectionRange
	focused        bool
	selectionStyle uv.Style
	focusStyle     uv.Style
	focusStyleSet  bool
}

// MakeSelectableAndFocusable wraps an element to make it both selectable and focusable.
func MakeSelectableAndFocusable(child Element) *selectableAndFocusableElement {
	return &selectableAndFocusableElement{
		child:          child,
		hasSelection:   false,
		focused:        false,
		selectionStyle: uv.Style{Attrs: uv.AttrReverse}, // Default terminal selection style
	}
}

// MakeSelectableAndFocusableStyled wraps an element with custom styles.
func MakeSelectableAndFocusableStyled(child Element, selectStyle, focusStyle uv.Style) *selectableAndFocusableElement {
	return &selectableAndFocusableElement{
		child:          child,
		hasSelection:   false,
		focused:        false,
		selectionStyle: selectStyle,
		focusStyle:     focusStyle,
		focusStyleSet:  true,
	}
}

// SetSelection sets the selection range.
func (s *selectableAndFocusableElement) SetSelection(selection SelectionRange) {
	s.selection = selection
	s.hasSelection = !selection.IsEmpty()
}

// GetSelection returns the selection range and whether there is a selection.
func (s *selectableAndFocusableElement) GetSelection() (SelectionRange, bool) {
	return s.selection, s.hasSelection
}

// ClearSelection clears the selection.
func (s *selectableAndFocusableElement) ClearSelection() {
	s.hasSelection = false
	s.selection = SelectionRange{}
}

// SetSelectionStyle sets the style to use for selected text.
func (s *selectableAndFocusableElement) SetSelectionStyle(style uv.Style) {
	s.selectionStyle = style
}

// SetFocused sets the focus state.
func (s *selectableAndFocusableElement) SetFocused(focused bool) {
	s.focused = focused
}

// IsFocused returns whether the element is focused.
func (s *selectableAndFocusableElement) IsFocused() bool {
	return s.focused
}

// Render implements the Element interface.
func (s *selectableAndFocusableElement) Render(scr uv.Screen, area uv.Rectangle) {
	// Render the child element
	s.child.Render(scr, area)

	// Apply selection highlighting if there is a selection
	if s.hasSelection {
		norm := s.selection.Normalize()
		
		for y := area.Min.Y; y < area.Max.Y; y++ {
			line := y - area.Min.Y
			for x := area.Min.X; x < area.Max.X; x++ {
				col := x - area.Min.X
				
				if norm.Contains(line, col) {
					if cell := scr.CellAt(x, y); cell != nil {
						newStyle := cell.Style
						newStyle.Attrs |= s.selectionStyle.Attrs
						
						if s.selectionStyle.Fg != nil {
							newStyle.Fg = s.selectionStyle.Fg
						}
						if s.selectionStyle.Bg != nil {
							newStyle.Bg = s.selectionStyle.Bg
						}
						
						cell.Style = newStyle
						scr.SetCell(x, y, cell)
					}
				}
			}
		}
	}

	// Apply focus styling if focused (on the left edge)
	if s.focused {
		for y := area.Min.Y; y < area.Max.Y; y++ {
			if cell := scr.CellAt(area.Min.X, y); cell != nil {
				if s.focusStyleSet {
					cell.Style = s.focusStyle
				} else {
					cell.Style.Attrs |= uv.AttrReverse
				}
				scr.SetCell(area.Min.X, y, cell)
			}
		}
	}
}

// MinSize implements the Element interface.
func (s *selectableAndFocusableElement) MinSize(scr uv.Screen) (width, height int) {
	return s.child.MinSize(scr)
}
