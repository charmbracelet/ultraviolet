// Package conformance drives the renderer against real terminal emulators and
// compares the resulting screen against ground truth.
//
// The renderer keeps a model of what the terminal is currently showing and
// writes only the difference between that model and the next frame. When the
// model drifts out of step with reality the output is still valid ANSI and
// still looks plausible; it just paints the wrong thing. Asserting on escape
// sequences cannot catch that, because there is no single correct sequence to
// compare against.
//
// So these tests compare against ground truth instead: run a sequence of
// frames incrementally on one terminal, paint the final frame from scratch on
// another, and require the two screens to match. Any disagreement is a bug in
// the incremental path.
//
// This lives in its own module because the reference emulator is a cgo package
// with a newer Go requirement, and neither cost belongs in the library that
// ships to users.
package conformance

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// The fuzzer needs to explore sequences of renderer operations, not just
// strings of text, because the bugs worth finding live in the interaction
// between frames. A single frame is nearly always painted correctly; it is the
// second frame, drawn as a diff against a model of the first, that goes wrong.
//
// So a fuzz input is decoded as a little program: a screen size, a set of
// renderer options, then a list of operations to run in order. Go's fuzzer
// mutates the raw bytes and coverage feedback steers it toward programs that
// reach new code in the renderer. That is the part a hand-written table of
// cases cannot do, since interesting programs are combinations nobody thinks
// to write down.

// Op is a single step in a fuzz program.
type Op struct {
	// Kind selects which operation to perform.
	Kind OpKind

	// Y is the row an OpDrawLine writes to, already reduced modulo the screen
	// height so it is always in range.
	Y int

	// Text is the content an OpDrawLine writes.
	Text string

	// N is a small integer argument, used by operations like OpMoveTo.
	N int

	// W and H are the new dimensions for an OpResize, already clamped to the
	// valid range.
	W, H int
}

// OpKind enumerates the renderer operations a fuzz program can perform.
type OpKind uint8

const (
	// OpDrawLine writes Text into row Y of the pending frame.
	OpDrawLine OpKind = iota

	// OpClear blanks the pending frame.
	OpClear

	// OpRender flushes the pending frame through the renderer's incremental
	// path. This is the operation under test; every other op exists to get
	// the renderer into an interesting state before one of these.
	OpRender

	// OpRedraw forces a full repaint, which should always resynchronise the
	// renderer's model with the terminal.
	OpRedraw

	// OpMoveTo repositions the cursor, which is a common source of drift
	// because it is where a wrong column model becomes a wrong write.
	OpMoveTo

	// OpResize changes the terminal and buffer dimensions to W by H. Resizing
	// is the biggest geometry change a terminal undergoes, and a renderer that
	// carries stale geometry across one will paint the wrong thing, so it is
	// worth fuzzing directly.
	OpResize

	opKindCount
)

// Resize bounds. Screens stay small so failures stay readable, but the range
// is wide enough to shrink below and grow above the starting size, which is
// where stale-geometry bugs show up. Exported so the decoder's own tests can
// check that every OpResize lands inside them.
const (
	MinResizeW, MaxResizeW = 4, 32
	MinResizeH, MaxResizeH = 2, 8
)

// String makes failures readable, since a raw OpKind number tells you nothing
// about what the renderer was asked to do.
func (k OpKind) String() string {
	switch k {
	case OpDrawLine:
		return "DrawLine"
	case OpClear:
		return "Clear"
	case OpRender:
		return "Render"
	case OpRedraw:
		return "Redraw"
	case OpMoveTo:
		return "MoveTo"
	case OpResize:
		return "Resize"
	default:
		return fmt.Sprintf("OpKind(%d)", uint8(k))
	}
}

// Program is a decoded fuzz input: a terminal size, renderer settings, and the
// operations to run.
type Program struct {
	// Width and Height are the terminal size, clamped to a small range. Small
	// screens are deliberate: they make wrapping and scrolling happen often,
	// and they keep failure output readable.
	Width, Height int

	// GraphemeWidth selects the renderer's width model, and is matched by the
	// emulator's own mode so both sides agree or disagree deliberately rather
	// than by accident.
	GraphemeWidth bool

	// ScrollOptim enables the hardware-scroll path, which is worth fuzzing
	// separately because it moves glyphs bodily and carries their painted
	// widths with them.
	ScrollOptim bool

	// Ops are the operations to run in order.
	Ops []Op
}

