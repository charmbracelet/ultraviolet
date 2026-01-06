package dom

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/clipperhouse/displaywidth"
)

// mockScreen is a simple mock implementation of uv.Screen for testing.
type mockScreen struct {
	cells map[string]*uv.Cell
}

func newMockScreen() *mockScreen {
	return &mockScreen{
		cells: make(map[string]*uv.Cell),
	}
}

// widthMethod is a simple implementation of WidthMethod using displaywidth
type widthMethod struct{}

func (w widthMethod) StringWidth(s string) int {
	return displaywidth.String(s)
}

func (m *mockScreen) WidthMethod() uv.WidthMethod {
	return widthMethod{}
}

func (m *mockScreen) Bounds() uv.Rectangle {
	return uv.Rect(0, 0, 80, 24) // Default terminal size
}

func (m *mockScreen) SetCell(x, y int, cell *uv.Cell) {
	key := string(rune(x))+ "," + string(rune(y))
	m.cells[key] = cell
}

func (m *mockScreen) CellAt(x, y int) *uv.Cell {
	key := string(rune(x)) + "," + string(rune(y))
	return m.cells[key]
}

func TestDocument_CreateElement(t *testing.T) {
	doc := NewDocument()
	elem := doc.CreateElement("div")
	
	if elem == nil {
		t.Fatal("CreateElement returned nil")
	}
	
	if elem.TagName() != "DIV" {
		t.Errorf("Expected tag name DIV, got %s", elem.TagName())
	}
	
	if elem.NodeType() != ElementNode {
		t.Errorf("Expected node type ElementNode, got %d", elem.NodeType())
	}
}

func TestDocument_CreateTextNode(t *testing.T) {
	doc := NewDocument()
	text := doc.CreateTextNode("Hello")
	
	if text == nil {
		t.Fatal("CreateTextNode returned nil")
	}
	
	if text.Data() != "Hello" {
		t.Errorf("Expected data 'Hello', got '%s'", text.Data())
	}
	
	if text.NodeType() != TextNode {
		t.Errorf("Expected node type TextNode, got %d", text.NodeType())
	}
}

func TestElement_AppendChild(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	text := doc.CreateTextNode("Hello")
	
	div.AppendChild(text)
	
	if len(div.ChildNodes()) != 1 {
		t.Errorf("Expected 1 child, got %d", len(div.ChildNodes()))
	}
	
	if div.FirstChild() != text {
		t.Error("FirstChild() did not return the appended text node")
	}
	
	if text.ParentNode() != div {
		t.Error("Text node's ParentNode() does not point to div")
	}
}

func TestElement_GetAttribute(t *testing.T) {
	doc := NewDocument()
	elem := doc.CreateElement("div")
	
	elem.SetAttribute("border", "rounded")
	
	if !elem.HasAttribute("border") {
		t.Error("HasAttribute returned false for 'border'")
	}
	
	if elem.GetAttribute("border") != "rounded" {
		t.Errorf("Expected attribute value 'rounded', got '%s'", elem.GetAttribute("border"))
	}
	
	if elem.GetAttribute("nonexistent") != "" {
		t.Error("GetAttribute should return empty string for nonexistent attribute")
	}
}

func TestElement_RemoveAttribute(t *testing.T) {
	doc := NewDocument()
	elem := doc.CreateElement("div")
	
	elem.SetAttribute("border", "rounded")
	elem.RemoveAttribute("border")
	
	if elem.HasAttribute("border") {
		t.Error("HasAttribute returned true after RemoveAttribute")
	}
	
	if elem.GetAttribute("border") != "" {
		t.Error("GetAttribute should return empty string after RemoveAttribute")
	}
}

func TestNode_RemoveChild(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	text1 := doc.CreateTextNode("First")
	text2 := doc.CreateTextNode("Second")
	
	div.AppendChild(text1)
	div.AppendChild(text2)
	
	if len(div.ChildNodes()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(div.ChildNodes()))
	}
	
	div.RemoveChild(text1)
	
	if len(div.ChildNodes()) != 1 {
		t.Errorf("Expected 1 child after removal, got %d", len(div.ChildNodes()))
	}
	
	if div.FirstChild() != text2 {
		t.Error("FirstChild() should be text2 after removing text1")
	}
	
	if text1.ParentNode() != nil {
		t.Error("Removed node should have nil ParentNode()")
	}
}

