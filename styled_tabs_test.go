package uv

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestStyledStringTabs(t *testing.T) {
	const link = "\x1b]8;;https://example.com/a\tb\x1b\\"
	const resetLink = "\x1b]8;;\x1b\\"
	const payload = "\x1b_payload\twith\nnewline\x1b\\"
	cases := []struct {
		name, input, expanded, expandedWc string
		width                             int
	}{
		{name: "reported view", input: "Hi\tthere", expanded: "Hi      there"},
		{name: "leading tab", input: "\tX", expanded: "        X"},
		{name: "trailing tab", input: "X\t", expanded: "X       "},
		{name: "at tab stop", input: "12345678\tX", expanded: "12345678        X"},
		{name: "multiple tabs", input: "a\t\tX", expanded: "a               X"},
		{name: "wide character", input: "界\tX", expanded: "界      X"},
		{name: "combining mark", input: "e\u0301\tX", expanded: "e\u0301       X"},
		{name: "styled combining mark", input: "e\x1b[31m\u0301\tX", expanded: "e\x1b[31m\u0301       X"},
		{name: "combining after tab", input: "\t\u0301X", expanded: "        \u0301X"},
		{name: "VS16 after tab", input: "\t\ufe0fX", expanded: "        \ufe0fX"},
		{name: "VS16 between tabs", input: "\t\ufe0f\tX", expanded: "        \ufe0f       X", expandedWc: "        \ufe0f        X"},
		{name: "modifier after tab", input: "\t🏽X", expanded: "        🏽X"},
		{name: "keycap", input: "1\ufe0f\u20e3\tX", expanded: "1\ufe0f\u20e3      X", expandedWc: "1\ufe0f\u20e3       X"},
		{name: "width methods", input: "👨‍👩‍👧‍👦\tX", expanded: "👨‍👩‍👧‍👦      X", expandedWc: "👨‍👩‍👧‍👦        X"},
		{name: "newlines", input: "abc\tX\n界\tY", expanded: "abc     X\n界      Y"},
		{name: "CRLF", input: "abc\tX\r\n界\tY", expanded: "abc     X\r\n界      Y"},
		{name: "carriage return", input: "ab\r\tX", expanded: "ab\r        X"},
		{name: "logical line", input: "123456\tX\tY", expanded: "123456  X       Y"},
		{name: "style and link", input: "\x1b[31m" + link + "Hi\tX" + resetLink + "\x1b[0m", expanded: "\x1b[31m" + link + "Hi      X" + resetLink + "\x1b[0m"},
		// A newline inside an opaque payload is not a text-line boundary.
		{name: "pass-through payload", input: "a" + payload + "\tX", expanded: "a" + payload + "       X", width: 9},
		// Keep the existing raw-line width calculation when no text tab exists.
		{name: "payload only tab", input: "a" + payload + "X", expanded: "a" + payload + "X", width: 8},
		{name: "dropped controls", input: "a\x1b[5C\tX", expanded: "a\x1b[5C       X"},
		// The decoder discards incomplete escape prefixes before the tab.
		// Expansion must not turn them into complete sequences that swallow X.
		{name: "incomplete CSI", input: "a\x1b[\tX", expanded: "a       X"},
		{name: "incomplete escape", input: "a\x1b\tX", expanded: "a       X"},
		{name: "incomplete intermediate", input: "a\x1b(\tX", expanded: "a       X"},
		{name: "incomplete CSI prefix", input: "a\x1b[?\tX", expanded: "a       X"},
		{name: "incomplete DCS", input: "a\x1bP\tX", expanded: "a       X"},
		{name: "incomplete C1 CSI", input: "a\x9b\tX", expanded: "a       X"},
		{name: "incomplete C1 DCS", input: "a\x90\tX", expanded: "a       X"},
		{name: "cancelled CSI", input: "a\x1b[\x1b\tX", expanded: "a       X"},
		{name: "cancelled APC", input: "a\x1b_payload\x1b\tX", expanded: "a       X"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			want := NewStyledString(tc.expanded)
			wantWc := want
			if tc.expandedWc != "" {
				wantWc = NewStyledString(tc.expandedWc)
			}
			wantWidth, wantWcWidth := want.UnicodeWidth(), wantWc.WcWidth()
			if tc.width != 0 {
				wantWidth, wantWcWidth = tc.width, tc.width
			}
			if got := ss.UnicodeWidth(); got != wantWidth {
				t.Errorf("UnicodeWidth() = %d, want %d", got, wantWidth)
			}
			if got := ss.WcWidth(); got != wantWcWidth {
				t.Errorf("WcWidth() = %d, want %d", got, wantWcWidth)
			}
			if got, expected := ss.Bounds(), Rect(0, 0, wantWidth, want.Height()); got != expected {
				t.Errorf("Bounds() = %v, want %v", got, expected)
			}
			if got := ss.Height(); got != want.Height() {
				t.Errorf("Height() = %d, want %d", got, want.Height())
			}

			for name, method := range map[string]ansi.Method{"grapheme": ansi.GraphemeWidth, "wcwidth": ansi.WcWidth} {
				t.Run(name, func(t *testing.T) {
					expanded := tc.expanded
					if method == ansi.WcWidth && tc.expandedWc != "" {
						expanded = tc.expandedWc
					}
					want := NewStyledString(expanded)
					if got, expected := ss.Lines(method), want.Lines(method); !reflect.DeepEqual(got, expected) {
						t.Errorf("Lines() = %#v, want %#v", got, expected)
					}

					for _, layout := range []struct {
						name string
						area Rectangle
						wrap bool
						tail string
					}{
						{name: "wide", area: Rect(0, 0, 28, 4)},
						{name: "offset", area: Rect(3, 1, 25, 3)},
						{name: "truncate", area: Rect(3, 1, 5, 3)},
						{name: "tail", area: Rect(3, 1, 5, 3), tail: "…"},
						{name: "wrap", area: Rect(3, 1, 5, 3), wrap: true},
					} {
						t.Run(layout.name, func(t *testing.T) {
							got, expected := NewScreenBuffer(32, 5), NewScreenBuffer(32, 5)
							got.Method, expected.Method = method, method
							for y := 0; y < got.Height(); y++ {
								for x := 0; x < got.Width(); x++ {
									cell := NewCell(method, "!")
									got.SetCell(x, y, cell)
									expected.SetCell(x, y, cell)
								}
							}
							// Direct construction must work as well as NewStyledString.
							actualString := StyledString{Text: tc.input, Wrap: layout.wrap, Tail: layout.tail}
							wantedString := StyledString{Text: expanded, Wrap: layout.wrap, Tail: layout.tail}
							actualString.Draw(got, layout.area)
							wantedString.Draw(expected, layout.area)
							if !reflect.DeepEqual(got.Lines, expected.Lines) {
								t.Errorf("Draw() differs from explicit spaces in %v", layout.area)
							}
							if actualString.String() != tc.input {
								t.Error("Draw changed the original text")
							}
						})
					}
				})
			}
			if ss.String() != tc.input {
				t.Error("measuring or decomposing changed the original text")
			}
		})
	}
}

