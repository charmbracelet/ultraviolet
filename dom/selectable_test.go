package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestSelectionRange(t *testing.T) {
	// Test normalization
	sel := SelectionRange{StartLine: 2, StartCol: 5, EndLine: 1, EndCol: 3}
	norm := sel.Normalize()
	if norm.StartLine != 1 || norm.StartCol != 3 || norm.EndLine != 2 || norm.EndCol != 5 {
		t.Errorf("Expected normalized (1,3)-(2,5), got (%d,%d)-(%d,%d)",
			norm.StartLine, norm.StartCol, norm.EndLine, norm.EndCol)
	}

	// Test IsEmpty
	empty := SelectionRange{StartLine: 1, StartCol: 1, EndLine: 1, EndCol: 1}
	if !empty.IsEmpty() {
		t.Error("Expected empty selection")
	}

	notEmpty := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 5}
	if notEmpty.IsEmpty() {
		t.Error("Expected non-empty selection")
	}

	// Test Contains
	sel = SelectionRange{StartLine: 1, StartCol: 2, EndLine: 3, EndCol: 4}
	if !sel.Contains(2, 0) {
		t.Error("Expected (2,0) to be in selection")
	}
	if !sel.Contains(1, 2) {
		t.Error("Expected (1,2) to be in selection")
	}
	if sel.Contains(1, 1) {
		t.Error("Expected (1,1) to not be in selection")
	}
	if sel.Contains(3, 4) {
		t.Error("Expected (3,4) to not be in selection (end is exclusive)")
	}
}

func TestSelectable(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	elem := MakeSelectable(text)

	// Test initial state
	_, hasSelection := elem.GetSelection()
	if hasSelection {
		t.Error("Expected no selection initially")
	}

	// Test setting selection
	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 4}
	elem.SetSelection(sel)
	got, hasSelection := elem.GetSelection()
	if !hasSelection {
		t.Error("Expected selection after SetSelection")
	}
	if got.StartLine != 0 || got.StartCol != 0 || got.EndLine != 0 || got.EndCol != 4 {
		t.Errorf("Expected selection (0,0)-(0,4), got (%d,%d)-(%d,%d)",
			got.StartLine, got.StartCol, got.EndLine, got.EndCol)
	}

	// Test rendering with selection
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}

	// Test clearing selection
	elem.ClearSelection()
	_, hasSelection = elem.GetSelection()
	if hasSelection {
		t.Error("Expected no selection after ClearSelection")
	}

	// Test min size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestFocusable(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	elem := MakeFocusable(text)

	// Test initial state
	if elem.IsFocused() {
		t.Error("Expected element to be unfocused initially")
	}

	// Test setting focused
	elem.SetFocused(true)
	if !elem.IsFocused() {
		t.Error("Expected element to be focused")
	}

	// Test rendering
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}

	// Test min size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestSelectableAndFocusable(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	elem := MakeSelectableAndFocusable(text)

	// Test initial state
	_, hasSelection := elem.GetSelection()
	if hasSelection {
		t.Error("Expected no selection initially")
	}
	if elem.IsFocused() {
		t.Error("Expected element to be unfocused initially")
	}

	// Test setting selection
	sel := SelectionRange{StartLine: 0, StartCol: 1, EndLine: 0, EndCol: 3}
	elem.SetSelection(sel)
	got, hasSelection := elem.GetSelection()
	if !hasSelection {
		t.Error("Expected selection after SetSelection")
	}
	if got.StartLine != 0 || got.StartCol != 1 {
		t.Errorf("Expected selection to start at (0,1), got (%d,%d)", got.StartLine, got.StartCol)
	}

	// Test setting focused
	elem.SetFocused(true)
	if !elem.IsFocused() {
		t.Error("Expected element to be focused")
	}

	// Test rendering with both
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}

	// Test min size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestSelectableStyled(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	style := uv.Style{Attrs: uv.AttrItalic}
	elem := MakeSelectableStyled(text, style)

	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 2}
	elem.SetSelection(sel)
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestFocusableStyled(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	style := uv.Style{Attrs: uv.AttrItalic}
	elem := MakeFocusableStyled(text, style)

	elem.SetFocused(true)
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestSelectableAndFocusableStyled(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Test")
	selectStyle := uv.Style{Attrs: uv.AttrBold}
	focusStyle := uv.Style{Attrs: uv.AttrReverse}
	elem := MakeSelectableAndFocusableStyled(text, selectStyle, focusStyle)

	sel := SelectionRange{StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 3}
	elem.SetSelection(sel)
	elem.SetFocused(true)
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestSelectionStyle(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Hello")
	elem := MakeSelectable(text)

	// Test setting custom selection style
	customStyle := uv.Style{Attrs: uv.AttrBold}
	elem.SetSelectionStyle(customStyle)

	sel := SelectionRange{StartLine: 0, StartCol: 1, EndLine: 0, EndCol: 4}
	elem.SetSelection(sel)
	area := uv.Rect(0, 0, 20, 1)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}

func TestMultilineSelection(t *testing.T) {
	scr := newMockScreen(80, 10)
	text := Text("Line1\nLine2\nLine3")
	elem := MakeSelectable(text)

	// Select from middle of line 0 to middle of line 2
	sel := SelectionRange{StartLine: 0, StartCol: 2, EndLine: 2, EndCol: 3}
	elem.SetSelection(sel)
	area := uv.Rect(0, 0, 20, 5)
	elem.Render(scr, area)

	// Verify cells were rendered
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell to be rendered")
	}
}
