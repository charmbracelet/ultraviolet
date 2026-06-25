package uv

import (
	"strings"
	"testing"
)

func TestDrawOverMatchesDraw(t *testing.T) {
	const w, h = 80, 24
	content := strings.Repeat("Hello, World! \x1b[1mBold\x1b[0m normal\n", h-1) + "last line"
	s := NewStyledString(content)
	area := Rect(0, 0, w, h)

	buf1 := NewScreenBuffer(w, h)
	s.Draw(buf1, area)

	buf2 := NewScreenBuffer(w, h)
	s.DrawOver(buf2, area)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c1 := buf1.CellAt(x, y)
			c2 := buf2.CellAt(x, y)
			if !cellEqual(c1, c2) {
				t.Errorf("cell mismatch at (%d, %d): Draw=%+v DrawOver=%+v", x, y, c1, c2)
			}
		}
	}
}

func TestDrawOverIncrementalTouch(t *testing.T) {
	const w, h = 80, 10
	content := strings.Repeat("Same line content here\n", h-1) + "Same line content here"
	s := NewStyledString(content)
	area := Rect(0, 0, w, h)

	buf := NewScreenBuffer(w, h)
	// Populate the buffer with the first draw.
	s.DrawOver(buf, area)

	// Reset touched tracking to simulate start of a new frame.
	buf.Touched = make([]*LineData, h)

	// DrawOver with identical content should not touch any lines.
	s.DrawOver(buf, area)

	touched := buf.TouchedLines()
	if touched != 0 {
		t.Errorf("DrawOver with identical content touched %d lines, expected 0", touched)
	}
}
