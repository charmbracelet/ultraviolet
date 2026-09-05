package conformance_test

import (
	"testing"

	"github.com/charmbracelet/ultraviolet/internal/conformance"
)

// A screen that shrinks narrow enough to rewrap a row, then grows again, is the
// case the nightly run kept finding and the corpus could not hold: the inputs
// it minimised to were long, and what they were really exercising is short.
//
// Shrinking makes the terminal rewrap a row too wide to fit, which pushes
// content onto rows that are off-screen at the time. Growing brings those rows
// back, still holding it. The renderer never painted them, so nothing in its
// model says they are dirty, and the residue used to survive every later frame.
func TestReflowAcrossShrinkAndGrow(t *testing.T) {
	const wave = "\U0001f44b\U0001f3ff" // four columns under legacy widths

	wide := ""
	for range 6 {
		wide += wave
	}

	progs := map[string]conformance.Program{
		// Rows wider than the screen, squeezed to a fifth of the width and
		// then opened out past the original height.
		"shrink then grow taller": {
			Width: 26, Height: 2,
			Ops: []conformance.Op{
				{Kind: conformance.OpDrawLine, Y: 0, Text: wide},
				{Kind: conformance.OpDrawLine, Y: 1, Text: wide},
				{Kind: conformance.OpRender},
				{Kind: conformance.OpResize, W: 5, H: 2},
				{Kind: conformance.OpRender},
				{Kind: conformance.OpResize, W: 23, H: 8},
			},
		},
		// The same shape with the rows drawn out of order and a gentler
		// shrink, which is how the fuzzer first hit it.
		"shrink then grow, rows out of order": {
			Width: 26, Height: 4,
			Ops: []conformance.Op{
				{Kind: conformance.OpDrawLine, Y: 2, Text: wide},
				{Kind: conformance.OpDrawLine, Y: 0, Text: wide},
				{Kind: conformance.OpDrawLine, Y: 1, Text: wide},
				{Kind: conformance.OpRender},
				{Kind: conformance.OpResize, W: 11, H: 2},
				{Kind: conformance.OpRender},
				{Kind: conformance.OpResize, W: 23, H: 8},
			},
		},
	}

	for name, prog := range progs {
		t.Run(name, func(t *testing.T) {
			for _, o := range oracles {
				got := runIncremental(t, prog, o.mk)
				want := runFullRepaint(t, prog, o.mk)
				compare(t, prog, o.name, got, want)
			}
		})
	}
}
