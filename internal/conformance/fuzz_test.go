package conformance_test

import (
	"slices"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/conformance"
	"github.com/charmbracelet/x/ansi"
)

// runner drives one fuzz program against one emulator.
type runner struct {
	prog conformance.Program
	rend *uv.TerminalRenderer
	term oracle
	buf  uv.ScreenBuffer

	// w and h track the live dimensions, which an OpResize changes. The buffer,
	// the draw rectangle, and the screen readback all follow them so a line
	// drawn after a resize is valid for the current screen, not the starting
	// one.
	w, h int
}

// newRunner wires a renderer to a fresh emulator sized for the program. The
// renderer's width model and the emulator's come from the same field, so the two
// agree unless a test deliberately makes them disagree.
func newRunner(t *testing.T, p conformance.Program, mk func(*testing.T, int, int, bool) oracle) *runner {
	t.Helper()

	term := mk(t, p.Width, p.Height, p.GraphemeWidth)
	rend := uv.NewTerminalRenderer(term, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	rend.SetFullscreen(true)
	rend.SetGraphemeWidth(p.GraphemeWidth)
	rend.SetScrollOptim(p.ScrollOptim)
	rend.Erase()

	// The buffer and the renderer must measure width the same way, otherwise
	// the cells the renderer is asked to draw have widths it did not expect.
	// The emulator follows the same flag, so all three agree.
	buf := uv.NewScreenBuffer(p.Width, p.Height)
	if p.GraphemeWidth {
		buf.Method = ansi.GraphemeWidth
	}

	return &runner{
		prog: p,
		rend: rend,
		term: term,
		buf:  buf,
		w:    p.Width,
		h:    p.Height,
	}
}

func (r *runner) close() { r.term.Close() }

func (r *runner) flush(t *testing.T) {
	t.Helper()
	if err := r.rend.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

// step applies a single operation.
func (r *runner) step(t *testing.T, op conformance.Op) {
	t.Helper()

	switch op.Kind {
	case conformance.OpDrawLine:
		uv.NewStyledString(op.Text).Draw(r.buf, uv.Rect(0, op.Y, r.w, 1))
	case conformance.OpClear:
		r.buf.Clear()
	case conformance.OpRender:
		r.rend.Render(r.buf.RenderBuffer)
		r.flush(t)
	case conformance.OpRedraw:
		r.rend.Redraw(r.buf.RenderBuffer)
		r.flush(t)
	case conformance.OpMoveTo:
		r.rend.MoveTo(op.N, 0)
		r.flush(t)
	case conformance.OpResize:
		// Resize the buffer, the emulator, and the renderer's tab stops in
		// lockstep, the way a real SIGWINCH handler would. The renderer's model
		// of the previous frame keeps its old dimensions, which is exactly the
		// stale-geometry case under test.
		r.w, r.h = op.W, op.H
		r.buf.Resize(op.W, op.H)
		r.term.Resize(op.W, op.H)
		r.rend.Resize(op.W, op.H)
	}
}

// screen returns every row, for comparison against another runner.
func (r *runner) screen(t *testing.T) []string {
	t.Helper()

	rows := make([]string, r.h)
	for y := range rows {
		rows[y] = r.term.Row(t, y)
	}
	return rows
}

// runIncremental replays the whole program through one renderer, so its model
// accumulates history exactly as it would in a running application.
func runIncremental(t *testing.T, p conformance.Program, mk func(*testing.T, int, int, bool) oracle) []string {
	t.Helper()

	r := newRunner(t, p, mk)
	defer r.close()

	for _, op := range p.Ops {
		r.step(t, op)
	}

	// Push the final state out even if the program did not end with a render,
	// otherwise trailing drawing ops would be dropped and the comparison would
	// run against a stale screen.
	r.rend.Render(r.buf.RenderBuffer)
	r.flush(t)

	return r.screen(t)
}

// runFullRepaint builds the same final buffer but paints it once onto a fresh
// terminal with no prior history to diff against. That makes its output an
// unconditional full paint, which is the ground truth the incremental path has
// to match.
func runFullRepaint(t *testing.T, p conformance.Program, mk func(*testing.T, int, int, bool) oracle) []string {
	t.Helper()

	r := newRunner(t, p, mk)
	defer r.close()

	// Replay only the operations that shape the buffer. Skipping the renders is
	// what makes this a reference: the buffer ends up identical, but the
	// renderer never sees an intermediate frame and so cannot drift. Resizes are
	// replayed too, so the final buffer has the same dimensions as the
	// incremental run.
	for _, op := range p.Ops {
		switch op.Kind {
		case conformance.OpDrawLine, conformance.OpClear, conformance.OpResize:
			r.step(t, op)
		}
	}

	r.rend.Render(r.buf.RenderBuffer)
	r.flush(t)

	return r.screen(t)
}

// compare reports the first differing row, with the program that produced it.
// Printing the program is what makes a fuzz failure actionable: the raw input is
// a byte string that says nothing about what the renderer was asked to do.
func compare(t *testing.T, p conformance.Program, what string, got, want []string) {
	t.Helper()

	for y := range want {
		if got[y] == want[y] {
			continue
		}
		t.Errorf("%s: screen disagrees with a full repaint at row %d\n"+
			"  got  %q\n"+
			"  want %q\n"+
			"program:\n%s",
			what, y, got[y], want[y], p)
		return
	}
}

// Codepoints an emulator is known to swallow, so [FuzzScreenShowsContent] can
// assert on everything else instead of giving up on whole clusters. Each entry
// is a characterised emulator behaviour, not a renderer bug: keep them as tight
// as the emulator forces and no tighter, since every rune listed here is one the
// target can no longer catch the renderer dropping.
const (
	// VS16 asks for emoji presentation. ghostty's legacy width mode reads it as
	// a hint and does not store it in the cell, so it never reads back.
	variationSelector16 = '\ufe0f'
	// The keycap combining mark. Listed for documentation; x/vt's trouble with
	// keycaps is handled per cluster rather than per codepoint, see skip below.
	combiningKeycap = '\u20e3'
)

// keycap is the cluster x/vt cannot place coherently. It splits the cluster
// across two cells but advances the cursor by only one, so its own idea of
// where the row ends disagrees with what it drew. Whether any part of the
// cluster survives then depends on where in the row it landed, which makes it
// useless as evidence about the renderer. ghostty places it correctly in both
// width modes, so the assertion still bites there.
const keycap = "1\ufe0f\u20e3"

// oracles are the emulators every fuzz target runs against.
// oracleSpec is one emulator the fuzz targets run against.
type oracleSpec struct {
	name string
	mk   func(*testing.T, int, int, bool) oracle
	// drops are codepoints this emulator swallows outright, so they never read
	// back from any cell.
	drops []rune
	// skip are whole clusters this emulator cannot place coherently, so where
	// their codepoints land says nothing about the renderer.
	skip []string
}

var oracles = []oracleSpec{
	{name: "ghostty", mk: newGhostty, drops: []rune{variationSelector16}},
	// x/vt has no legacy width mode, so its width model is fixed regardless of
	// what the program asked for.
	{
		name:  "vt",
		mk:    func(t *testing.T, w, h int, _ bool) oracle { return newVT(t, w, h) },
		drops: []rune{variationSelector16},
		skip:  []string{keycap},
	},
}

// addSeeds seeds a fuzz target from the shared corpus.
func addSeeds(f *testing.F) {
	f.Helper()

	for _, seed := range conformance.Seeds() {
		f.Add(seed)
	}

	// Short raw inputs too, so the fuzzer has small programs to grow rather
	// than having to shrink its way down from the structured seeds.
	f.Add([]byte(nil))
	f.Add([]byte{0, 0, 0})
	f.Add([]byte{2, 12, 2, 12, 2, 12})
}

// FuzzRenderer is the main coverage-guided target. It reads each input as a
// program of renderer operations, replays it incrementally, and requires the
// resulting screen to match a full repaint of the same final buffer.
//
// This is the property worth fuzzing because it is the renderer's whole
// contract: however it chooses to reach a frame, the frame must look the same as
// if it had been painted from scratch. Cursor tracking, dirty regions and scroll
// optimisation are all implementation details in service of that.
func FuzzRenderer(f *testing.F) {
	addSeeds(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		p := conformance.DecodeProgram(data)
		if len(p.Ops) == 0 {
			return
		}

		for _, o := range oracles {
			got := runIncremental(t, p, o.mk)
			want := runFullRepaint(t, p, o.mk)
			compare(t, p, o.name, got, want)
		}
	})
}

// FuzzRendererIdempotent checks that rendering an unchanged buffer twice leaves
// the screen alone.
//
// This is a weaker property than the one above, but it fails differently and on
// smaller inputs. A renderer whose model is wrong in a way that cancels out over
// a whole frame can slip past the differential test and still fail here.
func FuzzRendererIdempotent(f *testing.F) {
	addSeeds(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		p := conformance.DecodeProgram(data)
		if len(p.Ops) == 0 {
			return
		}

		for _, o := range oracles {
			r := newRunner(t, p, o.mk)

			for _, op := range p.Ops {
				r.step(t, op)
			}
			r.rend.Render(r.buf.RenderBuffer)
			r.flush(t)
			before := r.screen(t)

			// A second render of the same buffer should be a no-op.
			r.rend.Render(r.buf.RenderBuffer)
			r.flush(t)
			after := r.screen(t)

			compare(t, p, o.name+", second render of an unchanged buffer", after, before)
			r.close()
		}
	})
}

