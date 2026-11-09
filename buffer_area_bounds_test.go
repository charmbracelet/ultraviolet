package uv

import "testing"

// These tests cover the area-bounds invariant of InsertLineArea and
// DeleteLineArea: an area extending past the buffer bounds (e.g. a terminal
// scroll region computed before a resize) must be clamped instead of
// indexing past the end of b.Lines.

func TestInsertLineAreaClampsOversizedArea(t *testing.T) {
	b := NewBuffer(80, 63)
	area := Rect(0, 0, 80, 64) // Max.Y exceeds Height by one

	// Used to panic: index out of range [63] with length 63.
	b.InsertLineArea(0, 1, nil, area)
}

func TestDeleteLineAreaClampsOversizedArea(t *testing.T) {
	b := NewBuffer(80, 63)
	area := Rect(0, 0, 80, 64)

	b.DeleteLineArea(0, 1, nil, area)
}

func TestInsertLineAreaClampsOversizedWidth(t *testing.T) {
	b := NewBuffer(80, 24)
	area := Rect(0, 0, 100, 24) // Max.X exceeds Width

	b.InsertLineArea(0, 1, nil, area)
}

func TestInsertLineAreaStillShiftsLinesWithinBounds(t *testing.T) {
	b := NewBuffer(4, 3)
	c := &Cell{Content: "a", Width: 1}
	b.SetCell(0, 0, c)

	b.InsertLineArea(0, 1, nil, b.Bounds())

	if got := b.CellAt(0, 1); got == nil || got.Content != "a" {
		t.Errorf("line 0 was not shifted down to line 1, got %v", got)
	}
}
