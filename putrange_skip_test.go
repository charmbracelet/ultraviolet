package uv

import (
	"bytes"
	"strings"
	"testing"
)

// TestRendererSkipsUnchangedRunInsideLine checks that a line which changes
// at both ends is written as two short pieces, with a cursor move over the
// unchanged cells between them.
//
// putRange is ncurses' PutRange. It is meant to jump over a run of unchanged
// cells once the run is longer than a cursor move costs. The check compared
// the run with the whole range (same > end-start), which a run inside the
// range can never pass, so every cell from the first change to the last was
// written again. A growing circle changes each row only where its edge
// crosses it, and the whole row between the two crossings went out on every
// frame.
func TestRendererSkipsUnchangedRunInsideLine(t *testing.T) {
	const w = 80
	middle := "\x1b[38;2;10;20;30m" + strings.Repeat("ab", (w-2)/2) + "\x1b[m"
	old := "X" + middle + "Y"
	next := "Z" + middle + "W"

	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color", "COLORTERM=truecolor"})
	r.SetFullscreen(true)
	r.SetRelativeCursor(false)
	r.Resize(w, 1)
	scr := NewScreenBuffer(w, 1)

	frame := func(s string) string {
		buf.Reset()
		scr.Clear()
		NewStyledString(s).Draw(scr, scr.Bounds())
		r.Render(scr.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("flush: %v", err)
		}
		return buf.String()
	}
	frame(old)
	out := frame(next)

	if strings.Contains(out, "abab") {
		t.Errorf("the unchanged run between the two changes was written again: %q", out)
	}
	if !strings.Contains(out, "Z") || !strings.Contains(out, "W") {
		t.Errorf("a changed cell is missing from the output: %q", out)
	}
	if len(out) > 32 {
		t.Errorf("two changed cells took %d bytes: %q", len(out), out)
	}
}
