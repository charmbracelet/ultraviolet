package uv

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"unsafe"
)

// pointsInto reports whether the bytes of s lie inside those of frame.
func pointsInto(s, frame string) bool {
	if s == "" || frame == "" {
		return false
	}
	p := uintptr(unsafe.Pointer(unsafe.StringData(s)))
	lo := uintptr(unsafe.Pointer(unsafe.StringData(frame)))
	return p >= lo && p < lo+uintptr(len(frame))
}

// A cell drawn from a styled string holds a substring of that string, and the
// renderer keeps its model of the screen for as long as the screen shows it.
// A model cell that kept the substring would keep the whole frame alive until
// that cell changed again, so typing at the end of a line would keep one frame
// per key. Every way a row reaches the model is covered: a diffed row, a row
// whose last cell can't be erased with, a repainted row of wide cells, leading
// blanks erased with EL 1 and a tail erased with EL 0.
func TestRendererModelDoesNotKeepFrames(t *testing.T) {
	frames := []string{
		"hello     \nabcdefghij\n世界ab    \nabcdefghij\nabcdefghij\nhé        ",
		"hello!    \nabcdeXghij\n世界abc   \n        ij\na         \nhéé       ",
		"hello!?   \nabcdeYghij\n世界abcd  \n          \nab        \nhééé      ",
	}

	r := NewTerminalRenderer(io.Discard, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	r.Resize(10, 6)
	buf := NewScreenBuffer(10, 6)
	for i, frame := range frames {
		NewStyledString(frame).Draw(buf, buf.Bounds())
		r.Render(buf.RenderBuffer)
		if err := r.Flush(); err != nil {
			t.Fatalf("failed to flush renderer: %v", err)
		}
		if got, want := r.curbuf.String(), buf.String(); got != want {
			t.Fatalf("after frame %d the model is\n%q\nwant\n%q", i, got, want)
		}
		var kept int
		var first string
		for y := range r.curbuf.Height() {
			for x, c := range r.curbuf.Line(y) {
				for j, f := range frames[:i+1] {
					if pointsInto(c.Content, f) {
						if kept == 0 {
							first = fmt.Sprintf("cell %d,%d (%q) into frame %d", x, y, c.Content, j)
						}
						kept++
					}
				}
			}
		}
		if kept > 0 {
			t.Errorf("after frame %d, %d cells of the model point into a frame, the first %s", i, kept, first)
		}
	}
}

// typingFrames returns frames of a 200x50 editor in which a paragraph wrapped
// below a gutter grows by one character a frame, as when typing at its end,
// with a status bar that follows the cursor.
func typingFrames(n int) []string {
	const width, height, gutter = 200, 50, 7
	words := strings.Fields("the quick brown fox jumps over a lazy dog while every river runs down to the sea")
	var para strings.Builder
	for i := 0; para.Len() < 1500+n; i++ {
		para.WriteString(words[i%len(words)])
		para.WriteByte(' ')
	}
	text := para.String()

	frames := make([]string, n)
	var b strings.Builder
	for k := range n {
		b.Reset()
		typed := text[:1500+k]
		for y := range height - 1 {
			var row string
			if lo := y * (width - gutter); lo < len(typed) {
				row = typed[lo:min(lo+width-gutter, len(typed))]
			}
			fmt.Fprintf(&b, "\x1b[38;5;244m%4d │\x1b[m %s%s\n", y+1, row, strings.Repeat(" ", width-gutter-len(row)))
		}
		status := fmt.Sprintf(" NORMAL │ p.md … ln 1, col %d", len(typed))
		fmt.Fprintf(&b, "\x1b[7m%s%s\x1b[m", status, strings.Repeat(" ", width-len([]rune(status))))
		frames[k] = b.String()
	}
	return frames
}

// One key typed at the end of a wrapped paragraph on a 200x50 screen: the
// frame is parsed, diffed and written, as a program built on the renderer does
// for every key.
func BenchmarkRenderTyping(b *testing.B) {
	frames := typingFrames(256)
	r := NewTerminalRenderer(io.Discard, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	r.Resize(200, 50)
	buf := NewScreenBuffer(200, 50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewStyledString(frames[i%len(frames)]).Draw(buf, buf.Bounds())
		r.Render(buf.RenderBuffer)
		if err := r.Flush(); err != nil {
			b.Fatalf("failed to flush renderer: %v", err)
		}
	}
}
