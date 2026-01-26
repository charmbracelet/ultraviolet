package doc

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestInlineStyleColor(t *testing.T) {
	htmlStr := `<span style="color: red">Red text</span>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	span := body.Children()[0].(*node)
	r := NewRenderer(span)
	r.computeStyles(span)

	if span.computedStyle.Color != ansi.Red {
		t.Errorf("Expected color red, got %v", span.computedStyle.Color)
	}
}

func TestInlineStyleBackgroundColor(t *testing.T) {
	htmlStr := `<span style="background-color: blue">Blue bg</span>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	span := body.Children()[0].(*node)
	r := NewRenderer(span)
	r.computeStyles(span)

	if span.computedStyle.BackgroundColor != ansi.Blue {
		t.Errorf("Expected background color blue, got %v", span.computedStyle.BackgroundColor)
	}
}

func TestInlineStyleFontWeight(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected FontWeight
	}{
		{"bold keyword", "font-weight: bold", FontWeightBold},
		{"700 numeric", "font-weight: 700", FontWeightBold},
		{"normal keyword", "font-weight: normal", FontWeightNormal},
		{"400 numeric", "font-weight: 400", FontWeightNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<span style="` + tt.style + `">Text</span>`
			body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			span := body.Children()[0].(*node)
			r := NewRenderer(span)
			r.computeStyles(span)

			if span.computedStyle.FontWeight != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, span.computedStyle.FontWeight)
			}
		})
	}
}

func TestInlineStyleFontStyle(t *testing.T) {
	htmlStr := `<span style="font-style: italic">Italic text</span>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	span := body.Children()[0].(*node)
	r := NewRenderer(span)
	r.computeStyles(span)

	if span.computedStyle.FontStyle != FontStyleItalic {
		t.Errorf("Expected italic, got %v", span.computedStyle.FontStyle)
	}
}

func TestInlineStyleTextDecoration(t *testing.T) {
	tests := []struct {
		name     string
		style    string
		expected []TextDecorationType
	}{
		{
			name:     "underline",
			style:    "text-decoration: underline",
			expected: []TextDecorationType{TextDecorationUnderline},
		},
		{
			name:     "line-through",
			style:    "text-decoration: line-through",
			expected: []TextDecorationType{TextDecorationLineThrough},
		},
		{
			name:     "multiple decorations",
			style:    "text-decoration: underline line-through",
			expected: []TextDecorationType{TextDecorationUnderline, TextDecorationLineThrough},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<span style="` + tt.style + `">Text</span>`
			body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			span := body.Children()[0].(*node)
			r := NewRenderer(span)
			r.computeStyles(span)

			if len(span.computedStyle.TextDecoration) != len(tt.expected) {
				t.Fatalf("Expected %d decorations, got %d", len(tt.expected), len(span.computedStyle.TextDecoration))
			}

			for i, exp := range tt.expected {
				if span.computedStyle.TextDecoration[i] != exp {
					t.Errorf("Expected decoration[%d] = %v, got %v", i, exp, span.computedStyle.TextDecoration[i])
				}
			}
		})
	}
}

func TestInlineStyleHexColor(t *testing.T) {
	tests := []struct {
		name  string
		style string
		checkR uint8
		checkG uint8
		checkB uint8
	}{
		{
			name:   "6-digit hex",
			style:  "color: #FF5733",
			checkR: 0xFF,
			checkG: 0x57,
			checkB: 0x33,
		},
		{
			name:   "3-digit hex",
			style:  "color: #F53",
			checkR: 0xFF,
			checkG: 0x55,
			checkB: 0x33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<span style="` + tt.style + `">Text</span>`
			body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			span := body.Children()[0].(*node)
			r := NewRenderer(span)
			r.computeStyles(span)

			if span.computedStyle.Color == nil {
				t.Fatal("Expected color to be set")
			}

			rr, gg, bb, _ := span.computedStyle.Color.RGBA()
			// RGBA returns 16-bit values, convert to 8-bit
			r8, g8, b8 := uint8(rr>>8), uint8(gg>>8), uint8(bb>>8)

			if r8 != tt.checkR || g8 != tt.checkG || b8 != tt.checkB {
				t.Errorf("Expected RGB(%d,%d,%d), got RGB(%d,%d,%d)",
					tt.checkR, tt.checkG, tt.checkB, r8, g8, b8)
			}
		})
	}
}

