package uv

import (
	"context"
	"io"
	"testing"
	"time"
)

// A key held down repeats faster than EscTimeout. Keys that are ambiguous on
// their own, like Alt+F sent as ESC f, must still be delivered while the key
// is held rather than after it is released.
func TestStreamEventsDeliversHeldAltKeysWhileHeld(t *testing.T) {
	const n = 20
	gap := 30 * time.Millisecond

	pr, pw := io.Pipe()
	defer pw.Close()
	tr := NewTerminalReader(pr, "xterm-256color")
	events := make(chan Event, 4*n)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = tr.StreamEvents(ctx, events) }()

	sent := make([]time.Time, 0, n)
	got := make(chan time.Time, 4*n)
	go func() {
		for ev := range events {
			if _, ok := ev.(KeyPressEvent); ok {
				got <- time.Now()
			}
		}
	}()

	for i := 0; i < n; i++ {
		sent = append(sent, time.Now())
		if _, err := pw.Write([]byte("\x1bf")); err != nil {
			t.Fatal(err)
		}
		time.Sleep(gap)
	}

	for i := 0; i < n; i++ {
		select {
		case at := <-got:
			if delay := at.Sub(sent[i]); delay > 3*tr.EscTimeout {
				t.Fatalf("key %d reached the consumer %v after it was written; held keys must not wait for the key to be released", i, delay)
			}
		case <-time.After(time.Second):
			t.Fatalf("only %d of %d keys arrived", i, n)
		}
	}
}
