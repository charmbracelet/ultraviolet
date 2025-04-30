package tv

import (
	"context"
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
	err         error
	size        Size // The last known size of the terminal
}

// NewTerminal creates a new Terminal instance with the given input and output
// streams.
func NewTerminal(in io.Reader, out io.Writer) *Terminal {
	t := new(Terminal)
	t.in = in
	t.out = out
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

// Events returns an event channel that will receive events from the terminal.
func (t *Terminal) Events(ctx context.Context) <-chan Event {
	evch := make(chan Event)
	errch := make(chan error, 1)

	go func() {
		// Create default receivers.
		winchrcv, err := NewWinchReceiver(t.out)
		if err != nil {
			errch <- fmt.Errorf("failed to create winch receiver: %w", err)
			return
		}

		// Receive events from the terminal.
		man := NewInputManager(winchrcv)
		man.ReceiveEvents(ctx, evch, errch)
	}()

	go func() {
		// Wait for the context to be done or an error to occur.
		select {
		case <-ctx.Done():
			return
		case err := <-errch:
			if err != nil {
				t.err = err
			}
			return
		}
	}()

	go func() {
		// Listen for window size changes and store them in the terminal.
		for ev := range evch {
			switch ev := ev.(type) {
			case Size:
				if ev.Width != t.size.Width || ev.Height != t.size.Height {
					t.size = ev
				}
			}
			evch <- ev // Send the event back to the channel
		}
	}()

	return evch
}

// Err returns the last error that occurred while receiving events from the
// terminal.
func (t *Terminal) Err() error {
	if t.err != nil {
		return fmt.Errorf("terminal error: %w", t.err)
	}
	return nil
}
