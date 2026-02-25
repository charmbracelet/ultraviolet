package uv

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWrap(t *testing.T) {
	c := func(s string, w int) Cell { return Cell{Content: s, Width: w} }
	C := func(s string) Cell { return c(s, 1) }
	W := func(s string) Cell { return c(s, 2) }
	P := Cell{} // wide-char placeholder

	t.Run("nil input", func(t *testing.T) {
		result := wrap(t, nil, 10, "")
		if result != nil {
			t.Errorf("got %v, want nil", result)
		}
	})
	t.Run("empty slice input", func(t *testing.T) {
		result := wrap(t, []Line{}, 10, "")
		if result != nil {
			t.Errorf("got %v, want nil", result)
		}
	})
	t.Run("empty line", func(t *testing.T) {
		result := wrap(t, []Line{{}}, 10, "")
		want := []Line{{}}
		if !linesEqual(result, want) {
			t.Errorf("got %v, want %v", result, want)
		}
	})

	tests := []struct {
		name   string
		input  string
		width  int
		breaks string
		want   []Line
	}{
		{
			name:  "zero width",
			input: "hello",
			width: 0,
			want:  []Line{{C("h"), C("e"), C("l"), C("l"), C("o")}},
		},
		{
			name:  "negative width",
			input: "hello",
			width: -1,
			want:  []Line{{C("h"), C("e"), C("l"), C("l"), C("o")}},
		},
		{
			name:  "no wrap needed",
			input: "hello",
			width: 10,
			want:  []Line{{C("h"), C("e"), C("l"), C("l"), C("o")}},
		},
		{
			name:  "simple word wrap",
			input: "hello world",
			width: 6,
			want: []Line{
				{C("h"), C("e"), C("l"), C("l"), C("o")},
				{C("w"), C("o"), C("r"), C("l"), C("d")},
			},
		},
		{
			name:  "multiple words",
			input: "aa bb cc dd",
			width: 5,
			want: []Line{
				{C("a"), C("a"), C(" "), C("b"), C("b")},
				{C("c"), C("c"), C(" "), C("d"), C("d")},
			},
		},
		{
			name:  "hard wrap",
			input: "abcdefghij",
			width: 4,
			want: []Line{
				{C("a"), C("b"), C("c"), C("d")},
				{C("e"), C("f"), C("g"), C("h")},
				{C("i"), C("j")},
			},
		},
		{
			name:  "dash breakpoint",
			input: "self-aware",
			width: 6,
			want: []Line{
				{C("s"), C("e"), C("l"), C("f"), C("-")},
				{C("a"), C("w"), C("a"), C("r"), C("e")},
			},
		},
		{
			name:   "extra breakpoints",
			input:  "path/to/file",
			width:  6,
			breaks: "/",
			want: []Line{
				{C("p"), C("a"), C("t"), C("h"), C("/")},
				{C("t"), C("o"), C("/")},
				{C("f"), C("i"), C("l"), C("e")},
			},
		},
		{
			name:  "preserve line breaks",
			input: "hello\nworld",
			width: 10,
			want: []Line{
				{C("h"), C("e"), C("l"), C("l"), C("o")},
				{C("w"), C("o"), C("r"), C("l"), C("d")},
			},
		},
		{
			name:  "preserve leading whitespace",
			input: "  hello",
			width: 10,
			want:  []Line{{C(" "), C(" "), C("h"), C("e"), C("l"), C("l"), C("o")}},
		},
		{
			name:  "trailing whitespace trimmed by default",
			input: "hi   ",
			width: 10,
			want:  []Line{{C("h"), C("i")}},
		},
		{
			name:  "wide characters fit",
			input: "你好",
			width: 4,
			want:  []Line{{W("你"), P, W("好"), P}},
		},
		{
			name:  "wide char at boundary",
			input: "a你",
			width: 2,
			want: []Line{
				{C("a")},
				{W("你"), P},
			},
		},
		{
			name:  "tab character",
			input: "hello\tworld",
			width: 6,
			want: []Line{
				{C("h"), C("e"), C("l"), C("l"), C("o")},
				{},
				{C("w"), C("o"), C("r"), C("l"), C("d")},
			},
		},
		{
			name:  "mixed word wrap and hard wrap",
			input: "hi abcdefgh",
			width: 5,
			want: []Line{
				{C("h"), C("i")},
				{C("a"), C("b"), C("c"), C("d"), C("e")},
				{C("f"), C("g"), C("h")},
			},
		},
		{
			name:  "emoji",
			input: "👋hi",
			width: 4,
			want:  []Line{{W("👋"), P, C("h"), C("i")}},
		},
		{
			name:  "multiple dashes",
			input: "a--b",
			width: 3,
			want: []Line{
				{C("a"), C("-"), C("-")},
				{C("b")},
			},
		},
		{
			name:  "only whitespace trimmed",
			input: "     ",
			width: 3,
			want: []Line{
				{},
				{},
			},
		},
		{
			name:  "wide placeholder handling",
			input: "你好世界",
			width: 4,
			want: []Line{
				{W("你"), P, W("好"), P},
				{W("世"), P, W("界"), P},
			},
		},
		{
			name:  "long string ",
			input: "the quick brown foxxxxxxxxxxxxxxxx jumped over the lazy dog.",
			width: 16,
			want: []Line{
				{C("t"), C("h"), C("e"), C(" "), C("q"), C("u"), C("i"), C("c"), C("k"), C(" "), C("b"), C("r"), C("o"), C("w"), C("n")},
				{C("f"), C("o"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x"), C("x")},
				{C("x"), C("x"), C(" "), C("j"), C("u"), C("m"), C("p"), C("e"), C("d"), C(" "), C("o"), C("v"), C("e"), C("r")},
				{C("t"), C("h"), C("e"), C(" "), C("l"), C("a"), C("z"), C("y"), C(" "), C("d"), C("o"), C("g"), C(".")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Lines(tt.input, ansi.GraphemeWidth)
			result := wrap(t, lines, tt.width, tt.breaks)
			if !linesEqual(result, tt.want) {
				t.Errorf("got  \n%v\nwant \n%v", linesString(result), linesString(tt.want))
			}
		})
	}
}

func TestLines_EmptyInput(t *testing.T) {
	tests := []struct {
		name string
		wm   WidthMethod
	}{
		{"GraphemeWidth", ansi.GraphemeWidth},
		{"WcWidth", ansi.WcWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test empty string
			result := Lines("", tt.wm)
			if result != nil {
				t.Errorf("Lines(\"\") = %v, want nil", result)
			}

			// Test empty []byte
			result = Lines([]byte{}, tt.wm)
			if result != nil {
				t.Errorf("Lines([]byte{}) = %v, want nil", result)
			}
		})
	}
}

func TestLines_SingleLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wm    WidthMethod
		want  []Line
	}{
		{
			name:  "simple ASCII",
			input: "hello",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "h", Width: 1},
					{Content: "e", Width: 1},
					{Content: "l", Width: 1},
					{Content: "l", Width: 1},
					{Content: "o", Width: 1},
				},
			},
		},
		{
			name:  "wide characters (CJK)",
			input: "你好",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "你", Width: 2},
					{Content: "", Width: 0}, // placeholder
					{Content: "好", Width: 2},
					{Content: "", Width: 0}, // placeholder
				},
			},
		},
		{
			name:  "emoji",
			input: "👋🌍",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "👋", Width: 2},
					{Content: "", Width: 0}, // placeholder
					{Content: "🌍", Width: 2},
					{Content: "", Width: 0}, // placeholder
				},
			},
		},
		{
			name:  "emoji with ZWJ",
			input: "👨‍👩‍👧",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "👨‍👩‍👧", Width: 2},
					{Content: "", Width: 0}, // placeholder
				},
			},
		},
		{
			name:  "multi codepoint emoji with wcwidth",
			input: "🇸🇦",
			wm:    ansi.WcWidth,
			want: []Line{
				{
					{Content: "🇸🇦", Width: 1},
				},
			},
		},
		{
			name:  "multi codepoint emoji with grapheme width",
			input: "🇸🇦",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "🇸🇦", Width: 2},
					{Content: "", Width: 0}, // placeholder
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lines(tt.input, tt.wm)
			if !linesEqual(result, tt.want) {
				t.Errorf("Lines(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestLines_MultipleLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wm    WidthMethod
		want  []Line
	}{
		{
			name:  "two lines with LF",
			input: "ab\ncd",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "a", Width: 1},
					{Content: "b", Width: 1},
				},
				{
					{Content: "c", Width: 1},
					{Content: "d", Width: 1},
				},
			},
		},
		{
			name:  "two lines with CRLF",
			input: "ab\r\ncd",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "a", Width: 1},
					{Content: "b", Width: 1},
				},
				{
					{Content: "c", Width: 1},
					{Content: "d", Width: 1},
				},
			},
		},
		{
			name:  "trailing newline",
			input: "ab\n",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "a", Width: 1},
					{Content: "b", Width: 1},
				},
				{},
			},
		},
		{
			name:  "multiple empty lines",
			input: "a\n\nb",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{
					{Content: "a", Width: 1},
				},
				{},
				{
					{Content: "b", Width: 1},
				},
			},
		},
		{
			name:  "only newlines",
			input: "\n\n",
			wm:    ansi.GraphemeWidth,
			want: []Line{
				{},
				{},
				{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lines(tt.input, tt.wm)
			if !linesEqual(result, tt.want) {
				t.Errorf("Lines(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestLines_CarriageReturn(t *testing.T) {
	// Standalone \r should be skipped
	input := "a\rb"
	result := Lines(input, ansi.GraphemeWidth)
	want := []Line{
		{
			{Content: "a", Width: 1},
			{Content: "b", Width: 1},
		},
	}

	if !linesEqual(result, want) {
		t.Errorf("Lines(%q) = %v, want %v", input, result, want)
	}
}

func TestLines_ByteSlice(t *testing.T) {
	input := []byte("hello\nworld")
	result := Lines(input, ansi.GraphemeWidth)
	want := []Line{
		{
			{Content: "h", Width: 1},
			{Content: "e", Width: 1},
			{Content: "l", Width: 1},
			{Content: "l", Width: 1},
			{Content: "o", Width: 1},
		},
		{
			{Content: "w", Width: 1},
			{Content: "o", Width: 1},
			{Content: "r", Width: 1},
			{Content: "l", Width: 1},
			{Content: "d", Width: 1},
		},
	}

	if !linesEqual(result, want) {
		t.Errorf("Lines(%q) = %v, want %v", input, result, want)
	}
}

func TestLines_WcWidth(t *testing.T) {
	// Test that WcWidth method is used correctly
	// "你好" = 2 wide chars, each with width 2, so 4 cells total (2 content + 2 placeholder)
	input := "你好"
	result := Lines(input, ansi.WcWidth)

	if len(result) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result))
	}

	// 2 graphemes * 2 width = 4 cells (including placeholders)
	if len(result[0]) != 4 {
		t.Fatalf("expected 4 cells, got %d", len(result[0]))
	}

	// First cell should be content with width 2
	if result[0][0].Width != 2 {
		t.Errorf("cell[0].Width = %d, want 2", result[0][0].Width)
	}
	// Second cell should be placeholder
	if result[0][1].Width != 0 {
		t.Errorf("cell[1].Width = %d, want 0 (placeholder)", result[0][1].Width)
	}
}

func TestLines_GraphemeClusters(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCells int // total cells including placeholders
	}{
		{
			name:      "combining character",
			input:     "e\u0301", // e + combining acute accent (´)
			wantCells: 1,         // width 1, no placeholder
		},
		{
			name:      "flag emoji",
			input:     "🇺🇸",
			wantCells: 2, // width 2, 1 content + 1 placeholder
		},
		{
			name:      "skin tone modifier",
			input:     "👋🏽",
			wantCells: 2, // width 2, 1 content + 1 placeholder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lines(tt.input, ansi.GraphemeWidth)
			if len(result) != 1 {
				t.Fatalf("expected 1 line, got %d", len(result))
			}
			if len(result[0]) != tt.wantCells {
				t.Errorf("got %d cells, want %d", len(result[0]), tt.wantCells)
			}
		})
	}
}

