package uv

import (
	"image/color"
	"math"
	"reflect"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestParseSequence_Events(t *testing.T) {
	input := []byte("\x1b\x1b[Ztest\x00\x1b]10;rgb:1234/1234/1234\x07\x1b[27;2;27~\x1b[?1049;2$y\x1b[4;1$y")
	want := []Event{
		KeyPressEvent{Code: KeyTab, Mod: ModShift | ModAlt},
		KeyPressEvent{Code: 't', Text: "t"},
		KeyPressEvent{Code: 'e', Text: "e"},
		KeyPressEvent{Code: 's', Text: "s"},
		KeyPressEvent{Code: 't', Text: "t"},
		KeyPressEvent{Code: KeySpace, Mod: ModCtrl},
		ForegroundColorEvent{color.RGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xff}},
		KeyPressEvent{Code: KeyEscape, Mod: ModShift},
		ModeReportEvent{Mode: ansi.AltScreenSaveCursorMode, Value: ansi.ModeReset},
		ModeReportEvent{Mode: ansi.InsertReplaceMode, Value: ansi.ModeSet},
	}

	var p EventDecoder
	for i := 0; len(input) != 0; i++ {
		if i >= len(want) {
			t.Fatalf("reached end of want events")
		}
		n, got := p.Decode(input)
		if !reflect.DeepEqual(got, want[i]) {
			t.Errorf("got %#v (%T), want %#v (%T)", got, got, want[i], want[i])
		}
		input = input[n:]
	}
}

func BenchmarkParseSequence(b *testing.B) {
	var p EventDecoder
	input := []byte("\x1b\x1b[Ztest\x00\x1b]10;1234/1234/1234\x07\x1b[27;2;27~")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Decode(input)
	}
}

