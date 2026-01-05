package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// scrollableVBox represents a vertical box container with scrollback support.
type scrollableVBox struct {
	children     []Element
	scrollOffset int
	focusedIndex int
}

// ScrollableVBox creates a vertical box container with scrollback.
func ScrollableVBox(children ...Element) *scrollableVBox {
	return &scrollableVBox{
		children:     children,
		scrollOffset: 0,
		focusedIndex: -1,
	}
}

// ScrollUp scrolls the content up by the given number of lines.
func (v *scrollableVBox) ScrollUp(lines int) {
	v.scrollOffset = max(0, v.scrollOffset-lines)
}

// ScrollDown scrolls the content down by the given number of lines.
func (v *scrollableVBox) ScrollDown(lines int) {
	v.scrollOffset += lines
}

// SetScrollOffset sets the scroll offset to a specific value.
func (v *scrollableVBox) SetScrollOffset(offset int) {
	v.scrollOffset = max(0, offset)
}

// GetScrollOffset returns the current scroll offset.
func (v *scrollableVBox) GetScrollOffset() int {
	return v.scrollOffset
}

// FocusNext moves focus to the next element.
func (v *scrollableVBox) FocusNext() {
	if len(v.children) == 0 {
		return
	}
	
	// Clear focus on current element if it's focusable
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	v.focusedIndex = (v.focusedIndex + 1) % len(v.children)
	
	// Set focus on new element if it's focusable
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
			focusable.SetFocused(true)
		}
	}
}

// FocusPrevious moves focus to the previous element.
func (v *scrollableVBox) FocusPrevious() {
	if len(v.children) == 0 {
		return
	}
	
	// Clear focus on current element if it's focusable
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	if v.focusedIndex <= 0 {
		v.focusedIndex = len(v.children) - 1
	} else {
		v.focusedIndex--
	}
	
	// Set focus on new element if it's focusable
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
			focusable.SetFocused(true)
		}
	}
}

// SetFocus sets focus to a specific element by index.
func (v *scrollableVBox) SetFocus(index int) {
	if index < 0 || index >= len(v.children) {
		return
	}
	
	// Clear focus on current element if it's focusable
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	v.focusedIndex = index
	
	// Set focus on new element if it's focusable
	if focusable, ok := v.children[v.focusedIndex].(Focusable); ok {
		focusable.SetFocused(true)
	}
}

// GetFocusedIndex returns the index of the currently focused element.
func (v *scrollableVBox) GetFocusedIndex() int {
	return v.focusedIndex
}

// GetFocusedElement returns the currently focused element, or nil if none.
func (v *scrollableVBox) GetFocusedElement() Element {
	if v.focusedIndex >= 0 && v.focusedIndex < len(v.children) {
		return v.children[v.focusedIndex]
	}
	return nil
}

// Render implements the Element interface.
func (v *scrollableVBox) Render(scr uv.Screen, area uv.Rectangle) {
	if len(v.children) == 0 || area.Dy() <= 0 {
		return
	}

	// Calculate total content height
	totalHeight := 0
	childHeights := make([]int, len(v.children))
	for i, child := range v.children {
		_, h := child.MinSize(scr)
		if h == 0 {
			h = 1 // Default height for flexible elements
		}
		childHeights[i] = h
		totalHeight += h
	}

	// Adjust scroll offset if necessary
	maxScroll := max(0, totalHeight-area.Dy())
	if v.scrollOffset > maxScroll {
		v.scrollOffset = maxScroll
	}

	// Render visible children
	y := area.Min.Y
	currentLine := 0
	for i, child := range v.children {
		childHeight := childHeights[i]

		// Skip children that are scrolled out of view
		if currentLine+childHeight <= v.scrollOffset {
			currentLine += childHeight
			continue
		}

		// Stop if we've rendered everything that fits in the area
		if y >= area.Max.Y {
			break
		}

		// Calculate visible portion of this child
		skipLines := 0
		if currentLine < v.scrollOffset {
			skipLines = v.scrollOffset - currentLine
		}

		visibleHeight := min(childHeight-skipLines, area.Max.Y-y)
		if visibleHeight <= 0 {
			currentLine += childHeight
			continue
		}

		childArea := uv.Rect(area.Min.X, y, area.Dx(), visibleHeight)

		// Render child (selection/focus is handled by the element itself)
		child.Render(scr, childArea)

		y += visibleHeight
		currentLine += childHeight
	}
}

// MinSize implements the Element interface.
func (v *scrollableVBox) MinSize(scr uv.Screen) (width, height int) {
	for _, child := range v.children {
		w, h := child.MinSize(scr)
		if w > width {
			width = w
		}
		height += h
	}
	return width, height
}

