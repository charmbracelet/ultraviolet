package uv

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// renderRows renders rows the way Bubble Tea does each frame: Clear (which
// touches every line), Draw, Render, Flush. It returns the bytes written.
func renderRows(t *testing.T, r *TerminalRenderer, out *bytes.Buffer, buf *ScreenBuffer, rows []string) int {
	t.Helper()
	before := out.Len()
	buf.Clear()
	NewStyledString(strings.Join(rows, "\n")).Draw(buf, buf.Bounds())
	r.Render(buf.RenderBuffer)
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	return out.Len() - before
}

func wideRows(glyph string, n int) []string {
	var rows []string
	for i := 0; i < 30; i++ {
		rows = append(rows, fmt.Sprintf("row %02d static text %s", i, glyph))
	}
	return append(rows, fmt.Sprintf("counter %d", n))
}

// Reported in https://github.com/charmbracelet/ultraviolet/issues/210, from
// which this test is adapted.
func TestUnchangedWideRowsAreNotRepainted(t *testing.T) {
	bytesPerFrame := func(glyph string) int {
		var out bytes.Buffer
		r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
		r.SetFullscreen(true)
		r.Resize(80, 31)
		buf := NewScreenBuffer(80, 31)

		renderRows(t, r, &out, &buf, wideRows(glyph, 0))
		const frames = 10
		var total int
		for n := 1; n <= frames; n++ {
			total += renderRows(t, r, &out, &buf, wideRows(glyph, n))
		}
		return total / frames
	}

	ascii := bytesPerFrame("")
	for _, g := range []string{"日本語", "👍"} {
		got := bytesPerFrame(g)
		t.Logf("%q: %d bytes/frame (ascii: %d)", g, got, ascii)
		if got > ascii+100 {
			t.Errorf("%q on every row: %d bytes per frame against %d for ASCII; unchanged rows are repainted", g, got, ascii)
		}
	}
}

func TestChangedWideRowIsRepainted(t *testing.T) {
	var out bytes.Buffer
	r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	r.Resize(80, 31)
	buf := NewScreenBuffer(80, 31)

	rows := wideRows("日本語", 0)
	renderRows(t, r, &out, &buf, rows)

	rows[5] = "row 05 changed text 日本"
	before := out.Len()
	renderRows(t, r, &out, &buf, rows)
	if written := out.String()[before:]; !strings.Contains(written, "changed text 日本") {
		t.Errorf("changed wide row was not repainted, wrote %q", written)
	}
}
