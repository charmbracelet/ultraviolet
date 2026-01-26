package doc

import (
	"strings"
	"testing"
)

func TestBoxTreeStructure(t *testing.T) {
	htmlStr := `<p><strong><u>Bold underlined</u></strong> and <em><s>italic strikethrough</s></em>.</p>`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Get the root node (Document embeds *node)
	root := doc.node

	// Compute styles
	renderer := NewRenderer(root)
	renderer.computeStyles(root)

	// Build box tree
	boxTree := buildBoxTree(root)

	// Print box tree structure
	var printBoxTree func(*Box, int)
	printBoxTree = func(box *Box, depth int) {
		if box == nil {
			return
		}
		indent := strings.Repeat("  ", depth)
		nodeType := "nil"
		nodeData := ""
		if box.Node != nil {
			if box.Node.Type() == 3 { // TextNode
				nodeType = "TextNode"
				nodeData = box.Node.Data()
				if len(nodeData) > 30 {
					nodeData = nodeData[:30] + "..."
				}
			} else if box.Node.Type() == 1 { // ElementNode
				nodeType = "ElementNode"
				nodeData = box.Node.Data()
			}
		}
		
		boxTypeStr := "?"
		switch box.Type {
		case BlockBox:
			boxTypeStr = "BlockBox"
		case InlineBox:
			boxTypeStr = "InlineBox"
		case AnonymousBlockBox:
			boxTypeStr = "AnonymousBlockBox"
		case AnonymousInlineBox:
			boxTypeStr = "AnonymousInlineBox"
		}
		
		t.Logf("%s%s node=%s data=%q text=%q children=%d", 
			indent, boxTypeStr, nodeType, nodeData, box.Text, len(box.Children))
		
		for _, child := range box.Children {
			printBoxTree(child, depth+1)
		}
	}

	t.Log("Box tree structure:")
	printBoxTree(boxTree, 0)
}
