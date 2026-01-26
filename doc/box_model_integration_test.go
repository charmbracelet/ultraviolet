package doc

import (
	"strings"
	"testing"
)

// TestBoxModelExample demonstrates the box model with margin, padding, and border.
func TestBoxModelExample(t *testing.T) {
	htmlStr := `
		<div style="margin: 2; padding: 1; border: 1">
			<p style="margin: 1; padding: 2">First paragraph with spacing</p>
			<p style="margin: 1; padding: 2">Second paragraph with spacing</p>
		</div>
	`

	r := strings.NewReader(htmlStr)
	rootHTML, err := parseHTML(r)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create root node
	root := &node{
		n:      rootHTML,
		parent: nil,
	}
	root.initChildren()

	// Create renderer and compute styles
	renderer := NewRenderer(root)
	renderer.computeStyles(root)

	// Build box tree
	boxTree := buildBoxTree(root)
	if boxTree == nil {
		t.Fatal("Box tree is nil")
	}

	// Find the div box
	var findDiv func(*Box) *Box
	findDiv = func(box *Box) *Box {
		if box == nil {
			return nil
		}
		if box.Node != nil && box.Node.Data() == "div" {
			return box
		}
		for _, child := range box.Children {
			if found := findDiv(child); found != nil {
				return found
			}
		}
		return nil
	}

	divBox := findDiv(boxTree)
	if divBox == nil {
		t.Fatal("Could not find div box")
	}

	// Verify div has correct spacing
	if divBox.Style.MarginTop != 2 || divBox.Style.MarginLeft != 2 {
		t.Errorf("Div margin: got (%d, %d), want (2, 2)", divBox.Style.MarginTop, divBox.Style.MarginLeft)
	}
	if divBox.Style.PaddingTop != 1 || divBox.Style.PaddingLeft != 1 {
		t.Errorf("Div padding: got (%d, %d), want (1, 1)", divBox.Style.PaddingTop, divBox.Style.PaddingLeft)
	}
	if divBox.Style.BorderTop != 1 || divBox.Style.BorderLeft != 1 {
		t.Errorf("Div border: got (%d, %d), want (1, 1)", divBox.Style.BorderTop, divBox.Style.BorderLeft)
	}

	// Find first p box
	var findFirstP func(*Box) *Box
	findFirstP = func(box *Box) *Box {
		if box == nil {
			return nil
		}
		if box.Node != nil && box.Node.Data() == "p" {
			return box
		}
		for _, child := range box.Children {
			if found := findFirstP(child); found != nil {
				return found
			}
		}
		return nil
	}

	pBox := findFirstP(divBox)
	if pBox == nil {
		t.Fatal("Could not find p box")
	}

	// Verify p has correct spacing
	if pBox.Style.MarginTop != 1 || pBox.Style.MarginLeft != 1 {
		t.Errorf("P margin: got (%d, %d), want (1, 1)", pBox.Style.MarginTop, pBox.Style.MarginLeft)
	}
	if pBox.Style.PaddingTop != 2 || pBox.Style.PaddingLeft != 2 {
		t.Errorf("P padding: got (%d, %d), want (2, 2)", pBox.Style.PaddingTop, pBox.Style.PaddingLeft)
	}

	t.Logf("Box model test passed!")
	t.Logf("Div spacing: margin=%d, border=%d, padding=%d", divBox.Style.MarginTop, divBox.Style.BorderTop, divBox.Style.PaddingTop)
	t.Logf("P spacing: margin=%d, padding=%d", pBox.Style.MarginTop, pBox.Style.PaddingTop)
}

