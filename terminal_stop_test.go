package uv

import (
	"runtime"
	"testing"
)

// TestStopWithInputInFlightDoesNotPanic is the regression test for #177. Stop
// set t.pr to nil while the input loop could still be about to evaluate it, so
// the loop dereferenced nil, in ultraviolet's own goroutine where the
// application cannot recover.
//
// each round starts a terminal with a key in flight and stops it at once.
// there is nothing to assert beyond surviving the rounds.
//
// the test pins GOMAXPROCS to 1. with one processor the input loop has not
// been scheduled by the time Stop returns, so the bug panics in the first
// round on every run. with several, the loop usually gets to Read first on
// another processor, and the bug went unseen in every -race run tried.
func TestStopWithInputInFlightDoesNotPanic(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))

	for i := range 1000 {
		con := newPipeConsole()
		tm := NewTerminal(con, nil)
		if err := tm.Start(); err != nil {
			t.Fatalf("round %d: Start: %v", i, err)
		}
		go con.send("x")
		if err := tm.Stop(); err != nil {
			t.Fatalf("round %d: Stop: %v", i, err)
		}
		// with the fix the input loop may still be blocked on the pipe;
		// closing it fails that read and lets the goroutine exit.
		_ = con.feed.Close()
	}
}
