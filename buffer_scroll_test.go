package uv

import (
	"fmt"
	"image/color"
	"testing"
)

// numberedBuffer returns a buffer where every cell carries a content string
// unique to its position, so any misplaced cell is detectable.
func numberedBuffer(width, height int) *Buffer {
	b := NewBuffer(width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			b.SetCell(x, y, &Cell{Content: fmt.Sprintf("%d:%d", x, y), Width: 1})
		}
	}
	return b
}

// refDeleteLineArea is the pre-optimization cell-by-cell implementation of
// DeleteLineArea, kept as the behavioral reference.
func refDeleteLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}
	if n > area.Max.Y-y {
		n = area.Max.Y - y
	}
	for dst := y; dst < area.Max.Y-n; dst++ {
		src := dst + n
		for x := area.Min.X; x < area.Max.X; x++ {
			b.Lines[dst][x] = b.Lines[src][x]
		}
	}
	for i := area.Max.Y - n; i < area.Max.Y; i++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.SetCell(x, i, c)
		}
	}
}

// refInsertLineArea is the pre-optimization cell-by-cell implementation of
// InsertLineArea, kept as the behavioral reference.
func refInsertLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}
	if y+n > area.Max.Y {
		n = area.Max.Y - y
	}
	for i := area.Max.Y - 1; i >= y+n; i-- {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.Lines[i][x] = b.Lines[i-n][x]
		}
	}
	for i := y; i < y+n; i++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.SetCell(x, i, c)
		}
	}
}

func assertBuffersEqual(t *testing.T, got, want *Buffer) {
	t.Helper()
	if got.Width() != want.Width() || got.Height() != want.Height() {
		t.Fatalf("size mismatch: got %dx%d, want %dx%d",
			got.Width(), got.Height(), want.Width(), want.Height())
	}
	for y := 0; y < want.Height(); y++ {
		for x := 0; x < want.Width(); x++ {
			g, w := got.CellAt(x, y), want.CellAt(x, y)
			if !cellEqual(g, w) {
				t.Fatalf("cell (%d,%d): got %+v, want %+v", x, y, g, w)
			}
		}
	}
}

func TestLineScrollMatchesReference(t *testing.T) {
	const width, height = 8, 10
	blank := &Cell{Content: " ", Width: 1, Style: Style{Bg: color.Black}}

	cases := []struct {
		name string
		y, n int
		cell *Cell
		area Rectangle
	}{
		{"full buffer scroll one", 0, 1, nil, Rect(0, 0, width, height)},
		{"full buffer scroll many", 0, 3, blank, Rect(0, 0, width, height)},
		{"full width top margin", 2, 1, nil, Rect(0, 2, width, height-2)},
		{"full width both margins", 2, 2, blank, Rect(0, 2, width, 5)},
		{"mid-region start", 4, 2, nil, Rect(0, 2, width, 6)},
		{"count exceeds region", 2, 50, blank, Rect(0, 2, width, 5)},
		{"count equals region", 2, 5, nil, Rect(0, 2, width, 5)},
		{"narrow region keeps cell path", 1, 1, blank, Rect(2, 1, 4, 6)},
		{"left-anchored narrow region", 1, 1, nil, Rect(0, 1, 4, 6)},
	}

	for _, tc := range cases {
		t.Run("delete/"+tc.name, func(t *testing.T) {
			got := numberedBuffer(width, height)
			want := numberedBuffer(width, height)
			got.DeleteLineArea(tc.y, tc.n, tc.cell, tc.area)
			refDeleteLineArea(want, tc.y, tc.n, tc.cell, tc.area)
			assertBuffersEqual(t, got, want)
		})
		t.Run("insert/"+tc.name, func(t *testing.T) {
			got := numberedBuffer(width, height)
			want := numberedBuffer(width, height)
			got.InsertLineArea(tc.y, tc.n, tc.cell, tc.area)
			refInsertLineArea(want, tc.y, tc.n, tc.cell, tc.area)
			assertBuffersEqual(t, got, want)
		})
	}
}

// TestLineScrollKeepsRowWidths guards the recycling: rotated-out rows come
// back as blank rows of the same width, so later writes stay in bounds.
func TestLineScrollKeepsRowWidths(t *testing.T) {
	b := numberedBuffer(6, 5)
	for range 20 {
		b.DeleteLineArea(0, 1, nil, b.Bounds())
	}
	for y, line := range b.Lines {
		if len(line) != 6 {
			t.Fatalf("row %d width = %d, want 6", y, len(line))
		}
	}
}

func BenchmarkDeleteLineAreaFullWidth(b *testing.B) {
	buf := numberedBuffer(120, 40)
	area := buf.Bounds()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.DeleteLineArea(0, 1, nil, area)
	}
}

func BenchmarkDeleteLineAreaPartialWidth(b *testing.B) {
	buf := numberedBuffer(120, 40)
	area := Rect(1, 0, 118, 40)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.DeleteLineArea(0, 1, nil, area)
	}
}
