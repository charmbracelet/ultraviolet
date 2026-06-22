package uv

import (
	"io"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// warningSign is U+26A0 followed by a VS16 emoji-presentation selector
// (U+FE0F). Terminals in Unicode core mode (DEC mode 2027) measure it as a
// wide, width-2 glyph.
const warningSign = "\u26a0\ufe0f"

func TestGraphemeWidthNegotiation(t *testing.T) {
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})

	if got := s.StringWidth(warningSign); got != 1 {
		t.Fatalf("default width method: StringWidth(%q) = %d, want 1", warningSign, got)
	}

	s.enableGraphemeWidth()

	if got := s.StringWidth(warningSign); got != 2 {
		t.Fatalf("grapheme mode: StringWidth(%q) = %d, want 2", warningSign, got)
	}
	if s.WidthMethod() != ansi.GraphemeWidth {
		t.Fatalf("grapheme mode: WidthMethod() = %v, want GraphemeWidth", s.WidthMethod())
	}
	if !s.rend.flags.Contains(tGraphemeWidth) {
		t.Fatal("grapheme mode: renderer tGraphemeWidth flag not set")
	}
	if s.rend.method != ansi.GraphemeWidth {
		t.Fatalf("grapheme mode: renderer method = %v, want GraphemeWidth", s.rend.method)
	}
}

func TestGraphemeWidthOverride(t *testing.T) {
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})

	s.SetWidthMethod(ansi.GraphemeWidth)
	if got := s.StringWidth(warningSign); got != 2 {
		t.Fatalf("override: StringWidth(%q) = %d, want 2", warningSign, got)
	}
	if s.rend.method != ansi.GraphemeWidth {
		t.Fatalf("override: renderer method = %v, want GraphemeWidth", s.rend.method)
	}

	s.SetWidthMethod(ansi.WcWidth)
	if got := s.StringWidth(warningSign); got != 1 {
		t.Fatalf("override back: StringWidth(%q) = %d, want 1", warningSign, got)
	}
}

// TestGraphemeWidthGating verifies that the mode-2027 report only switches the
// width method to grapheme clustering for settable or active states, and never
// for not-recognized or permanently-reset terminals.
func TestGraphemeWidthGating(t *testing.T) {
	tests := []struct {
		name    string
		value   ansi.ModeSetting
		enabled bool
	}{
		{"set", ansi.ModeSet, true},
		{"reset", ansi.ModeReset, true},
		{"permanently set", ansi.ModePermanentlySet, true},
		{"not recognized", ansi.ModeNotRecognized, false},
		{"permanently reset", ansi.ModePermanentlyReset, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})
			tm := &Terminal{scr: s}

			tm.handleEvent(ModeReportEvent{Mode: ansi.ModeUnicodeCore, Value: tt.value})

			wantWidth := 1
			wantMethod := ansi.Method(ansi.WcWidth)
			if tt.enabled {
				wantWidth = 2
				wantMethod = ansi.GraphemeWidth
			}

			if got := s.StringWidth(warningSign); got != wantWidth {
				t.Fatalf("StringWidth(%q) = %d, want %d", warningSign, got, wantWidth)
			}
			if s.rend.method != wantMethod {
				t.Fatalf("renderer method = %v, want %v", s.rend.method, wantMethod)
			}
			if got := s.rend.flags.Contains(tGraphemeWidth); got != tt.enabled {
				t.Fatalf("renderer tGraphemeWidth flag = %v, want %v", got, tt.enabled)
			}
		})
	}
}

// TestGraphemeWidthGatingPreservesOverride verifies that a late mode-2027
// report does not clobber a width method the application set explicitly.
func TestGraphemeWidthGatingPreservesOverride(t *testing.T) {
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})
	tm := &Terminal{scr: s}

	s.SetWidthMethod(ansi.WcWidth)

	tm.handleEvent(ModeReportEvent{Mode: ansi.ModeUnicodeCore, Value: ansi.ModeSet})

	if got := s.StringWidth(warningSign); got != 1 {
		t.Fatalf("explicit override clobbered: StringWidth(%q) = %d, want 1", warningSign, got)
	}
	if s.rend.method != ansi.WcWidth {
		t.Fatalf("explicit override clobbered: renderer method = %v, want WcWidth", s.rend.method)
	}
	if s.rend.flags.Contains(tGraphemeWidth) {
		t.Fatal("explicit override clobbered: renderer tGraphemeWidth flag set")
	}
}
