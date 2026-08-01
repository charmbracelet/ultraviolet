package uv

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestDrawWcWidthCells checks that drawing lays cells out over the same number
// of columns the terminal advances over. A cell holds a whole grapheme
// cluster, but under wcwidth its width is the sum of the cluster's codepoints,
// which is what a terminal without Unicode core mode (DEC mode 2027) advances
// by. Get that wrong and every cell after it lands in the wrong column.
func TestDrawWcWidthCells(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		widths []int // width of each non-placeholder cell, in order
	}{
		{"ascii", "abc", []int{1, 1, 1}},
		{"combining acute", "éx", []int{1, 1}},
		{"devanagari", "नमस्ते", []int{1, 1, 2}},
		{"cjk", "世界", []int{2, 2}},
		{"vs16", "⚠️", []int{1}},
		{"zwj pair", "👨‍💻", []int{4}},
		{"zwj flag", "🏳️‍🌈", []int{3}},
		{"skin tone", "👍🏽", []int{4}},
		// Wider than any glyph, and wider than the four columns cells used to
		// be capped at. A capped cluster ends up zero width, which the
		// renderer skips as a wide-cell placeholder, so it never gets drawn.
		{"zwj family", "👨‍👩‍👧‍👦", []int{8}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewScreenBuffer(20, 1)
			buf.Method = ansi.WcWidth
			NewStyledString(tc.str).Draw(buf, buf.Bounds())

			var x int
			for i, want := range tc.widths {
				cell := buf.CellAt(x, 0)
				if cell == nil {
					t.Fatalf("cell %d: no cell at column %d", i, x)
				}
				if cell.Width != want {
					t.Errorf("cell %d: width %d, want %d", i, cell.Width, want)
				}
				// Every column the cell covers past the first has to be a
				// placeholder, otherwise the buffer and the terminal disagree
				// about what occupies them.
				for j := 1; j < cell.Width; j++ {
					if p := buf.CellAt(x+j, 0); p == nil || !p.isWidePlaceholder() {
						t.Errorf("column %d: got %+v, want a wide placeholder", x+j, p)
					}
				}
				x += max(cell.Width, 1)
			}

			if got := ansi.StringWidthWc(tc.str); got != x {
				t.Errorf("cells cover %d columns, but the string measures %d", x, got)
			}
		})
	}
}

// TestDrawGraphemeWidthCells covers the same strings under the width model of
// a terminal in Unicode core mode, which measures a cluster as the one glyph
// it draws.
func TestDrawGraphemeWidthCells(t *testing.T) {
	tests := []struct {
		str   string
		width int // width of the first cell
	}{
		{"नमस्ते", 1},
		{"⚠️", 2},
		{"👨‍👩‍👧‍👦", 2},
		{"👍🏽", 2},
	}

	for _, tc := range tests {
		buf := NewScreenBuffer(20, 1)
		buf.Method = ansi.GraphemeWidth
		NewStyledString(tc.str).Draw(buf, buf.Bounds())

		cell := buf.CellAt(0, 0)
		if cell == nil {
			t.Fatalf("%q: no cell at column 0", tc.str)
		}
		if cell.Width != tc.width {
			t.Errorf("%q: width %d, want %d", tc.str, cell.Width, tc.width)
		}
	}
}

// TestNewCellWidths checks that cells built directly agree with the cells
// [StyledString] lays out.
func TestNewCellWidths(t *testing.T) {
	for _, tc := range []struct {
		gr   string
		want int
	}{
		{"a", 1},
		{" ", 1},
		{"स्ते", 2},
		{"👨‍💻", 4},
	} {
		if got := NewCell(ansi.WcWidth, tc.gr).Width; got != tc.want {
			t.Errorf("NewCell(WcWidth, %q).Width = %d, want %d", tc.gr, got, tc.want)
		}
	}
}
