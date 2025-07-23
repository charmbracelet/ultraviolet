package uv

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/cancelreader"
)

// Logger is a simple logger interface.
type Logger interface {
	Printf(format string, v ...interface{})
}

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

// DefaultEscTimeout is the default timeout at which the [InputScanner] will
// process ESC sequences. It is set to 50 milliseconds.
const DefaultEscTimeout = 50 * time.Millisecond

// InputScanner represents an input event reader. It reads input events and
// parses escape sequences from the terminal input buffer and translates them
// into human-readable events.
type InputScanner struct {
	SequenceParser

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
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	lookup bool    // lookup indicates whether to use the lookup table for key sequences.
	buf    []byte  // buffer to hold the read data.
	events []Event // queued events to be sent.

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState

	// This indicates whether the reader is closed or not. It is used to
	// prevent	multiple calls to the Close() method.
	started   bool          // started indicates whether the reader has been started.
	runOnce   sync.Once     // runOnce is used to ensure that the reader is only started once.
	donec     chan struct{} // close is a channel used to signal the reader to close.
	closeOnce sync.Once
	notify    chan []byte // notify is a channel used to notify the reader of new input events.
	timeout   *time.Timer
	timedout  atomic.Bool
	esc       atomic.Bool
	err       atomic.Pointer[error] // err is the last error encountered by the reader.

	logger Logger // The logger to use for debugging.
}

// This is to silence the linter warning about the [win32InputState] not being
// used.
var _ any = &InputScanner{
	keyState: win32InputState{},
}

// NewInputScanner returns a new input event reader. The reader reads input
// events from the terminal and parses escape sequences into human-readable
// events. It supports reading Terminfo databases.
//
// Use [InputScanner.UseTerminfo] to use Terminfo defined key sequences.
// Use [InputScanner.Legacy] to control legacy key encoding behavior.
//
// Example:
//
//	r, _ := input.NewInputScanner(os.Stdin, os.Getenv("TERM"))
//	defer r.Close()
//	events, _ := r.ReadEvents()
//	for _, ev := range events {
//	  log.Printf("%v", ev)
//	}
func NewInputScanner(r io.Reader, termType string) (*InputScanner, error) {
	d := &InputScanner{
		EscTimeout: DefaultEscTimeout,
		r:          r,
		term:       termType,
		lookup:     true, // Use lookup table by default.
	}
	rd, err := newCancelreader(d.r)
	if err != nil {
		return nil, err
	}
	d.rd = rd
	if d.table == nil {
		d.table = buildKeysTable(d.Legacy, d.term, d.UseTerminfo)
	}
	d.started = true
	d.esc.Store(false)
	d.timeout = time.NewTimer(d.EscTimeout)
	d.notify = make(chan []byte)
	d.donec = make(chan struct{})
	d.closeOnce = sync.Once{}
	d.runOnce = sync.Once{}
	return d, nil
}

// SetLogger sets the logger to use for debugging. If nil, no logging will be
// performed.
func (d *InputScanner) SetLogger(logger Logger) {
	d.logger = logger
}

// Cancel cancels the underlying reader.
func (d *InputScanner) Cancel() bool {
	if d.rd == nil {
		return false
	}
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *InputScanner) Close() (rErr error) {
	if !d.started {
		return ErrReaderNotStarted
	}
	return d.close() //nolint:wrapcheck
}

func (d *InputScanner) close() error {
	if err := d.rd.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	d.started = false
	d.closeEvents()
	d.timedout.Store(true)
	errp := d.err.Load()
	if errp == nil {
		return nil
	}
	err := *errp
	if errors.Is(err, io.EOF) || errors.Is(err, cancelreader.ErrCanceled) {
		return nil
	}
	return err
}

func (d *InputScanner) closeEvents() {
	d.closeOnce.Do(func() {
		close(d.donec) // signal the reader to close
	})
}

// Event returns the last read event from the scanner. It returns nil if no
// event is available.
func (d *InputScanner) Event() Event {
	if len(d.events) == 0 {
		return nil
	}
	ev := d.events[0]
	return ev
}

