package uv

import (
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
		{
			// A sequence that carries data rather than driving the cursor
			// rides in the content of the cell that follows it, so it reaches
			// the terminal when the cell is painted. See #95.
			name:           "APC sequence rides the next cell",
			input:          "\x1b_foo\x1b\\bar",
			expectedWidth:  3,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("\x1b_foo\x1b\\b", nil, nil),
						newWcCell("a", nil, nil),
						newWcCell("r", nil, nil),
					},
				},
			},
		},
		{
			// At the end of the string there is no following cell, so it folds
			// into the last one. A width-0 cell one past the content would sit
			// outside the bounds whenever the string filled them.
			name:           "trailing APC folds into the last cell",
			input:          "bar\x1b_foo\x1b\\",
			expectedWidth:  3,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("b", nil, nil),
						newWcCell("a", nil, nil),
						newWcCell("r\x1b_foo\x1b\\", nil, nil),
					},
				},
			},
		},
		{
			// Cursor movement is not carried. Replaying it from inside a cell
			// would move the real cursor somewhere the renderer's model cannot
			// see, so it stays dropped.
			name:           "cursor movement is not carried",
			input:          "a\x1b[5Cb",
			expectedWidth:  2,
			expectedHeight: 1,
			expected: &Buffer{
				Lines: []Line{
					{
						newWcCell("a", nil, nil),
						newWcCell("b", nil, nil),
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

func TestStyledStringLinesUnboundedContent(t *testing.T) {
	const destination = "https://example.com"
	input := "\x1b[31mab\x1b[0m\n" + ansi.SetHyperlink(destination, "") + "界" + ansi.ResetHyperlink()
	lines := NewStyledString(input).Lines(ansi.GraphemeWidth)

	if got := Lines(lines).String(); got != "ab\n界" {
		t.Fatalf("Lines().String() = %q, want %q", got, "ab\n界")
	}
	if cell := lines[1][0]; cell.Width != 2 || cell.Link.URL != destination {
		t.Fatalf("linked cell = %#v, want width 2 and link %q", cell, destination)
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

func TestReadLinkAllowsSemicolonsInURL(t *testing.T) {
	var link Link
	ReadLink([]byte("8;id=123;Other;Guide.md"), &link)

	if link.Params != "id=123" || link.URL != "Other;Guide.md" {
		t.Fatalf("ReadLink() = %#v, want params %q and URL %q", link, "id=123", "Other;Guide.md")
	}
}

// ASCII-heavy line: the common case. Guards the printString re-decode
// pre-filter (str[1] >= 0xc0) from regressing plain-text draws.
func BenchmarkPrintStringASCII(b *testing.B) {
	line := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 4)
	buf := NewScreenBuffer(200, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStyledString(line).Draw(buf, Rect(0, 0, 200, 1))
	}
}

// GraphemeWidth variant of the ASCII draw, since the decoder selection
// differs.
func BenchmarkPrintStringASCIIGrapheme(b *testing.B) {
	line := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 4)
	buf := NewScreenBuffer(200, 4)
	buf.Method = ansi.GraphemeWidth
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStyledString(line).Draw(buf, Rect(0, 0, 200, 1))
	}
}

// Cluster-heavy line: keycaps, VS16 emoji, combining marks. Exercises the
// FirstGraphemeCluster re-decode path.
func BenchmarkPrintStringClusters(b *testing.B) {
	line := strings.Repeat("1️⃣❤️é☹️", 20)
	buf := NewScreenBuffer(200, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStyledString(line).Draw(buf, Rect(0, 0, 200, 1))
	}
}

func TestStyledStringCombiningMarks(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"combining mark folds into its base", "a\u0301b", []string{"a\u0301", "b"}},
		{"keycap stays in one cell", "1\ufe0f\u20e3x", []string{"1\ufe0f\u20e3", "x"}},
		{"style change between cells", "a\u0301\x1b[31mR", []string{"a\u0301", "R"}},
		// The re-decode cannot fold this one, since the escape sequence sits
		// between the base and the mark. The mark still belongs to the cell
		// before it, and dropping it would lose the glyph.
		{"style change between base and mark", "a\x1b[31m\u0301b", []string{"a\u0301", "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewScreenBuffer(10, 1)
			NewStyledString(tc.in).Draw(buf, Rect(0, 0, 10, 1))
			for x, want := range tc.want {
				if got := buf.CellAt(x, 0).Content; got != want {
					t.Errorf("cell %d = %q, want %q", x, got, want)
				}
			}
		})
	}
}

// A string-type sequence is only carried when it actually ended, and only when
// it is one of the kinds a cell has any business replaying.
func TestPassThroughTerminators(t *testing.T) {
	for name, tc := range map[string]struct {
		input   string
		carried bool
	}{
		"APC ends with ST": {"\x1b_x\x1b\\a", true},
		"DCS ends with ST": {"\x1bP1$r0m\x1b\\a", true},
		"SOS ends with ST": {"\x1bXfoo\x1b\\a", true},
		"PM ends with ST":  {"\x1b^bar\x1b\\a", true},

		// The parser returns what it has when the input runs out. An
		// unterminated introducer in a cell would have the terminal swallow
		// everything painted after it.
		"APC never terminated": {"a\x1b_x", false},
		"OSC never terminated": {"a\x1b]0;t", false},
		"DCS never terminated": {"a\x1bP1$r0m", false},

		// 0x9c is a UTF-8 continuation byte as well as C1 ST, so a sequence
		// that ran out partway through a character can end in a byte that
		// looks like a terminator. U+071C encodes as dc 9c.
		"APC cut off mid-character": {"a\x1b_x\u071c", false},
		"OSC cut off mid-character": {"a\x1b]0;\u071c", false},
		"APC ends with C1 ST":       {"\x1b_x\x9ca", false},

		// An OSC carries data but a cell is painted again on every repaint. A
		// title survives that; a clipboard write or a notification should not
		// fire again on each resize.
		"OSC ends with BEL": {"\x1b]0;t\x07a", false},
		"OSC ends with ST":  {"\x1b]0;t\x1b\\a", false},
	} {
		t.Run(name, func(t *testing.T) {
			ss := NewStyledString(tc.input)
			area := ss.Bounds()
			buf := NewScreenBuffer(area.Dx(), area.Dy())
			ss.Draw(buf, area)

			var content string
			for x := 0; x < buf.Width(); x++ {
				if c := buf.CellAt(x, 0); c != nil {
					content += c.Content
				}
			}
			// Every input above introduces its sequence with ESC, so an ESC
			// surviving into the cells is the sequence riding along.
			if carried := strings.ContainsRune(content, ansi.ESC); carried != tc.carried {
				t.Errorf("carried = %v, want %v (cells hold %q)", carried, tc.carried, content)
			}
		})
	}
}

// A carried sequence must not change how wide its cell measures.
// [lineHasDrift] works the width out from the content, so a sequence that
// counted would have the renderer treat every line holding one as drift-prone
// and repaint it whole.
func TestPassThroughDoesNotAffectWidth(t *testing.T) {
	ss := NewStyledString("\x1b_x\x1b\\abc")
	area := ss.Bounds()
	buf := NewScreenBuffer(area.Dx(), area.Dy())
	ss.Draw(buf, area)

	if c := buf.CellAt(0, 0); c == nil || c.Width != 1 {
		t.Errorf("cell holding a sequence has width %v, want 1", c)
	}
	if lineHasDrift(ansi.WcWidth, buf.Line(0)) {
		t.Error("a line holding a pass-through sequence was flagged drift-prone")
	}
}
