package tv

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

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
//
//nolint:all
type win32InputState struct {
	ansiBuf                    [256]byte
	ansiIdx                    int
	utf16Buf                   [2]rune
	utf16Half                  bool
	lastCks                    uint32 // the last control key state for the previous event
	lastMouseBtns              uint32 // the last mouse button state for the previous event
	lastWinsizeX, lastWinsizeY int16  // the last window size for the previous event to prevent multiple size events from firing
}

// TerminalReader represents an input event reader. It reads input events and
// parses escape sequences from the terminal input buffer and translates them
// into human-readable events.
type TerminalReader struct {
	SequenceParser

	// MouseMode determines whether mouse events are enabled or not. This is a
	// platform-specific feature and is only available on Windows. When this is
	// true, the reader will be initialized to read mouse events using the
	// Windows Console API.
	MouseMode *MouseMode

	r     io.Reader
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	buf [256]byte // do we need a larger buffer?

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState //nolint:all

	// This indicates whether the reader is closed or not. It is used to
	// prevent	multiple calls to the Close() method.
	closed bool

	logger Logger // The logger to use for debugging.
}

// NewTerminalReader returns a new input event reader. The reader reads input
// events from the terminal and parses escape sequences into human-readable
// events. It supports reading Terminfo databases.
//
// Use [TerminalReader.UseTerminfo] to use Terminfo defined key sequences.
// Use [TerminalReader.Legacy] to control legacy key encoding behavior.
//
// Example:
//
//	r, _ := input.NewTerminalReader(os.Stdin, os.Getenv("TERM"))
//	defer r.Close()
//	events, _ := r.ReadEvents()
//	for _, ev := range events {
//	  log.Printf("%v", ev)
//	}
func NewTerminalReader(r io.Reader, termType string) *TerminalReader {
	return &TerminalReader{
		r:    r,
		term: termType,
	}
}

// SetLogger sets the logger to use for debugging. If nil, no logging will be
// performed.
func (d *TerminalReader) SetLogger(logger Logger) {
	d.logger = logger
}

// Start initializes the reader and prepares it for reading input events. It
// sets up the cancel reader and the key sequence parser. It also sets up the
// lookup table for key sequences if it is not already set. This function
// should be called before reading input events.
func (r *TerminalReader) Start() (err error) {
	if r.rd == nil {
		r.rd, err = newCancelreader(r.r)
		if err != nil {
			return err
		}
	}
	if r.table == nil {
		r.table = buildKeysTable(r.Legacy, r.term, r.UseTerminfo)
	}
	r.closed = false
	return nil
}

// Read implements [io.Reader].
func (d *TerminalReader) Read(p []byte) (int, error) {
	if err := d.Start(); err != nil {
		return 0, err
	}
	return d.rd.Read(p) //nolint:wrapcheck
}

// Cancel cancels the underlying reader.
func (d *TerminalReader) Cancel() bool {
	if d.rd == nil {
		return false
	}
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *TerminalReader) Close() (rErr error) {
	if d.rd == nil {
		return fmt.Errorf("reader was not initialized")
	}
	if d.closed {
		return nil
	}
	defer func() {
		if rErr != nil {
			d.closed = true
		}
	}()
	return d.rd.Close() //nolint:wrapcheck
}

func (d *TerminalReader) logf(format string, v ...interface{}) {
	if d.logger == nil {
		return
	}
	d.logger.Printf(format, v...)
}

func (d *TerminalReader) readEvents() ([]Event, error) {
	if err := d.Start(); err != nil {
		return nil, err
	}

	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	var events []Event
	buf := d.buf[:nb]

	// Lookup table first
	if bytes.HasPrefix(buf, []byte{'\x1b'}) {
		if k, ok := d.table[string(buf)]; ok {
			d.logf("input: %q", buf)
			events = append(events, KeyPressEvent(k))
			return events, nil
		}
	}

	var i int
	for i < len(buf) {
		nb, ev := d.parseSequence(buf[i:])
		d.logf("input: %q", buf[i:i+nb])

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndEvent); !ok {
				d.paste = append(d.paste, buf[i])
				i++
				continue
			}
		}

		switch ev.(type) {
		case UnknownEvent:
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(buf[i:i+nb])]; ok {
				ev = KeyPressEvent(k)
			}
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
			d.paste = nil // reset the buffer
			events = append(events, PasteEvent(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(MultiEvent); ok {
			events = append(events, []Event(mevs)...)
		} else {
			events = append(events, ev)
		}
		i += nb
	}

	return events, nil
}
