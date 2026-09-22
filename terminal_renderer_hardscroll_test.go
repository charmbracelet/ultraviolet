package uv

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
)

// setTestFrame draws one string per row into buf, padding every row with
// spaces so each cell holds explicit content.
func setTestFrame(buf *RenderBuffer, lines []string, width, height int) {
	for y := range height {
		line := ""
		if y < len(lines) {
			line = lines[y]
		}
		for x := range width {
			content := " "
			if x < len(line) {
				content = string(line[x])
			}
			buf.SetCell(x, y, &Cell{Content: content, Width: 1})
		}
	}
}

// bufferRows returns the buffer's rows as strings with trailing blanks
// trimmed.
func bufferRows(buf *Buffer) []string {
	rows := make([]string, buf.Height())
	for y := 0; y < buf.Height(); y++ {
		var sb strings.Builder
		for x := 0; x < buf.Width(); x++ {
			if c := buf.CellAt(x, y); c != nil {
				sb.WriteString(c.Content)
			} else {
				sb.WriteByte(' ')
			}
		}
		rows[y] = strings.TrimRight(sb.String(), " ")
	}
	return rows
}

// TestRendererScrollOptimUnchangedRows is a regression test for
// https://github.com/charmbracelet/ultraviolet/issues/137. When
// scrollOptimize performs a hardware scroll, rows inside the scrolled region
// whose newbuf content is identical to the previous frame at the same index
// must still be repainted: the physical screen (and curbuf) changed beneath
// them. After every render, curbuf — the renderer's model of the physical
// screen — must match the rendered frame; a stale row shows up as a curbuf
// row still holding the scrolled-in content.
func TestRendererScrollOptimUnchangedRows(t *testing.T) {
	const width = 20
	alphabet := []string{"case-a", "case-b", "case-c", "skill-x", "skill-y", "", "rule----"}
	rng := rand.New(rand.NewSource(1))

	trials := 0
	for trial := range 500 {
		height := 6 + rng.Intn(10)
		total := height + 1 + rng.Intn(20)
		list := make([]string, total)
		for i := range list {
			list[i] = alphabet[rng.Intn(len(alphabet))]
		}
		off := rng.Intn(total - height)
		shift := 1 + rng.Intn(3)
		if off+shift+height > total {
			continue
		}
		trials++

		r := NewTerminalRenderer(io.Discard, []string{"TERM=xterm-256color"})
		r.SetFullscreen(true)
		r.SetScrollOptim(true)
		r.Resize(width, height)

		buf := NewRenderBuffer(width, height)
		frames := [][]string{
			list[off : off+height],             // frame A
			list[off+shift : off+shift+height], // frame B: frame A scrolled up by shift
		}
		for f, frame := range frames {
			setTestFrame(buf, frame, width, height)
			r.Render(buf)
			if err := r.Flush(); err != nil {
				t.Fatalf("trial %d: failed to flush renderer: %v", trial, err)
			}

			got := bufferRows(r.curbuf.Buffer)
			want := make([]string, height)
			copy(want, frame)
			for y := range want {
				if got[y] != want[y] {
					t.Errorf("trial %d frame %d (h=%d shift=%d): stale row %d on physical screen:\ngot  %q\nwant %q",
						trial, f, height, shift, y, got[y], want[y])
				}
			}
		}
		if t.Failed() {
			break
		}
	}
	if trials == 0 {
		t.Fatal("no trials executed")
	}
}

// benchmarkRendererScroll renders successive frames of a list scrolled up by
// one line per frame, the render pattern of a scrolling viewport.
func benchmarkRendererScroll(b *testing.B, lines []string) {
	const width, height = 80, 24
	var written int64
	w := countingWriter{n: &written}

	r := NewTerminalRenderer(w, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	r.SetScrollOptim(true)
	r.Resize(width, height)

	buf := NewRenderBuffer(width, height)
	setTestFrame(buf, lines[:height], width, height)
	r.Render(buf)
	if err := r.Flush(); err != nil {
		b.Fatalf("failed to flush renderer: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		off := 1 + i%(len(lines)-height-1)
		setTestFrame(buf, lines[off:off+height], width, height)
		r.Render(buf)
		if err := r.Flush(); err != nil {
			b.Fatalf("failed to flush renderer: %v", err)
		}
	}
	b.ReportMetric(float64(written)/float64(b.N), "bytes/frame")
}

type countingWriter struct{ n *int64 }

func (w countingWriter) Write(p []byte) (int, error) {
	*w.n += int64(len(p))
	return len(p), nil
}

// BenchmarkRendererScrollUniqueRows scrolls through content where every row
// is distinct, so every row in the viewport changes on each frame.
func BenchmarkRendererScrollUniqueRows(b *testing.B) {
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = fmt.Sprintf("line-%04d: the quick brown fox jumps over the lazy dog", i)
	}
	benchmarkRendererScroll(b, lines)
}

// BenchmarkRendererScrollRepeatedRows scrolls through content drawn from a
// small alphabet so rows frequently repeat — the case from issue #137 where
// rows inside the scrolled region are identical across frames at the same
// index.
func BenchmarkRendererScrollRepeatedRows(b *testing.B) {
	alphabet := []string{"case-a", "case-b", "case-c", "skill-x", "skill-y", "", "rule----"}
	rng := rand.New(rand.NewSource(1))
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = alphabet[rng.Intn(len(alphabet))]
	}
	benchmarkRendererScroll(b, lines)
}
