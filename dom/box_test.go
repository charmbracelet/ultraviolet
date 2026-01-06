package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestBox(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Test")
	box := StyledBox(text)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}

	// Test min size
	w, h := box.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestBoxWithBorder(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Bordered")
	box := StyledBox(text).WithBorder(NormalBorder())

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}

	// Min size should include border (2x2)
	w, h := box.MinSize(scr)
	if w < 2 || h < 2 {
		t.Errorf("Expected min size to include border, got (%d, %d)", w, h)
	}
}

func TestBoxWithPadding(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Padded")
	box := StyledBox(text).WithPadding(2)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Content should be rendered at (2, 2) due to padding
	if scr.CellAt(2, 2) == nil {
		t.Error("Expected cell to be rendered at content position")
	}

	// Min size should include padding (4x4 total)
	w, h := box.MinSize(scr)
	if w < 4 || h < 4 {
		t.Errorf("Expected min size to include padding, got (%d, %d)", w, h)
	}
}

func TestBoxWithMargin(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Margin")
	box := StyledBox(text).WithMargin(1)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Min size should include margin (2x2 total)
	w, h := box.MinSize(scr)
	if w < 2 || h < 2 {
		t.Errorf("Expected min size to include margin, got (%d, %d)", w, h)
	}
}

func TestBoxWithFocus(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Focusable")
	box := StyledBox(text).WithFocusable(true)

	// Test initial state
	if box.IsFocused() {
		t.Error("Expected box to be unfocused initially")
	}

	// Set focused
	box.SetFocused(true)
	if !box.IsFocused() {
		t.Error("Expected box to be focused")
	}

	// Render with focus
	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestBoxWithSelection(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Selectable")
	box := StyledBox(text).WithSelectable(true)

	// Test initial state
	_, hasSelection := box.GetSelection()
	if hasSelection {
		t.Error("Expected no selection initially")
	}

	// Set selection
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 5}
	box.SetSelection(sel)
	got, hasSelection := box.GetSelection()
	if !hasSelection {
		t.Error("Expected selection after SetSelection")
	}
	if got.StartCol != 0 || got.EndCol != 5 {
		t.Errorf("Expected selection (0,0)-(0,5), got (%d,%d)-(%d,%d)",
			got.StartLine, got.StartCol, got.EndLine, got.EndCol)
	}

	// Render with selection
	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Clear selection
	box.ClearSelection()
	_, hasSelection = box.GetSelection()
	if hasSelection {
		t.Error("Expected no selection after ClearSelection")
	}
}

func TestBoxWithFixedSize(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Fixed")
	box := StyledBox(text).WithSize(10, 3)

	w, h := box.MinSize(scr)
	if w != 10 || h != 3 {
		t.Errorf("Expected fixed size (10, 3), got (%d, %d)", w, h)
	}
}

func TestBoxWithMinMaxSize(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("X")
	box := StyledBox(text).WithMinSize(5, 3).WithMaxSize(20, 10)

	w, h := box.MinSize(scr)
	if w < 5 || w > 20 {
		t.Errorf("Expected width between 5 and 20, got %d", w)
	}
	if h < 3 || h > 10 {
		t.Errorf("Expected height between 3 and 10, got %d", h)
	}
}

func TestBoxWithBackground(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Background")
	bg := &uv.Cell{
		Content: " ",
		Width:   1,
		Style:   uv.Style{Bg: ansi.Blue},
	}
	box := StyledBox(text).WithBackground(bg)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestBoxChaining(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Chained")
	
	// Test method chaining
	box := StyledBox(text).
		WithBorder(RoundedBorder()).
		WithPadding(1).
		WithMargin(1).
		WithFocusable(true).
		WithSelectable(true).
		WithSize(20, 5)

	box.SetFocused(true)
	box.SetSelection(SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 3})

	area := uv.Rect(0, 0, 30, 10)
	box.Render(scr, area)

	// Content should be at margin(1) + border(1) + padding(1) = (3, 3)
	if scr.CellAt(3, 3) == nil {
		t.Error("Expected cell to be rendered at content position")
	}

	// Verify size
	w, h := box.MinSize(scr)
	if w != 20 || h != 5 {
		t.Errorf("Expected size (20, 5), got (%d, %d)", w, h)
	}
}

func TestBoxComplexLayout(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Create a complex nested structure using Box
	inner := StyledBox(Text("Inner")).
		WithBorder(NormalBorder()).
		WithPadding(1)

	outer := StyledBox(inner).
		WithBorder(RoundedBorder()).
		WithPadding(2).
		WithMargin(1)

	area := uv.Rect(0, 0, 40, 15)
	outer.Render(scr, area)

	// Outer margin(1) + border(1) should have content at least at (2, 2)
	cellFound := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if scr.CellAt(x, y) != nil {
				cellFound = true
				break
			}
		}
		if cellFound {
			break
		}
	}
	if !cellFound {
		t.Error("Expected some cells to be rendered")
	}

	// Min size should include all layers
	w, h := outer.MinSize(scr)
	if w < 10 || h < 6 {
		t.Errorf("Expected min size to include all layers, got (%d, %d)", w, h)
	}
}

