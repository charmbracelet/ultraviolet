package uv

import (
	"bytes"
	"strings"
	"testing"
)

func frameOf(width int, rows ...string) *RenderBuffer {
	b := NewRenderBuffer(width, len(rows))
	for y, row := range rows {
		for x, r := range row {
			b.SetCell(x, y, &Cell{Content: string(r), Width: 1})
		}
	}
	return b
}

func TestRendererScrollUpPaintsOnlyRowsScrolledIn(t *testing.T) {
	var out bytes.Buffer
	r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	r.SetRelativeCursor(true)
	r.Resize(10, 5)
	r.Render(frameOf(10, "r0", "r1", "r2", "r3", "r4"))
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	out.Reset()

	if !r.ScrollUp(2) {
		t.Fatal("a frame spanning the terminal must scroll")
	}
	r.Render(frameOf(10, "r2", "r3", "r4", "n0", "n1"))
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "\n\n") || !strings.Contains(got, "n0") || !strings.Contains(got, "n1") {
		t.Fatalf("scroll did not feed two lines and paint the new rows: %q", got)
	}
	for _, kept := range []string{"r2", "r3", "r4"} {
		if strings.Contains(got, kept) {
			t.Fatalf("row %s already on screen was repainted: %q", kept, got)
		}
	}
	if strings.Contains(got, "\x1b[J") || strings.Contains(got, "\x1b[2J") {
		t.Fatalf("scroll erased the screen: %q", got)
	}
}

func TestRendererScrollUpRefusesFrameShorterThanTerminal(t *testing.T) {
	var out bytes.Buffer
	r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	r.SetRelativeCursor(true)
	r.Resize(10, 5)
	r.Render(frameOf(10, "r0", "r1", "r2"))
	if err := r.Flush(); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if r.ScrollUp(1) {
		t.Fatal("a line feed below a short frame moves the cursor, it does not scroll")
	}
	if err := r.Flush(); err != nil || out.Len() != 0 {
		t.Fatalf("refused scroll wrote %q (err %v)", out.String(), err)
	}
	r.Resize(10, 3)
	if r.ScrollUp(1) {
		t.Fatal("scroll must wait for the next render to reconcile a resize")
	}
}

// Where a line feed would not scroll a known frame, nothing is written.
func TestRendererScrollUpRefusesWithoutAKnownFullHeightFrame(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(r *TerminalRenderer)
		n     int
	}{
		{"before the first render", func(r *TerminalRenderer) {}, 1},
		{"fullscreen", func(r *TerminalRenderer) {
			r.SetFullscreen(true)
			r.Render(frameOf(10, "r0", "r1", "r2", "r3", "r4"))
		}, 1},
		{"no rows", func(r *TerminalRenderer) {
			r.Render(frameOf(10, "r0", "r1", "r2", "r3", "r4"))
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
			r.SetRelativeCursor(true)
			r.Resize(10, 5)
			tc.setup(r)
			if err := r.Flush(); err != nil {
				t.Fatal(err)
			}
			out.Reset()
			if r.ScrollUp(tc.n) {
				t.Fatal("scrolled")
			}
			if err := r.Flush(); err != nil || out.Len() != 0 {
				t.Fatalf("refused scroll wrote %q (err %v)", out.String(), err)
			}
		})
	}
}
