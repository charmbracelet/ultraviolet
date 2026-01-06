package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestBox(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Test basic box with content
	box := NewBox(Text("Hello"))
	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Check content is rendered
	if scr.cells[0][0].Content != "H" {
		t.Errorf("Expected 'H' at (0, 0), got %q", scr.cells[0][0].Content)
	}
}

func TestBoxWithBorder(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Test box with border
	box := NewBox(Text("Content")).WithBorder(BorderStyleNormal())
	area := uv.Rect(0, 0, 15, 5)
	box.Render(scr, area)

	// Check border corners
	if scr.cells[0][0].Content != "┌" {
		t.Errorf("Expected top-left corner at (0, 0), got %q", scr.cells[0][0].Content)
	}
	if scr.cells[0][14].Content != "┐" {
		t.Errorf("Expected top-right corner at (14, 0), got %q", scr.cells[0][14].Content)
	}
	if scr.cells[4][0].Content != "└" {
		t.Errorf("Expected bottom-left corner at (0, 4), got %q", scr.cells[4][0].Content)
	}
	if scr.cells[4][14].Content != "┘" {
		t.Errorf("Expected bottom-right corner at (14, 4), got %q", scr.cells[4][14].Content)
	}

	// Check min size includes border
	w, h := box.MinSize(scr)
	if w != 9 { // "Content" (7) + 2 border
		t.Errorf("Expected width 9, got %d", w)
	}
	if h != 3 { // 1 line + 2 border
		t.Errorf("Expected height 3, got %d", h)
	}
}

func TestBoxWithPadding(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Test box with padding
	box := NewBox(Text("X")).WithPadding(1)
	area := uv.Rect(0, 0, 10, 10)
	box.Render(scr, area)

	// Content should be at (1, 1) due to padding
	if scr.cells[1][1].Content != "X" {
		t.Errorf("Expected 'X' at (1, 1), got %q", scr.cells[1][1].Content)
	}

	// Check min size includes padding
	w, h := box.MinSize(scr)
	if w != 3 { // 1 char + 2 padding
		t.Errorf("Expected width 3, got %d", w)
	}
	if h != 3 { // 1 line + 2 padding
		t.Errorf("Expected height 3, got %d", h)
	}
}

func TestBoxWithBorderAndPadding(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Test box with both border and padding
	box := NewBox(Text("Test")).
		WithBorder(BorderStyleRounded()).
		WithPadding(1)
	area := uv.Rect(0, 0, 20, 5)
	box.Render(scr, area)

	// Check rounded border corners
	if scr.cells[0][0].Content != "╭" {
		t.Errorf("Expected rounded top-left corner, got %q", scr.cells[0][0].Content)
	}

	// Content should be at (2, 2) due to border (1) + padding (1)
	if scr.cells[2][2].Content != "T" {
		t.Errorf("Expected 'T' at (2, 2), got %q", scr.cells[2][2].Content)
	}

	// Check min size includes both border and padding
	w, h := box.MinSize(scr)
	if w != 8 { // 4 chars + 2 padding + 2 border
		t.Errorf("Expected width 8, got %d", w)
	}
	if h != 5 { // 1 line + 2 padding + 2 border
		t.Errorf("Expected height 5, got %d", h)
	}
}

func TestBoxScrolling(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Create a box with content that's larger than the viewport
	content := VBox(
		Text("Line 1"),
		Text("Line 2"),
		Text("Line 3"),
		Text("Line 4"),
		Text("Line 5"),
	)

	box := NewBox(content)
	area := uv.Rect(0, 0, 10, 3) // Only 3 lines visible

	// Initial render - should show first 3 lines
	box.Render(scr, area)
	if scr.cells[0][0].Content != "L" {
		t.Errorf("Expected 'L' from Line 1 at (0, 0), got %q", scr.cells[0][0].Content)
	}

	// Scroll down
	box.ScrollDown(2)
	scr = newMockScreen(80, 24) // Clear screen
	box.Render(scr, area)

	// After scrolling, viewport should be shifted (but content still renders at same logical position)
	// The scrolling affects the offset area calculation
}

func TestBoxFocus(t *testing.T) {
	scr := newMockScreen(80, 24)

	box := NewBox(Text("Focused")).
		WithBorder(BorderStyleNormal()).
		WithFocus(true)

	area := uv.Rect(0, 0, 15, 5)
	box.Render(scr, area)

	// Focus should apply reverse attribute to border
	// The border cells should have the reverse attribute set
	if scr.cells[0][0].Style.Attrs&uv.AttrReverse == 0 {
		t.Error("Expected border to have reverse attribute when focused")
	}
}

func TestBoxSelection(t *testing.T) {
	scr := newMockScreen(80, 24)

	box := NewBox(Text("Selected")).WithSelection(true)
	area := uv.Rect(0, 0, 10, 3)
	box.Render(scr, area)

	// Selection should apply reverse attribute to all cells
	if scr.cells[0][0].Style.Attrs&uv.AttrReverse == 0 {
		t.Error("Expected cells to have reverse attribute when selected")
	}
}

func TestBorderStyles(t *testing.T) {
	tests := []struct {
		name  string
		style *BorderStyle
		want  string // top-left corner
	}{
		{"Normal", BorderStyleNormal(), "┌"},
		{"Rounded", BorderStyleRounded(), "╭"},
		{"Double", BorderStyleDouble(), "╔"},
		{"Thick", BorderStyleThick(), "┏"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scr := newMockScreen(80, 24)
			box := NewBox(Text("Test")).WithBorder(tt.style)
			area := uv.Rect(0, 0, 10, 5)
			box.Render(scr, area)

			if scr.cells[0][0].Content != tt.want {
				t.Errorf("Expected %q corner, got %q", tt.want, scr.cells[0][0].Content)
			}
		})
	}
}
