package uv

import (
	"math/rand"
	"testing"
)

// The line moves (rotateLines) must give the same cells as the per-cell copy
// they replace, for full-width and partial-width areas, and must never leave
// two rows sharing a line.

func cellCopyDeleteLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}
	if n > area.Max.Y-y {
		n = area.Max.Y - y
	}
	for dst := y; dst < area.Max.Y-n; dst++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.Lines[dst][x] = b.Lines[dst+n][x]
		}
	}
	for i := area.Max.Y - n; i < area.Max.Y; i++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			b.SetCell(x, i, c)
		}
	}
}

func cellCopyInsertLineArea(b *Buffer, y, n int, c *Cell, area Rectangle) {
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

func randomBuffer(w, h int, r *rand.Rand) *Buffer {
	b := NewBuffer(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b.SetCell(x, y, &Cell{Content: string(rune('a' + r.Intn(26))), Width: 1})
		}
	}
	return b
}

func assertSameCells(t *testing.T, what string, got, want *Buffer) {
	t.Helper()
	for y := range want.Lines {
		for x := range want.Lines[y] {
			if got.Lines[y][x].Content != want.Lines[y][x].Content {
				t.Fatalf("%s: (%d,%d) %q, want %q", what, x, y, got.Lines[y][x].Content, want.Lines[y][x].Content)
			}
		}
	}
	seen := map[*Cell]int{}
	for y, l := range got.Lines {
		if len(l) == 0 {
			continue
		}
		if o, ok := seen[&l[0]]; ok {
			t.Fatalf("%s: rows %d and %d share a line", what, o, y)
		}
		seen[&l[0]] = y
	}
}

func TestLineAreaMovesMatchCellCopy(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	blank := &Cell{Content: " ", Width: 1}
	for i := 0; i < 3000; i++ {
		w, h := 1+r.Intn(20), 1+r.Intn(15)
		minY := r.Intn(h)
		maxY := minY + 1 + r.Intn(h-minY)
		area := Rect(0, minY, w, maxY-minY)
		if r.Intn(3) == 0 { // a partial width: the cell path
			minX := r.Intn(w)
			area = Rect(minX, minY, 1+r.Intn(w-minX), maxY-minY)
		}
		y, n := minY+r.Intn(maxY-minY), r.Intn(h+2)
		seed := r.Int63()
		for _, op := range []string{"delete", "insert"} {
			got, want := randomBuffer(w, h, rand.New(rand.NewSource(seed))), randomBuffer(w, h, rand.New(rand.NewSource(seed)))
			if op == "delete" {
				got.DeleteLineArea(y, n, blank, area)
				cellCopyDeleteLineArea(want, y, n, blank, area)
			} else {
				got.InsertLineArea(y, n, blank, area)
				cellCopyInsertLineArea(want, y, n, blank, area)
			}
			assertSameCells(t, op, got, want)
		}
	}
}

func BenchmarkDeleteLineAreaScroll(b *testing.B) {
	for _, s := range []struct{ w, h int }{{80, 24}, {160, 45}, {240, 70}} {
		buf := randomBuffer(s.w, s.h, rand.New(rand.NewSource(1)))
		blank := &Cell{Content: " ", Width: 1}
		b.Run(Rect(0, 0, s.w, s.h).String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf.DeleteLineArea(0, 1, blank, buf.Bounds())
			}
		})
	}
}
