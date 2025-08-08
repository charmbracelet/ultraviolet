package uv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/cancelreader"
)

// win32InputState is a state machine for parsing key events from the Windows
// Console API into escape sequences and utf8 runes, and keeps track of the last
// control key state to determine modifier key changes. It also keeps track of
// the last mouse button state and window size changes to determine which mouse
// buttons were released and to prevent multiple size events from firing.
type win32InputState struct {
	ansiBuf                    [256]byte
	ansiIdx                    int
	utf16Buf                   [2]rune
	utf16Half                  bool
	lastCks                    uint32 // the last control key state for the previous event
	lastMouseBtns              uint32 // the last mouse button state for the previous event
	lastWinsizeX, lastWinsizeY int16  // the last window size for the previous event to prevent multiple size events from firing
}

// This is to silence the linter warning about the win32InputState not being
// used.
var _ any = win32InputState{
	ansiBuf:       [256]byte{},
	ansiIdx:       0,
	utf16Buf:      [2]rune{},
	utf16Half:     false,
	lastCks:       0,
	lastMouseBtns: 0,
	lastWinsizeX:  0,
	lastWinsizeY:  0,
}

// ErrReaderNotStarted is returned when the reader has not been started yet.
var ErrReaderNotStarted = fmt.Errorf("reader not started")

// DefaultEscTimeout is the default timeout at which the [TerminalReader] will
// process ESC sequences. It is set to 50 milliseconds.
const DefaultEscTimeout = 50 * time.Millisecond

// TerminalReader represents an input event loop that reads input events from
// a reader and parses them into human-readable events. It supports
// reading escape sequences, mouse events, and bracketed paste mode.
type TerminalReader struct {
	EventDecoder

	// MouseMode determines whether mouse events are enabled or not. This is a
	// platform-specific feature and is only available on Windows. When this is
	// true, the reader will be initialized to read mouse events using the
	// Windows Console API.
	MouseMode *MouseMode

	// EscTimeout is the escape character timeout duration. Most escape
	// sequences start with an escape character [ansi.ESC] and are followed by
	// one or more characters. If the next character is not received within
	// this timeout, the reader will assume that the escape sequence is
	// complete and will process the received characters as a complete escape
	// sequence.
	//
	// By default, this is set to [DefaultEscTimeout] (50 milliseconds).
	EscTimeout time.Duration

	r     io.Reader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	lookup bool   // lookup indicates whether to use the lookup table for key sequences.
	buf    []byte // buffer to hold the read data.

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState

	logger Logger // The logger to use for debugging.
}

// This is to silence the linter warning about the [win32InputState] not being
// used.
var _ any = &TerminalReader{
	keyState: win32InputState{},
}

// NewTerminalReader returns a new input event reader. The reader streams input
// events from the terminal and parses escape sequences into human-readable
// events. It supports reading Terminfo databases.
//
// Use [TerminalReader.UseTerminfo] to use Terminfo defined key sequences.
// Use [TerminalReader.Legacy] to control legacy key encoding behavior.
//
// Example:
//
//	```go
//	var cr cancelreader.CancelReader
//	var evc chan Event
//	sc := NewTerminalReader(cr, os.Getenv("TERM"))
//	go sc.StreamEvents(ctx, evc)
//	```
func NewTerminalReader(r io.Reader, termType string) *TerminalReader {
	d := &TerminalReader{
		EscTimeout: DefaultEscTimeout,
		r:          r,
		term:       termType,
		lookup:     true, // Use lookup table by default.
	}
	d.r = r
	if d.table == nil {
		d.table = buildKeysTable(d.Legacy, d.term, d.UseTerminfo)
	}
	return d
}

// readBufSize is the size of the read buffer used to read input events at a time.
const readBufSize = 4096

// sendBytes reads data from the reader and sends it to the provided channel.
// It stops when an error occurs or when the context is closed.
func (d *TerminalReader) sendBytes(ctx context.Context, readc chan []byte) error {
	for {
		var readBuf [readBufSize]byte
		n, err := d.r.Read(readBuf[:])
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case readc <- readBuf[:n]:
		}
	}
}

