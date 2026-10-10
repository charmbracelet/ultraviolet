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

func streamFrom(t *testing.T, timeout time.Duration) (*io.PipeWriter, <-chan Event) {
	t.Helper()
	pr, pw := io.Pipe()
	tr := NewTerminalReader(pr, "xterm-256color")
	tr.EscTimeout = timeout
	events := make(chan Event, 16)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel(); _ = pw.Close() })
	go func() { _ = tr.StreamEvents(ctx, events) }()
	return pw, events
}

func nextEvent(t *testing.T, events <-chan Event, within time.Duration) Event {
	t.Helper()
	select {
	case ev := <-events:
		return ev
	case <-time.After(within):
		t.Fatalf("no event within %v", within)
		return nil
	}
}

// ESC followed by a letter can only be Alt+letter: no escape sequence starts
// that way, so there is nothing to wait for.
func TestStreamEventsDeliversAltLetterWithoutWaitingForTimeout(t *testing.T) {
	pw, events := streamFrom(t, 500*time.Millisecond)
	if _, err := pw.Write([]byte("\x1bf")); err != nil {
		t.Fatal(err)
	}
	ev := nextEvent(t, events, 200*time.Millisecond)
	k, ok := ev.(KeyPressEvent)
	if !ok || k.Code != 'f' || k.Mod&ModAlt == 0 {
		t.Fatalf("got %#v, want Alt+f", ev)
	}
}

// ESC followed by a byte that opens a longer sequence must still wait for the
// rest, even when the rest arrives in a later read.
func TestStreamEventsStillWaitsWhenAnEscapeSequenceMayFollow(t *testing.T) {
	for name, parts := range map[string][2]string{
		"CSI arrow": {"\x1b[", "C"},
		"SS3 arrow": {"\x1bO", "C"},
	} {
		pw, events := streamFrom(t, 500*time.Millisecond)
		if _, err := pw.Write([]byte(parts[0])); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
		if _, err := pw.Write([]byte(parts[1])); err != nil {
			t.Fatal(err)
		}
		ev := nextEvent(t, events, 200*time.Millisecond)
		k, ok := ev.(KeyPressEvent)
		if !ok || k.Code != KeyRight || k.Mod != 0 {
			t.Errorf("%s: got %#v, want a plain Right arrow", name, ev)
		}
	}
}
