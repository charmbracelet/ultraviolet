package uv

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestStyleParser(t *testing.T) {
	tests := []struct {
		name     string
		params   []int
		hasMore  []bool // explicit hasMore for each param to simulate colon vs semicolon
		want     Style
	}{
		{
			name:    "RGB color (mode 2)",
			params:  []int{38, 2, 0, 255, 128, 64},
			hasMore: []bool{true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "RGB with tolerance params",
			params:  []int{38, 2, 0, 255, 128, 64, 10, 0},
			hasMore: []bool{true, true, true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "RGB without colorspace",
			params:  []int{38, 2, 255, 128, 64},
			hasMore: []bool{true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "indexed color (mode 5)",
			params:  []int{38, 5, 196},
			hasMore: []bool{true, true, false},
			want: Style{
				Fg: ansi.IndexedColor(196),
			},
		},
		{
			name:    "RGBA color (mode 6 - WezTerm)",
			params:  []int{48, 6, 255, 128, 64, 200},
			hasMore: []bool{true, true, true, true, true, false},
			want: Style{
				Bg: color.RGBA{R: 255, G: 128, B: 64, A: 200},
			},
		},
		{
			name:    "RGBA with colorspace",
			params:  []int{48, 6, 0, 255, 128, 64, 200},
			hasMore: []bool{true, true, true, true, true, true, false},
			want: Style{
				Bg: color.RGBA{R: 255, G: 128, B: 64, A: 200},
			},
		},
		{
			name:    "RGBA with tolerance params",
			params:  []int{48, 6, 0, 255, 128, 64, 200, 10, 0},
			hasMore: []bool{true, true, true, true, true, true, true, true, false},
			want: Style{
				Bg: color.RGBA{R: 255, G: 128, B: 64, A: 200},
			},
		},
		{
			name:    "transparent (mode 1)",
			params:  []int{38, 1},
			hasMore: []bool{true, false},
			want: Style{
				Fg: color.Transparent,
			},
		},
		{
			name:    "CMYK color (mode 4)",
			params:  []int{38, 4, 0, 100, 50, 25, 10},
			hasMore: []bool{true, true, true, true, true, true, false},
			want: Style{
				Fg: color.CMYK{C: 100, M: 50, Y: 25, K: 10},
			},
		},
		{
			name:    "CMYK without colorspace",
			params:  []int{38, 4, 100, 50, 25, 10},
			hasMore: []bool{true, true, true, true, true, false},
			want: Style{
				Fg: color.CMYK{C: 100, M: 50, Y: 25, K: 10},
			},
		},
		{
			name:    "CMY color (mode 3)",
			params:  []int{48, 3, 0, 100, 50, 25},
			hasMore: []bool{true, true, true, true, true, false},
			want: Style{
				Bg: color.CMYK{C: 100, M: 50, Y: 25, K: 0},
			},
		},
		{
			name:    "CMY without colorspace",
			params:  []int{48, 3, 100, 50, 25},
			hasMore: []bool{true, true, true, true, false},
			want: Style{
				Bg: color.CMYK{C: 100, M: 50, Y: 25, K: 0},
			},
		},
		{
			name:    "underline color RGB",
			params:  []int{58, 2, 0, 255, 0, 128},
			hasMore: []bool{true, true, true, true, true, false},
			want: Style{
				UnderlineColor: color.RGBA{R: 255, G: 0, B: 128, A: 255},
			},
		},
		{
			name:    "basic bold",
			params:  []int{1},
			hasMore: []bool{false},
			want: Style{
				Attrs: AttrBold,
			},
		},
		{
			name:    "complex sequence",
			params:  []int{1, 3, 38, 2, 0, 255, 128, 64, 48, 5, 196},
			hasMore: []bool{false, false, true, true, true, true, true, false, true, true, false},
			want: Style{
				Attrs: AttrBold | AttrItalic,
				Fg:    color.RGBA{R: 255, G: 128, B: 64, A: 255},
				Bg:    ansi.IndexedColor(196),
			},
		},
		// Semicolon-separated tests (hasMore=false for all color params)
		{
			name:    "RGB semicolon-separated (mode 2)",
			params:  []int{38, 2, 255, 128, 64},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "RGB semicolon-separated background",
			params:  []int{48, 2, 200, 100, 50},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Bg: color.RGBA{R: 200, G: 100, B: 50, A: 255},
			},
		},
		{
			name:    "CMY semicolon-separated (mode 3)",
			params:  []int{38, 3, 100, 50, 25},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Fg: color.CMYK{C: 100, M: 50, Y: 25, K: 0},
			},
		},
		{
			name:    "CMYK semicolon-separated (mode 4)",
			params:  []int{48, 4, 100, 50, 25, 10},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Bg: color.CMYK{C: 100, M: 50, Y: 25, K: 10},
			},
		},
		{
			name:    "RGBA semicolon-separated (mode 6)",
			params:  []int{38, 6, 255, 128, 64, 200},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 200},
			},
		},
		{
			name:    "indexed semicolon-separated (mode 5)",
			params:  []int{38, 5, 196},
			hasMore: []bool{false, false, false},
			want: Style{
				Fg: ansi.IndexedColor(196),
			},
		},
		{
			name:    "underline color semicolon-separated",
			params:  []int{58, 2, 255, 0, 128},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				UnderlineColor: color.RGBA{R: 255, G: 0, B: 128, A: 255},
			},
		},
		{
			name:    "mixed semicolon sequence with attributes",
			params:  []int{1, 3, 38, 2, 255, 128, 64, 48, 5, 196},
			hasMore: []bool{false, false, false, false, false, false, false, false, false, false},
			want: Style{
				Attrs: AttrBold | AttrItalic,
				Fg:    color.RGBA{R: 255, G: 128, B: 64, A: 255},
				Bg:    ansi.IndexedColor(196),
			},
		},
		// Colon-separated with full tolerance parameters
		{
			name:    "RGB colon-separated with full tolerance",
			params:  []int{38, 2, 0, 255, 128, 64, 10, 0},
			hasMore: []bool{true, true, true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "CMY colon-separated with colorspace and tolerance",
			params:  []int{48, 3, 1, 100, 50, 25, 5, 1},
			hasMore: []bool{true, true, true, true, true, true, true, false},
			want: Style{
				Bg: color.CMYK{C: 100, M: 50, Y: 25, K: 0},
			},
		},
		{
			name:    "CMYK colon-separated with colorspace and tolerance",
			params:  []int{38, 4, 0, 100, 50, 25, 10, 8, 0},
			hasMore: []bool{true, true, true, true, true, true, true, true, false},
			want: Style{
				Fg: color.CMYK{C: 100, M: 50, Y: 25, K: 10},
			},
		},
		{
			name:    "RGBA colon-separated with colorspace and tolerance",
			params:  []int{48, 6, 0, 255, 128, 64, 200, 10, 1},
			hasMore: []bool{true, true, true, true, true, true, true, true, false},
			want: Style{
				Bg: color.RGBA{R: 255, G: 128, B: 64, A: 200},
			},
		},
		// Edge cases
		{
			name:    "RGB colon-separated max tolerance params (8 total)",
			params:  []int{38, 2, 0, 255, 128, 64, 0, 0, 10, 0},
			hasMore: []bool{true, true, true, true, true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "transparent semicolon-separated",
			params:  []int{38, 1},
			hasMore: []bool{false, false},
			want: Style{
				Fg: color.Transparent,
			},
		},
		{
			name:    "RGB with zero values semicolon",
			params:  []int{38, 2, 0, 0, 0},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Fg: color.RGBA{R: 0, G: 0, B: 0, A: 255},
			},
		},
		{
			name:    "RGB with max values semicolon",
			params:  []int{48, 2, 255, 255, 255},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Bg: color.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
		},
		{
			name:    "RGBA with zero alpha semicolon",
			params:  []int{38, 6, 255, 128, 64, 0},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 0},
			},
		},
		{
			name:    "RGBA with zero alpha colon-separated with colorspace",
			params:  []int{38, 6, 0, 255, 128, 64, 0},
			hasMore: []bool{true, true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 0},
			},
		},
		{
			name:    "multiple color changes semicolon",
			params:  []int{38, 2, 255, 0, 0, 48, 2, 0, 255, 0, 58, 2, 0, 0, 255},
			hasMore: []bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
			want: Style{
				Fg:             color.RGBA{R: 255, G: 0, B: 0, A: 255},
				Bg:             color.RGBA{R: 0, G: 255, B: 0, A: 255},
				UnderlineColor: color.RGBA{R: 0, G: 0, B: 255, A: 255},
			},
		},
		{
			name:    "multiple color changes colon-separated",
			params:  []int{38, 2, 0, 255, 0, 0, 48, 2, 0, 0, 255, 0, 58, 2, 0, 0, 0, 255},
			hasMore: []bool{true, true, true, true, true, false, true, true, true, true, true, false, true, true, true, true, true, false},
			want: Style{
				Fg:             color.RGBA{R: 255, G: 0, B: 0, A: 255},
				Bg:             color.RGBA{R: 0, G: 255, B: 0, A: 255},
				UnderlineColor: color.RGBA{R: 0, G: 0, B: 255, A: 255},
			},
		},
		// Reset and color clearing
		{
			name:    "reset clears all styles",
			params:  []int{1, 3, 38, 2, 255, 0, 0, 0},
			hasMore: []bool{false, false, false, false, false, false, false, false},
			want:    Style{}, // Reset clears everything
		},
		{
			name:    "default foreground (39)",
			params:  []int{38, 2, 255, 0, 0, 39},
			hasMore: []bool{false, false, false, false, false, false},
			want:    Style{Fg: nil},
		},
		{
			name:    "default background (49)",
			params:  []int{48, 5, 196, 49},
			hasMore: []bool{false, false, false, false},
			want:    Style{Bg: nil},
		},
		{
			name:    "default underline color (59)",
			params:  []int{58, 2, 255, 0, 0, 59},
			hasMore: []bool{false, false, false, false, false, false},
			want:    Style{UnderlineColor: nil},
		},
		// Attribute combinations
		{
			name:    "all attributes set",
			params:  []int{1, 2, 3, 5, 7, 8, 9},
			hasMore: []bool{false, false, false, false, false, false, false},
			want: Style{
				Attrs: AttrBold | AttrFaint | AttrItalic | AttrBlink | AttrReverse | AttrConceal | AttrStrikethrough,
			},
		},
		{
			name:    "attribute reset combinations",
			params:  []int{1, 2, 3, 22, 23},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Attrs: 0, // Bold/Faint reset by 22, Italic reset by 23
			},
		},
		{
			name:    "underline styles",
			params:  []int{4, 1},
			hasMore: []bool{true, false},
			want: Style{
				Underline: UnderlineSingle,
			},
		},
		{
			name:    "underline double",
			params:  []int{4, 2},
			hasMore: []bool{true, false},
			want: Style{
				Underline: UnderlineDouble,
			},
		},
		{
			name:    "underline curly",
			params:  []int{4, 3},
			hasMore: []bool{true, false},
			want: Style{
				Underline: UnderlineCurly,
			},
		},
		{
			name:    "underline dotted",
			params:  []int{4, 4},
			hasMore: []bool{true, false},
			want: Style{
				Underline: UnderlineDotted,
			},
		},
		{
			name:    "underline dashed",
			params:  []int{4, 5},
			hasMore: []bool{true, false},
			want: Style{
				Underline: UnderlineDashed,
			},
		},
		{
			name:    "underline off (24)",
			params:  []int{4, 1, 24},
			hasMore: []bool{true, false, false},
			want: Style{
				Underline: UnderlineStyleNone,
			},
		},
		// Basic 16 colors
		{
			name:    "basic black foreground (30)",
			params:  []int{30},
			hasMore: []bool{false},
			want: Style{
				Fg: ansi.Black,
			},
		},
		{
			name:    "basic red foreground (31)",
			params:  []int{31},
			hasMore: []bool{false},
			want: Style{
				Fg: ansi.Red,
			},
		},
		{
			name:    "basic white foreground (37)",
			params:  []int{37},
			hasMore: []bool{false},
			want: Style{
				Fg: ansi.White,
			},
		},
		{
			name:    "basic black background (40)",
			params:  []int{40},
			hasMore: []bool{false},
			want: Style{
				Bg: ansi.Black,
			},
		},
		{
			name:    "basic white background (47)",
			params:  []int{47},
			hasMore: []bool{false},
			want: Style{
				Bg: ansi.White,
			},
		},
		{
			name:    "bright black foreground (90)",
			params:  []int{90},
			hasMore: []bool{false},
			want: Style{
				Fg: ansi.BrightBlack,
			},
		},
		{
			name:    "bright white foreground (97)",
			params:  []int{97},
			hasMore: []bool{false},
			want: Style{
				Fg: ansi.BrightWhite,
			},
		},
		{
			name:    "bright black background (100)",
			params:  []int{100},
			hasMore: []bool{false},
			want: Style{
				Bg: ansi.BrightBlack,
			},
		},
		{
			name:    "bright white background (107)",
			params:  []int{107},
			hasMore: []bool{false},
			want: Style{
				Bg: ansi.BrightWhite,
			},
		},
		// Mode 0 (implementation defined)
		{
			name:    "mode 0 foreground (no-op)",
			params:  []int{38, 0},
			hasMore: []bool{true, false},
			want:    Style{},
		},
		// Indexed color edge cases
		{
			name:    "indexed color 0",
			params:  []int{38, 5, 0},
			hasMore: []bool{false, false, false},
			want: Style{
				Fg: ansi.IndexedColor(0),
			},
		},
		{
			name:    "indexed color 255",
			params:  []int{48, 5, 255},
			hasMore: []bool{false, false, false},
			want: Style{
				Bg: ansi.IndexedColor(255),
			},
		},
		// Mixed separator types in same sequence
		{
			name:    "mixed colon and semicolon separators",
			params:  []int{1, 38, 2, 0, 255, 128, 64, 48, 5, 196},
			hasMore: []bool{false, true, true, true, true, true, false, false, false, false},
			want: Style{
				Attrs: AttrBold,
				Fg:    color.RGBA{R: 255, G: 128, B: 64, A: 255},
				Bg:    ansi.IndexedColor(196),
			},
		},
		// CMYK edge cases
		{
			name:    "CMYK all zeros",
			params:  []int{38, 4, 0, 0, 0, 0},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Fg: color.CMYK{C: 0, M: 0, Y: 0, K: 0},
			},
		},
		{
			name:    "CMYK all max",
			params:  []int{48, 4, 255, 255, 255, 255},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Bg: color.CMYK{C: 255, M: 255, Y: 255, K: 255},
			},
		},
		// CMY edge cases
		{
			name:    "CMY all zeros",
			params:  []int{38, 3, 0, 0, 0},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Fg: color.CMYK{C: 0, M: 0, Y: 0, K: 0},
			},
		},
		{
			name:    "CMY all max",
			params:  []int{48, 3, 255, 255, 255},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Bg: color.CMYK{C: 255, M: 255, Y: 255, K: 0},
			},
		},
		// Rapid blink
		{
			name:    "rapid blink (6)",
			params:  []int{6},
			hasMore: []bool{false},
			want: Style{
				Attrs: AttrRapidBlink,
			},
		},
		{
			name:    "blink off resets both blink types",
			params:  []int{5, 6, 25},
			hasMore: []bool{false, false, false},
			want:    Style{},
		},
		// Complex real-world scenarios
		{
			name:    "full styled text with all features",
			params:  []int{1, 3, 4, 1, 38, 2, 0, 255, 100, 50, 48, 5, 234, 58, 6, 200, 150, 100, 128},
			hasMore: []bool{false, false, true, false, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false},
			want: Style{
				Attrs:          AttrBold | AttrItalic,
				Underline:      UnderlineSingle,
				Fg:             color.RGBA{R: 255, G: 100, B: 50, A: 255},
				Bg:             ansi.IndexedColor(234),
				UnderlineColor: color.RGBA{R: 200, G: 150, B: 100, A: 128},
			},
		},
		{
			name:    "overwrite colors in sequence",
			params:  []int{38, 5, 196, 38, 2, 0, 255, 128},
			hasMore: []bool{false, false, false, false, false, false, false, false},
			want: Style{
				Fg: color.RGBA{R: 0, G: 255, B: 128, A: 255}, // Second color overwrites first
			},
		},
		{
			name:    "attributes then reset then new attributes",
			params:  []int{1, 3, 0, 2, 9},
			hasMore: []bool{false, false, false, false, false},
			want: Style{
				Attrs: AttrFaint | AttrStrikethrough,
			},
		},
		// Tolerance with different colorspaces
		{
			name:    "RGB with CIELAB tolerance colorspace (1)",
			params:  []int{38, 2, 0, 255, 128, 64, 5, 1},
			hasMore: []bool{true, true, true, true, true, true, true, false},
			want: Style{
				Fg: color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "CMYK with tolerance but no tolerance colorspace",
			params:  []int{48, 4, 0, 100, 50, 25, 10, 8},
			hasMore: []bool{true, true, true, true, true, true, true, false},
			want: Style{
				Bg: color.CMYK{C: 100, M: 50, Y: 25, K: 10},
			},
		},
		// Test state transitions
		{
			name:    "incomplete color sequence followed by attribute",
			params:  []int{38, 2, 255, 128, 64, 1},
			hasMore: []bool{false, false, false, false, false, false},
			want: Style{
				Attrs: AttrBold,
				Fg:    color.RGBA{R: 255, G: 128, B: 64, A: 255},
			},
		},
		{
			name:    "underline style without hasMore",
			params:  []int{4},
			hasMore: []bool{false},
			want: Style{
				Underline: UnderlineSingle,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &StyleParser{}
			for i, param := range tt.params {
				sp.Advance(param, tt.hasMore[i])
			}
			got := sp.Build()

			if got.Attrs != tt.want.Attrs {
				t.Errorf("Attrs = %v, want %v", got.Attrs, tt.want.Attrs)
			}
			if !colorsEqual(got.Fg, tt.want.Fg) {
				t.Errorf("Fg = %v, want %v", got.Fg, tt.want.Fg)
			}
			if !colorsEqual(got.Bg, tt.want.Bg) {
				t.Errorf("Bg = %v, want %v", got.Bg, tt.want.Bg)
			}
			if !colorsEqual(got.UnderlineColor, tt.want.UnderlineColor) {
				t.Errorf("UnderlineColor = %v, want %v", got.UnderlineColor, tt.want.UnderlineColor)
			}
		})
	}
}

func colorsEqual(a, b color.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	
	// Compare RGBA values
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
