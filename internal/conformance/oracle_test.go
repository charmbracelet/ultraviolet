package conformance_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/x/vt"
	"go.mitchellh.com/libghostty"
)

// Two reference emulators, because they disagree about how wide a grapheme
// cluster is and that disagreement is the whole subject of these tests.
//
// libghostty is ghostty's real VT engine. It models legacy widths per
// codepoint, which is what stock terminals actually do, and it implements DEC
// mode 2027 so both width models can be exercised. x/vt always measures whole
// clusters, so it reproduces a mode-2027 terminal only.
//
// Neither is redundant. A renderer bug that only shows up under legacy widths
// is invisible to x/vt, and keeping both means a disagreement between them is
// itself a signal worth investigating.

// oracle is a terminal emulator the renderer can write to and whose screen can
// be read back column by column.
//
// Reading by column is the point. A formatter that returns the screen as a
// string collapses the spacer cells that follow a wide glyph, which is exactly
// the information these tests need: it cannot distinguish a cluster at column 0
// with a marker at column 2 from the same cluster with a marker at column 1.
type oracle interface {
	// Write feeds renderer output into the emulator.
	Write(p []byte) (n int, err error)

	// Row returns the visible text of a row, trailing blanks removed.
	Row(t *testing.T, y int) string

	// CellAt returns the content of one cell. A cell that holds nothing, or
	// that continues a wide glyph starting to its left, returns "".
	CellAt(t *testing.T, x, y int) string

	// Size reports the emulator's dimensions.
	Size() (w, h int)

	// Resize changes the emulator's dimensions, mirroring a SIGWINCH.
	Resize(w, h int)

	// Close releases emulator resources.
	Close()

	// Name identifies the emulator in failure messages.
	Name() string
}

// ghosttyOracle wraps ghostty's VT engine.
type ghosttyOracle struct {
	term *libghostty.Terminal
	w, h int
}

// newGhostty creates a ghostty emulator. When grapheme is false it measures
// clusters with legacy per-codepoint widths, the way an ordinary terminal that
// has not negotiated mode 2027 does.
func newGhostty(t *testing.T, w, h int, grapheme bool) oracle {
	t.Helper()

	term, err := libghostty.NewTerminal(libghostty.WithSize(uint16(w), uint16(h)))
	if err != nil {
		t.Fatalf("libghostty.NewTerminal(%d, %d): %v", w, h, err)
	}
	if err := term.ModeSet(libghostty.ModeGraphemeCluster, grapheme); err != nil {
		term.Close()
		t.Fatalf("ModeSet(ModeGraphemeCluster, %v): %v", grapheme, err)
	}
	return &ghosttyOracle{term: term, w: w, h: h}
}

func (g *ghosttyOracle) Write(p []byte) (int, error) { return g.term.Write(p) }
func (g *ghosttyOracle) Size() (int, int)            { return g.w, g.h }
func (g *ghosttyOracle) Close()                      { g.term.Close() }
func (g *ghosttyOracle) Name() string                { return "ghostty" }

func (g *ghosttyOracle) Resize(w, h int) {
	g.w, g.h = w, h
	// Cell pixel dimensions are irrelevant to these tests, so pass zero.
	_ = g.term.Resize(uint16(w), uint16(h), 0, 0)
}

func (g *ghosttyOracle) CellAt(t *testing.T, x, y int) string {
	t.Helper()

	ref, err := g.term.GridRef(libghostty.Point{
		Tag: libghostty.PointTagActive,
		X:   uint16(x),
		Y:   uint32(y),
	})
	if err != nil {
		t.Fatalf("GridRef(%d, %d): %v", x, y, err)
	}
	cell, err := ref.Cell()
	if err != nil {
		t.Fatalf("Cell(%d, %d): %v", x, y, err)
	}

	// Graphemes returns the whole cluster including its base codepoint, so
	// reading Codepoint as well would duplicate the base. Fall back to
	// Codepoint only for cells that hold a single rune.
	var b strings.Builder
	if cps, err := ref.Graphemes(); err == nil && len(cps) > 0 {
		for _, cp := range cps {
			b.WriteRune(rune(cp))
		}
		return b.String()
	}
	if cp, err := cell.Codepoint(); err == nil && cp != 0 {
		b.WriteRune(rune(cp))
	}
	return b.String()
}

func (g *ghosttyOracle) Row(t *testing.T, y int) string {
	t.Helper()

	var b strings.Builder
	for x := range g.w {
		cell := g.CellAt(t, x, y)
		if cell == "" {
			// A cell that was erased reports no codepoint, while a cell
			// someone wrote a space into reports one. They look the same on
			// screen and mean the same thing, so read them back the same way.
			// Otherwise a row painted by erasing compares unequal to the same
			// row painted with spaces.
			cell = " "
		}
		b.WriteString(cell)
	}

	// Empty cells read back as NUL rather than a space, so trim both.
	return strings.TrimRight(b.String(), " \x00")
}

// clusterWidth reports how many columns an emulator advances when it prints one
// grapheme cluster.
//
// Measured rather than modelled. How wide a cluster is happens to be the exact
// thing these tests are about, so a width table here would be a second opinion
// standing in for the emulator's, and wrong in the cases that matter: ghostty's
// legacy mode paints a regional indicator pair in four columns where wcwidth
// says two.
//
// The measurement writes the cluster followed by a marker and reports the
// column the marker landed in, which is the advance width by definition. The
// answers are cached because a fuzz target asks for the same handful of
// clusters millions of times.
var clusterWidths sync.Map

func clusterWidth(t *testing.T, o oracleSpec, grapheme bool, cluster string) int {
	t.Helper()

	type key struct {
		name     string
		grapheme bool
		cluster  string
	}
	k := key{o.name, grapheme, cluster}
	if w, ok := clusterWidths.Load(k); ok {
		return w.(int)
	}

	// Wide enough for any cluster in the corpus, with room for the marker.
	const scratch = 64
	const marker = "|"

	term := o.mk(t, scratch, 1, grapheme)
	defer term.Close()
	if _, err := term.Write([]byte(cluster + marker)); err != nil {
		t.Fatalf("measuring %q: %v", cluster, err)
	}

	w := 0
	for x := range scratch {
		if term.CellAt(t, x, 0) == marker {
			w = x
			break
		}
	}
	clusterWidths.Store(k, w)
	return w
}

// vtOracle wraps x/vt, which always measures whole grapheme clusters and so
// behaves like a terminal with mode 2027 enabled.
type vtOracle struct {
	em *vt.Emulator
}

func newVT(t *testing.T, w, h int) oracle {
	t.Helper()
	return &vtOracle{em: vt.NewEmulator(w, h)}
}

func (v *vtOracle) Write(p []byte) (int, error) { return v.em.Write(p) }
func (v *vtOracle) Size() (int, int)            { return v.em.Width(), v.em.Height() }
func (v *vtOracle) Close()                      {}
func (v *vtOracle) Name() string                { return "vt" }
func (v *vtOracle) Resize(w, h int)             { v.em.Resize(w, h) }

func (v *vtOracle) CellAt(t *testing.T, x, y int) string {
	t.Helper()

	if cell := v.em.CellAt(x, y); cell != nil {
		return cell.Content
	}
	return ""
}

func (v *vtOracle) Row(t *testing.T, y int) string {
	t.Helper()

	var b strings.Builder
	for x := range v.em.Width() {
		b.WriteString(v.CellAt(t, x, y))
	}
	return strings.TrimRight(b.String(), " ")
}
