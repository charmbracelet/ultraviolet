package uv

import (
	"fmt"
	"testing"
	"time"
)

// TestInputLoopDoesNotReuseTheReadBuffer is the regression test for #176. the
// input loop handed the event loop a slice of its own read buffer, and its
// next read could overwrite those bytes before the event loop had copied them,
// so one input sequence arrived as the bytes of the next.
//
// two different replies go in back to back, the way a terminal answers two
// queries, and every round must produce them in order.
func TestInputLoopDoesNotReuseTheReadBuffer(t *testing.T) {
	const (
		rounds = 2000
		decrpm = "\x1b[?1016;2$y"      // a mode report, eleven bytes
		kitty  = "\x1b_Gi=31;OK\x1b\\" // a kitty graphics reply; its first eleven bytes differ
	)

	con := newPipeConsole()
	tm := NewTerminal(con, nil)
	if err := tm.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer stopAfterInputEnds(t, tm, con)

	for i := range rounds {
		con.send(decrpm)
		con.send(kitty)
		got := nextReplies(tm, 2)
		if len(got) != 2 || !isModeReport(got[0]) || !isKittyReply(got[1]) {
			t.Fatalf("round %d: replies arrived as %s, want [uv.ModeReportEvent uv.KittyGraphicsEvent]",
				i, typesOf(got))
		}
	}
}

// nextReplies collects up to n events other than the window size Start
// reports on its own, and returns early with what it has if they stop
// coming. a reply the corruption swallowed shows up as a short result, not a
// hang.
func nextReplies(tm *Terminal, n int) []Event {
	deadline := time.After(time.Second)
	var got []Event
	for len(got) < n {
		select {
		case ev := <-tm.Events():
			if _, ok := ev.(WindowSizeEvent); ok {
				continue
			}
			got = append(got, ev)
		case <-deadline:
			return got
		}
	}
	return got
}

func isModeReport(ev Event) bool {
	_, ok := ev.(ModeReportEvent)
	return ok
}

func isKittyReply(ev Event) bool {
	_, ok := ev.(KittyGraphicsEvent)
	return ok
}

func typesOf(evs []Event) string {
	names := make([]string, len(evs))
	for i, ev := range evs {
		names[i] = fmt.Sprintf("%T", ev)
	}
	return fmt.Sprint(names)
}