// FuzzRedrawResyncs checks that a forced repaint recovers from any state the
// renderer managed to get itself into.
//
// Redraw is the escape hatch the whole design leans on. If it does not reliably
// resynchronise then no amount of care in the incremental path is enough,
// because there is no way back from a bad state.
func FuzzRedrawResyncs(f *testing.F) {
	addSeeds(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		p := conformance.DecodeProgram(data)
		if len(p.Ops) == 0 {
			return
		}

		for _, o := range oracles {
			r := newRunner(t, p, o.mk)

			for _, op := range p.Ops {
				r.step(t, op)
			}

			// However confused the renderer now is, a redraw must land on the
			// same screen as painting this buffer onto a fresh terminal.
			r.rend.Redraw(r.buf.RenderBuffer)
			r.flush(t)
			got := r.screen(t)
			r.close()

			want := runFullRepaint(t, p, o.mk)
			compare(t, p, o.name+", after Redraw", got, want)
		}
	})
}

// drawnRow returns the text a buffer row holds that the emulator has room to
// show: the content of every cell in order, skipping the placeholders that
// continue a wide cell, and stopping at the first cluster that runs past the
// right margin.
//
// This, and not the string the program asked to draw, is what the renderer was
// given to paint and could paint. A cluster whose start column falls past the
// right edge is never placed in the buffer at all. One that starts inside the
// row but runs off the end cannot be shown whole, and the renderer paints such
// a row with autowrap off, so it clips rather than spilling onto the next row
// and everything after it is lost too.
//
// The walk stops at the first cluster the emulator measures differently than
// the renderer does, and that cluster is not included either.
//
// Up to that point the two agree on every column, so the renderer put each
// cluster exactly where the emulator was expecting it and it had to land. The
// first disagreement ends that. The renderer goes on positioning by its own
// model, so the very next write lands inside the cluster that disagreed and
// destroys it, and every write after that is similarly adrift. A codepoint
// missing from there on says the two disagree about width, which is a given
// here, not that the renderer dropped it.
//
// This still bites where it matters. ghostty measures the drift-prone clusters
// exactly as the renderer's matching width model does, in both legacy and mode
// 2027, so the assertion covers all of them there. It is x/vt driven in legacy
// mode that mostly falls out, and only because x/vt has no legacy mode to
// drive.
func drawnRow(buf uv.ScreenBuffer, y int, width func(string) int) string {
	var b strings.Builder
	for x := 0; x < buf.Width(); x++ {
		c := buf.CellAt(x, y)
		if c == nil || c.Width == 0 {
			continue
		}
		if width(c.Content) != c.Width || x+c.Width > buf.Width() {
			break
		}
		b.WriteString(c.Content)
	}
	return b.String()
}