func TestInlineStyleOverridesTagDefaults(t *testing.T) {
	// <strong> has bold by default, inline style should override to normal
	htmlStr := `<strong style="font-weight: normal">Not bold</strong>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	strong := body.Children()[0].(*node)
	r := NewRenderer(strong)
	r.computeStyles(strong)

	if strong.computedStyle.FontWeight != FontWeightNormal {
		t.Errorf("Expected normal (inline style should override tag default), got %v", strong.computedStyle.FontWeight)
	}
}

func TestInlineStyleMultipleProperties(t *testing.T) {
	htmlStr := `<span style="color: red; font-weight: bold; font-style: italic">Styled</span>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	span := body.Children()[0].(*node)
	r := NewRenderer(span)
	r.computeStyles(span)

	if span.computedStyle.Color != ansi.Red {
		t.Error("Expected color red")
	}
	if span.computedStyle.FontWeight != FontWeightBold {
		t.Error("Expected font-weight bold")
	}
	if span.computedStyle.FontStyle != FontStyleItalic {
		t.Error("Expected font-style italic")
	}
}

func TestInlineStyleTabSize(t *testing.T) {
	htmlStr := `<pre style="tab-size: 4">Text</pre>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	pre := body.Children()[0].(*node)
	r := NewRenderer(pre)
	r.computeStyles(pre)

	if pre.computedStyle.TabSize != 4 {
		t.Errorf("Expected tab-size 4, got %d", pre.computedStyle.TabSize)
	}
}

func TestInlineStyleIndexedColor(t *testing.T) {
	tests := []struct {
		name  string
		value string
		index uint8
	}{
		{"color 0", "0", 0},
		{"color 15", "15", 15},
		{"color 128", "128", 128},
		{"color 255", "255", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlStr := `<span style="color: ` + tt.value + `">Text</span>`
			body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			span := body.Children()[0].(*node)
			r := NewRenderer(span)
			r.computeStyles(span)

			if span.computedStyle.Color == nil {
				t.Fatal("Expected color to be set")
			}

			// Check it's an ExtendedColor
			expected := ansi.ExtendedColor(tt.index)
			if span.computedStyle.Color != expected {
				t.Errorf("Expected indexed color %d, got %v", tt.index, span.computedStyle.Color)
			}
		})
	}
}

func TestBackgroundColorInheritsInTerminal(t *testing.T) {
	// In terminals (unlike web), background should inherit to ensure readability
	// Parent has background-color, child SHOULD inherit it
	htmlStr := `<div style="background-color: red"><span>child text</span></div>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	div := body.Children()[0].(*node)
	r := NewRenderer(div)
	r.computeStyles(div)

	// Div should have red background
	if div.computedStyle.BackgroundColor != ansi.Red {
		t.Errorf("Expected div to have red background, got %v", div.computedStyle.BackgroundColor)
	}

	// Span (child) should inherit red background (terminal-specific behavior)
	if len(div.Children()) == 0 {
		t.Fatal("No children in div")
	}
	span := div.Children()[0].(*node)
	if span.computedStyle.BackgroundColor != ansi.Red {
		t.Errorf("Expected span to inherit red background, got %v", span.computedStyle.BackgroundColor)
	}
}

func TestChildBackgroundOverridesParent(t *testing.T) {
	// Child with explicit background should override parent
	htmlStr := `<div style="background-color: red"><span style="background-color: blue">child text</span></div>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	div := body.Children()[0].(*node)
	r := NewRenderer(div)
	r.computeStyles(div)

	// Div should have red background
	if div.computedStyle.BackgroundColor != ansi.Red {
		t.Errorf("Expected div to have red background, got %v", div.computedStyle.BackgroundColor)
	}

	// Span should have blue background (explicit style overrides inheritance)
	if len(div.Children()) == 0 {
		t.Fatal("No children in div")
	}
	span := div.Children()[0].(*node)
	if span.computedStyle.BackgroundColor != ansi.Blue {
		t.Errorf("Expected span to have blue background (override), got %v", span.computedStyle.BackgroundColor)
	}
}

func TestColorInherits(t *testing.T) {
	// Parent has color, child SHOULD inherit it
	htmlStr := `<div style="color: blue"><span>child text</span></div>`
	body, err := ParseFragment(strings.NewReader(htmlStr), nil, nil)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	div := body.Children()[0].(*node)
	r := NewRenderer(div)
	r.computeStyles(div)

	// Div should have blue color
	if div.computedStyle.Color != ansi.Blue {
		t.Errorf("Expected div to have blue color, got %v", div.computedStyle.Color)
	}

	// Span (child) should inherit blue color
	if len(div.Children()) == 0 {
		t.Fatal("No children in div")
	}
	span := div.Children()[0].(*node)
	if span.computedStyle.Color != ansi.Blue {
		t.Errorf("Expected span to inherit blue color, got %v", span.computedStyle.Color)
	}
}