// String renders a program as reproducible Go-ish pseudocode. A fuzz failure is
// only useful if you can see what it did, and a hex dump of the input does not
// tell you that.
func (p Program) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "size=%dx%d graphemeWidth=%v scrollOptim=%v\n",
		p.Width, p.Height, p.GraphemeWidth, p.ScrollOptim)
	for i, op := range p.Ops {
		switch op.Kind {
		case OpDrawLine:
			fmt.Fprintf(&b, "  %2d. DrawLine(y=%d, %q)\n", i, op.Y, op.Text)
		case OpMoveTo:
			fmt.Fprintf(&b, "  %2d. MoveTo(%d)\n", i, op.N)
		case OpResize:
			fmt.Fprintf(&b, "  %2d. Resize(%dx%d)\n", i, op.W, op.H)
		default:
			fmt.Fprintf(&b, "  %2d. %s\n", i, op.Kind)
		}
	}
	return b.String()
}

// corpusAlphabet is the set of clusters a fuzz program can draw. The fuzzer
// picks from this by index rather than consuming raw UTF-8, because letting it
// invent its own bytes wastes almost all its time on invalid encodings and
// widthless control characters. Choosing from a curated set means every
// generated line is made of clusters that actually stress the width model.
//
// The clusters are grouped by width behaviour. Everything from
// driftClusterOffset onward is a cluster where legacy wcwidth and grapheme
// width disagree, and those are the ones that provoke column drift. The
// clusters before that point exercise distinct code paths (combining marks,
// text presentation, a different wide script) without drifting, so the fuzzer
// explores them too.
var corpusAlphabet = []string{
	// Narrow ASCII.
	"a", "b", "c", "X", " ", ".", "|", "#",
	// Wide CJK.
	"世", "界",
	// Wide Hangul, a different script in the same width class.
	"가", "한",
	// Plain emoji, width 2, no modifiers.
	"\U0001fae0", // melting face
	// Combining marks: a base plus a zero-width combiner, width 1 total.
	"e\u0301", // é
	"n\u0303", // ñ
	// VS15 text presentation, forces narrow rendering, width 1.
	"\u2639\ufe0e", // frowning face, text style
	// ZWJ sequences where both width models agree at 2.
	"\U0001f469\u200d\U0001f4bb", // woman technologist
	"\U0001f426\u200d\U0001f525", // phoenix
	// Skin tone modifiers. Legacy wcwidth measures these at 4, two columns for
	// the base and two for the modifier, which is what ghostty's legacy mode
	// paints. They sit before driftClusterOffset only so the seeds stay pointed
	// at the sequences terminals disagree about; the fuzzer still reaches them,
	// and a four-column cluster is worth reaching.
	"\U0001f44d\U0001f3fd", // thumbs up, medium skin
	"\U0001f44b\U0001f3ff", // waving hand, dark skin

	// --- drift-prone clusters below: legacy wcwidth != grapheme width ---

	// VS16 emoji presentation: legacy says 1, grapheme says 2.
	"\u2639\ufe0f", // frowning face
	"\u2764\ufe0f", // heart
	"\u2708\ufe0f", // airplane
	// ZWJ with VS16: legacy says 1, grapheme says 2.
	"\u26d3\ufe0f\u200d\U0001f4a5", // broken chain
	// Regional indicators: legacy width varies wildly across terminals,
	// grapheme says 2. A single unpaired indicator is the most ambiguous case.
	"\U0001f1fa",           // single regional indicator U
	"\U0001f1fa\U0001f1f8", // flag US
	"\U0001f1ef\U0001f1f5", // flag JP
	// Keycap: legacy says 1, grapheme says 2.
	"1\ufe0f\u20e3", // keycap 1
}

// driftClusterOffset is where the drift-prone clusters begin in corpusAlphabet.
// Everything from here on is a VS16, ZWJ, regional-indicator, or keycap sequence
// where legacy wcwidth and grapheme width disagree.
const driftClusterOffset = 20

