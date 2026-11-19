package uv

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// refMoveLines is the cell-by-cell move InsertLineArea and DeleteLineArea did
// before whole lines moved: the reference the line moves must match.
func refMoveLines(b *Buffer, y, n int, c *Cell, area Rectangle, insert bool) {
	if n <= 0 || y < area.Min.Y || y >= area.Max.Y || y >= b.Height() {
		return
	}

	if insert {
		n = min(n, area.Max.Y-y)
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

		return
	}

	n = min(n, area.Max.Y-y)
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

func randomBuffer(r *rand.Rand, w, h int) *Buffer {
	b := NewBuffer(w, h)
	for y := range h {
		for x := range w {
			if r.IntN(3) > 0 {
				b.SetCell(x, y, &Cell{Content: string(rune('a' + r.IntN(26))), Width: 1, Style: Style{Attrs: uint8(r.IntN(4))}})
			}
		}
	}

	return b
}

func TestMoveLinesMatchesCellMoves(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	blank := &Cell{Content: " ", Width: 1, Style: Style{Attrs: 1}}

	for i := range 3000 {
		w, h := 1+r.IntN(12), 1+r.IntN(12)
		top := r.IntN(h)
		bottom := top + 1 + r.IntN(h-top)
		area := Rect(0, top, w, bottom-top)
		if r.IntN(4) == 0 { // a narrower region keeps the cell moves
			x0 := r.IntN(w)
			area = Rect(x0, top, 1+r.IntN(w-x0), bottom-top)
		}

		y, n, insert := r.IntN(h+1)-1+0, r.IntN(h+2), r.IntN(2) == 0
		c := blank
		if r.IntN(2) == 0 {
			c = nil
		}

		seed := r.Uint64()
		got := randomBuffer(rand.New(rand.NewPCG(seed, 0)), w, h)
		want := randomBuffer(rand.New(rand.NewPCG(seed, 0)), w, h)

		if insert {
			got.InsertLineArea(y, n, c, area)
		} else {
			got.DeleteLineArea(y, n, c, area)
		}
		refMoveLines(want, y, n, c, area, insert)

		if g, wv := fmt.Sprint(got.Lines), fmt.Sprint(want.Lines); g != wv {
			t.Fatalf("case %d: insert=%v y=%d n=%d area=%v\ngot  %s\nwant %s", i, insert, y, n, area, g, wv)
		}
	}
}

func BenchmarkScrollFullWidth(b *testing.B) {
	buf := randomBuffer(rand.New(rand.NewPCG(3, 4)), 120, 40)
	area := buf.Bounds()

	b.ReportAllocs()
	for b.Loop() {
		buf.DeleteLineArea(0, 1, nil, area)
	}
}
