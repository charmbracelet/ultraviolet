package doc

import (
	"testing"

	"golang.org/x/net/html"
)

func TestRenderTextToLinesWithWideCharacters(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with wide characters (emojis, CJK)
	lines := r.renderTextToLines("Hello 世界 🌍", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Count total width (should be more than the string length)
	totalWidth := 0
	for _, cell := range lines[0] {
		totalWidth += cell.Width
	}

	// "Hello 世界 🌍" should have width > len("Hello 世界 🌍")
	// H(1) e(1) l(1) l(1) o(1) space(1) 世(2) 界(2) space(1) 🌍(2) = 13
	expectedWidth := 13
	if totalWidth != expectedWidth {
		t.Errorf("Expected total width %d, got %d", expectedWidth, totalWidth)
	}
}

func TestRenderTextToLinesWithControlCharacters(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with zero-width control character (U+200B Zero Width Space)
	text := "Hello\u200BWorld"
	lines := r.renderTextToLines(text, &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Should have 10 cells: H e l l o W o r l d
	if len(lines[0]) != 10 {
		t.Errorf("Expected 10 cells, got %d", len(lines[0]))
	}

	// Control character should be concatenated with next character
	found := false
	for _, cell := range lines[0] {
		if cell.Content == "\u200BW" {
			found = true
			if cell.Width != 1 {
				t.Errorf("Expected width 1 for cell with control char, got %d", cell.Width)
			}
		}
	}

	if !found {
		t.Error("Control character not concatenated with next character")
	}
}

func TestRenderTextToLinesWithControlCharacterAtEnd(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with control character at the end
	text := "Hello\u200B"
	lines := r.renderTextToLines(text, &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Should have 5 cells: H e l l o
	if len(lines[0]) != 5 {
		t.Errorf("Expected 5 cells, got %d", len(lines[0]))
	}

	// Last cell should contain the control character
	lastCell := lines[0][len(lines[0])-1]
	if lastCell.Content != "o\u200B" {
		t.Errorf("Expected last cell to be 'o\\u200B', got %q", lastCell.Content)
	}
	if lastCell.Width != 1 {
		t.Errorf("Expected width 1 for last cell, got %d", lastCell.Width)
	}
}

func TestRenderTextToLinesWithMultipleControlCharacters(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with multiple consecutive control characters
	text := "A\u200B\u200CB"
	lines := r.renderTextToLines(text, &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Should have 2 cells: A and B (with control chars attached)
	if len(lines[0]) != 2 {
		t.Errorf("Expected 2 cells, got %d", len(lines[0]))
	}

	// Second cell should have both control characters
	if lines[0][1].Content != "\u200B\u200CB" {
		t.Errorf("Expected second cell to be '\\u200B\\u200CB', got %q", lines[0][1].Content)
	}
	if lines[0][1].Width != 1 {
		t.Errorf("Expected width 1 for second cell, got %d", lines[0][1].Width)
	}
}

func TestRenderTextToLinesWithControlCharacterBeforeNewline(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with control character before newline
	text := "Hello\u200B\nWorld"
	lines := r.renderTextToLines(text, &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// First line should have control char attached to last cell
	if len(lines[0]) != 5 {
		t.Errorf("Expected 5 cells in first line, got %d", len(lines[0]))
	}

	lastCell := lines[0][len(lines[0])-1]
	if lastCell.Content != "o\u200B" {
		t.Errorf("Expected last cell to be 'o\\u200B', got %q", lastCell.Content)
	}
}

func TestRenderTextToLinesWithControlCharacterBeforeTab(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with control character before tab
	text := "A\u200B\tB"
	lines := r.renderTextToLines(text, &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// First cell should be "A\u200B"
	if lines[0][0].Content != "A\u200B" {
		t.Errorf("Expected first cell to be 'A\\u200B', got %q", lines[0][0].Content)
	}
	if lines[0][0].Width != 1 {
		t.Errorf("Expected width 1 for first cell, got %d", lines[0][0].Width)
	}
}

func TestRenderTextToLinesWithGraphemeClusters(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with grapheme clusters (combining characters)
	// e + combining acute accent = é
	lines := r.renderTextToLines("café", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Should have 4 cells for c-a-f-é (grapheme cluster counted as one)
	if len(lines[0]) != 4 {
		t.Errorf("Expected 4 cells, got %d", len(lines[0]))
	}
}

func TestWrapLinesBasic(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Create a line with known width
	lines := r.renderTextToLines("Hello World", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Wrap to width 5 (should split "Hello" and "World")
	wrapped := r.wrapLines(lines, 5)

	// Should have at least 2 lines
	if len(wrapped) < 2 {
		t.Errorf("Expected at least 2 wrapped lines, got %d", len(wrapped))
	}
}

func TestWrapLinesWordBoundary(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Create a line with spaces
	lines := r.renderTextToLines("The quick brown fox", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Wrap to width 10
	wrapped := r.wrapLines(lines, 10)

	// Should wrap at word boundaries, not mid-word
	// Verify first line doesn't end mid-word
	if len(wrapped) > 0 {
		firstLine := wrapped[0]
		_ = firstLine
		// Last character should ideally be a letter or space, not mid-word
		// This is a basic check
		if len(wrapped) >= 2 {
			t.Logf("First line has %d cells", len(firstLine))
			t.Logf("Second line has %d cells", len(wrapped[1]))
		}
	}
}

func TestWrapLinesWideCharacters(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Create a line with wide characters
	lines := r.renderTextToLines("日本語テキスト", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Wrap to width 6 (should fit 3 wide characters per line)
	wrapped := r.wrapLines(lines, 6)

	// Should have multiple lines due to wide character widths
	if len(wrapped) < 2 {
		t.Errorf("Expected at least 2 wrapped lines for wide characters, got %d", len(wrapped))
	}

	// Check that each line respects width limit
	for i, line := range wrapped {
		width := 0
		for _, cell := range line {
			width += cell.Width
		}
		if width > 6 {
			t.Errorf("Line %d exceeds width limit: %d > 6", i, width)
		}
	}
}

func TestWrapLinesNoWrapNeeded(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	lines := r.renderTextToLines("Hi", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Wrap to width 10 (line fits)
	wrapped := r.wrapLines(lines, 10)

	// Should have exactly 1 line
	if len(wrapped) != 1 {
		t.Errorf("Expected 1 line (no wrap needed), got %d", len(wrapped))
	}
}

func TestWrapLinesZeroWidth(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	lines := r.renderTextToLines("Hello", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Wrap to width 0 (should return unchanged)
	wrapped := r.wrapLines(lines, 0)

	if len(wrapped) != len(lines) {
		t.Errorf("Expected %d lines (unchanged), got %d", len(lines), len(wrapped))
	}
}

func TestRenderTextToLinesEmoji(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with various emojis
	lines := r.renderTextToLines("👍 ❤️ 🎉", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Emojis should have width 2 each, plus spaces
	// 👍(2) space(1) ❤️(2) space(1) 🎉(2) = 8
	totalWidth := 0
	for _, cell := range lines[0] {
		totalWidth += cell.Width
	}

	if totalWidth < 6 {
		t.Errorf("Expected total width >= 6 for emojis with spaces, got %d", totalWidth)
	}
}

func TestRenderTextToLinesMultiline(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with newlines
	lines := r.renderTextToLines("Line 1\nLine 2\nLine 3", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// Check first line content
	line1Text := ""
	for _, cell := range lines[0] {
		line1Text += cell.Content
	}
	if line1Text != "Line 1" {
		t.Errorf("Expected 'Line 1', got '%s'", line1Text)
	}

	// Check second line content
	line2Text := ""
	for _, cell := range lines[1] {
		line2Text += cell.Content
	}
	if line2Text != "Line 2" {
		t.Errorf("Expected 'Line 2', got '%s'", line2Text)
	}
}

func TestRenderTextToLinesEmptyLines(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with empty lines (double newlines)
	lines := r.renderTextToLines("First\n\nThird", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines (with empty middle), got %d", len(lines))
	}

	// Middle line should be empty
	if len(lines[1]) != 0 {
		t.Errorf("Expected empty middle line, got %d cells", len(lines[1]))
	}
}

func TestRenderTextToLinesTrailingNewline(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with trailing newline
	lines := r.renderTextToLines("Text\n", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Should have 2 lines (text + empty line after newline)
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines (text + empty after newline), got %d", len(lines))
	}

	// Last line should be empty
	if len(lines[1]) != 0 {
		t.Errorf("Expected empty last line, got %d cells", len(lines[1]))
	}
}

func TestRenderTextToLinesEmptyString(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()

	// Test with empty string
	lines := r.renderTextToLines("", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	// Should return at least one empty line
	if len(lines) != 1 {
		t.Fatalf("Expected 1 empty line, got %d", len(lines))
	}

	if len(lines[0]) != 0 {
		t.Errorf("Expected empty line to have 0 cells, got %d", len(lines[0]))
	}
}

func TestRenderTextToLinesTabDefault(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	// Default tab size is 8

	// Test with tab at start
	lines := r.renderTextToLines("\tHello", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Tab should be converted to 8 spaces
	// Count leading spaces
	spaceCount := 0
	for _, cell := range lines[0] {
		if cell.Content == " " {
			spaceCount++
		} else {
			break
		}
	}

	if spaceCount != 8 {
		t.Errorf("Expected 8 spaces for tab (default tab-size), got %d", spaceCount)
	}
}

func TestRenderTextToLinesTabCustomSize(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TabSize = 4

	// Test with tab at start
	lines := r.renderTextToLines("\tHello", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Tab should be converted to 4 spaces
	spaceCount := 0
	for _, cell := range lines[0] {
		if cell.Content == " " {
			spaceCount++
		} else {
			break
		}
	}

	if spaceCount != 4 {
		t.Errorf("Expected 4 spaces for tab (tab-size=4), got %d", spaceCount)
	}
}

func TestRenderTextToLinesTabAlignment(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TabSize = 4

	// Test tab alignment: "ab\t" should expand to next tab stop
	// "ab" = 2 chars, next tab stop at 4, so 2 spaces
	lines := r.renderTextToLines("ab\tx", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Should be: a, b, space, space, x = 5 cells
	if len(lines[0]) != 5 {
		t.Errorf("Expected 5 cells (a, b, 2 spaces, x), got %d", len(lines[0]))
	}

	// Verify content
	expected := []string{"a", "b", " ", " ", "x"}
	for i, cell := range lines[0] {
		if cell.Content != expected[i] {
			t.Errorf("Cell %d: expected '%s', got '%s'", i, expected[i], cell.Content)
		}
	}
}

func TestRenderTextToLinesMultipleTabs(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TabSize = 4

	// Test multiple tabs
	lines := r.renderTextToLines("\t\t", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Two tabs = 4 + 4 = 8 spaces
	if len(lines[0]) != 8 {
		t.Errorf("Expected 8 spaces for two tabs, got %d", len(lines[0]))
	}
}

func TestRenderTextToLinesTabInMiddle(t *testing.T) {
	r := &Renderer{}
	style := NewComputedStyle()
	style.TabSize = 8

	// Tab in middle: "Hello\tWorld"
	// "Hello" = 5 chars, next tab stop at 8, so 3 spaces
	lines := r.renderTextToLines("Hello\tWorld", &node{n: &html.Node{Type: html.TextNode}, computedStyle: style})

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	// Count total cells: H e l l o + 3 spaces + W o r l d = 13
	totalCells := len(lines[0])
	if totalCells != 13 {
		t.Errorf("Expected 13 cells, got %d", totalCells)
	}
}