// FuzzScreenShowsContent checks that everything drawn actually reaches the
// screen, without reference to any other render.
//
// The three targets above are all differential: they compare one render against
// another. That makes them blind to a renderer that paints the same wrong thing
// every time, and writing a cluster into a column the terminal has already
// filled is exactly that kind of consistent wrongness. A first paint and a full
// repaint agree perfectly while both drop a glyph.
//
// So this target asserts something absolute instead. The most recently drawn
// line must appear on its row, with every codepoint of every drift-prone
// cluster it contained present at least as many times as it was drawn.
//
// Codepoints are counted one at a time rather than as whole clusters because
// the emulators disagree about how a cluster lands in cells: x/vt splits a
// keycap across two, so the bytes are on the row but not contiguous. Counting
// them separately keeps the assertion blind to layout and sharp about loss,
// which matters, since dropping a combining mark is the bug this target was
// written to catch. Where an emulator swallows a codepoint outright it is
// listed in that oracle's drops, so the tolerance is per-emulator and named
// rather than applied to everyone. Extra occurrences are tolerated as
// emulator-defined residue from earlier draws. "The glyph reached the screen at
// all" is a floor that holds under every width model.
//
// Only the last draw is checked, not every draw. A resize can clip rows drawn
// earlier or shrink the screen below them, and whether clipped content should
// reappear is emulator-defined, so asserting on it would be noise. The last
// draw is always made at the current width and within the current height, so it
// is always fully visible and always fair to check.
func FuzzScreenShowsContent(f *testing.F) {
	addSeeds(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		p := conformance.DecodeProgram(data)
		if len(p.Ops) == 0 {
			return
		}

		// Track the row of the last line drawn. A Clear invalidates it, since
		// nothing is drawn after a clear unless a later DrawLine says so.
		lastY := -1
		for _, op := range p.Ops {
			switch op.Kind {
			case conformance.OpDrawLine:
				lastY = op.Y
			case conformance.OpClear:
				lastY = -1
			}
		}
		if lastY < 0 {
			return
		}

		for _, o := range oracles {
			r := newRunner(t, p, o.mk)
			for _, op := range p.Ops {
				r.step(t, op)
			}
			r.rend.Render(r.buf.RenderBuffer)
			r.flush(t)

			// A resize after the last draw can shrink the screen below that
			// row, in which case it is off-screen and there is nothing to
			// assert.
			if lastY >= r.h {
				r.close()
				continue
			}

			screen := r.screen(t)
			drawn := drawnRow(r.buf, lastY, func(cluster string) int {
				return clusterWidth(t, o, p.GraphemeWidth, cluster)
			})
			r.close()

			for _, cluster := range conformance.DriftClusters() {
				want := strings.Count(drawn, cluster)
				if want == 0 || slices.Contains(o.skip, cluster) {
					continue
				}
				// Count codepoints, not clusters; extras are residue.
				for _, r := range cluster {
					if slices.Contains(o.drops, r) {
						continue
					}
					wantRune := strings.Count(drawn, string(r))
					gotRune := strings.Count(screen[lastY], string(r))
					if gotRune < wantRune {
						t.Errorf("%s: row %d shows %q of cluster %q %d times but at least %d were drawn\n"+
							"  screen %q\n"+
							"  drawn  %q\n"+
							"program:\n%s",
							o.name, lastY, string(r), cluster, gotRune, wantRune, screen[lastY], drawn, p)
						return
					}
				}
			}
		}
	})
}
