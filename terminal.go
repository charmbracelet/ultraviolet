package tv

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

// Terminal represents a terminal screen that can be manipulated and drawn to.
// It handles reading events from the terminal using [WinchReceiver],
// [SequenceReceiver], and [ConReceiver].
type Terminal struct {
	// Terminal I/O streams and state.
	in          io.Reader
	out         io.Writer
	inTty       term.File
	inTtyState  *term.State
	outTty      term.File
	outTtyState *term.State

	// Terminal type, screen and buffer.
	termtype string     // The $TERM type.
	environ  Environ    // The environment variables.
	scr      *tScreen   // The actual screen to be drawn to.
	size     WindowSize // The last known size of the terminal.
	profile  colorprofile.Profile

	err   error
	evch  chan Event
	errch chan error
	once  sync.Once
}

// DefaultTerminal returns a new default terminal instance that uses
// [os.Stdin], [os.Stdout], and [os.Environ].
func DefaultTerminal() *Terminal {
	return NewTerminal(os.Stdin, os.Stdout, os.Environ())
}

// NewTerminal creates a new Terminal instance with the given terminal size.
// Use [term.GetSize] to get the size of the output screen.
func NewTerminal(in io.Reader, out io.Writer, env []string) *Terminal {
	t := new(Terminal)
	t.in = in
	t.out = out
	t.evch = make(chan Event)
	t.errch = make(chan error, 1)
	t.environ = env
	t.termtype = t.environ.Getenv("TERM")
	t.profile = colorprofile.Detect(out, env)
	return t
}

// SetColorProfile sets a custom color profile for the terminal. This is useful
// for forcing a specific color output. By default, the terminal will use the
// system's color profile inferred by the environment variables.
func (t *Terminal) SetColorProfile(p colorprofile.Profile) {
	t.profile = p
}

var _ Screen = (*Terminal)(nil)

// CellAt returns the cell at the given x, y position in the terminal buffer.
func (t *Terminal) CellAt(x int, y int) *Cell {
	return t.scr.CellAt(x, y)
}

// ColorModel returns the color model of the terminal screen.
func (t *Terminal) ColorModel() color.Model {
	return t.profile
}

// GetSize returns the size of the terminal screen. It errors if the size
// cannot be determined.
func (t *Terminal) GetSize() (width, height int, err error) {
	f, ok := t.out.(term.File)
	if !ok {
		return 0, 0, fmt.Errorf("output is not a terminal")
	}
	w, h, err := term.GetSize(f.Fd())
	if err != nil {
		return 0, 0, fmt.Errorf("error getting terminal size: %w", err)
	}
	// Cache the last known size.
	t.size.Width = w
	t.size.Height = h
	return w, h, nil
}

var _ Displayer = (*Terminal)(nil)

func (t *Terminal) newScreen() *tScreen {
	s := newTScreen(t.out, t.size.Width, t.size.Height)
	s.SetTermType(t.termtype)
	s.SetColorProfile(t.profile)
	return s
}

// Display displays the given frame on the terminal screen. It returns an
// error if the display fails.
func (t *Terminal) Display(f *Frame) error {
	if t.scr == nil {
		// Initialize the screen for the first time.
		t.scr = t.newScreen()
		t.optimizeMovements()
	}

	t.scr.SetBuffer(f.Buffer)
	width, height := f.Area.Dx(), f.Area.Dy()
	if width != t.scr.Width() || height != t.scr.Height() {
		t.scr.Resize(f.Area.Dx(), f.Area.Dy())
	}

	switch f.Viewport.(type) {
	case FullViewport:
		t.scr.EnterAltScreen()
		t.scr.SetRelativeCursor(false)
	case InlineViewport:
		t.scr.ExitAltScreen()
		t.scr.SetRelativeCursor(true)
	case FixedViewport:
		t.scr.ExitAltScreen()
		t.scr.SetRelativeCursor(false)
	}

	// BUG: Hide/Show cursor doesn't take effect unless we call them before
	// Render.
	if f.Position == nil {
		t.scr.HideCursor()
	} else {
		t.scr.ShowCursor()
	}

	// XXX: We want to render the changes before moving the cursor to ensure
	// the cursor is at the position specified in the frame.
	t.scr.Render()

	if f.Position != nil && f.Position.X >= 0 && f.Position.Y >= 0 {
		t.scr.MoveTo(f.Position.X, f.Position.Y)
	}

	return t.scr.Flush()
}

