package uv

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

func TestBufferUniseg(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty buffer",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "multiple lines",
			input:    "Hello, World!\nThis is a test.\nGoodbye!",
			expected: "Hello, World!\r\nThis is a test.\r\nGoodbye!",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := NewBuffer(width(c.input), height(c.input))
			for y, line := range strings.Split(c.input, "\n") {
				var x int
				seg := uniseg.NewGraphemes(line)
				for seg.Next() {
					cell := &Cell{
						Content: seg.Str(),
						Width:   seg.Width(),
					}
					buf.SetCell(x, y, cell)
					x += cell.Width
				}
			}

			if buf.String() != c.expected {
				t.Errorf("expected %q, got %q", c.expected, buf.String())
			}
		})
	}
}

func width(s string) int {
	width := 0
	for _, line := range strings.Split(s, "\n") {
		width = max(width, ansi.StringWidth(line))
	}
	return width
}

func height(s string) int {
	return strings.Count(s, "\n") + 1
}