// scrollableHBox represents a horizontal box container with scrollback support.
type scrollableHBox struct {
	children     []Element
	scrollOffset int
	focusedIndex int
}

// ScrollableHBox creates a horizontal box container with scrollback.
func ScrollableHBox(children ...Element) *scrollableHBox {
	return &scrollableHBox{
		children:     children,
		scrollOffset: 0,
		focusedIndex: -1,
	}
}

// ScrollLeft scrolls the content left by the given number of columns.
func (h *scrollableHBox) ScrollLeft(cols int) {
	h.scrollOffset = max(0, h.scrollOffset-cols)
}

// ScrollRight scrolls the content right by the given number of columns.
func (h *scrollableHBox) ScrollRight(cols int) {
	h.scrollOffset += cols
}

// SetScrollOffset sets the scroll offset to a specific value.
func (h *scrollableHBox) SetScrollOffset(offset int) {
	h.scrollOffset = max(0, offset)
}

// GetScrollOffset returns the current scroll offset.
func (h *scrollableHBox) GetScrollOffset() int {
	return h.scrollOffset
}

// FocusNext moves focus to the next element.
func (h *scrollableHBox) FocusNext() {
	if len(h.children) == 0 {
		return
	}
	
	// Clear focus on current element if it's focusable
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	h.focusedIndex = (h.focusedIndex + 1) % len(h.children)
	
	// Set focus on new element if it's focusable
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
			focusable.SetFocused(true)
		}
	}
}

// FocusPrevious moves focus to the previous element.
func (h *scrollableHBox) FocusPrevious() {
	if len(h.children) == 0 {
		return
	}
	
	// Clear focus on current element if it's focusable
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	if h.focusedIndex <= 0 {
		h.focusedIndex = len(h.children) - 1
	} else {
		h.focusedIndex--
	}
	
	// Set focus on new element if it's focusable
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
			focusable.SetFocused(true)
		}
	}
}

// SetFocus sets focus to a specific element by index.
func (h *scrollableHBox) SetFocus(index int) {
	if index < 0 || index >= len(h.children) {
		return
	}
	
	// Clear focus on current element if it's focusable
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
			focusable.SetFocused(false)
		}
	}
	
	h.focusedIndex = index
	
	// Set focus on new element if it's focusable
	if focusable, ok := h.children[h.focusedIndex].(Focusable); ok {
		focusable.SetFocused(true)
	}
}

// GetFocusedIndex returns the index of the currently focused element.
func (h *scrollableHBox) GetFocusedIndex() int {
	return h.focusedIndex
}

// GetFocusedElement returns the currently focused element, or nil if none.
func (h *scrollableHBox) GetFocusedElement() Element {
	if h.focusedIndex >= 0 && h.focusedIndex < len(h.children) {
		return h.children[h.focusedIndex]
	}
	return nil
}

// Render implements the Element interface.
func (h *scrollableHBox) Render(scr uv.Screen, area uv.Rectangle) {
	if len(h.children) == 0 || area.Dx() <= 0 {
		return
	}

	// Calculate total content width
	totalWidth := 0
	childWidths := make([]int, len(h.children))
	for i, child := range h.children {
		w, _ := child.MinSize(scr)
		if w == 0 {
			w = 1 // Default width for flexible elements
		}
		childWidths[i] = w
		totalWidth += w
	}

	// Adjust scroll offset if necessary
	maxScroll := max(0, totalWidth-area.Dx())
	if h.scrollOffset > maxScroll {
		h.scrollOffset = maxScroll
	}

	// Render visible children
	x := area.Min.X
	currentCol := 0
	for i, child := range h.children {
		childWidth := childWidths[i]

		// Skip children that are scrolled out of view
		if currentCol+childWidth <= h.scrollOffset {
			currentCol += childWidth
			continue
		}

		// Stop if we've rendered everything that fits in the area
		if x >= area.Max.X {
			break
		}

		// Calculate visible portion of this child
		skipCols := 0
		if currentCol < h.scrollOffset {
			skipCols = h.scrollOffset - currentCol
		}

		visibleWidth := min(childWidth-skipCols, area.Max.X-x)
		if visibleWidth <= 0 {
			currentCol += childWidth
			continue
		}

		childArea := uv.Rect(x, area.Min.Y, visibleWidth, area.Dy())

		// Render child (selection/focus is handled by the element itself)
		child.Render(scr, childArea)

		x += visibleWidth
		currentCol += childWidth
	}
}

// MinSize implements the Element interface.
func (h *scrollableHBox) MinSize(scr uv.Screen) (width, height int) {
	for _, child := range h.children {
		w, h := child.MinSize(scr)
		width += w
		if h > height {
			height = h
		}
	}
	return width, height
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
