package uv

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestStreamEventsCanceledDelivery(t *testing.T) {
	for _, tc := range []struct {
		name, input     string
		buffer, receive int
	}{
		{name: "ordinary key", input: "q"},
		{name: "escape", input: "\x1b"},
		{name: "partial CSI", input: "\x1b["},
		{name: "partial paste", input: "q\x1b[200~pending"},
		{name: "consumer exits on quit", input: "qx", receive: 1},
		{name: "full buffer", input: "qxy", buffer: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				events := make(chan Event, tc.buffer)
				done := make(chan error, 1)
				go func() { done <- NewTerminalReader(strings.NewReader(tc.input), "xterm").StreamEvents(ctx, events) }()
				for i := 0; i < tc.receive; i++ {
					<-events
				}
				synctest.Wait()
				time.Sleep(2 * DefaultEscTimeout)
				synctest.Wait()
				cancel()
				synctest.Wait()
				select {
				case err := <-done:
					if err != nil {
						t.Fatal(err)
					}
				default:
					t.Error("event delivery remained blocked after cancellation")
					// Release the old implementation after recording the regression.
					finished := make(chan struct{})
					go func() {
						defer close(finished)
						for {
							select {
							case <-events:
							case <-done:
								return
							}
						}
					}()
					synctest.Wait()
					<-finished
				}
			})
		})
	}
}

func TestStreamEventsEOFDeliversKeys(t *testing.T) {
	events := make(chan Event, 4)
	if err := NewTerminalReader(strings.NewReader("qx"), "xterm").StreamEvents(t.Context(), events); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"q", "x"} {
		select {
		case ev := <-events:
			key, ok := ev.(KeyPressEvent)
			if !ok || key.String() != want {
				t.Fatalf("got %#v, want key %q", ev, want)
			}
		default:
			t.Fatalf("missing key %q at EOF", want)
		}
	}
	select {
	case ev, open := <-events:
		if !open {
			t.Fatal("reader closed the caller's event channel")
		}
		t.Fatalf("unexpected event: %#v", ev)
	default:
	}
}

type blockedStreamReader struct{ started, release chan struct{} }

func (r blockedStreamReader) Read([]byte) (int, error) {
	close(r.started)
	<-r.release
	return 0, io.EOF
}

func TestStreamEventsCancellationJoinsReader(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		r := blockedStreamReader{make(chan struct{}), make(chan struct{})}
		events := make(chan Event)
		done := make(chan error, 1)
		go func() { done <- NewTerminalReader(r, "xterm").StreamEvents(ctx, events) }()
		<-r.started
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Error("StreamEvents returned before its read worker exited")
		default:
		}
		close(r.release)
		synctest.Wait()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		default:
			t.Error("StreamEvents did not finish after its read worker exited")
		}
	})
}
