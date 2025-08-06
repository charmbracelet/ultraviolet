package uv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
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

	ctx       context.Context    // ctx controls the reader's loop.
	cancel    context.CancelFunc // cancel can be used to cancel the reader's loop.
	eventc    chan Event         // eventc is a channel for sending events to.
	lookup    bool               // lookup indicates whether to use the lookup table for key sequences.
	buf       []byte             // buffer to hold the read data.
	events    []Event            // queued events to be sent.
	eventsIdx int                // eventsIdx is the index of the next event to be sent.
	ttimeout  time.Time          // ttimeout is the time at which the last input was received.
	runOnce   sync.Once          // runOnce is used to ensure that the reader is only started once.

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState

	// This indicates whether the reader is closed or not. It is used to
	// prevent	multiple calls to the Close() method.
	donec     chan struct{} // close is a channel used to signal the reader to close.
	closeOnce sync.Once
	closed    atomic.Bool // closed indicates whether the scanner has been closed.
	notify    chan []byte // notify is a channel used to notify the reader of new input events.
	timeout   *time.Timer
	err       atomic.Pointer[error] // err is the last error encountered by the reader.

	logger Logger // The logger to use for debugging.
}

// This is to silence the linter warning about the [win32InputState] not being
// used.
var _ any = &TerminalReader{
	keyState: win32InputState{},
}

// NewTerminalReader returns a new input event scanner. The scanner scans input
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
//	sc := NewTerminalReader(cr, os.Getenv("TERM"))
//	for sc.Scan() {
//	    log.Printf("event: %v", sc.Event())
//	}
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
	d.eventc = make(chan Event)
	d.timeout = time.NewTimer(d.EscTimeout)
	d.notify = make(chan []byte)
	d.donec = make(chan struct{})
	d.closeOnce = sync.Once{}
	d.runOnce = sync.Once{}
	d.eventsIdx = -1 // indicates that no events have been read yet.
	return d
}

// Events returns a channel that receives events from the scanner.
// This channel is never closed, so it is safe to use in a loop.
func (d *TerminalReader) Events() chan Event {
	return d.eventc
}

// Start starts the event loop. It returns an error if the event loop has
// already been started.
func (d *TerminalReader) Start() error {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-d.ctx.Done():
				return
			default:
				var readBuf [4096]byte
				n, err := d.r.Read(readBuf[:])
				if err != nil {
					if errors.Is(err, io.EOF) || errors.Is(err, cancelreader.ErrCanceled) {
						return
					}
					d.err.Store(&err)
					return
				}

				d.logf("input: %q", readBuf[:n])
				// This handles small inputs that start with an ESC like:
				// - "\x1b" (escape key press)
				// - "\x1b\x1b" (alt+escape key press)
				// - "\x1b[" (alt+[ key press)
				// - "\x1bP" (alt+shift+p key press)
				// - "\x1bX" (alt+shift+x key press)
				// - "\x1bO" (alt+shift+o key press)
				// - "\x1b_" (alt+_ key press)
				// - "\x1b^" (alt+^ key press)
				select {
				case <-d.ctx.Done():
					return
				case d.notify <- readBuf[:n]:
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-d.ctx.Done():
				return
			}
		}
	}()

	d.timeout.Reset(d.EscTimeout)
	d.ttimeout = time.Now().Add(d.EscTimeout)

	return nil
}

// SetLogger sets the logger to use for debugging. If nil, no logging will be
// performed.
func (d *TerminalReader) SetLogger(logger Logger) {
	d.logger = logger
}

// Close closes the scanner loop if it is still running. It also returns the
// first non-EOF error encountered by the scanner.
func (d *TerminalReader) Close() error {
	d.closeEvents()
	errp := d.err.Load()
	if errp == nil {
		return nil
	}
	return *errp
}

func (d *TerminalReader) closeEvents() {
	d.closeOnce.Do(func() {
		d.closed.Store(true)
		close(d.donec) // signal the reader to close
	})
}

// Event returns the last read event from the scanner. It returns nil if no
// event is available.
func (d *TerminalReader) Event() Event {
	if len(d.events) == 0 || d.eventsIdx >= len(d.events) {
		return nil
	}
	return d.events[d.eventsIdx]
}

// Err returns the first non-EOF error encountered by the scanner. If no error
// has been encountered, it returns nil.
func (d *TerminalReader) Err() error {
	errp := d.err.Load()
	if errp == nil {
		return nil // no error encountered
	}
	return *errp
}

