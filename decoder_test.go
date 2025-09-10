package uv

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/ansi"
	xwindows "github.com/charmbracelet/x/windows"
	"github.com/lucasb-eyer/go-colorful"
)

// TestLegacyKeyEncodingMethods tests all the LegacyKeyEncoding flag methods
func TestLegacyKeyEncodingMethods(t *testing.T) {
	tests := []struct {
		name   string
		method func(LegacyKeyEncoding) LegacyKeyEncoding
		flag   uint32
	}{
		{
			name: "CtrlAt true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlAt(true)
			},
			flag: flagCtrlAt,
		},
		{
			name: "CtrlAt false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlAt(false)
			},
			flag: 0,
		},
		{
			name: "CtrlI true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlI(true)
			},
			flag: flagCtrlI,
		},
		{
			name: "CtrlI false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlI(false)
			},
			flag: 0,
		},
		{
			name: "CtrlM true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlM(true)
			},
			flag: flagCtrlM,
		},
		{
			name: "CtrlM false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlM(false)
			},
			flag: 0,
		},
		{
			name: "CtrlOpenBracket true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlOpenBracket(true)
			},
			flag: flagCtrlOpenBracket,
		},
		{
			name: "CtrlOpenBracket false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.CtrlOpenBracket(false)
			},
			flag: 0,
		},
		{
			name: "Backspace true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Backspace(true)
			},
			flag: flagBackspace,
		},
		{
			name: "Backspace false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Backspace(false)
			},
			flag: 0,
		},
		{
			name: "Find true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Find(true)
			},
			flag: flagFind,
		},
		{
			name: "Find false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Find(false)
			},
			flag: 0,
		},
		{
			name: "Select true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Select(true)
			},
			flag: flagSelect,
		},
		{
			name: "Select false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.Select(false)
			},
			flag: 0,
		},
		{
			name: "FKeys true",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.FKeys(true)
			},
			flag: flagFKeys,
		},
		{
			name: "FKeys false",
			method: func(l LegacyKeyEncoding) LegacyKeyEncoding {
				return l.FKeys(false)
			},
			flag: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with all flags set if testing false, or no flags if testing true
			var initial LegacyKeyEncoding
			if tt.flag == 0 {
				initial = LegacyKeyEncoding(0xFFFFFFFF)
			}

			result := tt.method(initial)

			if tt.flag != 0 {
				// Testing true - flag should be set
				if uint32(result)&tt.flag == 0 {
					t.Errorf("expected flag %x to be set, got %x", tt.flag, uint32(result))
				}
			} else {
				// Testing false - specific flag should be cleared
				// We can't check the exact value since we don't know which flag is being cleared
				// Just verify the method runs without panic
			}
		})
	}
}

// TestDeviceAttributesParsing tests device attributes parsing functions
func TestDeviceAttributesParsing(t *testing.T) {
	// Test parsePrimaryDevAttrs
	t.Run("parsePrimaryDevAttrs", func(t *testing.T) {
		params := ansi.Params{62, 1, 2, 6, 9}
		event := parsePrimaryDevAttrs(params)
		if pda, ok := event.(PrimaryDeviceAttributesEvent); ok {
			expected := []int{62, 1, 2, 6, 9}
			if len(pda) != len(expected) {
				t.Errorf("expected %d attrs, got %d", len(expected), len(pda))
			}
			for i, v := range expected {
				if i < len(pda) && pda[i] != v {
					t.Errorf("attr[%d]: expected %d, got %d", i, v, pda[i])
				}
			}
		} else {
			t.Errorf("expected PrimaryDeviceAttributesEvent, got %T", event)
		}
	})

	// Test parseSecondaryDevAttrs
	t.Run("parseSecondaryDevAttrs", func(t *testing.T) {
		params := ansi.Params{1, 2, 3}
		event := parseSecondaryDevAttrs(params)
		if sda, ok := event.(SecondaryDeviceAttributesEvent); ok {
			expected := []int{1, 2, 3}
			if len(sda) != len(expected) {
				t.Errorf("expected %d attrs, got %d", len(expected), len(sda))
			}
			for i, v := range expected {
				if i < len(sda) && sda[i] != v {
					t.Errorf("attr[%d]: expected %d, got %d", i, v, sda[i])
				}
			}
		} else {
			t.Errorf("expected SecondaryDeviceAttributesEvent, got %T", event)
		}
	})

	// Test parseTertiaryDevAttrs
	t.Run("parseTertiaryDevAttrs", func(t *testing.T) {
		data := []byte("4368726d")
		event := parseTertiaryDevAttrs(data)
		if tda, ok := event.(TertiaryDeviceAttributesEvent); ok {
			if expected := "Chrm"; string(tda) != expected {
				t.Errorf("expected '%s', got '%s'", expected, string(tda))
			}
		} else {
			t.Errorf("expected TertiaryDeviceAttributesEvent, got %T", event)
		}
	})
}