func TestElement_TextContent(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	text1 := doc.CreateTextNode("Hello ")
	text2 := doc.CreateTextNode("World")
	
	div.AppendChild(text1)
	div.AppendChild(text2)
	
	if div.TextContent() != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", div.TextContent())
	}
}

func TestElement_SetTextContent(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	text1 := doc.CreateTextNode("Old")
	div.AppendChild(text1)
	
	div.SetTextContent("New")
	
	if len(div.ChildNodes()) != 1 {
		t.Errorf("Expected 1 child after SetTextContent, got %d", len(div.ChildNodes()))
	}
	
	if div.TextContent() != "New" {
		t.Errorf("Expected 'New', got '%s'", div.TextContent())
	}
}

func TestElement_GetElementsByTagName(t *testing.T) {
	doc := NewDocument()
	root := doc.CreateElement("div")
	child1 := doc.CreateElement("span")
	child2 := doc.CreateElement("div")
	grandchild := doc.CreateElement("span")
	
	root.AppendChild(child1)
	root.AppendChild(child2)
	child2.AppendChild(grandchild)
	
	spans := root.GetElementsByTagName("span")
	
	if len(spans) != 2 {
		t.Errorf("Expected 2 span elements, got %d", len(spans))
	}
}

func TestElement_CloneNode(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	div.SetAttribute("border", "rounded")
	text := doc.CreateTextNode("Hello")
	div.AppendChild(text)
	
	// Shallow clone
	shallowClone := div.CloneNode(false).(*Element)
	if len(shallowClone.ChildNodes()) != 0 {
		t.Error("Shallow clone should have no children")
	}
	if shallowClone.GetAttribute("border") != "rounded" {
		t.Error("Shallow clone should preserve attributes")
	}
	
	// Deep clone
	deepClone := div.CloneNode(true).(*Element)
	if len(deepClone.ChildNodes()) != 1 {
		t.Error("Deep clone should have children")
	}
	if deepClone.TextContent() != "Hello" {
		t.Error("Deep clone should preserve text content")
	}
}

func TestDocument_Render(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	text := doc.CreateTextNode("Test")
	div.AppendChild(text)
	doc.AppendChild(div)
	
	screen := newMockScreen()
	area := uv.Rect(0, 0, 10, 10)
	
	// Should not panic
	doc.Render(screen, area)
}

func TestElement_RenderWithBorder(t *testing.T) {
	doc := NewDocument()
	div := doc.CreateElement("div")
	div.SetAttribute("border", "rounded")
	text := doc.CreateTextNode("Test")
	div.AppendChild(text)
	
	screen := newMockScreen()
	area := uv.Rect(0, 0, 10, 10)
	
	// Should not panic
	div.Render(screen, area)
}

func TestVBox_Layout(t *testing.T) {
	doc := NewDocument()
	vbox := doc.CreateElement("vbox")
	text1 := doc.CreateTextNode("Line 1")
	text2 := doc.CreateTextNode("Line 2")
	vbox.AppendChild(text1)
	vbox.AppendChild(text2)
	
	screen := newMockScreen()
	
	_, h := vbox.MinSize(screen)
	if h != 2 {
		t.Errorf("Expected height 2, got %d", h)
	}
}

func TestHBox_Layout(t *testing.T) {
	doc := NewDocument()
	hbox := doc.CreateElement("hbox")
	text1 := doc.CreateTextNode("A")
	text2 := doc.CreateTextNode("B")
	hbox.AppendChild(text1)
	hbox.AppendChild(text2)
	
	screen := newMockScreen()
	
	w, _ := hbox.MinSize(screen)
	if w != 2 {
		t.Errorf("Expected width 2, got %d", w)
	}
}