// StreamEvents sends events to the provided channel. It stops when the context
// is closed or when an error occurs.
func (d *TerminalReader) StreamEvents(ctx context.Context, eventc chan<- Event) error {
	var buf bytes.Buffer
	errc := make(chan error, 1)
	readc := make(chan []byte)
	recordc := make(chan []inputRecord)
	timeout := time.NewTimer(d.EscTimeout)
	ttimeout := time.Now().Add(d.EscTimeout)

	go func() {
		if err := d.streamData(ctx, readc, recordc); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, cancelreader.ErrCanceled) {
				errc <- nil
				return
			}
			errc <- err
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			d.sendEvents(buf.Bytes(), true, eventc)
			return nil
		case err := <-errc:
			d.sendEvents(buf.Bytes(), true, eventc)
			return err // return the first error encountered
		case <-timeout.C:
			d.logf("timeout reached")

			// Timeout reached process the buffer including any incomplete sequences.
			var n int // n is the number of bytes processed
			if buf.Len() > 0 {
				if time.Now().After(ttimeout) {
					d.logf("timeout expired, processing buffer")
					n = d.sendEvents(buf.Bytes(), true, eventc)
				}
			}

			if buf.Len() > 0 {
				if !timeout.Stop() {
					// drain the channel if it was already running
					select {
					case <-timeout.C:
					default:
					}
				}

				d.logf("resetting timeout for remaining buffer")
				timeout.Reset(d.EscTimeout)
			}

			if n > 0 {
				buf.Next(n)
			}
		case records := <-recordc:
			d.processRecords(records, eventc)
		case read := <-readc:
			d.logf("input: %q", read)
			buf.Write(read)
			ttimeout = time.Now().Add(d.EscTimeout)
			n := d.sendEvents(buf.Bytes(), false, eventc)
			if !timeout.Stop() {
				// drain the channel if it was already running
				select {
				case <-timeout.C:
				default:
				}
			}

			if len(d.buf) > 0 {
				d.logf("resetting timeout for remaining buffer after parse")
				timeout.Reset(d.EscTimeout)
			}

			if n > 0 {
				buf.Next(n)
			}
		}
	}
}

// SetLogger sets the logger to use for debugging. If nil, no logging will be
// performed.
func (d *TerminalReader) SetLogger(logger Logger) {
	d.logger = logger
}

func (d *TerminalReader) sendEvents(buf []byte, expired bool, eventc chan<- Event) int {
	// Lookup table first
	if d.lookup && len(buf) > 2 && buf[0] == ansi.ESC {
		if k, ok := d.table[string(d.buf)]; ok {
			eventc <- KeyPressEvent(k)
			return len(buf)
		}
	}

	// total is the total number of bytes processed
	var total int
	for len(buf) > 0 {
		esc := buf[0] == ansi.ESC
		n, event := d.Decode(buf)

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := event.(PasteEndEvent); !ok {
				d.paste = append(d.paste, buf[:n]...)
				buf = buf[n:]
				total += n
				continue
			}
		}

		var isUnknown bool
		switch event.(type) {
		case ignoredEvent:
			// ignore this event
			event = nil
		case UnknownEvent:
			isUnknown = true
			// Try to look up the event in the table.
			if !expired {
				return total
			}

			if k, ok := d.table[string(buf[:n])]; ok {
				eventc <- KeyPressEvent(k)
				return total + n
			}

			eventc <- event
		case PasteStartEvent:
			d.paste = []byte{} // reset the paste buffer
		case PasteEndEvent:
			var paste []rune
			for len(d.paste) > 0 {
				r, w := utf8.DecodeRune(d.paste)
				if r != utf8.RuneError {
					paste = append(paste, r)
				}
				d.paste = d.paste[w:]
			}
			d.paste = nil // reset the paste buffer
			eventc <- PasteEvent(paste)
		}

		if !isUnknown && event != nil {
			if esc && n <= 2 && !expired {
				// Wait for more input
				return total
			}

			if m, ok := event.(MultiEvent); ok {
				// If the event is a MultiEvent, append all events to the queue.
				for _, e := range m {
					eventc <- e
				}
			} else {
				// Otherwise, just append the event to the queue.
				eventc <- event
			}
		}

		buf = buf[n:]
		total += n
	}

	return total
}

func (d *TerminalReader) logf(format string, v ...interface{}) {
	if d.logger == nil {
		return
	}
	d.logger.Printf(format, v...)
}