// driftIndices are the positions in corpusAlphabet whose clusters legacy
// wcwidth and grapheme width disagree about.
//
// Computed rather than written down, because the width tables underneath move
// with Unicode releases. A cluster that provokes drift today may not tomorrow,
// and a hardcoded list would keep pointing at it while quietly testing nothing.
// TestDriftPremise fails if the list ever comes out empty.
func driftIndices() []int {
	var idx []int
	for i := driftClusterOffset; i < len(corpusAlphabet); i++ {
		c := corpusAlphabet[i]
		if ansi.StringWidthWc(c) != ansi.StringWidth(c) {
			idx = append(idx, i)
		}
	}
	return idx
}

// DriftClusters are the clusters that legacy wcwidth and grapheme width
// disagree about, and so the ones that provoke column drift.
func DriftClusters() []string {
	var out []string
	for _, i := range driftIndices() {
		out = append(out, corpusAlphabet[i])
	}
	return out
}

// CorpusAlphabet returns the clusters a fuzz program can draw, for tests that
// need to reason about the corpus.
func CorpusAlphabet() []string {
	return append([]string(nil), corpusAlphabet...)
}

// Seeds is the starting corpus for the fuzz targets.
//
// Seeds matter more than they look. Coverage-guided fuzzing explores outward
// from what it is given, and random bytes mostly decode into programs that draw
// plain ASCII and never reach the width-handling code at all. These put the
// fuzzer inside the interesting region immediately: each one draws a
// drift-provoking cluster and renders more than once, so mutation starts from
// programs that already exercise the incremental path against awkward content.
func Seeds() [][]byte {
	// endOfLine is the alphabet index that terminates a drawn line.
	endOfLine := byte(len(corpusAlphabet))

	var seeds [][]byte
	for _, i := range driftIndices() {
		cluster := byte(i)

		// Cover both width models crossed with both scroll settings, since a
		// bug can hide in any one of the four combinations.
		for mode := byte(0); mode < 4; mode++ {
			seeds = append(seeds, []byte{
				14, 2, // 20x4
				mode,
				// Two adjacent clusters, then render: drift accumulates across
				// neighbours, so adjacency is the productive case.
				byte(OpDrawLine), 0, cluster, cluster, endOfLine,
				byte(OpRender),
				// A second frame, so the renderer has to diff rather than
				// paint, which is where the interesting bugs live.
				byte(OpDrawLine), 1, cluster, 0, cluster, endOfLine,
				byte(OpRender),
				// Shrink row 0, which leaves residue behind if the renderer's
				// column model is wrong.
				byte(OpDrawLine), 0, 0, cluster, endOfLine,
				byte(OpRender),
			})
		}
	}

	// Structural seeds, so the fuzzer does not have to discover Clear, Redraw
	// and MoveTo by chance. They still draw drifting clusters, so they explore
	// those operations somewhere worth being.
	if drift := driftIndices(); len(drift) > 0 {
		first := byte(drift[0])
		last := byte(drift[len(drift)-1])

		seeds = append(seeds,
			[]byte{
				14, 2, 0,
				byte(OpDrawLine), 0, first, last, 8, endOfLine,
				byte(OpRender),
				byte(OpClear),
				byte(OpRender),
				byte(OpRedraw),
			},
			[]byte{
				10, 1, 1,
				byte(OpDrawLine), 0, last, last, endOfLine,
				byte(OpRender),
				byte(OpMoveTo), 3,
				byte(OpDrawLine), 0, 0, endOfLine,
				byte(OpRender),
			},
		)

		// Resize seeds. Resizing is the biggest geometry change a terminal
		// undergoes, and the interesting failures come from carrying stale
		// geometry across one, so each seed draws, resizes, and draws again.
		// The dimensions are encoded relative to the resize bounds: a byte b
		// maps to minResize + b % (max-min+1).
		resizeW := func(w int) byte { return byte((w - MinResizeW) % (MaxResizeW - MinResizeW + 1)) }
		resizeH := func(h int) byte { return byte((h - MinResizeH) % (MaxResizeH - MinResizeH + 1)) }
		for _, mode := range []byte{0, 1, 2, 3} {
			seeds = append(seeds, []byte{
				14, 2, // 20x4
				mode,
				byte(OpDrawLine), 0, first, first, last, last, endOfLine,
				byte(OpRender),
				// Shrink, then draw a line that only fits the smaller screen.
				byte(OpResize), resizeW(8), resizeH(3),
				byte(OpDrawLine), 0, first, 0, endOfLine,
				byte(OpRender),
				// Grow back, which is where clipped content can leave residue.
				byte(OpResize), resizeW(24), resizeH(6),
				byte(OpDrawLine), 0, last, last, last, endOfLine,
				byte(OpRender),
			})
		}
	}

	return seeds
}

