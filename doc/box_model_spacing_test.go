package doc

import (
	"testing"
)

func TestParseSpacingValue(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"10", 10},
		{"5px", 5},
		{"3em", 3},
		{"2rem", 2},
		{"10%", 10},
		{"  5  ", 5},
		{"invalid", 0},
		{"-5", 0}, // Negative values should be 0
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseSpacingValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseSpacingValue(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSpacingShorthand(t *testing.T) {
	tests := []struct {
		input  string
		top    int
		right  int
		bottom int
		left   int
	}{
		// One value: all sides
		{"1", 1, 1, 1, 1},
		{"5px", 5, 5, 5, 5},
		
		// Two values: vertical horizontal
		{"1 2", 1, 2, 1, 2},
		{"10px 20px", 10, 20, 10, 20},
		
		// Three values: top horizontal bottom
		{"1 2 3", 1, 2, 3, 2},
		{"5px 10px 15px", 5, 10, 15, 10},
		
		// Four values: top right bottom left
		{"1 2 3 4", 1, 2, 3, 4},
		{"5px 10px 15px 20px", 5, 10, 15, 20},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var top, right, bottom, left int
			parseSpacingShorthand(tt.input, &top, &right, &bottom, &left)
			
			if top != tt.top || right != tt.right || bottom != tt.bottom || left != tt.left {
				t.Errorf("parseSpacingShorthand(%q) = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
					tt.input, top, right, bottom, left, tt.top, tt.right, tt.bottom, tt.left)
			}
		})
	}
}

func TestApplyStylePropertyMargin(t *testing.T) {
	style := NewComputedStyle()
	
	// Test margin shorthand
	applyStyleProperty("margin", "10px", style)
	if style.MarginTop != 10 || style.MarginRight != 10 || style.MarginBottom != 10 || style.MarginLeft != 10 {
		t.Errorf("margin shorthand failed: got (%d, %d, %d, %d), want (10, 10, 10, 10)",
			style.MarginTop, style.MarginRight, style.MarginBottom, style.MarginLeft)
	}
	
	// Test individual margins
	style = NewComputedStyle()
	applyStyleProperty("margin-top", "5px", style)
	applyStyleProperty("margin-right", "10px", style)
	applyStyleProperty("margin-bottom", "15px", style)
	applyStyleProperty("margin-left", "20px", style)
	
	if style.MarginTop != 5 || style.MarginRight != 10 || style.MarginBottom != 15 || style.MarginLeft != 20 {
		t.Errorf("individual margins failed: got (%d, %d, %d, %d), want (5, 10, 15, 20)",
			style.MarginTop, style.MarginRight, style.MarginBottom, style.MarginLeft)
	}
}

func TestApplyStylePropertyPadding(t *testing.T) {
	style := NewComputedStyle()
	
	// Test padding shorthand
	applyStyleProperty("padding", "5 10", style)
	if style.PaddingTop != 5 || style.PaddingRight != 10 || style.PaddingBottom != 5 || style.PaddingLeft != 10 {
		t.Errorf("padding shorthand failed: got (%d, %d, %d, %d), want (5, 10, 5, 10)",
			style.PaddingTop, style.PaddingRight, style.PaddingBottom, style.PaddingLeft)
	}
}

func TestApplyStylePropertyBorder(t *testing.T) {
	style := NewComputedStyle()
	
	// Test border shorthand
	applyStyleProperty("border", "2px solid red", style)
	if style.BorderTop != 2 || style.BorderRight != 2 || style.BorderBottom != 2 || style.BorderLeft != 2 {
		t.Errorf("border shorthand failed: got (%d, %d, %d, %d), want (2, 2, 2, 2)",
			style.BorderTop, style.BorderRight, style.BorderBottom, style.BorderLeft)
	}
	
	// Test border-width
	style = NewComputedStyle()
	applyStyleProperty("border-width", "1 2 3 4", style)
	if style.BorderTop != 1 || style.BorderRight != 2 || style.BorderBottom != 3 || style.BorderLeft != 4 {
		t.Errorf("border-width failed: got (%d, %d, %d, %d), want (1, 2, 3, 4)",
			style.BorderTop, style.BorderRight, style.BorderBottom, style.BorderLeft)
	}
}

func TestDisplayInlineBlock(t *testing.T) {
	style := NewComputedStyle()
	applyStyleProperty("display", "inline-block", style)
	
	if style.Display != DisplayInlineBlock {
		t.Errorf("display inline-block parsing failed: got %q, want %q", style.Display, DisplayInlineBlock)
	}
}

func TestCalculateContentWidth(t *testing.T) {
	style := &ComputedStyle{
		MarginLeft:   1,
		MarginRight:  1,
		BorderLeft:   1,
		BorderRight:  1,
		PaddingLeft:  1,
		PaddingRight: 1,
	}
	
	// Available width 80, spacing 6, content should be 74
	contentWidth := calculateContentWidth(80, style)
	if contentWidth != 74 {
		t.Errorf("calculateContentWidth(80, style) = %d, want 74", contentWidth)
	}
}

func TestCalculateTotalWidth(t *testing.T) {
	style := &ComputedStyle{
		MarginLeft:   1,
		MarginRight:  1,
		BorderLeft:   1,
		BorderRight:  1,
		PaddingLeft:  1,
		PaddingRight: 1,
	}
	
	// Content width 74, spacing 6, total should be 80
	totalWidth := calculateTotalWidth(74, style)
	if totalWidth != 80 {
		t.Errorf("calculateTotalWidth(74, style) = %d, want 80", totalWidth)
	}
}
