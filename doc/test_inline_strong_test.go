package doc

import (
	"strings"
	"testing"
	
	uv "github.com/charmbracelet/ultraviolet"
)

func TestInlineStrongBold(t *testing.T) {
	htmlStr := `<p><strong>Bold</strong> normal</p>`
	d, err := Parse(strings.NewReader(htmlStr), nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p := d.QuerySelector("p").(*node)
	r := NewRenderer(p)
	r.computeStyles(p)

	// Layout
	rect := r.layout(p, uv.Rect(0, 0, 80, 10))
	
	t.Logf("Layout rect: %v", rect)
	
	// Get the text node inside <strong>
	strong := d.QuerySelector("strong").(*node)
	if strong == nil {
		t.Fatal("strong element not found")
	}
	
	if len(strong.Children()) == 0 {
		t.Fatal("strong has no children")
	}
	
	textNode := strong.Children()[0].(*node)
	t.Logf("Text node wrappedLines: %d", len(textNode.wrappedLines))
	t.Logf("Text node computedStyle.FontWeight: %v", textNode.computedStyle.FontWeight)
	t.Logf("Strong computedStyle.FontWeight: %v", strong.computedStyle.FontWeight)
	
	if len(textNode.wrappedLines) == 0 {
		t.Fatal("No wrapped lines in text node")
	}
	
	// Check first word "Bold" has bold attribute
	line := textNode.wrappedLines[0]
	if len(line) < 4 {
		t.Fatalf("Line too short: %d cells", len(line))
	}
	
	t.Logf("First 4 cells:")
	for i := 0; i < 4 && i < len(line); i++ {
		t.Logf("  [%d] Content=%q Attrs=%v Bold=%v", i, line[i].Content, line[i].Style.Attrs, line[i].Style.Attrs&uv.AttrBold != 0)
	}
	
	// "Bold" should have bold attr
	if line[0].Style.Attrs&uv.AttrBold == 0 {
		t.Error("First char 'B' should be bold")
	}
}