// Err returns the first non-EOF error encountered by the scanner. If no error
// has been encountered, it returns nil.
func (d *InputScanner) Err() error {
	errp := d.err.Load()
	if errp == nil {
		return nil // no error encountered
	}
	err := *errp
	return err // return the first non-EOF error
}

func (d *InputScanner) scan() bool {
	// Start the reader loop if it hasn't been started yet.
	d.runOnce.Do(func() {
		go func() {
			defer d.closeEvents() // ensure we close the events channel when done
			d.run()
		}()
	})

	if len(d.events) > 0 {
		// Advance the event queue if there are events available.
		d.events = d.events[1:]
		return true
	}

	select {
	case <-d.donec:
		return false
	case <-d.timeout.C:
		d.timedout.Store(true)
		d.sendEvents()
		d.esc.Store(false)
	case buf := <-d.notify:
		d.buf = append(d.buf, buf...)
		if !d.esc.Load() {
			d.sendEvents()
			d.timedout.Store(false)
		}
	}

	return true
}

func (d *InputScanner) run() {
	for {
		var readBuf [256]byte
		n, err := d.rd.Read(readBuf[:])
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
		esc := n > 0 && n <= 2 && readBuf[0] == ansi.ESC
		if esc {
			d.resetEsc()
		}

		d.notify <- readBuf[:n]
	}
}

func (d *InputScanner) resetEsc() {
	// Reset the escape sequence state and timer.
	d.esc.Store(true)
	d.timeout.Reset(d.EscTimeout)
}

func (d *InputScanner) sendEvents() {
	// Lookup table first
	if d.lookup && d.timedout.Load() && len(d.buf) > 2 && d.buf[0] == ansi.ESC {
		if k, ok := d.table[string(d.buf)]; ok {
			d.events = append(d.events, KeyPressEvent(k))
			d.buf = d.buf[:0]
			return
		}
	}

LOOP:
	for len(d.buf) > 0 {
		nb, ev := d.parseSequence(d.buf)

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndEvent); !ok {
				d.paste = append(d.paste, d.buf[0])
				d.buf = d.buf[1:]
				continue
			}
		}

		var isUnknownEvent bool
		switch ev.(type) {
		case ignoredEvent:
			ev = nil // ignore this event
		case UnknownEvent:
			isUnknownEvent = true

			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(d.buf[:nb])]; ok {
				ev = KeyPressEvent(k)
			}

			d.logf("unknown sequence: %q", d.buf[:nb])
			if !d.timedout.Load() {
				if nb > 0 {
					// This handles unknown escape sequences that might be incomplete.
					if slices.Contains([]byte{
						ansi.ESC, ansi.CSI, ansi.OSC, ansi.DCS, ansi.APC, ansi.SOS, ansi.PM,
					}, d.buf[0]) {
						d.resetEsc()
					}
				}
				// If this is the entire buffer, we can break and assume this
				// is an incomplete sequence.
				break LOOP
			}
			d.logf("timed out, skipping unknown sequence: %q", d.buf[:nb])
		case PasteStartEvent:
			d.paste = []byte{}
		case PasteEndEvent:
			// Decode the captured data into runes.
			var paste []rune
			for len(d.paste) > 0 {
				r, w := utf8.DecodeRune(d.paste)
				if r != utf8.RuneError {
					paste = append(paste, r)
				}
				d.paste = d.paste[w:]
			}
			d.paste = nil // reset the d.buffer
			d.events = append(d.events, PasteEvent(paste))
		}

		if ev != nil {
			if !isUnknownEvent && d.esc.Load() {
				// If we are in an escape sequence, and the event is a valid
				// one, we need to reset the escape state.
				d.esc.Store(false)
			}

			if mevs, ok := ev.(MultiEvent); ok {
				for _, mev := range mevs {
					d.events = append(d.events, mev)
				}
			} else {
				d.events = append(d.events, ev)
			}
		}

		d.buf = d.buf[nb:]
	}
}

func (d *InputScanner) logf(format string, v ...interface{}) {
	if d.logger == nil {
		return
	}
	d.logger.Printf(format, v...)
}