func TestBoxWithStyle(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Styled")
	style := uv.Style{
		Fg:    ansi.Red,
		Attrs: uv.AttrBold,
	}
	box := StyledBox(text).WithStyle(style)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestBoxGetSetStyle(t *testing.T) {
	text := Text("Test")
	box := StyledBox(text)

	// Get initial style
	style := box.GetStyle()
	if style.Wrap {
		t.Error("Expected wrap to be false by default")
	}

	// Modify and set style
	style.Wrap = true
	style.Truncate = false
	box.SetStyle(style)

	// Verify changes
	newStyle := box.GetStyle()
	if !newStyle.Wrap {
		t.Error("Expected wrap to be true")
	}
	if newStyle.Truncate {
		t.Error("Expected truncate to be false")
	}
}

func TestInlineBox(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Inline")
	box := InlineBox(text).
		WithPadding(2).
		WithMargin(2)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Min size for inline box should ignore vertical padding and margins
	w, h := box.MinSize(scr)
	
	// Width should include horizontal padding/margins: 2(left pad) + 2(right pad) + 2(left margin) + 2(right margin) = 8
	// But actual text width is also included
	if w < 8 {
		t.Errorf("Expected width to include horizontal padding/margins, got %d", w)
	}
	
	// Height should NOT include vertical padding/margins for inline
	// It should just be the text height
	if h >= 8 {
		t.Errorf("Expected height to ignore vertical padding/margins for inline, got %d", h)
	}
}

func TestInlineBlockBox(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("InlineBlock")
	box := InlineBlockBox(text).
		WithPadding(2).
		WithMargin(2)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Min size for inline-block box should include all padding and margins
	w, h := box.MinSize(scr)
	
	// Should include both horizontal and vertical spacing
	if w < 8 || h < 8 {
		t.Errorf("Expected size to include all padding/margins for inline-block, got (%d, %d)", w, h)
	}
}

func TestBlockBox(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Block")
	box := BlockBox(text).
		WithPadding(2).
		WithMargin(2)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Min size for block box should include all padding and margins
	w, h := box.MinSize(scr)
	
	// Should include both horizontal and vertical spacing
	if w < 8 || h < 8 {
		t.Errorf("Expected size to include all padding/margins for block, got (%d, %d)", w, h)
	}
}

func TestBoxDisplayModes(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Test")
	
	// Test changing display mode
	box := BlockBox(text).WithDisplay(DisplayInline)
	
	style := box.GetStyle()
	if style.Display != DisplayInline {
		t.Errorf("Expected display to be Inline, got %v", style.Display)
	}

	// Render to ensure it doesn't panic
	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)
}

func TestInlineBoxWithBorder(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Bordered")
	box := InlineBox(text).WithBorder(NormalBorder())

	// Min size should include border width but not height for inline
	w, h := box.MinSize(scr)
	
	// Width includes border (2)
	if w < 2 {
		t.Errorf("Expected width to include border, got %d", w)
	}
	
	// Height should NOT include border for inline display
	if h >= 4 {
		t.Errorf("Expected height to ignore border for inline, got %d", h)
	}

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)
}

func TestBorderWithSides(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Partial Border")
	
	// Create border with only top and bottom
	border := NormalBorder().WithSides(true, false, true, false)
	box := StyledBox(text).WithBorder(border)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Check that top border is rendered (should be at y=0, x > 0)
	if scr.CellAt(1, 0) == nil {
		t.Error("Expected top border cell to be rendered")
	}
	
	// Check that bottom border is rendered
	if scr.CellAt(1, 4) == nil {
		t.Error("Expected bottom border cell to be rendered")
	}
}

func TestBorderWithGradient(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Gradient Border")
	
	// Create border with horizontal gradient
	border := RoundedBorder().WithGradient(ansi.Red, ansi.Blue)
	box := StyledBox(text).WithBorder(border)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestBorderWithVerticalGradient(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Vertical Gradient")
	
	// Create border with vertical gradient
	border := DoubleBorder().WithVerticalGradient(ansi.Green, ansi.Yellow)
	box := StyledBox(text).WithBorder(border)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestBorderWithSideStyles(t *testing.T) {
	scr := newMockScreen(80, 24)
	text := Text("Styled Sides")
	
	// Create border with different colors per side
	border := NormalBorder().
		TopForeground(ansi.Red).
		BottomForeground(ansi.Blue).
		LeftForeground(ansi.Green).
		RightForeground(ansi.Yellow)
	box := StyledBox(text).WithBorder(border)

	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}