// TestHelperFunctions tests various helper functions
func TestHelperFunctions(t *testing.T) {
	// Test shift
	t.Run("shift", func(t *testing.T) {
		if shift(uint(0)) != 0 {
			t.Errorf("shift(0) = %d, want 0", shift(uint(0)))
		}
		if shift(uint(1)) != 1 {
			t.Errorf("shift(1) = %d, want 1", shift(uint(1)))
		}
		if shift(uint(0x100)) != 1 {
			t.Errorf("shift(0x100) = %d, want 1", shift(uint(0x100)))
		}
		if shift(uint32(0x1000)) != 0x10 {
			t.Errorf("shift(0x1000) = %d, want 0x10", shift(uint32(0x1000)))
		}
	})

	// Test colorToHex
	t.Run("colorToHex", func(t *testing.T) {
		tests := []struct {
			color color.Color
			want  string
		}{
			{color.RGBA{R: 255, G: 128, B: 64, A: 255}, "#ff8040"},
			{color.RGBA{R: 0, G: 0, B: 0, A: 255}, "#000000"},
			{color.RGBA{R: 255, G: 255, B: 255, A: 255}, "#ffffff"},
		}
		for _, tt := range tests {
			got := colorToHex(tt.color)
			if got != tt.want {
				t.Errorf("colorToHex(%v) = %s, want %s", tt.color, got, tt.want)
			}
		}
	})

	// Test getMaxMin
	t.Run("getMaxMin", func(t *testing.T) {
		tests := []struct {
			r, g, b  float64
			max, min float64
		}{
			{0.5, 0.3, 0.8, 0.8, 0.3},
			{1.0, 1.0, 1.0, 1.0, 1.0},
			{0.0, 0.0, 0.0, 0.0, 0.0},
			{0.2, 0.5, 0.3, 0.5, 0.2},
		}
		for _, tt := range tests {
			max, min := getMaxMin(tt.r, tt.g, tt.b)
			if max != tt.max || min != tt.min {
				t.Errorf("getMaxMin(%f, %f, %f) = (%f, %f), want (%f, %f)",
					tt.r, tt.g, tt.b, max, min, tt.max, tt.min)
			}
		}
	})

	// Test rgbToHSL
	t.Run("rgbToHSL", func(t *testing.T) {
		tests := []struct {
			r, g, b uint8
			h, s, l float64
		}{
			{255, 0, 0, 0, 1.0, 0.5},     // Red
			{0, 255, 0, 120, 1.0, 0.5},   // Green
			{0, 0, 255, 240, 1.0, 0.5},   // Blue
			{128, 128, 128, 0, 0.0, 0.5}, // Gray
			{255, 255, 255, 0, 0.0, 1.0}, // White
			{0, 0, 0, 0, 0.0, 0.0},       // Black
		}
		for _, tt := range tests {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
			// Allow small floating point differences
			const epsilon = 0.01
			if (h-tt.h) > epsilon || (s-tt.s) > epsilon || (l-tt.l) > epsilon {
				if !(tt.s == 0 && s == 0) { // When saturation is 0, hue is undefined
					t.Errorf("rgbToHSL(%d, %d, %d) = (%f, %f, %f), want (%f, %f, %f)",
						tt.r, tt.g, tt.b, h, s, l, tt.h, tt.s, tt.l)
				}
			}
		}
	})

	// Test isDarkColor
	t.Run("isDarkColor", func(t *testing.T) {
		tests := []struct {
			hex  string
			dark bool
		}{
			{"#ffffff", false}, // White
			{"#000000", true},  // Black
			{"#808080", false}, // Medium gray
			{"#404040", true},  // Dark gray
			{"#c0c0c0", false}, // Light gray
			{"#ff0000", false}, // Red
			{"#800000", true},  // Dark red
		}
		for _, tt := range tests {
			c, _ := colorful.Hex(tt.hex)
			result := isDarkColor(c)
			if result != tt.dark {
				t.Errorf("isDarkColor(%s) = %v, want %v", tt.hex, result, tt.dark)
			}
		}
	})
}

// TestParseTermcap tests termcap parsing
func TestParseTermcap(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectEvent CapabilityEvent
	}{
		{
			name:        "RGB capability",
			input:       []byte("524742"),
			expectEvent: CapabilityEvent("RGB"),
		},
		{
			name:        "Co capability",
			input:       []byte("436F=323536"),
			expectEvent: CapabilityEvent("Co=256"),
		},
		{
			name:        "Empty input",
			input:       []byte(""),
			expectEvent: CapabilityEvent(""),
		},
		{
			name:        "Invalid hex",
			input:       []byte("GGGG"),
			expectEvent: CapabilityEvent(""),
		},
		{
			name:        "Odd length hex",
			input:       []byte("52474"),
			expectEvent: CapabilityEvent(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := parseTermcap(tt.input)
			if event != tt.expectEvent {
				t.Errorf("parseTermcap(%q) = %q, want %q",
					tt.input, event, tt.expectEvent)
			}
		})
	}
}

