package uv

import (
	"bytes"
	"testing"
)

// TestRendererInlinePartialUpdateEmitsCRLF is a regression test for
// https://github.com/charmbracelet/ultraviolet/issues/61.
//
// In inline (non-fullscreen) mode, when a single frame touches two
// non-adjacent rows (e.g. redrawing a box near the top of the screen and
// another box further down, leaving the rows between them untouched), the
// renderer moves the cursor down to the next touched row using bare line
// feeds ("\n") as a cost optimization in [relativeCursorMove]. Programs
// using this renderer run the terminal in raw mode, which disables the
// OS-level LF->CRLF translation a cooked terminal performs, so every line
// feed used to move the cursor down must be paired with a carriage return.
// Otherwise the terminal's actual cursor column drifts from what the
// renderer assumes it is, corrupting the following in-place update.
func TestRendererInlinePartialUpdateEmitsCRLF(t *testing.T) {
	var buf bytes.Buffer
	s := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})

	// Match the defaults [TerminalScreen] uses for inline rendering: not
	// fullscreen, relative cursor tracking, and no newline auto-mapping
	// (raw mode, see terminal_screen.go).
	s.SetFullscreen(false)
	s.SetRelativeCursor(true)
	s.SetMapNewline(false)

	const width, height = 10, 6
	scr := NewScreenBuffer(width, height)

	// Draw an initial full frame occupying every row, similar to the
	// 3-layer canvas in the issue rendering a full screen of boxes.
	for y := range height {
		for x := range width {
			scr.SetCell(x, y, NewCell(scr.WidthMethod(), "x"))
		}
	}
	s.Render(scr.RenderBuffer)
	if err := s.Flush(); err != nil {
		t.Fatalf("initial Flush failed: %v", err)
	}

	// Mutate two non-adjacent rows in a single frame: row 0 (the "top box")
	// and row 3 (the "bottom-left box"), leaving rows 1-2 untouched. This
	// forces the renderer to jump the cursor down multiple rows in one
	// move, exercising the newline fast path in [relativeCursorMove].
	buf.Reset()
	scr.SetCell(5, 0, NewCell(scr.WidthMethod(), "A"))
	scr.SetCell(0, 3, NewCell(scr.WidthMethod(), "B"))

	s.Render(scr.RenderBuffer)
	if err := s.Flush(); err != nil {
		t.Fatalf("partial update Flush failed: %v", err)
	}

	out := buf.Bytes()
	if !bytes.Contains(out, []byte("\n")) {
		t.Fatalf("expected the partial update to move the cursor across rows, got no line feed in %q", out)
	}
	for i, b := range out {
		if b == '\n' && (i == 0 || out[i-1] != '\r') {
			t.Fatalf("bare LF without a preceding CR at output byte %d: %q", i, out)
		}
	}
}