// EnterAltScreen enters the alternate screen buffer. This is typically used
// for applications that want to take over the entire terminal screen.
func (t *Terminal) EnterAltScreen() error {
	t.scr.EnterAltScreen()
	return t.scr.Flush()
}

// ExitAltScreen exits the alternate screen buffer and returns to the normal
// screen buffer.
func (t *Terminal) ExitAltScreen() error {
	t.scr.ExitAltScreen()
	return t.scr.Flush()
}

// SHowCursor shows the terminal cursor.
func (t *Terminal) ShowCursor() error {
	t.scr.ShowCursor()
	return t.scr.Flush()
}

// HideCursor hides the terminal cursor.
func (t *Terminal) HideCursor() error {
	t.scr.HideCursor()
	return t.scr.Flush()
}

// SetTitle sets the title of the terminal window. This is typically used to
// set the title of the terminal window to the name of the application.
func (t *Terminal) SetTitle(title string) error {
	_, err := io.WriteString(t.out, ansi.SetWindowTitle(title))
	return err
}

// MakeRaw puts the terminal in raw mode, which disables line buffering and
// echoing. The terminal will automatically be restored to its original state
// on [Terminal.Close] or by calling [Terminal.Restore].
func (t *Terminal) MakeRaw() error {
	return t.makeRaw()
}

// Restore restores the terminal to its original state. This should be called
// after [MakeRaw] to restore the terminal to its original state. Otherwise, it
// is a no-op.
func (t *Terminal) Restore() error {
	if t.inTtyState != nil {
		if err := term.Restore(t.inTty.Fd(), t.inTtyState); err != nil {
			return err
		}
		t.inTtyState = nil
	}
	if t.outTtyState != nil {
		if err := term.Restore(t.outTty.Fd(), t.outTtyState); err != nil {
			return err
		}
		t.outTtyState = nil
	}
	return nil
}

// Close restores the terminal to its original state.
func (t *Terminal) Close() error {
	if err := t.Restore(); err != nil {
		return fmt.Errorf("error restoring terminal state: %w", err)
	}

	// Reset the terminal state.
	close(t.evch)
	close(t.errch)
	t.once = sync.Once{}

	if t.scr != nil {
		if err := t.scr.Close(); err != nil {
			return fmt.Errorf("error closing terminal screen: %w", err)
		}
	}

	return nil
}

// Events returns an event channel that will receive events from the terminal.
func (t *Terminal) Events(ctx context.Context) <-chan Event {
	t.once.Do(func() {
		go func() {
			// Create default receivers.
			winchrcv, err := NewWinchReceiver(t.out)
			if err != nil {
				t.errch <- fmt.Errorf("failed to create winch receiver: %w", err)
				return
			}

			// Receive events from the terminal.
			man := NewInputManager(winchrcv)
			man.ReceiveEvents(ctx, t.evch, t.errch)
		}()

		go func() {
			// Wait for the context to be done or an error to occur.
			select {
			case <-ctx.Done():
				return
			case err := <-t.errch:
				if err != nil {
					t.err = err
				}
				return
			}
		}()
	})

	return t.evch
}

// Err returns the last error that occurred while receiving events from the
// terminal.
func (t *Terminal) Err() error {
	if t.err != nil {
		return fmt.Errorf("terminal error: %w", t.err)
	}
	return nil
}
