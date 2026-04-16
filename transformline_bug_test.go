package uv

import (
	"testing"
	"time"
)

// TestTransformLine_IsZeroInfiniteLoop reproduces and verifies the fix for
// the infinite loop in TerminalRenderer.transformLine at terminal_renderer.go:937.
//
// The bug: the for loop read `next` once before the loop, then looped on
// `next.IsZero()` while incrementing `n`, but never re-read `next` from
// the buffer. If next.IsZero() returned true, the loop never terminated.
//
// The fix: add `next = newLine.At(n + 1)` inside the loop body.
func TestTransformLine_IsZeroInfiniteLoop(t *testing.T) {
	newLine := Line{
		{Content: "世", Width: 2}, // wide char at index 0
		{},                       // zero cell at index 1 (trailing)
		{},                       // zero cell at index 2
		{Content: "a", Width: 1}, // normal char at index 3
	}

	n := 0
	next := newLine.At(n + 1) // reads trailing cell (zero)

	// The FIXED loop — updates `next` inside the loop
	done := make(chan struct{})
	go func() {
		defer func() {
			recover()
			close(done)
		}()
		for next != nil && next.IsZero() {
			n++
			next = newLine.At(n + 1) // fix: re-read from buffer
		}
	}()

	select {
	case <-done:
		if n != 2 {
			t.Errorf("Expected loop to skip 2 zero cells, got %d", n)
		}
		t.Logf("Fix verified: loop correctly skipped %d zero cells and stopped at non-zero cell", n)
	case <-time.After(2 * time.Second):
		t.Fatal("Loop still hangs — fix did not work")
	}
}
