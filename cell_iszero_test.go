package uv

import (
	"image/color"
	"testing"
	"time"
)

// recursiveColor is a color.Color implementation whose underlying struct
// contains an interface field, triggering infinite recursion in
// runtime.ifaceeq when compared with ==.
//
// This reproduces the bug where Cell.IsZero() and Style.IsZero() use
// struct == comparison on types containing color.Color interface fields.
type recursiveColor struct {
	color.Color // embedded interface — ifaceeq recurses into this
	r, g, b, a  uint8
}

func (c *recursiveColor) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(c.r), uint32(c.g), uint32(c.b), uint32(c.a)
}

// TestCellIsZero_RecursiveColorPanicsOrHangs tests that Cell.IsZero()
// does not hang when the cell's style contains a color.Color implementation
// with an embedded interface.
//
// Before the fix, this test would hang forever (CPU 100%) because:
//
//	Cell.IsZero() → *c == Cell{}
//	  → .eq.Cell → .eq.Style → runtime.ifaceeq(color.Color, nil)
//	    → recurse into embedded interface → infinite loop
func TestCellIsZero_RecursiveColorPanicsOrHangs(t *testing.T) {
	// Use a concrete non-nil color with embedded interface
	c := &Cell{
		Content: "x",
		Width:   1,
		Style: Style{
			Fg: &recursiveColor{r: 255, g: 0, b: 0, a: 255},
		},
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			recover() // in case it panics instead of hanging
			close(done)
		}()
		_ = c.IsZero()
	}()

	select {
	case <-done:
		// Good: returned without hanging
	case <-time.After(2 * time.Second):
		t.Fatal("Cell.IsZero() hung — likely infinite recursion in runtime.ifaceeq")
	}
}

// TestStyleIsZero_RecursiveColorPanicsOrHangs tests the same issue for Style.IsZero().
func TestStyleIsZero_RecursiveColorPanicsOrHangs(t *testing.T) {
	s := &Style{
		Fg: &recursiveColor{r: 255, g: 0, b: 0, a: 255},
	}

	done := make(chan struct{})
	go func() {
		defer func() {
			recover()
			close(done)
		}()
		_ = s.IsZero()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Style.IsZero() hung — likely infinite recursion in runtime.ifaceeq")
	}
}

// TestCellIsZero_NilColorDoesNotHang verifies the common case still works.
func TestCellIsZero_NilColorDoesNotHang(t *testing.T) {
	tests := []struct {
		name string
		cell *Cell
		want bool
	}{
		{"zero cell", &Cell{}, true},
		{"nil style colors", &Cell{Content: "x", Width: 1}, false},
		{"empty content with nil colors", &Cell{Content: " ", Width: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cell.IsZero()
			if got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