func TestLines_LenEqualsDisplayWidth(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantWidth int
	}{
		{"ASCII", "hello", 5},
		{"CJK", "你好", 4},
		{"Mixed", "hi你", 4},
		{"Emoji", "👋", 2},
		{"ZWJ emoji", "👨‍👩‍👧", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lines(tt.input, ansi.GraphemeWidth)
			if len(result) != 1 {
				t.Fatalf("expected 1 line, got %d", len(result))
			}
			if len(result[0]) != tt.wantWidth {
				t.Errorf("len(line) = %d, want %d (display width)", len(result[0]), tt.wantWidth)
			}
		})
	}
}

func wrap(t *testing.T, lines []Line, width int, extraBreakpoints string) []Line {
	t.Helper()
	var breakpoints []string
	for _, r := range extraBreakpoints {
		breakpoints = append(breakpoints, string(r))
	}
	lw := NewWrapper(breakpoints...)
	return lw.Wrap(lines, width)
}

func TestWrapPreserveSpaces(t *testing.T) {
	C := func(s string) Cell { return Cell{Content: s, Width: 1} }

	tests := []struct {
		name  string
		input string
		width int
		want  []Line
	}{
		{
			name:  "trailing spaces preserved",
			input: "hi   ",
			width: 10,
			want:  []Line{{C("h"), C("i"), C(" "), C(" "), C(" ")}},
		},
		{
			name:  "trailing space at wrap boundary preserved",
			input: "hello world",
			width: 6,
			want: []Line{
				{C("h"), C("e"), C("l"), C("l"), C("o"), C(" ")},
				{C("w"), C("o"), C("r"), C("l"), C("d")},
			},
		},
		{
			name:  "only whitespace preserved",
			input: "     ",
			width: 3,
			want: []Line{
				{C(" "), C(" "), C(" ")},
				{C(" ")},
			},
		},
		{
			name:  "mixed wrap trailing space preserved",
			input: "hi abcdefgh",
			width: 5,
			want: []Line{
				{C("h"), C("i"), C(" ")},
				{C("a"), C("b"), C("c"), C("d"), C("e")},
				{C("f"), C("g"), C("h")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Lines(tt.input, ansi.GraphemeWidth)
			lw := NewWrapper()
			lw.PreserveSpaces = true
			result := lw.Wrap(lines, tt.width)
			if !linesEqual(result, tt.want) {
				t.Errorf("got  %v\nwant %v", result, tt.want)
			}
		})
	}
}

// linesEqual compares two [Line] slices for equality.
func linesEqual(a, b []Line) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// linesString returns a string representation of a slice of [Line]s for debugging.
func linesString(lines []Line) string {
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString("[")
		sb.WriteString(line.String())
		sb.WriteString("]\n")
	}
	return sb.String()
}
