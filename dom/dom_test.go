package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/clipperhouse/displaywidth"
)

// widthMethod implements uv.WidthMethod using displaywidth.
type widthMethod struct{}

func (w widthMethod) StringWidth(s string) int {
	return displaywidth.String(s)
}

// mockScreen is a simple mock implementation of uv.Screen for testing.
type mockScreen struct {
	width  int
	height int
	cells  map[int]map[int]*uv.Cell
}

func newMockScreen(width, height int) *mockScreen {
	return &mockScreen{
		width:  width,
		height: height,
		cells:  make(map[int]map[int]*uv.Cell),
	}
}

func (m *mockScreen) Bounds() uv.Rectangle {
	return uv.Rect(0, 0, m.width, m.height)
}

func (m *mockScreen) CellAt(x, y int) *uv.Cell {
	if row, ok := m.cells[y]; ok {
		return row[x]
	}
	return nil
}

func (m *mockScreen) SetCell(x, y int, c *uv.Cell) {
	if m.cells[y] == nil {
		m.cells[y] = make(map[int]*uv.Cell)
	}
	m.cells[y][x] = c
}

func (m *mockScreen) WidthMethod() uv.WidthMethod {
	return widthMethod{}
}

func TestTextElement(t *testing.T) {
	scr := newMockScreen(80, 24)
	elem := Text("Hello")

	area := uv.Rect(0, 0, 10, 1)
	elem.Render(scr, area)

	// Check that cells were set
	if scr.CellAt(0, 0) == nil {
		t.Error("Expected cell at (0, 0)")
	}

	// Check minimum size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestVBoxLayout(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := VBox(
		Text("Line 1"),
		Text("Line 2"),
		Text("Line 3"),
	)

	area := uv.Rect(0, 0, 20, 10)
	elem.Render(scr, area)

	// Verify min size calculation
	_, h := elem.MinSize(scr)
	if h != 3 {
		t.Errorf("Expected height 3 for 3 text elements, got %d", h)
	}
}

func TestHBoxLayout(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := HBox(
		Text("A"),
		Text("B"),
		Text("C"),
	)

	area := uv.Rect(0, 0, 20, 5)
	elem.Render(scr, area)

	// Verify min size calculation
	w, _ := elem.MinSize(scr)
	if w <= 0 {
		t.Errorf("Expected positive width for 3 text elements, got %d", w)
	}
}

func TestBorder(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := NewBox(Text("Bordered")).WithBorder(BorderStyleNormal())

	area := uv.Rect(0, 0, 20, 5)
	elem.Render(scr, area)

	// Check min size includes border
	w, h := elem.MinSize(scr)
	if w < 2 || h < 2 {
		t.Errorf("Expected border to add at least 2 to dimensions, got (%d, %d)", w, h)
	}
}

func TestPadding(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := PaddingAll(Text("Padded"), 1)

	area := uv.Rect(0, 0, 20, 5)
	elem.Render(scr, area)

	// Check min size includes padding
	w, h := elem.MinSize(scr)
	if w < 2 || h < 2 {
		t.Errorf("Expected padding to add at least 2 to dimensions, got (%d, %d)", w, h)
	}
}

func TestSeparator(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Test horizontal separator
	elem := Separator()
	area := uv.Rect(0, 0, 10, 1)
	elem.Render(scr, area)

	_, h := elem.MinSize(scr)
	if h != 1 {
		t.Errorf("Expected horizontal separator height of 1, got %d", h)
	}

	// Test vertical separator
	elem2 := SeparatorVertical()
	w2, h2 := elem2.MinSize(scr)
	if w2 != 1 {
		t.Errorf("Expected vertical separator width of 1, got %d", w2)
	}
	if h2 != 0 {
		t.Errorf("Expected vertical separator height of 0 (flexible), got %d", h2)
	}
}

func TestParagraph(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := Paragraph("This is a test paragraph that should wrap.")
	area := uv.Rect(0, 0, 20, 10)
	elem.Render(scr, area)

	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size for paragraph, got (%d, %d)", w, h)
	}
}

func TestCenter(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := Center(Text("Centered"))
	area := uv.Rect(0, 0, 40, 10)
	elem.Render(scr, area)

	// Center should report child's min size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size, got (%d, %d)", w, h)
	}
}

func TestFlex(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := Flex()

	// Flex should have zero min size (flexible)
	w, h := elem.MinSize(scr)
	if w != 0 || h != 0 {
		t.Errorf("Expected flex to have zero min size, got (%d, %d)", w, h)
	}
}

func TestSpacer(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := Spacer(5, 3)

	w, h := elem.MinSize(scr)
	if w != 5 || h != 3 {
		t.Errorf("Expected spacer size (5, 3), got (%d, %d)", w, h)
	}
}

func TestEmpty(t *testing.T) {
	scr := newMockScreen(80, 24)

	elem := Empty()
	area := uv.Rect(0, 0, 10, 10)

	// Should not panic
	elem.Render(scr, area)

	w, h := elem.MinSize(scr)
	if w != 0 || h != 0 {
		t.Errorf("Expected empty to have zero min size, got (%d, %d)", w, h)
	}
}

func TestComplexLayout(t *testing.T) {
	scr := newMockScreen(80, 24)

	// Create a complex nested layout
	elem := VBox(
		NewBox(Text("Header")).WithBorder(BorderStyleNormal()),
		HBox(
			VBox(
				Text("Left"),
				Separator(),
				Text("Content"),
			),
			SeparatorVertical(),
			VBox(
				Text("Right"),
				Spacer(0, 1),
				Text("More content"),
			),
		),
		Separator(),
		PaddingAll(Text("Footer"), 1),
	)

	area := uv.Rect(0, 0, 60, 20)

	// Should render without panicking
	elem.Render(scr, area)

	// Should calculate min size
	w, h := elem.MinSize(scr)
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive min size for complex layout, got (%d, %d)", w, h)
	}
}
