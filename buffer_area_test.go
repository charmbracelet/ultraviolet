package uv

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// TestLineAreaOpsClampToBuffer: a scroll region taller or wider than
// the buffer must not index past it. Regression for a panic in
// InsertLineArea when a terminal's DECSTBM region outlived a shrink.
func TestLineAreaOpsClampToBuffer(t *testing.T) {
	b := NewBuffer(10, 5)
	for y := 0; y < 5; y++ {
		b.SetCell(0, y, &Cell{Content: string(rune('a' + y)), Width: 1})
	}
	tall := Rect(0, 0, 10, 6) // one row past the end
	wide := Rect(0, 0, 11, 5) // one column past the end

	b.InsertLineArea(0, 1, nil, tall)
	if got := b.CellAt(0, 1).Content; got != "a" {
		t.Fatalf("after insert with oversized region, row 1 = %q, want a", got)
	}
	b.DeleteLineArea(0, 1, nil, tall)
	if got := b.CellAt(0, 0).Content; got != "a" {
		t.Fatalf("after delete with oversized region, row 0 = %q, want a", got)
	}
	b.InsertLineArea(0, 1, nil, wide)
	b.DeleteLineArea(0, 1, nil, wide)

	// A region entirely outside the buffer is a no-op, not a panic.
	b.InsertLineArea(7, 1, nil, Rect(0, 6, 10, 9))
	b.DeleteLineArea(7, 1, nil, Rect(0, 6, 10, 9))
}

// touchedRows returns the rows the renderer would repaint, by the same test
// Render applies to each entry of the touch list.
func touchedRows(b *RenderBuffer) []int {
	var rows []int
	for y, ld := range b.Touched {
		if ld != nil && (ld.FirstCell != -1 || ld.LastCell != -1) {
			rows = append(rows, y)
		}
	}
	return rows
}

func setRow(b interface{ SetCell(x, y int, c *Cell) }, y int, s string) {
	for x, r := range s {
		b.SetCell(x, y, &Cell{Content: string(r), Width: 1})
	}
}

func rowText(b *Buffer, y int) string {
	return TrimSpace(b.Line(y).String())
}

func bufferRows(b *Buffer) []string {
	rows := make([]string, b.Height())
	for y := range rows {
		rows[y] = rowText(b, y)
	}
	return rows
}

// TestRenderBufferShrinkUnderScrollRegion: a terminal sets a scroll region,
// then the buffer shrinks below its bottom margin before the region is reset.
// Scrolling that region must clamp to the rows that remain, without a panic,
// and touch every row it changed.
func TestRenderBufferShrinkUnderScrollRegion(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int // size after the shrink
		region        Rectangle
		afterDelete   []string
		afterInsert   []string
	}{
		{
			// Rows 2..7 of the 10x8 buffer, full width: the fast path.
			name: "height", width: 10, height: 5, region: Rect(0, 2, 10, 6),
			afterDelete: []string{"aaaaaaaaaa", "bbbbbbbbbb", "dddddddddd", "eeeeeeeeee", ""},
			afterInsert: []string{"aaaaaaaaaa", "bbbbbbbbbb", "", "", "dddddddddd"},
		},
		{
			// Wider than the buffer once it narrows: clamps to full width, so
			// still the fast path.
			name: "height and width", width: 6, height: 5, region: Rect(0, 2, 10, 6),
			afterDelete: []string{"aaaaaa", "bbbbbb", "dddddd", "eeeeee", ""},
			afterInsert: []string{"aaaaaa", "bbbbbb", "", "", "dddddd"},
		},
		{
			// Columns 2..7: the shift path, its right margin cut to column 5.
			name: "partial width", width: 6, height: 5, region: Rect(2, 2, 6, 6),
			afterDelete: []string{"aaaaaa", "bbbbbb", "ccdddd", "ddeeee", "ee"},
			afterInsert: []string{"aaaaaa", "bbbbbb", "cc", "dd", "eedddd"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := NewRenderBuffer(10, 8)
			for y := 0; y < 8; y++ {
				setRow(b, y, strings.Repeat(string(rune('a'+y)), 10))
			}
			b.Resize(tc.width, tc.height)
			if b.Width() != tc.width || b.Height() != tc.height {
				t.Fatalf("size = %dx%d, want %dx%d", b.Width(), b.Height(), tc.width, tc.height)
			}

			check := func(op string, want []string) {
				t.Helper()
				if got := bufferRows(b.Buffer); !slices.Equal(got, want) {
					t.Errorf("after %s, rows = %q, want %q", op, got, want)
				}
				rows := touchedRows(b)
				for _, y := range rows {
					if y >= b.Height() {
						t.Errorf("after %s, row %d past the buffer is touched", op, y)
					}
				}
				for y := 2; y < tc.height; y++ {
					if !slices.Contains(rows, y) {
						t.Errorf("after %s, region row %d is not touched (touched %v)", op, y, rows)
					}
				}
			}

			resetTouched(b.Touched)
			b.DeleteLineArea(2, 1, nil, tc.region)
			check("delete", tc.afterDelete)

			resetTouched(b.Touched)
			b.InsertLineArea(2, 2, nil, tc.region)
			check("insert", tc.afterInsert)

			// A cursor row left below the new bottom is a no-op.
			b.DeleteLineArea(6, 1, nil, tc.region)
			b.InsertLineArea(6, 1, nil, tc.region)
			if got := bufferRows(b.Buffer); !slices.Equal(got, tc.afterInsert) {
				t.Errorf("after ops below the bottom, rows = %q, want %q", got, tc.afterInsert)
			}
		})
	}
}