func TestStyledStringTabCells(t *testing.T) {
	ss := NewStyledString("\x1b[31m" + ansi.SetHyperlink("https://example.com", "") + "Hi\tthere")
	lines := ss.Lines(ansi.GraphemeWidth)
	if len(lines) != 1 || len(lines[0]) != 13 {
		t.Fatalf("Lines() has %d lines, want one 13-cell line: %#v", len(lines), lines)
	}
	for x := 2; x < 8; x++ {
		cell := lines[0][x]
		if cell.Content != " " || cell.Width != 1 || !cell.Style.Equal(&lines[0][0].Style) || !cell.Link.Equal(&lines[0][0].Link) {
			t.Errorf("tab cell %d = %#v, want a styled, linked single space", x, cell)
		}
	}
	for _, cell := range lines[0] {
		if strings.ContainsRune(cell.Content, '\t') {
			t.Errorf("raw tab was carried into a cell: %#v", cell)
		}
	}
}

func TestStyledStringTabPayloads(t *testing.T) {
	const payload = "\x1b_payload\twith\nnewline\x1b\\"
	const url = "https://example.com/a\tb"
	ss := NewStyledString(ansi.SetHyperlink(url, "") + "a" + payload + "\tX")
	lines := ss.Lines(ansi.GraphemeWidth)
	if len(lines) != 1 || len(lines[0]) != 9 {
		t.Fatalf("Lines() = %#v, want one 9-cell line", lines)
	}
	if got := lines[0][1].Content; got != payload+" " {
		t.Errorf("first tab cell = %q, want unchanged payload plus a space", got)
	}
	for x, cell := range lines[0] {
		if cell.Link.URL != url {
			t.Errorf("cell %d URL = %q, want %q", x, cell.Link.URL, url)
		}
	}
}
