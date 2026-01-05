package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestTextHardWrap(t *testing.T) {
	scr := newMockScreen(10, 5)

	// Test hard-wrap with long text
	elem := TextHardWrap("This is a very long line that should wrap")
	area := uv.Rect(0, 0, 10, 5)
	elem.Render(scr, area)

	// Verify text was rendered (cells should be set)
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell at (0, 0)")
	}
}

func TestScrollableVBox(t *testing.T) {
	scr := newMockScreen(80, 5)

	// Create a scrollable vbox with more content than fits
	vbox := ScrollableVBox(
		Text("Line 1"),
		Text("Line 2"),
		Text("Line 3"),
		Text("Line 4"),
		Text("Line 5"),
		Text("Line 6"),
		Text("Line 7"),
	)

	area := uv.Rect(0, 0, 80, 5)

	// Test initial render (should show first 5 lines)
	vbox.Render(scr, area)

	// Test scrolling down
	vbox.ScrollDown(2)
	if vbox.GetScrollOffset() != 2 {
		t.Errorf("Expected scroll offset 2, got %d", vbox.GetScrollOffset())
	}

	// Test scrolling up
	vbox.ScrollUp(1)
	if vbox.GetScrollOffset() != 1 {
		t.Errorf("Expected scroll offset 1, got %d", vbox.GetScrollOffset())
	}

	// Test scroll offset bounds
	vbox.ScrollUp(10)
	if vbox.GetScrollOffset() != 0 {
		t.Errorf("Expected scroll offset 0 (bounded), got %d", vbox.GetScrollOffset())
	}
}

func TestScrollableVBoxFocus(t *testing.T) {
	_ = newMockScreen(80, 10)

	vbox := ScrollableVBox(
		Text("Item 1"),
		Text("Item 2"),
		Text("Item 3"),
	)

	// Test initial focus
	if vbox.GetFocusedIndex() != -1 {
		t.Errorf("Expected initial focus -1, got %d", vbox.GetFocusedIndex())
	}

	// Test setting focus
	vbox.SetFocus(1)
	if vbox.GetFocusedIndex() != 1 {
		t.Errorf("Expected focus 1, got %d", vbox.GetFocusedIndex())
	}

	// Test focus next
	vbox.FocusNext()
	if vbox.GetFocusedIndex() != 2 {
		t.Errorf("Expected focus 2, got %d", vbox.GetFocusedIndex())
	}

	// Test focus wrapping
	vbox.FocusNext()
	if vbox.GetFocusedIndex() != 0 {
		t.Errorf("Expected focus 0 (wrapped), got %d", vbox.GetFocusedIndex())
	}

	// Test focus previous
	vbox.FocusPrevious()
	if vbox.GetFocusedIndex() != 2 {
		t.Errorf("Expected focus 2, got %d", vbox.GetFocusedIndex())
	}
}

func TestScrollableVBoxSelection(t *testing.T) {
	// Selection is now at the element level with character positions
	elem1 := MakeSelectable(Text("Item 1"))
	elem2 := MakeSelectable(Text("Item 2"))
	elem3 := MakeSelectable(Text("Item 3"))

	vbox := ScrollableVBox(elem1, elem2, elem3)

	// Test element selection with character ranges
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 4}
	elem1.SetSelection(sel)
	got, hasSelection := elem1.GetSelection()
	if !hasSelection {
		t.Error("Expected element to have selection")
	}
	if got.StartCol != 0 || got.EndCol != 4 {
		t.Errorf("Expected selection (0,0)-(0,4), got (%d,%d)-(%d,%d)",
			got.StartLine, got.StartCol, got.EndLine, got.EndCol)
	}

	// Clear selection
	elem1.ClearSelection()
	_, hasSelection = elem1.GetSelection()
	if hasSelection {
		t.Error("Expected no selection after clear")
	}

	// Test rendering with selection
	scr := newMockScreen(80, 10)
	area := uv.Rect(0, 0, 80, 10)
	sel2 := SelectionRange{StartLine: 0, StartCol: 2, EndLine: 0, EndCol: 5}
	elem2.SetSelection(sel2)
	vbox.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cells to be rendered")
	}
}