// decoder hands out small values from the fuzzer's byte string. Running off the
// end is normal and expected: the fuzzer supplies inputs of every length, so
// the decoder reports exhaustion rather than treating a short input as an
// error, and the caller simply stops building the program.
type decoder struct {
	buf []byte
	pos int
}

// next returns the next byte, or false once the input is exhausted.
func (d *decoder) next() (byte, bool) {
	if d.pos >= len(d.buf) {
		return 0, false
	}
	b := d.buf[d.pos]
	d.pos++
	return b, true
}

// intn returns a value in [0,n), or false if the input is exhausted. Taking one
// byte per value keeps the mapping from input bytes to program structure simple
// enough that the fuzzer's byte-level mutations translate into small, targeted
// program changes, which is what makes coverage feedback effective.
func (d *decoder) intn(n int) (int, bool) {
	b, ok := d.next()
	if !ok || n <= 0 {
		return 0, false
	}
	return int(b) % n, true
}

// intrange returns a value in [lo,hi], or false if the input is exhausted.
func (d *decoder) intrange(lo, hi int) (int, bool) {
	v, ok := d.intn(hi - lo + 1)
	if !ok {
		return 0, false
	}
	return lo + v, true
}

// DecodeProgram interprets fuzzer bytes as a program.
//
// Every byte string maps to some valid program, and the decoder never fails.
// That matters: if malformed inputs were rejected the fuzzer would spend its
// budget rediscovering the input format instead of exploring renderer
// behaviour.
func DecodeProgram(data []byte) Program {
	d := &decoder{buf: data}

	// Sizes stay small and odd-ish so wrapping and scrolling are common. A
	// wide cluster landing on the last column is one of the most productive
	// cases, and narrow screens hit it constantly.
	w, _ := d.intn(24)
	h, _ := d.intn(6)
	p := Program{Width: w + 6, Height: h + 2}

	if b, ok := d.next(); ok {
		// Two independent bits, so the fuzzer can flip either width model or
		// the scroll path without disturbing the rest of the program.
		p.GraphemeWidth = b&1 != 0
		p.ScrollOptim = b&2 != 0
	}

	// Bounded so a single input cannot run for an unreasonable time; the
	// fuzzer is far more effective with many short programs than a few long
	// ones.
	const maxOps = 24

	// The live dimensions, which an OpResize changes. DrawLine and MoveTo
	// bounds are computed from these so a line drawn after a resize is always
	// valid for the current screen, not the one the program started with.
	curW, curH := p.Width, p.Height

	for len(p.Ops) < maxOps {
		kind, ok := d.intn(int(opKindCount))
		if !ok {
			break
		}

		op := Op{Kind: OpKind(kind)}
		switch op.Kind {
		case OpDrawLine:
			y, ok := d.intn(curH)
			if !ok {
				return p
			}
			op.Y = y

			// Build the line from whole clusters and stop before the right
			// margin, so content cannot wrap onto the next row. Wrapping is a
			// separate concern, and content spilling across rows would make
			// failures ambiguous about which row was actually wrong.
			var sb strings.Builder
			used := 0
			for {
				idx, ok := d.intn(len(corpusAlphabet) + 1)
				if !ok {
					return p
				}
				// The extra index terminates the line, letting the fuzzer
				// control length instead of always filling the row.
				if idx == len(corpusAlphabet) {
					break
				}
				g := corpusAlphabet[idx]
				gw := ansi.StringWidth(g)
				if used+gw > curW-1 {
					break
				}
				sb.WriteString(g)
				used += gw
			}
			op.Text = sb.String()

		case OpMoveTo:
			n, ok := d.intn(curW)
			if !ok {
				return p
			}
			op.N = n

		case OpResize:
			nw, ok := d.intrange(MinResizeW, MaxResizeW)
			if !ok {
				return p
			}
			nh, ok := d.intrange(MinResizeH, MaxResizeH)
			if !ok {
				return p
			}
			op.W, op.H = nw, nh
			curW, curH = nw, nh
		}

		p.Ops = append(p.Ops, op)
	}

	return p
}
