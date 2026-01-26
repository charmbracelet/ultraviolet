package doc

import (
	"strings"
	"testing"
)

func TestRenderToStringWithBoxModel(t *testing.T) {
	htmlStr := `
		<div>
			<p>Standard <u>single underline</u> text.</p>
			<p>Link with <a>blue underline</a> by default.</p>
			<p>Strikethrough with <s>deleted text</s> and <del>removed content</del>.</p>
			<p><strong>Bold text</strong> with <em>italic text</em> combined.</p>
			<p><strong><u>Bold underlined</u></strong> and <em><s>italic strikethrough</s></em>.</p>
		</div>
	`

	r := strings.NewReader(htmlStr)
	doc, err := Parse(r, nil)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Render to string with color profile 0 (no color)
	output, err := doc.RenderToString(80, 25, 0)
	if err != nil {
		t.Fatalf("RenderToString error: %v", err)
	}

	t.Logf("Output:\n%s", output)

	// Count lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	t.Logf("Number of lines: %d", len(lines))

	// Check that we have content
	if len(lines) < 4 {
		t.Errorf("Expected at least 4 lines, got %d", len(lines))
	}

	// Check for expected text fragments
	fullOutput := strings.Join(lines, " ")
	
	expectations := []string{
		"Standard",
		"single underline",
		"text",
		"Link with",
		"blue underline",
		"Strikethrough with",
		"deleted text",
		"removed content",
		"Bold text",
		"italic text",
		"combined",
		"Bold underlined",
		"italic strikethrough",
	}

	for _, expected := range expectations {
		if !strings.Contains(fullOutput, expected) {
			t.Errorf("Output missing expected text: %q", expected)
		}
	}
}