func TestScrollableHBox(t *testing.T) {
	scr := newMockScreen(10, 24)

	// Create a scrollable hbox with more content than fits
	hbox := ScrollableHBox(
		Text("Col1"),
		Text("Col2"),
		Text("Col3"),
		Text("Col4"),
		Text("Col5"),
	)

	area := uv.Rect(0, 0, 10, 24)

	// Test initial render
	hbox.Render(scr, area)

	// Test scrolling right
	hbox.ScrollRight(5)
	if hbox.GetScrollOffset() != 5 {
		t.Errorf("Expected scroll offset 5, got %d", hbox.GetScrollOffset())
	}

	// Test scrolling left
	hbox.ScrollLeft(2)
	if hbox.GetScrollOffset() != 3 {
		t.Errorf("Expected scroll offset 3, got %d", hbox.GetScrollOffset())
	}

	// Test scroll offset bounds
	hbox.ScrollLeft(10)
	if hbox.GetScrollOffset() != 0 {
		t.Errorf("Expected scroll offset 0 (bounded), got %d", hbox.GetScrollOffset())
	}
}

func TestScrollableHBoxFocus(t *testing.T) {
	hbox := ScrollableHBox(
		Text("Item 1"),
		Text("Item 2"),
		Text("Item 3"),
	)

	// Test focus methods work same as vbox
	hbox.SetFocus(1)
	if hbox.GetFocusedIndex() != 1 {
		t.Errorf("Expected focus 1, got %d", hbox.GetFocusedIndex())
	}

	hbox.FocusNext()
	if hbox.GetFocusedIndex() != 2 {
		t.Errorf("Expected focus 2, got %d", hbox.GetFocusedIndex())
	}
}

func TestScrollableHBoxSelection(t *testing.T) {
	// Selection is now at the element level with character positions
	elem1 := MakeSelectable(Text("Item 1"))
	elem2 := MakeSelectable(Text("Item 2"))
	elem3 := MakeSelectable(Text("Item 3"))

	hbox := ScrollableHBox(elem1, elem2, elem3)

	// Test element selection with character range
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 6}
	elem2.SetSelection(sel)
	got, hasSelection := elem2.GetSelection()
	if !hasSelection {
		t.Error("Expected element to have selection")
	}
	if got.EndCol != 6 {
		t.Errorf("Expected selection end at col 6, got %d", got.EndCol)
	}

	// Test rendering with selection
	scr := newMockScreen(20, 24)
	area := uv.Rect(0, 0, 20, 24)
	hbox.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cells to be rendered")
	}
}

func TestScrollableVBoxRender(t *testing.T) {
	scr := newMockScreen(80, 5)

	elem1 := MakeFocusable(Text("Line 1"))
	elem2 := MakeSelectableAndFocusable(Text("Line 2"))
	_ = MakeSelectable(Text("Line 3"))

	vbox := ScrollableVBox(
		elem1,
		elem2,
		Text("Line 4"),
		Text("Line 5"),
		Text("Line 6"),
	)

	// Set focus and selection at element level with character positions
	vbox.SetFocus(1)
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 5}
	elem2.SetSelection(sel)

	area := uv.Rect(0, 0, 80, 5)
	vbox.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cells to be rendered")
	}
}

func TestScrollableHBoxRender(t *testing.T) {
	scr := newMockScreen(20, 24)

	elem1 := MakeFocusable(Text("A"))
	elem2 := MakeSelectableAndFocusable(Text("B"))
	elem3 := MakeSelectable(Text("C"))

	hbox := ScrollableHBox(
		elem1,
		elem2,
		elem3,
		Text("D"),
		Text("E"),
	)

	// Set focus and selection at element level with character positions
	hbox.SetFocus(1)
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 1}
	elem2.SetSelection(sel)

	area := uv.Rect(0, 0, 20, 24)
	hbox.Render(scr, area)

	// Verify something was rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cells to be rendered")
	}
}
