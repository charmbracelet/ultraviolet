package uv

import (
	"strings"
	"testing"
)

// newTestBuffer builds a buffer from ASCII rows, each rune becoming a
// single-width cell.
func newTestBuffer(rows ...string) *Buffer {
	h := len(rows)
	var w int
	for _, r := range rows {
		if len(r) > w {
			w = len(r)
		}
	}
	b := NewBuffer(w, h)
	for y, row := range rows {
		for x, r := range row {
			b.SetCell(x, y, &Cell{Content: string(r), Width: 1})
		}
	}
	return b
}

func TestDeleteLineArea(t *testing.T) {
	cases := []struct {
		name string
		rows []string
		y, n int
		cell *Cell
		area func(b *Buffer) Rectangle
		want []string
	}{
		{
			name: "scroll the whole buffer up",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    0, n: 1,
			area: (*Buffer).Bounds,
			want: []string{"bbb", "ccc", "ddd", ""},
		},
		{
			name: "delete several lines",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    1, n: 2,
			area: (*Buffer).Bounds,
			want: []string{"aaa", "ddd", "", ""},
		},
		{
			name: "delete more lines than the area holds",
			rows: []string{"aaa", "bbb", "ccc"},
			y:    1, n: 99,
			area: (*Buffer).Bounds,
			want: []string{"aaa", "", ""},
		},
		{
			name: "stay inside a scroll region",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    1, n: 1,
			area: func(b *Buffer) Rectangle { return Rect(0, 1, b.Width(), 2) },
			want: []string{"aaa", "ccc", "", "ddd"},
		},
		{
			name: "fill with the given cell",
			rows: []string{"aaa", "bbb"},
			y:    0, n: 1,
			cell: &Cell{Content: "-", Width: 1},
			area: (*Buffer).Bounds,
			want: []string{"bbb", "---"},
		},
		{
			name: "only touch cells inside a narrow area",
			rows: []string{"abc", "def", "ghi"},
			y:    0, n: 1,
			area: func(b *Buffer) Rectangle { return Rect(1, 0, 1, b.Height()) },
			want: []string{"aec", "dhf", "g i"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := newTestBuffer(c.rows...)
			b.DeleteLineArea(c.y, c.n, c.cell, c.area(b))
			if got := b.String(); got != strings.Join(c.want, "\n") {
				t.Errorf("got:\n%s\nwant:\n%s", got, strings.Join(c.want, "\n"))
			}
		})
	}
}

func TestInsertLineArea(t *testing.T) {
	cases := []struct {
		name string
		rows []string
		y, n int
		cell *Cell
		area func(b *Buffer) Rectangle
		want []string
	}{
		{
			name: "scroll the whole buffer down",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    0, n: 1,
			area: (*Buffer).Bounds,
			want: []string{"", "aaa", "bbb", "ccc"},
		},
		{
			name: "insert several lines",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    1, n: 2,
			area: (*Buffer).Bounds,
			want: []string{"aaa", "", "", "bbb"},
		},
		{
			name: "insert more lines than the area holds",
			rows: []string{"aaa", "bbb", "ccc"},
			y:    1, n: 99,
			area: (*Buffer).Bounds,
			want: []string{"aaa", "", ""},
		},
		{
			name: "stay inside a scroll region",
			rows: []string{"aaa", "bbb", "ccc", "ddd"},
			y:    1, n: 1,
			area: func(b *Buffer) Rectangle { return Rect(0, 1, b.Width(), 2) },
			want: []string{"aaa", "", "bbb", "ddd"},
		},
		{
			name: "fill with the given cell",
			rows: []string{"aaa", "bbb"},
			y:    0, n: 1,
			cell: &Cell{Content: "-", Width: 1},
			area: (*Buffer).Bounds,
			want: []string{"---", "aaa"},
		},
		{
			name: "only touch cells inside a narrow area",
			rows: []string{"abc", "def", "ghi"},
			y:    0, n: 1,
			area: func(b *Buffer) Rectangle { return Rect(1, 0, 1, b.Height()) },
			want: []string{"a c", "dbf", "gei"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := newTestBuffer(c.rows...)
			b.InsertLineArea(c.y, c.n, c.cell, c.area(b))
			if got := b.String(); got != strings.Join(c.want, "\n") {
				t.Errorf("got:\n%s\nwant:\n%s", got, strings.Join(c.want, "\n"))
			}
		})
	}
}

// TestDeleteLineAreaWideCell checks that a wide fill cell still goes through
// [Line.Set], so its zero-width placeholders are written.
func TestDeleteLineAreaWideCell(t *testing.T) {
	wide := &Cell{Content: "你", Width: 2}

	b := newTestBuffer("abcd", "efgh")
	b.DeleteLineArea(0, 1, wide, b.Bounds())

	if got := b.Lines[0].String(); got != "efgh" {
		t.Errorf("line was not scrolled up: got %q, want %q", got, "efgh")
	}

	// The filled line must look like one written cell by cell through Line.Set.
	want := make(Line, b.Width())
	for x := range want {
		want.Set(x, wide)
	}
	for x := range want {
		if got := b.Lines[1][x]; !got.Equal(&want[x]) {
			t.Errorf("cell (%d,1): got %+v, want %+v", x, got, want[x])
		}
	}
}

func BenchmarkDeleteLineArea(b *testing.B) {
	buf := NewBuffer(120, 40)
	cell := &Cell{Content: "x", Width: 1}
	for y := range buf.Height() {
		for x := range buf.Width() {
			buf.SetCell(x, y, cell)
		}
	}
	area := buf.Bounds()

	b.ReportAllocs()
	for b.Loop() {
		buf.DeleteLineArea(0, 1, nil, area)
	}
}
