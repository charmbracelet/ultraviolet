package tv

import (
	"fmt"
	"io"

	"github.com/charmbracelet/x/term"
)

// Terminal represents a terminal screen that can be manipulated and drawn to.
// It handles reading events from the terminal using [WinchReceiver],
// [SequenceReceiver], and [ConReceiver].
type Terminal struct {
	in          io.Reader
	out         io.Writer
	inTty       term.File
	inTtyState  *term.State
	outTty      term.File
	outTtyState *term.State
	evch        chan Event
	errch       chan error
}

// NewTerminal creates a new Terminal instance with the given input and output
// streams.
func NewTerminal(in io.Reader, out io.Writer) *Terminal {
	t := new(Terminal)
	t.in = in
	t.out = out
	t.evch = make(chan Event)
	t.errch = make(chan error)
	return t
}

// MakeRaw puts the terminal in raw mode, which disables line buffering and
// echoing. The terminal will be restored to its original state on
// [Terminal.Close].
func (t *Terminal) MakeRaw() error {
	return t.makeRaw()
}

// Close restores the terminal to its original state.
func (t *Terminal) Close() error {
	if t.inTtyState != nil {
		if err := term.Restore(t.inTty.Fd(), t.inTtyState); err != nil {
			return fmt.Errorf("error restoring terminal state: %w", err)
		}
	}

	if t.outTtyState != nil {
		if err := term.Restore(t.outTty.Fd(), t.outTtyState); err != nil {
			return fmt.Errorf("error restoring terminal state: %w", err)
		}
	}

	return nil
}
