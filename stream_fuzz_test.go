package uv_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// chunkReader is an io.Reader that hands out the data in fixed-size chunks,
// instead of everything at once. It is synchronous and never blocks, so the
// only thing it changes is where the reader's buffer boundaries fall.
type chunkReader struct {
	data  []byte
	chunk int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := r.chunk
	if n <= 0 {
		n = len(r.data)
	}
	if n > len(p) {
		n = len(p)
	}
	if n > len(r.data) {
		n = len(r.data)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

// collect reads every event a TerminalReader produces from r, and returns
// when StreamEvents has returned. A stream that neither finishes nor panics
// within the deadline fails the test.
func collect(t *testing.T, r io.Reader) []string {
	t.Helper()
	d := uv.NewTerminalReader(r, "xterm-256color")
	// A large timeout keeps the run deterministic: incomplete sequences are
	// flushed by the EOF path, never by a timer firing mid-stream.
	d.EscTimeout = 10 * time.Second

	evc := make(chan uv.Event, 4096)
	done := make(chan error, 1)
	go func() { done <- d.StreamEvents(context.Background(), evc) }()

	var evs []string
	deadline := time.After(30 * time.Second)
loop:
	for {
		select {
		case ev := <-evc:
			evs = append(evs, fmt.Sprintf("%T:%+v", ev, ev))
		case err := <-done:
			if err != nil {
				t.Fatalf("StreamEvents returned an error: %v", err)
			}
			break loop
		case <-deadline:
			t.Fatal("StreamEvents did not terminate after the input ended")
		}
	}
	for {
		select {
		case ev := <-evc:
			evs = append(evs, fmt.Sprintf("%T:%+v", ev, ev))
		default:
			return evs
		}
	}
}

// FuzzEventStream drives the TerminalReader with arbitrary byte streams and
// asserts two properties:
//
//  1. The stream terminates and does not panic, whatever the bytes are.
//  2. The decoded event sequence does not depend on where the read buffer
//     boundaries fall: reading the same bytes in one chunk or many must
//     produce the same events in the same order.
//
// This is the path a terminal, an SSH session, or anything else feeding
// untrusted bytes into a program takes, and it is where the escape-sequence
// bugs that reach bubbletea live (window-size reports, mouse, paste, kitty
// keyboard, OSC and DCS sequences).
func FuzzEventStream(f *testing.F) {
	f.Add([]byte("\x1b[8;25;81t"))            // window size report
	f.Add([]byte("\x1b[8;23000;23000t"))      // oversized window size report
	f.Add([]byte("\x1b[8;0;0t"))              // zero-sized window report
	f.Add([]byte("\x1b[<0;33;11M"))           // SGR mouse press
	f.Add([]byte("\x1b[<64;1;1M"))            // mouse wheel
	f.Add([]byte("\x1b[200~pasted\x1b[201~")) // bracketed paste
	f.Add([]byte("\x1b[I\x1b[O"))             // focus gain/loss
	f.Add([]byte("\x1b[?1u\x1b[27u"))         // kitty keyboard query and reply
	f.Add([]byte("\x1b[27;2;65~"))            // CSI ~ key
	f.Add([]byte("\x1bOP"))                   // SS3 key
	f.Add([]byte("\x1b]8;;https://charm.sh\x1b\\link\x1b]8;;\x1b\\"))
	f.Add([]byte("\x1bP>|charm(0.1.2)\x1b\\")) // DCS
	f.Add([]byte("\x1b_Ai=31;f=24\x1b\\"))     // APC
	f.Add([]byte("hello\r\n\t\x7f\x00\xff"))   // text and control bytes
	f.Add([]byte("\x1b["))                     // truncated CSI
	f.Add([]byte("\x1b]8;;no terminator"))     // truncated OSC
	f.Add([]byte("\x1b"))                      // lone escape
	f.Add([]byte("\x1b\x1b\x1b"))              // escapes only

	f.Fuzz(func(t *testing.T, data []byte) {
		whole := collect(t, bytes.NewReader(data))

		// Use the first byte to pick a chunk size, so the fuzzer explores
		// boundaries on its own.
		chunk := 1
		if len(data) > 0 {
			chunk = int(data[0]%3) + 1
		}
		chunked := collect(t, &chunkReader{data: data, chunk: chunk})

		if len(whole) != len(chunked) {
			t.Fatalf("event count depends on chunking (chunk=%d): %d vs %d\nwhole: %v\nchunked: %v",
				chunk, len(whole), len(chunked), whole, chunked)
		}
		for i := range whole {
			if whole[i] != chunked[i] {
				t.Fatalf("event %d differs with chunking (chunk=%d):\nwhole:   %v\nchunked: %v",
					i, chunk, whole, chunked)
			}
		}
	})
}
