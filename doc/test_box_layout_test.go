package doc

import (
	"strings"
	"testing"
)

func TestBoxTreeLayout(t *testing.T) {
	htmlStr := `<p><strong><u>Bold underlined</u></strong> and <em><s>italic strikethrough</s></em>.</p>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Get the root node (Document embeds *node)
	root := doc.node

	// Create renderer and compute styles
	renderer := NewRenderer(root)
	renderer.computeStyles(root)

	// Build and layout box tree
	boxTree := buildBoxTree(root)
	
	// Layout
	rect := renderer.layoutBox(boxTree, renderer.viewport)
	
	t.Logf("Layout rect: %v", rect)
	
	// Find the p element box
	var findPBox func(*Box) *Box
	findPBox = func(box *Box) *Box {
		if box == nil {
			return nil
		}
		if box.Node != nil && box.Node.Data() == "p" {
			return box
		}
		for _, child := range box.Children {
			if found := findPBox(child); found != nil {
				return found
			}
		}
		return nil
	}
	
	pBox := findPBox(boxTree)
	if pBox == nil {
		t.Fatal("Could not find p box")
	}
	
	t.Logf("P box rect: %v, children: %d", pBox.Rect, len(pBox.Children))
	
	if len(pBox.Children) > 0 {
		anonBox := pBox.Children[0]
		t.Logf("Anonymous box type: %d (BlockBox=0, InlineBox=1, AnonBlock=2)", anonBox.Type)
		t.Logf("Anonymous box rect: %v, children: %d", anonBox.Rect, len(anonBox.Children))
		
		// Check if any child has wrapped lines
		var countWrappedLines func(*Box) int
		countWrappedLines = func(box *Box) int {
			count := len(box.WrappedLines)
			for _, child := range box.Children {
				count += countWrappedLines(child)
			}
			return count
		}
		
		totalLines := countWrappedLines(anonBox)
		t.Logf("Total wrapped lines in anonymous box tree: %d", totalLines)
		
		// Find the first box with wrapped lines
		var findBoxWithLines func(*Box) *Box
		findBoxWithLines = func(box *Box) *Box {
			if len(box.WrappedLines) > 0 {
				return box
			}
			for _, child := range box.Children {
				if found := findBoxWithLines(child); found != nil {
					return found
				}
			}
			return nil
		}
		
		boxWithLines := findBoxWithLines(anonBox)
		if boxWithLines != nil {
			t.Logf("Found box with %d wrapped lines", len(boxWithLines.WrappedLines))
			if len(boxWithLines.WrappedLines) > 0 {
				t.Logf("First line has %d cells", len(boxWithLines.WrappedLines[0]))
			}
		} else {
			t.Log("No box has wrapped lines!")
		}
	}
}
