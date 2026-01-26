package doc

import (
	"testing"
	"strings"
)

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Text   with    spaces", "Text with spaces"},
		{"Multiple\n\nnewlines", "Multiple newlines"},
		{"Tab\t\ttabs", "Tab tabs"},
	}
	
	for _, tt := range tests {
		result := collapseWhitespace(tt.input)
		if result != tt.expected {
			t.Errorf("collapseWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestProcessWhitespace(t *testing.T) {
	tests := []struct {
		input          string
		whiteSpace     WhiteSpace
		isInlineContext bool
		expected       string
	}{
		// Block context: trim leading/trailing
		{"Text   with    spaces", WhiteSpaceNormal, false, "Text with spaces"},
		{" Leading and trailing ", WhiteSpaceNormal, false, "Leading and trailing"},
		
		// Inline context: preserve leading/trailing
		{" Leading and trailing ", WhiteSpaceNormal, true, " Leading and trailing "},
		{" and italic", WhiteSpaceNormal, true, " and italic"},
		
		// Pre preserves everything regardless of context
		{"Text   with    spaces", WhiteSpacePre, false, "Text   with    spaces"},
		{"Text   with    spaces", WhiteSpacePre, true, "Text   with    spaces"},
	}
	
	for _, tt := range tests {
		result := processWhitespace(tt.input, tt.whiteSpace, tt.isInlineContext)
		if result != tt.expected {
			t.Errorf("processWhitespace(%q, %v, %v) = %q, want %q", tt.input, tt.whiteSpace, tt.isInlineContext, result, tt.expected)
		}
	}
}

func TestPreTagWhiteSpace(t *testing.T) {
	htmlStr := `<pre>Text   with    spaces</pre>`
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

	// Check pre element has white-space: pre
	if preNode.computedStyle.WhiteSpace != WhiteSpacePre {
		t.Errorf("Pre element expected white-space: pre, got %v", preNode.computedStyle.WhiteSpace)
	}
	
	// Check text node inherits white-space: pre
	if len(preNode.Children()) == 0 {
		t.Fatal("Pre has no children")
	}

	textNode := preNode.Children()[0].(*node)
	r.computeStyles(textNode)

	if textNode.computedStyle.WhiteSpace != WhiteSpacePre {
		t.Errorf("Text node expected white-space: pre, got %v", textNode.computedStyle.WhiteSpace)
	}
}
