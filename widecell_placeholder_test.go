package uv

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestStyledWideCellPlaceholderDetection reproduces the placeholder-detection
// inconsistency introduced by PR #124.
//
// Line.Set marks the continuation columns of a wide cell with Width == 0
// placeholders. The PR makes those placeholders inherit the wide cell's Style
// and Link. The renderer detects placeholders two different ways:
//
//   - putAttrCell (terminal_renderer.go:490) uses cell.Width == 0
//   - putRange / transformLine (terminal_renderer.go:694,942,950) use IsZero()
//
// For an unstyled wide cell both agree, because the placeholder equals Cell{}.
// For a styled wide cell the placeholder is no longer zero, so IsZero() stops
// recognizing it while Width == 0 still does. The two checks disagree, which is
// how transformLine can pick a move/insert point inside a wide glyph.
func TestStyledWideCellPlaceholderDetection(t *testing.T) {
	t.Run("unstyled wide cell: both checks agree", func(t *testing.T) {
		l := make(Line, 5)
		l.Set(0, &Cell{Content: "你", Width: 2})

		ph := l.At(1)
		if ph.Width != 0 {
			t.Fatalf("expected placeholder Width == 0, got %d", ph.Width)
		}
		if !ph.IsZero() {
			t.Errorf("expected unstyled placeholder to be IsZero()")
		}
	})

	t.Run("styled wide cell: checks disagree (regression)", func(t *testing.T) {
		l := make(Line, 5)
		l.Set(0, &Cell{Content: "你", Width: 2, Style: Style{Bg: ansi.Red}})

		ph := l.At(1)
		if ph.Width != 0 {
			t.Fatalf("expected placeholder Width == 0, got %d", ph.Width)
		}
		// This is the bug: the Width==0 placeholder is no longer IsZero(), so
		// every renderer path still keyed on IsZero() fails to skip it.
		if ph.IsZero() {
			t.Skip("placeholder is zero; PR #124 style inheritance not present")
		}
		t.Errorf("styled wide-cell placeholder has Width==0 but IsZero()==false: "+
			"renderer paths using IsZero() (terminal_renderer.go:694,942,950) "+
			"will not recognize it; placeholder=%+v", *ph)
	})
}
