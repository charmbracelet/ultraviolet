package uv

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestStyledString(t *testing.T) {
	cases := []struct {
		name           string
		input          string
		expected       *Buffer
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "single line",
			input:          "Hello, World!",
			expectedWidth:  13,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("H", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell(",", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("W", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell("r", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("d", nil, nil),
						newWcCell("!", nil, nil),
					},
				},
			},
		},
		{
			name:           "multiple lines",
			input:          "Hello,\nWorld!",
			expectedWidth:  6,
			expectedHeight: 2,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("H", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell(",", nil, nil),
					},
					{
						newWcCell("W", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell("r", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("d", nil, nil),
						newWcCell("!", nil, nil),
					},
				},
			},
		},
		{
			name:           "empty string",
			input:          "",
			expectedWidth:  0,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{{}},
			},
		},
		{
			name:           "multiple lines different width",
			input:          "Hello,\nWorld!\nThis is a test.",
			expectedWidth:  15,
			expectedHeight: 3,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("H", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell(",", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
					},
					{
						newWcCell("W", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell("r", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("d", nil, nil),
						newWcCell("!", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell(" ", nil, nil),
					},
					{
						newWcCell("T", nil, nil),
						newWcCell("h", nil, nil),
						newWcCell("i", nil, nil),
						newWcCell("s", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("i", nil, nil),
						newWcCell("s", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("a", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("t", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("s", nil, nil),
						newWcCell("t", nil, nil),
						newWcCell(".", nil, nil),
					},
				},
			},
		},
		{
			name:           "unicode characters",
			input:          "Hello, 世界!",
			expectedWidth:  12,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("H", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell(",", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("世", nil, nil),
						Cell{},
						newWcCell("界", nil, nil),
						Cell{},
						newWcCell("!", nil, nil),
					},
				},
			},
		},
		{
			name:           "styled hello world string",
			input:          "\x1b[31;1;4mHello, \x1b[32;22;4mWorld!\x1b[0m",
			expectedWidth:  13,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("H", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("e", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("l", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("l", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("o", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell(",", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell(" ", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("W", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
						newWcCell("o", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
						newWcCell("r", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
						newWcCell("l", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
						newWcCell("d", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
						newWcCell("!", &Style{Fg: ansi.Green, Underline: UnderlineStyleSingle}, nil),
					},
				},
			},
		},
		{
			name:           "complex styling with multiple SGR sequences",
			input:          "\x1b[31;1;2;4mR\x1b[22;1med\x1b[0m \x1b[32;3mGreen\x1b[0m \x1b[34;9mBlue\x1b[0m \x1b[33;7mYellow\x1b[0m \x1b[35;5mPurple\x1b[0m",
			expectedWidth:  28,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("R", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold | AttrFaint}, nil),
						newWcCell("e", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell("d", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("G", &Style{Fg: ansi.Green, Attrs: AttrItalic}, nil),
						newWcCell("r", &Style{Fg: ansi.Green, Attrs: AttrItalic}, nil),
						newWcCell("e", &Style{Fg: ansi.Green, Attrs: AttrItalic}, nil),
						newWcCell("e", &Style{Fg: ansi.Green, Attrs: AttrItalic}, nil),
						newWcCell("n", &Style{Fg: ansi.Green, Attrs: AttrItalic}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("B", &Style{Fg: ansi.Blue, Attrs: AttrStrikethrough}, nil),
						newWcCell("l", &Style{Fg: ansi.Blue, Attrs: AttrStrikethrough}, nil),
						newWcCell("u", &Style{Fg: ansi.Blue, Attrs: AttrStrikethrough}, nil),
						newWcCell("e", &Style{Fg: ansi.Blue, Attrs: AttrStrikethrough}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("Y", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell("e", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell("l", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell("l", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell("o", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell("w", &Style{Fg: ansi.Yellow, Attrs: AttrReverse}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("P", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
						newWcCell("u", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
						newWcCell("r", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
						newWcCell("p", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
						newWcCell("l", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
						newWcCell("e", &Style{Fg: ansi.Magenta, Attrs: AttrBlink}, nil),
					},
				},
			},
		},
		{
			name:           "different underline styles",
			input:          "\x1b[4:1mSingle\x1b[0m \x1b[4:2mDouble\x1b[0m \x1b[4:3mCurly\x1b[0m \x1b[4:4mDotted\x1b[0m \x1b[4:5mDashed\x1b[0m",
			expectedWidth:  33,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("S", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell("i", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell("n", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell("g", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell("l", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell("e", &Style{Underline: UnderlineStyleSingle}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("D", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell("o", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell("u", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell("b", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell("l", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell("e", &Style{Underline: UnderlineStyleDouble}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("C", &Style{Underline: UnderlineStyleCurly}, nil),
						newWcCell("u", &Style{Underline: UnderlineStyleCurly}, nil),
						newWcCell("r", &Style{Underline: UnderlineStyleCurly}, nil),
						newWcCell("l", &Style{Underline: UnderlineStyleCurly}, nil),
						newWcCell("y", &Style{Underline: UnderlineStyleCurly}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("D", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell("o", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell("t", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell("t", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell("e", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell("d", &Style{Underline: UnderlineStyleDotted}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("D", &Style{Underline: UnderlineStyleDashed}, nil),
						newWcCell("a", &Style{Underline: UnderlineStyleDashed}, nil),
						newWcCell("s", &Style{Underline: UnderlineStyleDashed}, nil),
						newWcCell("h", &Style{Underline: UnderlineStyleDashed}, nil),
						newWcCell("e", &Style{Underline: UnderlineStyleDashed}, nil),
						newWcCell("d", &Style{Underline: UnderlineStyleDashed}, nil),
					},
				},
			},
		},
		{
			name:           "truecolor and 256 color support",
			input:          "\x1b[38;2;255;0;0mRGB Red\x1b[0m \x1b[48;2;0;255;0mRGB Green BG\x1b[0m \x1b[38;5;33m256 Blue\x1b[0m",
			expectedWidth:  29,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("R", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell("G", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell("B", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell(" ", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell("R", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell("e", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell("d", &Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("R", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("G", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("B", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell(" ", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("G", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("r", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("e", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("e", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("n", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell(" ", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("B", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell("G", &Style{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}}, nil),
						newWcCell(" ", nil, nil),
						newWcCell("2", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("5", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("6", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell(" ", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("B", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("l", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("u", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("e", &Style{Fg: ansi.IndexedColor(33)}, nil),
						newWcCell("e", &Style{Fg: ansi.IndexedColor(33)}, nil),
					},
				},
			},
		},
		{
			name:           "hyperlink support",
			input:          "Normal \x1b]8;;https://charm.sh\x1b\\Charm\x1b]8;;\x1b\\ Text \x1b]8;;https://github.com/charmbracelet\x1b\\GitHub\x1b]8;;\x1b\\",
			expectedWidth:  24,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("N", nil, nil),
						newWcCell("o", nil, nil),
						newWcCell("r", nil, nil),
						newWcCell("m", nil, nil),
						newWcCell("a", nil, nil),
						newWcCell("l", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("C", nil, &Link{URL: "https://charm.sh"}),
						newWcCell("h", nil, &Link{URL: "https://charm.sh"}),
						newWcCell("a", nil, &Link{URL: "https://charm.sh"}),
						newWcCell("r", nil, &Link{URL: "https://charm.sh"}),
						newWcCell("m", nil, &Link{URL: "https://charm.sh"}),
						newWcCell(" ", nil, nil),
						newWcCell("T", nil, nil),
						newWcCell("e", nil, nil),
						newWcCell("x", nil, nil),
						newWcCell("t", nil, nil),
						newWcCell(" ", nil, nil),
						newWcCell("G", nil, &Link{URL: "https://github.com/charmbracelet"}),
						newWcCell("i", nil, &Link{URL: "https://github.com/charmbracelet"}),
						newWcCell("t", nil, &Link{URL: "https://github.com/charmbracelet"}),
						newWcCell("H", nil, &Link{URL: "https://github.com/charmbracelet"}),
						newWcCell("u", nil, &Link{URL: "https://github.com/charmbracelet"}),
						newWcCell("b", nil, &Link{URL: "https://github.com/charmbracelet"}),
					},
				},
			},
		},
		{
			name:           "complex mixed styling with hyperlinks",
			input:          "\x1b[31;1;2;3mR\x1b[22;23;1med \x1b]8;;https://charm.sh\x1b\\\x1b[4mCharm\x1b]8;;\x1b\\\x1b[0m \x1b[38;5;33;48;2;0;100;0m\x1b]8;;https://github.com\x1b\\GitHub\x1b]8;;\x1b\\\x1b[0m",
			expectedWidth:  16,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("R", &Style{Fg: ansi.Red, Attrs: AttrBold | AttrFaint | AttrItalic}, nil),
						newWcCell("e", &Style{Fg: ansi.Red, Attrs: AttrBold}, nil),
						newWcCell("d", &Style{Fg: ansi.Red, Attrs: AttrBold}, nil),
						newWcCell(" ", &Style{Fg: ansi.Red, Attrs: AttrBold}, nil),
						newWcCell("C", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, &Link{URL: "https://charm.sh"}),
						newWcCell("h", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, &Link{URL: "https://charm.sh"}),
						newWcCell("a", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, &Link{URL: "https://charm.sh"}),
						newWcCell("r", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, &Link{URL: "https://charm.sh"}),
						newWcCell("m", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, &Link{URL: "https://charm.sh"}),
						newWcCell(" ", nil, nil),
						newWcCell("G", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
						newWcCell("i", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
						newWcCell("t", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
						newWcCell("H", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
						newWcCell("u", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
						newWcCell("b", &Style{Fg: ansi.IndexedColor(33), Bg: color.RGBA{R: 0, G: 100, B: 0, A: 255}}, &Link{URL: "https://github.com"}),
					},
				},
			},
		},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running case %d: %s for %q", i+1, tc.name, tc.input)
			ss := NewStyledString(tc.input)
			area := ss.Bounds()
			buf := NewScreenBuffer(area.Dx(), area.Dy())
			ss.Draw(buf, area)
			if buf.Width() != tc.expectedWidth {
				t.Errorf("case %d expected width %d, got %d", i+1, tc.expectedWidth, buf.Width())
			}
			if buf.Height() != tc.expectedHeight {
				t.Errorf("case %d expected height %d, got %d", i+1, tc.expectedHeight, buf.Height())
			}
			for y, line := range buf.Lines {
				for x, cell := range line {
					if !cellEqual(tc.expected.CellAt(x, y), &cell) {
						t.Errorf("case %d expected cell (%d, %d) %#v, got %#v", y+1, x, y, tc.expected.CellAt(x, y), &cell)
					}
				}
			}
		})
	}
}

func TestStyledStringEmptyLines(t *testing.T) {
	// This test uses an input that results in empty lines when drawn to a smaller
	// screen buffer.
	input := "\x1b[31;1;4mHello, \x1b[32;22;4mWorld!\x1b[0m"
	ss := NewStyledString(input)
	scr := NewScreenBuffer(5, 3)
	ss.Draw(scr, scr.Bounds())
	expected := &Buffer{
		Lines: []Line{
			{
				newWcCell("H", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
				newWcCell("e", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
				newWcCell("l", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
				newWcCell("l", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
				newWcCell("o", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle, Attrs: AttrBold}, nil),
			},
			NewLine(5),
			NewLine(5),
		},
	}
	for y, line := range scr.Lines {
		for x, cell := range line {
			if !cellEqual(expected.CellAt(x, y), &cell) {
				t.Errorf("expected cell (%d, %d) %#v, got %#v", x, y, expected.CellAt(x, y), &cell)
			}
		}
	}
}

// TestStyledStringKittyTextSizing exercises the kitty OSC 66 path added to
// printString. The protocol embeds the rendered glyph inside the OSC payload
// (https://sw.kovidgoyal.net/kitty/text-sizing-protocol/), so the round-trip
// has to (a) preserve the full escape verbatim in the cell content and
// (b) reserve the correct number of cells based on the metadata.
func TestStyledStringKittyTextSizing(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		atX       int
		wantWidth int
		wantBytes string
	}{
		{
			name:      "scale 2",
			input:     "\x1b]66;s=2;X\x1b\\",
			atX:       0,
			wantWidth: 2,
			wantBytes: "\x1b]66;s=2;X\x1b\\",
		},
		{
			name:      "explicit width",
			input:     "\x1b]66;w=3;X\x1b\\",
			atX:       0,
			wantWidth: 3,
			wantBytes: "\x1b]66;w=3;X\x1b\\",
		},
		{
			name:      "explicit width wins over scale",
			input:     "\x1b]66;w=4:s=2;X\x1b\\",
			atX:       0,
			wantWidth: 4,
			wantBytes: "\x1b]66;w=4:s=2;X\x1b\\",
		},
		{
			name:      "no metadata defaults to 1",
			input:     "\x1b]66;;X\x1b\\",
			atX:       0,
			wantWidth: 1,
			wantBytes: "\x1b]66;;X\x1b\\",
		},
		{
			name:      "follows preceding char",
			input:     "A\x1b]66;s=2;X\x1b\\",
			atX:       1,
			wantWidth: 2,
			wantBytes: "\x1b]66;s=2;X\x1b\\",
		},
		{
			// Callers that emit OSC 66 followed directly by an SGR are
			// advised to use BEL as the terminator instead of ST, because
			// ST is ESC+\\ and the trailing ESC can get concatenated with
			// the following ESC[... into a single ambiguous blob for some
			// terminal parsers. The cell buffer must still recognise and
			// forward a BEL-terminated OSC 66 unchanged.
			name:      "BEL terminator",
			input:     "\x1b]66;s=2;X\x07",
			atX:       0,
			wantWidth: 2,
			wantBytes: "\x1b]66;s=2;X\x07",
		},
		{
			name:      "BEL terminator followed by SGR",
			input:     "\x1b]66;s=2;X\x07\x1b[31m",
			atX:       0,
			wantWidth: 2,
			wantBytes: "\x1b]66;s=2;X\x07",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scr := NewScreenBuffer(8, 1)
			NewStyledString(tc.input).Draw(scr, scr.Bounds())
			got := scr.CellAt(tc.atX, 0)
			if got == nil {
				t.Fatalf("no cell at (%d, 0)", tc.atX)
			}
			if got.Width != tc.wantWidth {
				t.Errorf("width: want %d, got %d", tc.wantWidth, got.Width)
			}
			if got.Content != tc.wantBytes {
				t.Errorf("content:\n  want %q\n  got  %q", tc.wantBytes, got.Content)
			}
		})
	}
}

// TestStyledStringKittyTextSizingRoundTrip verifies the Cell makes it all
// the way through the renderer, so a downstream terminal would actually see
// the OSC bytes that drive the scaling.
func TestStyledStringKittyTextSizingRoundTrip(t *testing.T) {
	const osc = "\x1b]66;s=2;X\x1b\\"
	scr := NewScreenBuffer(4, 1)
	NewStyledString(osc).Draw(scr, scr.Bounds())
	out := scr.Render()
	if !strings.Contains(out, osc) {
		t.Fatalf("OSC 66 escape not found in render output:\n  want substring %q\n  got           %q", osc, out)
	}
}

// TestStyledStringKittyTextSizingMultiRow asserts that a scaled glyph also
// reserves CUF placeholder cells in the rows it spills into, and that a
// trailing space in the input does not overwrite those placeholders.
func TestStyledStringKittyTextSizingMultiRow(t *testing.T) {
	// Layout: "Xose<NL>spaces" - the X is the OSC 66 anchor at scale 2;
	// the spaces line below should be partially short-circuited where
	// the spill cells sit.
	const osc = "\x1b]66;s=2;X\x1b\\"
	// Pad the icon line to fill 4 cells so the spill is aligned with
	// known columns; spacer line is 4 spaces.
	input := osc + "\n    "

	scr := NewScreenBuffer(4, 2)
	NewStyledString(input).Draw(scr, scr.Bounds())

	anchor := scr.CellAt(0, 0)
	if anchor == nil || anchor.Width != 2 || anchor.Content != osc {
		t.Fatalf("anchor cell wrong: %#v", anchor)
	}

	spill := scr.CellAt(0, 1)
	if spill == nil {
		t.Fatal("spill cell not set; expected CUF placeholder")
	}
	if spill.Width != 2 {
		t.Errorf("spill cell width: want 2, got %d", spill.Width)
	}
	wantCUF := "\x1b[2C"
	if spill.Content != wantCUF {
		t.Errorf("spill cell content: want %q, got %q", wantCUF, spill.Content)
	}

	// The spaces written to (0, 1) and (1, 1) should NOT have overwritten
	// the placeholder. (2, 1) and (3, 1) should be regular space cells.
	if got := scr.CellAt(2, 1); got == nil || got.Content != " " {
		t.Errorf("(2, 1) want space, got %#v", got)
	}
	if got := scr.CellAt(3, 1); got == nil || got.Content != " " {
		t.Errorf("(3, 1) want space, got %#v", got)
	}

	// The spill CUF cell must carry *no* style. Emitting an SGR change
	// adjacent to a multicell extension is what some kitty builds
	// mis-render as literal SGR text ("[38;2;...m"). A zero style means
	// renderLine emits at most a plain ResetStyle around the CUF.
	if !spill.Style.IsZero() {
		t.Errorf("spill cell style must be zero; got %#v", spill.Style)
	}
}

// TestStyledStringKittyTextSizingUTF8Continuation covers OSC 66 payloads
// whose UTF-8 text contains byte 0x9C (C1 ST). Nerd-Font icons like
// U+E738 (0xEE 0x9C 0xB8) land here; the shared ansi decoder treats the
// middle 0x9C as a C1 String Terminator and truncates the OSC payload
// mid-UTF-8, which causes kitty to see malformed bytes and render the
// tail of the emission as literal text. The fix in printString must
// recognise the pattern and rescan for the real BEL/ESC+\ terminator.
func TestStyledStringKittyTextSizingUTF8Continuation(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "BEL terminator, 0x9C continuation byte",
			input: "\x1b]66;s=2;\ue738\x07",
		},
		{
			name:  "ST terminator, 0x9C continuation byte",
			input: "\x1b]66;s=2;\ue738\x1b\\",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scr := NewScreenBuffer(4, 1)
			NewStyledString(tc.input).Draw(scr, scr.Bounds())
			cell := scr.CellAt(0, 0)
			if cell == nil {
				t.Fatal("no cell at (0, 0)")
			}
			if cell.Width != 2 {
				t.Errorf("width: want 2, got %d", cell.Width)
			}
			if cell.Content != tc.input {
				t.Errorf("content truncated mid-UTF-8:\n  want %q\n  got  %q", tc.input, cell.Content)
			}
		})
	}
}

// TestTerminalRendererForwardsKittyOSC66 mirrors the path bubbletea uses:
// View.Content goes into a styled string, the styled string draws into the
// cell buffer, and TerminalRenderer.Render flushes that buffer to the
// terminal. We assert the OSC bytes show up in the bytes the renderer
// would have written to the terminal.
func TestTerminalRendererForwardsKittyOSC66(t *testing.T) {
	const osc = "\x1b]66;s=2;X\x1b\\"

	// Two frames: empty first, then content. The renderer diffs against
	// the previous frame, so emitting on the very first paint requires a
	// non-empty frame after an initialised state.
	cellbuf := NewScreenBuffer(4, 1)
	NewStyledString(osc).Draw(cellbuf, cellbuf.Bounds())

	var out bytes.Buffer
	tr := NewTerminalRenderer(&out, nil)
	tr.Resize(4, 1)
	tr.Render(cellbuf.RenderBuffer)
	if err := tr.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, osc) {
		t.Fatalf("TerminalRenderer dropped OSC 66\n  want substring %q\n  got           %q", osc, got)
	}
}

func newWcCell(s string, style *Style, link *Link) Cell {
	c := NewCell(ansi.WcWidth, s)
	if style != nil {
		c.Style = *style
	}
	if link != nil {
		c.Link = *link
	}
	return *c
}
