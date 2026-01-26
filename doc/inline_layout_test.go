package doc

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/net/html"
)

func TestInlineLayout(t *testing.T) {
	htmlStr := `<p><span>First</span><span>Second</span></p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Get the p element
	p := d.QuerySelector("p")
	if p == nil {
		t.Fatal("p element not found")
	}

	pNode := p.(*node)

	// Get the span elements
	spans := d.QuerySelectorAll("span")
	if len(spans) != 2 {
		t.Fatalf("Expected 2 span elements, got %d", len(spans))
	}

	// Create renderer and compute styles
	r := NewRenderer(pNode)
	r.computeStyles(pNode)

	// Layout the p element
	r.layout(pNode, uv.Rect(0, 0, 80, 25))

	// Verify the layout height is 1 (inline children should be on same line)
	if pNode.layout.Rect.Dy() != 1 {
		t.Errorf("Expected p element height to be 1, got %d", pNode.layout.Rect.Dy())
	}

	// Find the text node with wrapped lines (should be first text node)
	span1 := spans[0].(*node)
	if len(span1.Children()) == 0 {
		t.Fatal("Span has no children")
	}

	textNode1 := span1.Children()[0].(*node)

	// Check that wrapped lines were created
	if len(textNode1.wrappedLines) == 0 {
		t.Error("No wrapped lines created for inline text")
	}

	// Verify all text is on one line
	if len(textNode1.wrappedLines) != 1 {
		t.Errorf("Expected 1 wrapped line, got %d", len(textNode1.wrappedLines))
	}

	// Count total cells and verify content
	var content string
	for _, line := range textNode1.wrappedLines {
		for _, cell := range line {
			content += cell.Content
		}
	}

	// Should have "FirstSecond" (no space since HTML has no space)
	if content != "FirstSecond" {
		t.Errorf("Expected 'FirstSecond', got %q", content)
	}
}

func TestInlineLayoutWithStyling(t *testing.T) {
	htmlStr := `<p><span style="color: red">Red</span><span style="color: blue">Blue</span></p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	p := d.QuerySelector("p")
	if p == nil {
		t.Fatal("p element not found")
	}

	pNode := p.(*node)
	r := NewRenderer(pNode)
	r.computeStyles(pNode)
	r.layout(pNode, uv.Rect(0, 0, 80, 25))

	// Should be 1 line
	if pNode.layout.Rect.Dy() != 1 {
		t.Errorf("Expected height 1 for inline elements, got %d", pNode.layout.Rect.Dy())
	}

	// Get the first text node
	span1 := d.QuerySelector("span").(*node)
	textNode := span1.Children()[0].(*node)

	// Check that we have cells with content
	if len(textNode.wrappedLines) == 0 {
		t.Fatal("No wrapped lines")
	}

	line := textNode.wrappedLines[0]
	if len(line) == 0 {
		t.Fatal("No cells in wrapped line")
	}

	// Verify we have content (should have "RedBlue")
	var content string
	for _, cell := range line {
		content += cell.Content
	}

	if len(content) < 7 {
		t.Errorf("Expected at least 'RedBlue' (7+ chars), got %q", content)
	}
}

func TestInlineLayoutWrapping(t *testing.T) {
	// Test that inline elements wrap when they exceed width
	htmlStr := `<p><span>This is a very long text that should wrap</span> <span>when it exceeds the available width</span></p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	p := d.QuerySelector("p")
	if p == nil {
		t.Fatal("p element not found")
	}

	pNode := p.(*node)
	// Create renderer with small width
	r := NewRenderer(pNode)
	r.computeStyles(pNode)
	r.layout(pNode, uv.Rect(0, 0, 20, 25))

	// Should be multiple lines due to wrapping
	if pNode.layout.Rect.Dy() <= 1 {
		t.Errorf("Expected multiple lines due to wrapping, got height %d", pNode.layout.Rect.Dy())
	}

	// Get text node and verify wrapped lines
	span1 := d.QuerySelector("span").(*node)
	textNode := span1.Children()[0].(*node)

	if len(textNode.wrappedLines) <= 1 {
		t.Errorf("Expected multiple wrapped lines, got %d", len(textNode.wrappedLines))
	}
}

func TestMixedBlockAndInline(t *testing.T) {
	// Test that block elements between inline elements cause line breaks
	htmlStr := `<div><span>Inline1</span><div>Block</div><span>Inline2</span></div>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	div := d.QuerySelector("div")
	if div == nil {
		t.Fatal("div element not found")
	}

	divNode := div.(*node)
	r := NewRenderer(divNode)
	r.computeStyles(divNode)
	r.layout(divNode, uv.Rect(0, 0, 80, 25))

	// Should be at least 3 lines: inline1, block, inline2
	if divNode.layout.Rect.Dy() < 3 {
		t.Errorf("Expected at least 3 lines (inline, block, inline), got height %d", divNode.layout.Rect.Dy())
	}
}

func TestInlineLayoutWithWhitespace(t *testing.T) {
	// HTML with whitespace between spans  
	htmlStr := `<p>
		<span>Red</span>
		<span>Blue</span>
	</p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	p := d.QuerySelector("p")
	if p == nil {
		t.Fatal("p element not found")
	}

	pNode := p.(*node)
	r := NewRenderer(pNode)
	r.computeStyles(pNode)
	r.layout(pNode, uv.Rect(0, 0, 80, 25))

	// Should be 1 line (whitespace-only text nodes skipped)
	if pNode.layout.Rect.Dy() != 1 {
		t.Errorf("Expected height 1 for inline elements (whitespace ignored), got %d", pNode.layout.Rect.Dy())
	}
}

func TestTabSizeInlineStyleRenders(t *testing.T) {
	htmlStr := `<pre style="tab-size: 2">A	B</pre>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	pre := d.QuerySelector("pre")
	if pre == nil {
		t.Fatal("pre element not found")
	}

	preNode := pre.(*node)
	r := NewRenderer(preNode)
	r.computeStyles(preNode)

	// Check pre element has tab-size 2
	if preNode.computedStyle.TabSize != 2 {
		t.Errorf("Pre element expected tab-size 2, got %d", preNode.computedStyle.TabSize)
	}

	// Check text node inherits tab-size
	if len(preNode.Children()) == 0 {
		t.Fatal("Pre has no children")
	}

	textNode := preNode.Children()[0].(*node)
	if textNode.Type() != html.TextNode {
		t.Fatal("First child is not a text node")
	}

	// Compute styles for text node too
	r.computeStyles(textNode)

	if textNode.computedStyle.TabSize != 2 {
		t.Errorf("Text node expected tab-size 2, got %d", textNode.computedStyle.TabSize)
	}
}