func (d *TerminalReader) runDefault() {
	defer d.closeEvents() // close events channel when done
	for {
		var readBuf [4096]byte
		n, err := d.r.Read(readBuf[:])
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, cancelreader.ErrCanceled) {
				return
			}
			d.err.Store(&err)
			return
		}

		d.logf("input: %q", readBuf[:n])
		// This handles small inputs that start with an ESC like:
		// - "\x1b" (escape key press)
		// - "\x1b\x1b" (alt+escape key press)
		// - "\x1b[" (alt+[ key press)
		// - "\x1bP" (alt+shift+p key press)
		// - "\x1bX" (alt+shift+x key press)
		// - "\x1bO" (alt+shift+o key press)
		// - "\x1b_" (alt+_ key press)
		// - "\x1b^" (alt+^ key press)
		d.notify <- readBuf[:n]
	}
}

func (d *TerminalReader) scanLast() bool {
	// Process any remaining events before closing.
	if len(d.buf) > 0 && d.processEvents(true) {
		d.eventsIdx++
		return true
	}
	return false
}

func (d *TerminalReader) scan() bool {
	d.runOnce.Do(func() {
		go d.run()
		d.timeout.Reset(d.EscTimeout)
		d.ttimeout = time.Now().Add(d.EscTimeout)
	})
	for {
		if d.eventsIdx >= len(d.events) {
			// Reset the buffer if we have processed all events.
			d.events = d.events[:0]
			d.eventsIdx = -1
		} else if len(d.events) > 0 && d.eventsIdx+1 < len(d.events) {
			// If there are events available, increment the index and return true.
			d.eventsIdx++
			return true
		} else if d.closed.Load() {
			return d.scanLast()
		}
		select {
		case <-d.donec:
			return d.scanLast()
		case <-d.timeout.C:
			// Timeout reached process the buffer including any incomplete sequences.
			var hasEvents bool
			if len(d.buf) > 0 {
				if time.Now().After(d.ttimeout) {
					hasEvents = d.processEvents(true)
				}
			}

			if len(d.buf) > 0 {
				if !d.timeout.Stop() {
					// drain the channel if it was already running
					select {
					case <-d.timeout.C:
					default:
					}
				}

				d.timeout.Reset(d.EscTimeout)
			}

			if hasEvents {
				d.eventsIdx++
				return true
			}

		case buf := <-d.notify:
			d.buf = append(d.buf, buf...)
			d.ttimeout = time.Now().Add(d.EscTimeout)
			hasEvents := d.processEvents(false)
			if !d.timeout.Stop() {
				// drain the channel if it was already running
				select {
				case <-d.timeout.C:
				default:
				}
			}

			if len(d.buf) > 0 {
				d.timeout.Reset(d.EscTimeout)
			}

			if hasEvents {
				d.eventsIdx++
				return true
			}
		}
	}
}

// processEventsDefault processes the events in the queue and returns true if an event
// was processed.
func (d *TerminalReader) processEventsDefault(expired bool) bool {
	// Lookup table first
	if d.lookup && len(d.buf) > 2 && d.buf[0] == ansi.ESC {
		if k, ok := d.table[string(d.buf)]; ok {
			d.events = append(d.events, KeyPressEvent(k))
			d.buf = d.buf[:0]
			return true
		}
	}

	for len(d.buf) > 0 {
		esc := d.buf[0] == ansi.ESC
		n, event := d.Decode(d.buf)

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := event.(PasteEndEvent); !ok {
				d.paste = append(d.paste, d.buf[0])
				d.buf = d.buf[1:]
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
				return false // wait for more input
			}

			if k, ok := d.table[string(d.buf[:n])]; ok {
				d.events = append(d.events, KeyPressEvent(k))
				d.buf = d.buf[n:]
				return true
			}

			d.events = append(d.events, event)
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
			d.events = append(d.events, PasteEvent(paste))
		}

		if !isUnknown && event != nil {
			if esc && n <= 2 && !expired {
				// Wait for more input
				return false
			}

			if m, ok := event.(MultiEvent); ok {
				// If the event is a MultiEvent, append all events to the queue.
				d.events = append(d.events, m...)
			} else {
				// Otherwise, just append the event to the queue.
				d.events = append(d.events, event)
			}
		}

		d.buf = d.buf[n:]
	}

	return len(d.events) > 0
}

func (d *TerminalReader) logf(format string, v ...interface{}) {
	if d.logger == nil {
		return
	}
	d.logger.Printf(format, v...)
}