// TestInlineBlockExample demonstrates inline-block display type.
func TestInlineBlockExample(t *testing.T) {
	htmlStr := `
		<div>
			<span style="display: inline-block; margin: 2; padding: 1">Inline Block 1</span>
			<span style="display: inline-block; margin: 2; padding: 1">Inline Block 2</span>
		</div>
	`

	r := strings.NewReader(htmlStr)
	rootHTML, err := parseHTML(r)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create root node
	root := &node{
		n:      rootHTML,
		parent: nil,
	}
	root.initChildren()

	// Create renderer and compute styles
	renderer := NewRenderer(root)
	renderer.computeStyles(root)

	// Build box tree
	boxTree := buildBoxTree(root)
	if boxTree == nil {
		t.Fatal("Box tree is nil")
	}

	// Find span boxes
	var findSpans func(*Box) []*Box
	findSpans = func(box *Box) []*Box {
		if box == nil {
			return nil
		}
		var spans []*Box
		if box.Node != nil && box.Node.Data() == "span" {
			spans = append(spans, box)
		}
		for _, child := range box.Children {
			spans = append(spans, findSpans(child)...)
		}
		return spans
	}

	spans := findSpans(boxTree)
	if len(spans) < 2 {
		t.Fatalf("Expected at least 2 span boxes, got %d", len(spans))
	}

	// Verify first span is inline-block
	span1 := spans[0]
	if span1.Style.Display != DisplayInlineBlock {
		t.Errorf("Span1 display: got %q, want %q", span1.Style.Display, DisplayInlineBlock)
	}
	if span1.Type != InlineBlockBox {
		t.Errorf("Span1 box type: got %d, want %d (InlineBlockBox)", span1.Type, InlineBlockBox)
	}
	if !span1.IsInline() {
		t.Error("Span1 should be treated as inline-level box")
	}

	// Verify spacing
	if span1.Style.MarginTop != 2 || span1.Style.PaddingTop != 1 {
		t.Errorf("Span1 spacing: margin=%d, padding=%d, want margin=2, padding=1",
			span1.Style.MarginTop, span1.Style.PaddingTop)
	}

	t.Logf("Inline-block test passed!")
	t.Logf("Found %d inline-block spans", len(spans))
}

// TestShorthandParsing demonstrates CSS shorthand syntax.
func TestShorthandParsing(t *testing.T) {
	tests := []struct {
		html     string
		property string
		expected [4]int // top, right, bottom, left
	}{
		{
			html:     `<div style="margin: 10px">Test</div>`,
			property: "margin",
			expected: [4]int{10, 10, 10, 10},
		},
		{
			html:     `<div style="padding: 5px 10px">Test</div>`,
			property: "padding",
			expected: [4]int{5, 10, 5, 10},
		},
		{
			html:     `<div style="margin: 1 2 3">Test</div>`,
			property: "margin",
			expected: [4]int{1, 2, 3, 2},
		},
		{
			html:     `<div style="padding: 1 2 3 4">Test</div>`,
			property: "padding",
			expected: [4]int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.html, func(t *testing.T) {
			r := strings.NewReader(tt.html)
			rootHTML, err := parseHTML(r)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			root := &node{
				n:      rootHTML,
				parent: nil,
			}
			root.initChildren()

			renderer := NewRenderer(root)
			renderer.computeStyles(root)

			boxTree := buildBoxTree(root)

			var findDiv func(*Box) *Box
			findDiv = func(box *Box) *Box {
				if box == nil {
					return nil
				}
				if box.Node != nil && box.Node.Data() == "div" {
					return box
				}
				for _, child := range box.Children {
					if found := findDiv(child); found != nil {
						return found
					}
				}
				return nil
			}

			divBox := findDiv(boxTree)
			if divBox == nil {
				t.Fatal("Could not find div box")
			}

			var actual [4]int
			if tt.property == "margin" {
				actual = [4]int{divBox.Style.MarginTop, divBox.Style.MarginRight, divBox.Style.MarginBottom, divBox.Style.MarginLeft}
			} else {
				actual = [4]int{divBox.Style.PaddingTop, divBox.Style.PaddingRight, divBox.Style.PaddingBottom, divBox.Style.PaddingLeft}
			}

			if actual != tt.expected {
				t.Errorf("Property %s: got %v, want %v", tt.property, actual, tt.expected)
			}
		})
	}
}
