package uv

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

// TestMulticellSurvivesRowRedraw verifies that a kitty text-sizing
// multicell sitting in the middle of an otherwise-clearable line is
// not erased by EL right when the line is redrawn with shifted
// trailing content. The trailing-blank-only canClearWith check would
// have let the fast path fire here and zap the multicell; the per-cell
// non-clearable scan added to transformLine forces the slow path
// instead, preserving the cell.
func TestMulticellSurvivesRowRedraw(t *testing.T) {
	cellbuf := NewScreenBuffer(20, 1)

	// Frame 1: spaces, then an OSC 66 multicell at col 8 (width 2),
	// then trailing spaces. Last cell is a clearable space — so the
	// fast path used to be eligible.
	for x := 0; x < 8; x++ {
		cellbuf.SetCell(x, 0, &Cell{Content: " ", Width: 1})
	}
	cellbuf.SetCell(8, 0, &Cell{
		Content: "\x1b]66;s=2;X\x07",
		Width:   2,
	})
	for x := 10; x < 20; x++ {
		cellbuf.SetCell(x, 0, &Cell{Content: " ", Width: 1})
	}

	var out bytes.Buffer
	tr := NewTerminalRenderer(&out, []string{"COLORTERM=truecolor", "TERM=xterm-kitty"})
	tr.Resize(20, 1)
	tr.SetColorProfile(colorprofile.TrueColor)
	tr.Render(cellbuf.RenderBuffer)
	tr.Flush()
	out.Reset()

	// Frame 2: a single trailing letter at col 19 disappears, leaving
	// the rest of the line as before. This is exactly the kind of
	// edit that historically tempted EL right and would have wiped the
	// multicell at col 8.
	cellbuf.SetCell(19, 0, &Cell{Content: " ", Width: 1})
	tr.Render(cellbuf.RenderBuffer)
	if err := tr.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	got := out.String()
	if strings.Contains(got, "\x1b[K") || strings.Contains(got, "\x1b[0K") {
		t.Errorf("EL right was emitted on a row containing a multicell; "+
			"could destroy the glyph\n  got: %q", got)
	}
}

// Emit a spill row containing: 8 styled nbsps, CUF spill cell (no-style),
// 8 styled nbsps. Verify no SGR ends up at the CUF column.
func TestSpillRowCUFHasNoAdjacentSGR(t *testing.T) {
	cellbuf := NewScreenBuffer(20, 1)

	nbspStyle := Style{
		Fg: color.RGBA{R: 255, G: 250, B: 241, A: 255},
		Bg: color.RGBA{R: 77, G: 76, B: 87, A: 255},
	}

	// 8 styled nbsps at cols 0..7.
	for x := 0; x < 8; x++ {
		cellbuf.SetCell(x, 0, &Cell{Content: "\u00a0", Width: 1, Style: nbspStyle})
	}
	// CUF skip cell at col 8, mimicking what styled.go seeds for the
	// spill row of a kitty text-sizing multicell anchor.
	cellbuf.SetCell(8, 0, &Cell{Content: "\x1b[2C", Width: 2})
	// 8 more styled nbsps at cols 10..17.
	for x := 10; x < 18; x++ {
		cellbuf.SetCell(x, 0, &Cell{Content: "\u00a0", Width: 1, Style: nbspStyle})
	}

	var out bytes.Buffer
	tr := NewTerminalRenderer(&out, []string{"COLORTERM=truecolor", "TERM=xterm-kitty"})
	tr.Resize(20, 1)
	tr.SetColorProfile(colorprofile.TrueColor)
	tr.Render(cellbuf.RenderBuffer)
	tr.Flush()

	got := out.String()
	t.Logf("emitted: %q", got)

	// The CUF must appear.
	if !strings.Contains(got, "\x1b[2C") {
		t.Fatalf("CUF missing from output: %q", got)
	}
	// No SGR ("\x1b[...m") may appear immediately adjacent to the CUF on
	// either side. The whole point of the CUF cell is to leave the
	// multicell extension columns completely untouched; a pen change
	// before or after the CUF lands an SGR byte there.
	if strings.Contains(got, "\x1b[m\x1b[2C") || strings.Contains(got, "m\x1b[2C\x1b[") {
		t.Errorf("SGR emitted adjacent to CUF; pen state leaked into multicell extension\n  got: %q", got)
	}
}
