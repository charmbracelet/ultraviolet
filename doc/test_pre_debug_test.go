package doc

import (
	"strings"
	"testing"
	uv "github.com/charmbracelet/ultraviolet"
)

func TestPrePreservesSpaces(t *testing.T) {
	htmlStr := `<pre>A   B</pre>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	pre := d.QuerySelector("pre")
	preNode := pre.(*node)
	r := NewRenderer(preNode)
	r.computeStyles(preNode)
	
	// Get text node
	textNode := preNode.Children()[0].(*node)
	r.computeStyles(textNode)
	
	t.Logf("Text node data: %q", textNode.Data())
	t.Logf("Text node WhiteSpace: %v", textNode.computedStyle.WhiteSpace)
	
	// Layout
	r.layout(preNode, uv.Rect(0, 0, 80, 25))
	
	// Check wrapped lines
	if len(textNode.wrappedLines) == 0 {
		t.Fatal("No wrapped lines")
	}
	
	var content string
	for _, line := range textNode.wrappedLines {
		for _, cell := range line {
			content += cell.Content
		}
	}
	
	t.Logf("Rendered content: %q", content)
	
	if content != "A   B" {
		t.Errorf("Expected %q, got %q", "A   B", content)
	}
}
