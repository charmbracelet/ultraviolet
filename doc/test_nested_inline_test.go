package doc

import (
	"strings"
	"testing"
	
	uv "github.com/charmbracelet/ultraviolet"
)

func TestInlineDoublyNested(t *testing.T) {
	htmlStr := `<p><strong><u>Text</u></strong></p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p := d.QuerySelector("p").(*node)
	r := NewRenderer(p)
	r.computeStyles(p)
	r.layout(p, uv.Rect(0, 0, 80, 10))
	
	// Find the text node
	strong := d.QuerySelector("strong").(*node)
	u := d.QuerySelector("u").(*node)
	textNode := u.Children()[0].(*node)
	
	t.Logf("strong TextDecoration: %v", strong.computedStyle.TextDecoration)
	t.Logf("u TextDecoration: %v TextDecorationStyle=%v TextDecorationColor=%v", 
		u.computedStyle.TextDecoration, u.computedStyle.TextDecorationStyle, u.computedStyle.TextDecorationColor)
	t.Logf("text TextDecoration: %v TextDecorationStyle=%v TextDecorationColor=%v", 
		textNode.computedStyle.TextDecoration, textNode.computedStyle.TextDecorationStyle, textNode.computedStyle.TextDecorationColor)
	
	t.Logf("Text node has %d wrapped lines", len(textNode.wrappedLines))
	
	if len(textNode.wrappedLines) == 0 {
		t.Fatal("Text node should have wrapped lines")
	}
	
	line := textNode.wrappedLines[0]
	if len(line) != 4 {
		t.Fatalf("Expected 4 cells for 'Text', got %d", len(line))
	}
	
	// Check for underline (from <u>) and bold (from <strong>)
	cell := line[0]
	t.Logf("First cell: Content=%q Attrs=%v Underline=%v UnderlineColor=%v", 
		cell.Content, cell.Style.Attrs, cell.Style.Underline, cell.Style.UnderlineColor)
	
	if cell.Style.Attrs&uv.AttrBold == 0 {
		t.Error("Should have bold from <strong>")
	}
	if cell.Style.Underline == 0 {
		t.Error("Should have underline from <u>")
	}
}