// TestLineAreaOpsStaleRegion: the clamp on a plain [Buffer], for a region
// that outlived a shrink in both directions, on the fast and shift paths.
func TestLineAreaOpsStaleRegion(t *testing.T) {
	stale := Rect(0, 1, 8, 5) // rows 1..5, set when the buffer was 8x6
	for _, tc := range []struct {
		name string
		op   func(b *Buffer)
		want []string
	}{
		{"insert full width", func(b *Buffer) { b.InsertLineArea(1, 1, nil, stale) },
			[]string{"aaaa", "", "bbbb", "cccc"}},
		{"delete full width", func(b *Buffer) { b.DeleteLineArea(1, 1, nil, stale) },
			[]string{"aaaa", "cccc", "dddd", ""}},
		{"insert past the bottom margin", func(b *Buffer) { b.InsertLineArea(1, 5, nil, stale) },
			[]string{"aaaa", "", "", ""}},
		{"delete past the bottom margin", func(b *Buffer) { b.DeleteLineArea(2, 5, nil, stale) },
			[]string{"aaaa", "bbbb", "", ""}},
		{"insert with margins", func(b *Buffer) { b.InsertLineArea(1, 1, nil, Rect(1, 1, 6, 5)) },
			[]string{"aaaa", "b", "cbbb", "dccc"}},
		{"delete with margins", func(b *Buffer) { b.DeleteLineArea(1, 1, nil, Rect(1, 1, 6, 5)) },
			[]string{"aaaa", "bccc", "cddd", "d"}},
		{"insert with margins inside the buffer", func(b *Buffer) { b.InsertLineArea(1, 1, nil, Rect(1, 1, 2, 5)) },
			[]string{"aaaa", "b  b", "cbbc", "dccd"}},
		{"cursor row below the bottom", func(b *Buffer) {
			b.InsertLineArea(5, 1, nil, stale)
			b.DeleteLineArea(4, 1, nil, stale)
		}, []string{"aaaa", "bbbb", "cccc", "dddd"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBuffer(8, 6)
			for y := 0; y < 6; y++ {
				setRow(b, y, strings.Repeat(string(rune('a'+y)), 8))
			}
			b.Resize(4, 4)
			tc.op(b)
			if got := bufferRows(b); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestBulkNewlineFastPath drives the fast path the way a terminal emulator
// does under bulk output: a full-width scroll region below a fixed status
// row, one or more newlines at the bottom margin per write, each new line
// written to the row the scroll freed. The contents, the touch list, and the
// renderer's copy of the screen must all keep up.
func TestBulkNewlineFastPath(t *testing.T) {
	const width, height = 12, 6
	region := Rect(0, 1, width, height-1) // DECSTBM 2;6
	regionRows := []int{1, 2, 3, 4, 5}

	b := NewRenderBuffer(width, height)
	setRow(b, 0, "status")

	var out bytes.Buffer
	r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
	r.SetFullscreen(true)
	r.Resize(width, height)
	render := func() {
		t.Helper()
		r.Render(b)
		if err := r.Flush(); err != nil {
			t.Fatalf("flush: %v", err)
		}
		if got, want := bufferRows(r.curbuf.Buffer), bufferRows(b.Buffer); !slices.Equal(got, want) {
			t.Fatalf("renderer has %q, frame has %q", got, want)
		}
	}
	render()

	lines := 0
	for i, batch := range []int{1, 1, 1, 3, 1, 5, 2, 1, 1, 4, 7, 1} {
		// Rendering reset the touch list, so these are this batch's alone.
		b.DeleteLineArea(region.Min.Y, batch, nil, region)
		got := touchedRows(b)
		if !slices.Equal(got, regionRows) {
			t.Fatalf("batch %d: scroll touched rows %v, want %v", i, got, regionRows)
		}
		for _, y := range got {
			if ld := b.Touched[y]; ld.LastCell < width {
				t.Fatalf("batch %d: row %d touched up to column %d, want %d", i, y, ld.LastCell, width)
			}
		}

		// Write the lines that are still on screen into the rows the scroll
		// freed at the bottom.
		n := min(batch, region.Dy())
		for k := 0; k < n; k++ {
			setRow(b, height-n+k, fmt.Sprintf("line %d", lines+batch-n+1+k))
		}
		lines += batch
		render()
	}

	want := []string{"status"}
	for k := lines - len(regionRows) + 1; k <= lines; k++ {
		want = append(want, fmt.Sprintf("line %d", k))
	}
	if got := bufferRows(b.Buffer); !slices.Equal(got, want) {
		t.Errorf("rows = %q, want %q", got, want)
	}

	// Rotation moves lines and must never alias them: every row needs its
	// own storage, or writing one row would show up in another.
	seen := map[*Cell]int{}
	for y, line := range b.Lines {
		if prev, ok := seen[&line[0]]; ok {
			t.Fatalf("rows %d and %d share storage", prev, y)
		}
		seen[&line[0]] = y
	}
}
