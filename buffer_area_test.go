package uv

import "testing"

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