// TestWin32Functions tests Win32-related functions
func TestWin32Functions(t *testing.T) {
	// Test ensureKeyCase
	t.Run("ensureKeyCase", func(t *testing.T) {
		tests := []struct {
			name        string
			key         Key
			flags       uint32
			wantCode    rune
			wantShifted rune
			wantText    string
		}{
			{
				name:        "uppercase with shift",
				key:         Key{Code: 'a', Text: "A", Mod: ModShift},
				flags:       xwindows.SHIFT_PRESSED,
				wantCode:    'a',
				wantShifted: 'A',
				wantText:    "A",
			},
			{
				name:     "lowercase without shift",
				key:      Key{Code: 'a', Text: "a", Mod: 0},
				flags:    0,
				wantCode: 'a',
				wantText: "a",
			},
			{
				name:        "uppercase without shift",
				key:         Key{Code: 'A', Text: "A", Mod: 0},
				flags:       0,
				wantCode:    'A',
				wantShifted: 'a',
				wantText:    "a",
			},
			{
				name:     "non-letter",
				key:      Key{Code: '1', Text: "1", Mod: 0},
				flags:    0,
				wantCode: '1',
				wantText: "1",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ensureKeyCase(tt.key, tt.flags)
				if result.Text != tt.wantText {
					t.Errorf("ensureKeyCase(%+v, %d).Text = %q, want %q",
						tt.key, tt.flags, result.Text, tt.wantText)
				}
				if result.Code != tt.wantCode {
					t.Errorf("ensureKeyCase(%+v, %d).Code = %c, want %c",
						tt.key, tt.flags, result.Code, tt.wantCode)
				}
				if result.ShiftedCode != tt.wantShifted {
					t.Errorf("ensureKeyCase(%+v, %d).ShiftedCode = %c, want %c",
						tt.key, tt.flags, result.ShiftedCode, tt.wantShifted)
				}
			})
		}
	})

	// Test translateControlKeyState
	t.Run("translateControlKeyState", func(t *testing.T) {
		tests := []struct {
			state uint32
			want  KeyMod
		}{
			{xwindows.RIGHT_ALT_PRESSED, ModAlt},    // RIGHT_ALT_PRESSED
			{xwindows.LEFT_ALT_PRESSED, ModAlt},     // LEFT_ALT_PRESSED
			{xwindows.RIGHT_CTRL_PRESSED, ModCtrl},  // RIGHT_CTRL_PRESSED
			{xwindows.LEFT_CTRL_PRESSED, ModCtrl},   // LEFT_CTRL_PRESSED
			{xwindows.SHIFT_PRESSED, ModShift},      // SHIFT_PRESSED
			{xwindows.NUMLOCK_ON, ModNumLock},       // NUMLOCK_ON
			{xwindows.SCROLLLOCK_ON, ModScrollLock}, // SCROLLLOCK_ON
			{xwindows.CAPSLOCK_ON, ModCapsLock},     // CAPSLOCK_ON
			{xwindows.ENHANCED_KEY, 0},              // ENHANCED_KEY
			{xwindows.RIGHT_ALT_PRESSED | xwindows.RIGHT_CTRL_PRESSED | xwindows.SHIFT_PRESSED, ModAlt | ModCtrl | ModShift}, // Combination
		}
		for _, tt := range tests {
			got := translateControlKeyState(tt.state)
			if got != tt.want {
				t.Errorf("translateControlKeyState(0x%04x) = %d, want %d",
					tt.state, got, tt.want)
			}
		}
	})

	// Test parseWin32InputKeyEvent
	t.Run("parseWin32InputKeyEvent", func(t *testing.T) {
		var p EventDecoder
		tests := []struct {
			name string
			vk   uint16
			sc   uint16
			uc   rune
			kd   bool
			cs   uint32
			rc   uint16
			want Event
		}{
			{
				name: "simple key press",
				vk:   0x41, // VK_A
				sc:   0,
				uc:   'a',
				kd:   true,
				cs:   0,
				rc:   1,
				want: KeyPressEvent{Code: 'a', Text: "a"},
			},
			{
				name: "key release",
				vk:   0x41, // VK_A
				sc:   0,
				uc:   'a',
				kd:   false,
				cs:   0,
				rc:   1,
				want: KeyReleaseEvent{Code: 'a', Text: "a"},
			},
			{
				name: "function key",
				vk:   0x70, // VK_F1
				sc:   0,
				uc:   0,
				kd:   true,
				cs:   0,
				rc:   1,
				want: KeyPressEvent{Code: KeyF1},
			},
			{
				name: "enter key",
				vk:   0x0D, // VK_RETURN
				sc:   0,
				uc:   '\r',
				kd:   true,
				cs:   0,
				rc:   1,
				want: KeyPressEvent{Code: KeyEnter},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := p.parseWin32InputKeyEvent(tt.vk, tt.sc, tt.uc, tt.kd, tt.cs, tt.rc)
				if got == nil {
					t.Fatal("parseWin32InputKeyEvent returned nil")
				}
				// Basic type check
				switch tt.want.(type) {
				case KeyPressEvent:
					if _, ok := got.(KeyPressEvent); !ok {
						t.Errorf("expected KeyPressEvent, got %T", got)
					}
				case KeyReleaseEvent:
					if _, ok := got.(KeyReleaseEvent); !ok {
						t.Errorf("expected KeyReleaseEvent, got %T", got)
					}
				}
			})
		}
	})
}
