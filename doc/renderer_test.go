package doc

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/net/html"
)

func TestRenderTextToLinesBasic(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.Color = ansi.White

	lines := r.renderTextToLines("Hello", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if len(lines[0]) != 5 {
		t.Fatalf("Expected 5 cells, got %d", len(lines[0]))
	}
	if lines[0][0].Content != "H" {
		t.Errorf("Expected 'H', got '%s'", lines[0][0].Content)
	}
}

func TestRenderTextToLinesBold(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.FontWeight = FontWeightBold

	lines := r.renderTextToLines("Bold", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Attrs&uv.AttrBold == 0 {
		t.Error("Expected bold attribute to be set")
	}
}

func TestRenderTextToLinesItalic(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.FontStyle = FontStyleItalic

	lines := r.renderTextToLines("Italic", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Attrs&uv.AttrItalic == 0 {
		t.Error("Expected italic attribute to be set")
	}
}

func TestRenderTextToLinesFaint(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.Faint = true

	lines := r.renderTextToLines("Faint", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Attrs&uv.AttrFaint == 0 {
		t.Error("Expected faint attribute to be set")
	}
}

func TestRenderTextToLinesUnderlineSingle(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleSingle

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Underline != uv.UnderlineSingle {
		t.Errorf("Expected UnderlineSingle, got %d", lines[0][0].Style.Underline)
	}
}

func TestRenderTextToLinesUnderlineDouble(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleDouble

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Underline != uv.UnderlineDouble {
		t.Errorf("Expected UnderlineDouble, got %d", lines[0][0].Style.Underline)
	}
}

func TestRenderTextToLinesUnderlineCurly(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleCurly

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Underline != uv.UnderlineCurly {
		t.Errorf("Expected UnderlineCurly, got %d", lines[0][0].Style.Underline)
	}
}

func TestRenderTextToLinesUnderlineDotted(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleDotted

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Underline != uv.UnderlineDotted {
		t.Errorf("Expected UnderlineDotted, got %d", lines[0][0].Style.Underline)
	}
}

func TestRenderTextToLinesUnderlineDashed(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleDashed

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Underline != uv.UnderlineDashed {
		t.Errorf("Expected UnderlineDashed, got %d", lines[0][0].Style.Underline)
	}
}

func TestRenderTextToLinesUnderlineColor(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleSingle
	style.TextDecorationColor = ansi.BrightRed

	lines := r.renderTextToLines("Underline", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.UnderlineColor != ansi.BrightRed {
		t.Error("Expected underline color to be BrightRed")
	}
}

func TestRenderTextToLinesLineThrough(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationLineThrough}

	lines := r.renderTextToLines("Strike", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if lines[0][0].Style.Attrs&uv.AttrStrikethrough == 0 {
		t.Error("Expected strikethrough attribute to be set")
	}
}

func TestRenderTextToLinesCombined(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.FontWeight = FontWeightBold
	style.FontStyle = FontStyleItalic
	style.TextDecoration = TextDecoration{TextDecorationUnderline}
	style.TextDecorationStyle = UnderlineStyleDouble
	style.Faint = true

	lines := r.renderTextToLines("Combined", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	cell := lines[0][0]
	if cell.Style.Attrs&uv.AttrBold == 0 {
		t.Error("Expected bold attribute")
	}
	if cell.Style.Attrs&uv.AttrItalic == 0 {
		t.Error("Expected italic attribute")
	}
	if cell.Style.Attrs&uv.AttrFaint == 0 {
		t.Error("Expected faint attribute")
	}
	if cell.Style.Underline != uv.UnderlineDouble {
		t.Error("Expected double underline")
	}
}

func TestRenderTextToLinesMultipleDecorations(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TextDecoration = TextDecoration{TextDecorationUnderline, TextDecorationLineThrough}
	style.TextDecorationStyle = UnderlineStyleSingle

	lines := r.renderTextToLines("Both", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	cell := lines[0][0]
	if cell.Style.Underline != uv.UnderlineSingle {
		t.Error("Expected underline")
	}
	if cell.Style.Attrs&uv.AttrStrikethrough == 0 {
		t.Error("Expected strikethrough attribute")
	}
}

func TestTagDefaultsUnderline(t *testing.T) {
	htmlStr := `<html><body><u>Underlined</u></body></html>`
	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	u := doc.QuerySelector("u")
	if u == nil {
		t.Fatal("Could not find <u> element")
	}

	if node, ok := u.(*node); ok {
		style := node.computeStyle()
		if !style.TextDecoration.Has(TextDecorationUnderline) {
			t.Error("Expected <u> tag to have underline decoration")
		}
		if style.TextDecorationStyle != UnderlineStyleSingle {
			t.Error("Expected <u> tag to have single underline style")
		}
	}
}

func TestTagDefaultsLink(t *testing.T) {
	htmlStr := `<html><body><a>Link</a></body></html>`
	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	a := doc.QuerySelector("a")
	if a == nil {
		t.Fatal("Could not find <a> element")
	}

	if node, ok := a.(*node); ok {
		style := node.computeStyle()
		if !style.TextDecoration.Has(TextDecorationUnderline) {
			t.Error("Expected <a> tag to have underline decoration")
		}
		if style.Color != ansi.BrightBlue {
			t.Error("Expected <a> tag to have bright blue color")
		}
	}
}

func TestTagDefaultsStrikethrough(t *testing.T) {
	tests := []string{"s", "strike", "del"}

	for _, tag := range tests {
		t.Run(tag, func(t *testing.T) {
			htmlStr := `<html><body><` + tag + `>Strike</` + tag + `></body></html>`
			r := strings.NewReader(htmlStr)
			doc, err := Parse(r, nil)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			elem := doc.QuerySelector(tag)
			if elem == nil {
				t.Fatalf("Could not find <%s> element", tag)
			}

			if node, ok := elem.(*node); ok {
				style := node.computeStyle()
				if !style.TextDecoration.Has(TextDecorationLineThrough) {
					t.Errorf("Expected <%s> tag to have line-through decoration", tag)
				}
			}
		})
	}
}

func TestLinkBuilding(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		expectedURL    string
		expectedParams string
	}{
		{
			name:           "anchor with href",
			html:           `<a href="https://example.com">link</a>`,
			expectedURL:    "https://example.com",
			expectedParams: "",
		},
		{
			name:           "anchor with href and id",
			html:           `<a href="https://example.com" id="mylink">link</a>`,
			expectedURL:    "https://example.com",
			expectedParams: "id=mylink",
		},
		{
			name:           "element with id only",
			html:           `<span id="myspan">text</span>`,
			expectedURL:    "",
			expectedParams: "id=myspan",
		},
		{
			name:           "nested text inherits link from parent anchor",
			html:           `<a href="/path"><strong>bold link</strong></a>`,
			expectedURL:    "/path",
			expectedParams: "",
		},
		{
			name:           "nested element with parent id",
			html:           `<div id="container"><span>text</span></div>`,
			expectedURL:    "",
			expectedParams: "id=container",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the HTML fragment
			body, err := ParseFragment(strings.NewReader(tt.html), nil, nil)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(body.Children()) == 0 {
				t.Fatal("No children in parsed fragment")
			}

			// Get the first child as the root element
			root := body.Children()[0].(*node)

			// Create renderer and compute styles
			r := NewRenderer(root)
			r.computeStyles(root)

			// Find the text node and render it
			var textNode *node
			var findText func(Node)
			findText = func(n Node) {
				if n.Type() == html.TextNode && strings.TrimSpace(n.Data()) != "" {
					textNode = n.(*node)
					return
				}
				for _, child := range n.Children() {
					findText(child)
					if textNode != nil {
						return
					}
				}
			}
			findText(root)

			if textNode == nil {
				t.Fatal("No text node found")
			}

			lines := r.renderTextToLines(textNode.n.Data, textNode)
			if len(lines) == 0 || len(lines[0]) == 0 {
				t.Fatal("No cells rendered")
			}

			// Check first cell's link
			cell := lines[0][0]
			if cell.Link.URL != tt.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tt.expectedURL, cell.Link.URL)
			}
			if cell.Link.Params != tt.expectedParams {
				t.Errorf("Expected Params '%s', got '%s'", tt.expectedParams, cell.Link.Params)
			}
		})
	}
}
