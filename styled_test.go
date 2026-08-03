package uv

import (
	"image/color"
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

func TestStyledStringHeight(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected int
	}{
		{"single line", "hello", 1},
		{"two lines", "ab\ncd", 2},
		{"empty", "", 1},
		{"leading newline", "\nhello", 2},
		{"trailing newline", "hello\n", 2},
		{"multiple consecutive newlines", "a\n\nb", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			got := ss.Height()
			if got != tc.expected {
				t.Errorf("Height() = %d, want %d", got, tc.expected)
			}
		})
	}
}

func TestStyledStringBounds(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantW    int
		wantH    int
	}{
		{"single line", "hello", 5, 1},
		{"two lines same width", "abc\ndef", 3, 2},
		{"different widths", "hi\nhello", 5, 2},
		{"with ANSI", "\x1b[31mhi\x1b[0m", 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			b := ss.Bounds()
			if b.Dx() != tc.wantW {
				t.Errorf("Bounds().Dx() = %d, want %d", b.Dx(), tc.wantW)
			}
			if b.Dy() != tc.wantH {
				t.Errorf("Bounds().Dy() = %d, want %d", b.Dy(), tc.wantH)
			}
		})
	}
}

func TestStyledStringWidthMethods(t *testing.T) {
	cases := []struct {
		name  string
		input string
		wantW int
	}{
		{"simple", "hello", 5},
		{"with CJK", "a世b", 4},
	}
	for _, tc := range cases {
		t.Run(tc.name+" UnicodeWidth", func(t *testing.T) {
			ss := NewStyledString(tc.input)
			got := ss.UnicodeWidth()
			if got != tc.wantW {
				t.Errorf("UnicodeWidth() = %d, want %d", got, tc.wantW)
			}
		})
		t.Run(tc.name+" WcWidth", func(t *testing.T) {
			ss := NewStyledString(tc.input)
			got := ss.WcWidth()
			if got != tc.wantW {
				t.Errorf("WcWidth() = %d, want %d", got, tc.wantW)
			}
		})
	}
}

func TestStyledStringString(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"plain", "hello"},
		{"with ANSI", "\x1b[31mhi\x1b[0m"},
		{"multi-line", "a\nb"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			got := ss.String()
			if got != tc.input {
				t.Errorf("String() = %q, want %q", got, tc.input)
			}
		})
	}
}

func TestStyledStringLines(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		method   ansi.Method
		expected []Line
	}{
		{
			name:   "single line plain text",
			input:  "Hello",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("H", nil, nil),
					newWcCell("e", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("o", nil, nil),
				},
			},
		},
		{
			name:   "multiple lines",
			input:  "ab\ncd",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("a", nil, nil),
					newWcCell("b", nil, nil),
				},
				{
					newWcCell("c", nil, nil),
					newWcCell("d", nil, nil),
				},
			},
		},
		{
			name:     "empty string",
			input:    "",
			method:   ansi.WcWidth,
			expected: []Line{},
		},
		{
			name:   "ansi styled single line",
			input:  "\x1b[31;1mHi\x1b[0m",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("H", &Style{Fg: ansi.Red, Attrs: AttrBold}, nil),
					newWcCell("i", &Style{Fg: ansi.Red, Attrs: AttrBold}, nil),
				},
			},
		},
		{
			name:   "ansi across multiple lines",
			input:  "\x1b[31mLine1\nLine2\x1b[0m",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("L", &Style{Fg: ansi.Red}, nil),
					newWcCell("i", &Style{Fg: ansi.Red}, nil),
					newWcCell("n", &Style{Fg: ansi.Red}, nil),
					newWcCell("e", &Style{Fg: ansi.Red}, nil),
					newWcCell("1", &Style{Fg: ansi.Red}, nil),
				},
				{
					newWcCell("L", &Style{Fg: ansi.Red}, nil),
					newWcCell("i", &Style{Fg: ansi.Red}, nil),
					newWcCell("n", &Style{Fg: ansi.Red}, nil),
					newWcCell("e", &Style{Fg: ansi.Red}, nil),
					newWcCell("2", &Style{Fg: ansi.Red}, nil),
				},
			},
		},
		{
			name:   "hyperlink",
			input:  "\x1b]8;;https://charm.sh\x1b\\C\x1b]8;;\x1b\\",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("C", nil, &Link{URL: "https://charm.sh"}),
				},
			},
		},
		{
			name:   "mixed styling with hyperlinks",
			input:  "\x1b[31mR\x1b]8;;https://charm.sh\x1b\\\x1b[4mC\x1b]8;;\x1b\\\x1b[0m",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("R", &Style{Fg: ansi.Red}, nil),
					newWcCell("C", &Style{Fg: ansi.Red, Underline: UnderlineStyleSingle}, &Link{URL: "https://charm.sh"}),
				},
			},
		},
		{
			name:   "unicode characters",
			input:  "a世b",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("a", nil, nil),
					newWcCell("世", nil, nil),
					newWcCell("b", nil, nil),
				},
			},
		},
		{
			name:   "grapheme width method",
			input:  "Hi",
			method: ansi.GraphemeWidth,
			expected: []Line{
				{
					{Content: "H", Width: 1},
					{Content: "i", Width: 1},
				},
			},
		},
		{
			name:   "leading newline",
			input:  "\nhello",
			method: ansi.WcWidth,
			expected: []Line{
				{},
				{
					newWcCell("h", nil, nil),
					newWcCell("e", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("o", nil, nil),
				},
			},
		},
		{
			name:   "trailing newline",
			input:  "hello\n",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("h", nil, nil),
					newWcCell("e", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("l", nil, nil),
					newWcCell("o", nil, nil),
				},
				{},
			},
		},
		{
			name:   "multiple consecutive newlines",
			input:  "a\n\nb",
			method: ansi.WcWidth,
			expected: []Line{
				{
					newWcCell("a", nil, nil),
				},
				{},
				{
					newWcCell("b", nil, nil),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			got := ss.Lines(tc.method)

			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d lines, got %d", len(tc.expected), len(got))
			}
			for y := range tc.expected {
				if len(got[y]) != len(tc.expected[y]) {
					t.Fatalf("line %d: expected %d cells, got %d", y, len(tc.expected[y]), len(got[y]))
				}
				for x := range tc.expected[y] {
					gotCell := &got[y][x]
					wantCell := &tc.expected[y][x]
					if !cellEqual(wantCell, gotCell) {
						t.Errorf("cell (%d, %d): expected %#v, got %#v", x, y, wantCell, gotCell)
					}
				}
			}
		})
	}
}