// TestLegacyKeyEncoding tests all LegacyKeyEncoding flag methods to achieve 100% coverage
func TestLegacyKeyEncoding(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(LegacyKeyEncoding, bool) LegacyKeyEncoding
		flag     uint32
		expected uint32
	}{
		{"CtrlAt true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlAt(v) }, flagCtrlAt, flagCtrlAt},
		{"CtrlAt false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlAt(v) }, flagCtrlAt, 0},
		{"CtrlI true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlI(v) }, flagCtrlI, flagCtrlI},
		{"CtrlI false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlI(v) }, flagCtrlI, 0},
		{"CtrlM true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlM(v) }, flagCtrlM, flagCtrlM},
		{"CtrlM false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlM(v) }, flagCtrlM, 0},
		{"CtrlOpenBracket true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlOpenBracket(v) }, flagCtrlOpenBracket, flagCtrlOpenBracket},
		{"CtrlOpenBracket false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.CtrlOpenBracket(v) }, flagCtrlOpenBracket, 0},
		{"Backspace true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Backspace(v) }, flagBackspace, flagBackspace},
		{"Backspace false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Backspace(v) }, flagBackspace, 0},
		{"Find true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Find(v) }, flagFind, flagFind},
		{"Find false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Find(v) }, flagFind, 0},
		{"Select true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Select(v) }, flagSelect, flagSelect},
		{"Select false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.Select(v) }, flagSelect, 0},
		{"FKeys true", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.FKeys(v) }, flagFKeys, flagFKeys},
		{"FKeys false", func(l LegacyKeyEncoding, v bool) LegacyKeyEncoding { return l.FKeys(v) }, flagFKeys, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var l LegacyKeyEncoding = LegacyKeyEncoding(tt.flag) // Start with flag set
			result := tt.fn(l, tt.expected != 0)
			
			if tt.expected == 0 {
				// Test that flag is cleared when v=false
				if uint32(result)&tt.flag != 0 {
					t.Errorf("Expected flag %d to be cleared, but got %d", tt.flag, uint32(result))
				}
			} else {
				// Test that flag is set when v=true
				if uint32(result)&tt.flag == 0 {
					t.Errorf("Expected flag %d to be set, but got %d", tt.flag, uint32(result))
				}
			}
		})
	}
}

// TestControlCodes tests control code parsing with different legacy settings
func TestControlCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		legacy   LegacyKeyEncoding
		expected Event
	}{
		// Test basic control codes without legacy flags
		{"NUL default", []byte{0x00}, 0, KeyPressEvent{Code: KeySpace, Mod: ModCtrl}},
		{"BS", []byte{0x08}, 0, KeyPressEvent{Code: 'h', Mod: ModCtrl}},
		{"HT default", []byte{0x09}, 0, KeyPressEvent{Code: KeyTab}},
		{"CR default", []byte{0x0D}, 0, KeyPressEvent{Code: KeyEnter}},
		{"ESC default", []byte{0x1B}, 0, KeyPressEvent{Code: KeyEscape}},
		{"DEL default", []byte{0x7F}, 0, KeyPressEvent{Code: KeyBackspace}},
		{"SP", []byte{0x20}, 0, KeyPressEvent{Code: KeySpace, Text: " "}},
		
		// Test with legacy flags
		{"NUL with CtrlAt", []byte{0x00}, LegacyKeyEncoding(0).CtrlAt(true), KeyPressEvent{Code: '@', Mod: ModCtrl}},
		{"HT with CtrlI", []byte{0x09}, LegacyKeyEncoding(0).CtrlI(true), KeyPressEvent{Code: 'i', Mod: ModCtrl}},
		{"CR with CtrlM", []byte{0x0D}, LegacyKeyEncoding(0).CtrlM(true), KeyPressEvent{Code: 'm', Mod: ModCtrl}},
		{"ESC with CtrlOpenBracket", []byte{0x1B}, LegacyKeyEncoding(0).CtrlOpenBracket(true), KeyPressEvent{Code: KeyEscape}}, // Single ESC always returns KeyEscape in Decode function
		{"DEL with Backspace", []byte{0x7F}, LegacyKeyEncoding(0).Backspace(true), KeyPressEvent{Code: KeyDelete}},
		
		// Test SOH-SUB range (0x01-0x1A)
		{"SOH", []byte{0x01}, 0, KeyPressEvent{Code: 'a', Mod: ModCtrl}},
		{"ETX", []byte{0x03}, 0, KeyPressEvent{Code: 'c', Mod: ModCtrl}},
		{"SUB", []byte{0x1A}, 0, KeyPressEvent{Code: 'z', Mod: ModCtrl}},
		
		// Test FS-US range (0x1C-0x1F)
		{"FS", []byte{0x1C}, 0, KeyPressEvent{Code: '\\', Mod: ModCtrl}},
		{"GS", []byte{0x1D}, 0, KeyPressEvent{Code: ']', Mod: ModCtrl}},
		{"RS", []byte{0x1E}, 0, KeyPressEvent{Code: '^', Mod: ModCtrl}},
		{"US", []byte{0x1F}, 0, KeyPressEvent{Code: '_', Mod: ModCtrl}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := EventDecoder{Legacy: tt.legacy}
			n, got := p.Decode(tt.input)
			
			if n != 1 {
				t.Errorf("Expected width 1, got %d", n)
			}
			
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

// TestUtf8Decoding tests UTF-8 character decoding
func TestUtf8Decoding(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Event
		width    int
	}{
		// Empty input
		{"empty", []byte{}, nil, 0},
		
		// ASCII printable characters
		{"lowercase a", []byte{'a'}, KeyPressEvent{Code: 'a', Text: "a"}, 1},
		{"uppercase A", []byte{'A'}, KeyPressEvent{Code: 'a', ShiftedCode: 'A', Text: "A", Mod: ModShift}, 1},
		{"digit 5", []byte{'5'}, KeyPressEvent{Code: '5', Text: "5"}, 1},
		{"space", []byte{' '}, KeyPressEvent{Code: KeySpace, Text: " "}, 1},
		
		// UTF-8 characters
		{"emoji", []byte("🔥"), KeyPressEvent{Code: '🔥', Text: "🔥"}, 4},
		{"accented", []byte("é"), KeyPressEvent{Code: 'é', Text: "é"}, 2},
		{"chinese", []byte("你"), KeyPressEvent{Code: '你', Text: "你"}, 3},
		
		// Multi-rune graphemes - actually tests the first character only
		{"combining", []byte("e\u0301"), KeyPressEvent{Code: 'e', Text: "e"}, 1}, // The decoder processes e first, combining char comes later
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p EventDecoder
			n, got := p.Decode(tt.input)
			
			if n != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, n)
			}
			
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

// TestC1ControlCharacters tests C1 control character handling
func TestC1ControlCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Event
	}{
		// C1 control characters (0x80-0x9F) should be encoded as Ctrl+Alt+<code - 0x40>
		// But some might be valid UTF-8 on certain platforms, so test the known ones
		{"C1 0x80", []byte{0x80}, KeyPressEvent{Code: '@', Mod: ModCtrl | ModAlt}},
		{"C1 0x81", []byte{0x81}, KeyPressEvent{Code: 'A', Mod: ModCtrl | ModAlt}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p EventDecoder
			n, got := p.Decode(tt.input)
			
			if n != 1 {
				t.Errorf("Expected width 1, got %d", n)
			}
			
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

// TestColorFunctions tests color utility functions
func TestColorFunctions(t *testing.T) {
	t.Run("colorToHex", func(t *testing.T) {
		tests := []struct {
			name     string
			color    color.Color
			expected string
		}{
			{"nil color", nil, ""},
			{"black", color.RGBA{R: 0, G: 0, B: 0, A: 255}, "#000000"},
			{"white", color.RGBA{R: 255, G: 255, B: 255, A: 255}, "#ffffff"},
			{"red", color.RGBA{R: 255, G: 0, B: 0, A: 255}, "#ff0000"},
			{"green", color.RGBA{R: 0, G: 255, B: 0, A: 255}, "#00ff00"},
			{"blue", color.RGBA{R: 0, G: 0, B: 255, A: 255}, "#0000ff"},
			{"medium gray", color.RGBA{R: 128, G: 128, B: 128, A: 255}, "#808080"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := colorToHex(tt.color)
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			})
		}
	})

	t.Run("shift", func(t *testing.T) {
		tests := []struct {
			name     string
			input    uint32
			expected uint32
		}{
			{"small value", 100, 100},
			{"large value", 0xFF00, 0xFF},
			{"very large value", 0xFF0000, 0xFF00},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := shift(tt.input)
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			})
		}
	})

	t.Run("getMaxMin", func(t *testing.T) {
		tests := []struct {
			name     string
			a, b, c  float64
			max, min float64
		}{
			{"ascending", 1.0, 2.0, 3.0, 3.0, 1.0},
			{"descending", 3.0, 2.0, 1.0, 3.0, 1.0},
			{"mixed", 2.0, 3.0, 1.0, 3.0, 1.0},
			{"equal", 2.0, 2.0, 2.0, 2.0, 2.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				max, min := getMaxMin(tt.a, tt.b, tt.c)
				if max != tt.max || min != tt.min {
					t.Errorf("Expected max=%f, min=%f, got max=%f, min=%f", tt.max, tt.min, max, min)
				}
			})
		}
	})

	t.Run("round", func(t *testing.T) {
		tests := []struct {
			name     string
			input    float64
			expected float64
		}{
			{"simple", 1.2345, 1.235},
			{"round up", 1.2346, 1.235},
			{"round down", 1.2344, 1.234},
			{"no decimals", 5.0, 5.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := round(tt.input)
				if result != tt.expected {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			})
		}
	})

	t.Run("rgbToHSL", func(t *testing.T) {
		tests := []struct {
			name        string
			r, g, b     uint8
			h, s, l     float64
		}{
			{"black", 0, 0, 0, 0, 0, 0},
			{"white", 255, 255, 255, 0, 0, 1},
			{"red", 255, 0, 0, 0, 1, 0.5},
			{"green", 0, 255, 0, 120, 1, 0.5},
			{"blue", 0, 0, 255, 240, 1, 0.5},
			{"gray", 128, 128, 128, 0, 0, 0.502}, // approximately 0.502
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
				// Allow small floating point differences
				const tolerance = 0.01
				if math.Abs(h-tt.h) > tolerance || math.Abs(s-tt.s) > tolerance || math.Abs(l-tt.l) > tolerance {
					t.Errorf("Expected h=%f, s=%f, l=%f, got h=%f, s=%f, l=%f", tt.h, tt.s, tt.l, h, s, l)
				}
			})
		}
	})

	t.Run("isDarkColor", func(t *testing.T) {
		tests := []struct {
			name     string
			color    color.Color
			expected bool
		}{
			{"nil color", nil, true},
			{"black", color.RGBA{R: 0, G: 0, B: 0, A: 255}, true},
			{"white", color.RGBA{R: 255, G: 255, B: 255, A: 255}, false},
			{"dark gray", color.RGBA{R: 64, G: 64, B: 64, A: 255}, true},
			{"light gray", color.RGBA{R: 192, G: 192, B: 192, A: 255}, false},
			{"red", color.RGBA{R: 255, G: 0, B: 0, A: 255}, false}, // Red is considered light
			{"dark red", color.RGBA{R: 128, G: 0, B: 0, A: 255}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := isDarkColor(tt.color)
				if result != tt.expected {
					t.Errorf("Expected %t, got %t", tt.expected, result)
				}
			})
		}
	})
}



// TestParseTermcap tests the parseTermcap function directly
func TestParseTermcap(t *testing.T) {
tests := []struct {
name     string
input    []byte
expected CapabilityEvent
}{
{"empty data", []byte{}, CapabilityEvent("")},
{"valid hex", []byte("636f=3830"), CapabilityEvent("co=80")}, // "co" in hex = 636f, "80" in hex = 3830
{"no value", []byte("636f"), CapabilityEvent("co")},
{"invalid hex", []byte("zz=41"), CapabilityEvent("")},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := parseTermcap(tt.input)
if result != tt.expected {
t.Errorf("Expected %q, got %q", tt.expected, result)
}
})
}
}

// TestAdditionalEdgeCases tests additional edge cases
func TestAdditionalEdgeCases(t *testing.T) {
t.Run("rgbToHSL with negative hue", func(t *testing.T) {
// Test cyan-like color where hue calculation might go negative
h, s, l := rgbToHSL(0, 255, 128)
if h < 0 {
t.Errorf("Hue should not be negative, got %f", h)
}
// Basic sanity checks
if s < 0 || s > 1 {
t.Errorf("Saturation should be 0-1, got %f", s)
}
if l < 0 || l > 1 {
t.Errorf("Lightness should be 0-1, got %f", l)
}
})

t.Run("parseControl with boundary values", func(t *testing.T) {
var p EventDecoder

// Test boundary values that might not be covered
result := p.parseControl(0x1C) // FS
if k, ok := result.(KeyPressEvent); ok {
if k.Code != '\\' || k.Mod != ModCtrl {
t.Errorf("Expected ctrl+\\, got %+v", k)
}
}

result = p.parseControl(0x1F) // US
if k, ok := result.(KeyPressEvent); ok {
if k.Code != '_' || k.Mod != ModCtrl {
t.Errorf("Expected ctrl+_, got %+v", k)
}
}
})
}
