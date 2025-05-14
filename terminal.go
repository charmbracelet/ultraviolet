package tv

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

var (
	// ErrNotTerminal is returned when one of the I/O streams is not a terminal.
	ErrNotTerminal = fmt.Errorf("not a terminal")

	// ErrPlatformNotSupported is returned when the platform is not supported.
	ErrPlatformNotSupported = fmt.Errorf("platform not supported")
)

// Terminal represents a terminal screen that can be manipulated and drawn to.
// It handles reading events from the terminal using [WinChReceiver],
// [SequenceReceiver], and [ConReceiver].
type Terminal struct {
	// Terminal I/O streams and state.
	in          io.Reader
	out         io.Writer
	inTty       term.File
	inTtyState  *term.State
	outTty      term.File
	outTtyState *term.State
	winchTty    term.File // The terminal to receive window size changes from.

	// Terminal type, screen and buffer.
	termtype     string          // The $TERM type.
	environ      Environ         // The environment variables.
	buf          *Buffer         // Reference to the last buffer used.
	vp           Viewport        // Reference to the last viewport used.
	scr          *terminalWriter // The actual screen to be drawn to.
	size         Size            // The last known size of the terminal.
	profile      colorprofile.Profile
	useTabs      bool // Whether to use hard tabs or not.
	useBspace    bool // Whether to use backspace or not.
	altScreen    bool // Whether to use the alternate screen or not.
	showCursor   bool // Whether we want to show the cursor or not.
	cursorHidden bool // The current cursor visibility state.

	// Terminal input stream.
	rd        *TerminalReader
	err       error
	evch      chan Event
	once      sync.Once
	mouseMode MouseMode // The mouse mode for the terminal.

	logger Logger // The debug logger for I/O.
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
	if f, ok := in.(term.File); ok {
		t.inTty = f
	}
	if f, ok := out.(term.File); ok {
		t.outTty = f
	}
	t.environ = env
	t.termtype = t.environ.Getenv("TERM")
	// We need to call [Terminal.optimizeMovements] before creating the screen
	// to populate [Terminal.useBspace] and [Terminal.useTabs].
	t.optimizeMovements()
	t.scr = t.newScreen()
	t.SetColorProfile(colorprofile.Detect(out, env))
	t.rd = NewTerminalReader(t.in, t.termtype)
	t.rd.MouseMode = &t.mouseMode
	t.evch = make(chan Event)
	t.once = sync.Once{}

	// Handle debugging I/O.
	debug, ok := os.LookupEnv("TV_DEBUG")
	if ok && len(debug) > 0 {
		f, err := os.OpenFile(debug, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			panic("failed to open debug file: " + err.Error())
		}

		logger := log.New(f, "tv: ", log.LstdFlags|log.Lshortfile)
		t.SetLogger(logger)
	}

	return t
}

// SetLogger sets the debug logger for the terminal. This is used to log debug
// information about the terminal I/O. By default, it is set to a no-op logger.
func (t *Terminal) SetLogger(logger Logger) {
	t.logger = logger
	t.rd.SetLogger(logger)
	t.scr.SetLogger(logger)
}

// ColorProfile returns the currently used color profile for the terminal.
func (t *Terminal) ColorProfile() colorprofile.Profile {
	return t.profile
}

// SetColorProfile sets a custom color profile for the terminal. This is useful
// for forcing a specific color output. By default, the terminal will use the
// system's color profile inferred by the environment variables.
func (t *Terminal) SetColorProfile(p colorprofile.Profile) {
	t.profile = p
	t.scr.SetColorProfile(p)
}

var _ Screen = (*Terminal)(nil)

// CellAt returns the cell at the given x, y position in the terminal buffer.
func (t *Terminal) CellAt(x int, y int) *Cell {
	if t.buf == nil {
		return nil
	}
	return t.buf.CellAt(x, y)
}

// ColorModel returns the color model of the terminal screen.
func (t *Terminal) ColorModel() color.Model {
	return t.ColorProfile()
}

// GetSize returns the size of the terminal screen. It errors if the size
// cannot be determined.
func (t *Terminal) GetSize() (width, height int, err error) {
	w, h, err := t.getSize()
	if err != nil {
		return 0, 0, fmt.Errorf("error getting terminal size: %w", err)
	}
	// Cache the last known size.
	t.size.Width = w
	t.size.Height = h
	return w, h, nil
}

var _ Displayer = (*Terminal)(nil)

func (t *Terminal) newScreen() *terminalWriter {
	s := newTerminalWriter(t.out, t.termtype, t.size.Width)
	s.SetColorProfile(t.profile)
	s.SetHardTabs(t.useTabs)
	s.SetBackspace(t.useBspace)
	s.SetAltScreen(t.altScreen)
	s.SetLogger(t.logger)
	return s
}

// Display displays the given frame on the terminal screen. It returns an
// error if the display fails.
func (t *Terminal) Display(f *Frame) error {
	firstRender := t.buf == nil

	// Cache the last buffer used.
	t.buf = f.Buffer

	// Are we using the alternate screen?
	if f.Viewport != t.vp {
		switch f.Viewport.(type) {
		case FullViewport:
			if !t.altScreen {
				t.altScreen = true
				t.scr.WriteString(ansi.SetAltScreenSaveCursorMode) //nolint:errcheck
			}
		case InlineViewport:
			if t.altScreen {
				t.altScreen = false
				t.scr.WriteString(ansi.ResetAltScreenSaveCursorMode) //nolint:errcheck
			}
		}
		// These must happen before resizing the screen.
		t.scr.SetAltScreen(t.altScreen)
		t.scr.SetRelativeCursor(!t.altScreen)
		switch f.Viewport.(type) {
		case FullViewport:
			// Force clear the screen when we enter altscreen mode.
			t.scr.Clear()
		}
		if !t.altScreen {
			// Inline mode resizes the screen based on the frame height
			// and terminal width. This is because the frame height can
			// change based on the content of the frame. For example,
			// if the frame contains a list of items, the height of the
			// frame will be the number of items in the list. This is
			// different from the alt screen buffer, which has a fixed
			// height and width.
			area := f.ComputeArea()
			t.scr.Resize(t.buf, area.Dx(), area.Dy())
		}

		t.vp = f.Viewport
	}

	// Hide the cursor during rendering to avoid flickering. We need to do this
	// before toggling the alternate screen to ensure we target the altscreen
	// cursor.
	shouldHideCursor := (firstRender && !t.showCursor) || f.Position == nil
	hasChanges := t.scr.Touched() > 0
	if (shouldHideCursor || hasChanges) && !t.cursorHidden {
		t.scr.WriteString(ansi.HideCursor) //nolint:errcheck
		t.cursorHidden = true
	}

	// XXX: We want to render the changes before moving the cursor to ensure
	// the cursor is at the position specified in the frame.
	t.scr.Render(f.Buffer)

	if f.Position != nil && f.Position.X >= 0 && f.Position.Y >= 0 {
		t.scr.MoveTo(f.Buffer, f.Position.X, f.Position.Y)
	}

	shouldShowCursor := t.showCursor || f.Position != nil
	if (shouldShowCursor || (shouldHideCursor && hasChanges)) && t.cursorHidden {
		t.scr.WriteString(ansi.ShowCursor) //nolint:errcheck
		t.cursorHidden = false
	}

	return nil
}

// Flush flushes any pending renders to the terminal screen. This is typically
// used to flush the underlying screen buffer to the terminal.
func (t *Terminal) Flush() error {
	return t.scr.Flush()
}

// EnableMode enables the given modes on the terminal. This is typically used
// to enable mouse support, bracketed paste mode, and other terminal features.
func (t *Terminal) EnableMode(modes ...ansi.Mode) error {
	if len(modes) == 0 {
		return nil
	}
	_, err := t.WriteString(ansi.SetMode(modes...))
	return err
}

// DisableMode disables the given modes on the terminal. This is typically
// used to disable mouse support, bracketed paste mode, and other terminal
// features.
func (t *Terminal) DisableMode(modes ...ansi.Mode) error {
	if len(modes) == 0 {
		return nil
	}
	_, err := t.WriteString(ansi.ResetMode(modes...))
	return err
}

// MouseMode represents the mouse mode for the terminal. It is used to enable or
// disable mouse support on the terminal.
//
// It is a bitmask of the following values:
//   - [ReleasesMouseMode]: Enables mouse release events.
//   - [AllMotionMouseMode]: Enables all mouse motion events.
type MouseMode byte

const (
	// ReleasesMouseMode enables mouse release events.
	ReleasesMouseMode MouseMode = 1 << iota
	// AllMotionMouseMode enables all mouse motion events.
	AllMotionMouseMode
)

// EnableMouse enables mouse support on the terminal. This will enable basic
// mouse button and button motion events. To enable release events and all
// motion events, use [Terminal.EnableMouse] with the appropriate flags. See
// [MouseMode] for more information.
func (t *Terminal) EnableMouse(modes ...MouseMode) (err error) {
	var mode MouseMode
	for _, m := range modes {
		mode |= m
	}
	t.mouseMode = mode
	if runtime.GOOS != "windows" {
		modes := []ansi.Mode{}
		modes = append(modes, ansi.ButtonEventMouseMode)
		if t.mouseMode&AllMotionMouseMode != 0 {
			modes = append(modes, ansi.AnyEventMouseMode)
		}
		modes = append(modes, ansi.SgrExtMouseMode)
		if err := t.EnableMode(modes...); err != nil {
			return err
		}
	}
	return t.enableWindowsMouse()
}

// DisableMouse disables mouse support on the terminal. This will disable mouse
// button and button motion events.
func (t *Terminal) DisableMouse() (err error) {
	t.mouseMode = 0
	if runtime.GOOS != "windows" {
		return t.DisableMode(
			ansi.ButtonEventMouseMode,
			ansi.AnyEventMouseMode,
			ansi.SgrExtMouseMode,
		)
	}
	return t.disableWindowsMouse()
}

// EnableBracketedPaste enables bracketed paste mode on the terminal. This is
// typically used to enable support for pasting text into the terminal without
// interfering with the terminal's input handling.
func (t *Terminal) EnableBracketedPaste() error {
	return t.EnableMode(ansi.BracketedPasteMode)
}

// DisableBracketedPaste disables bracketed paste mode on the terminal. This is
// typically used to disable support for pasting text into the terminal.
func (t *Terminal) DisableBracketedPaste() error {
	return t.DisableMode(ansi.BracketedPasteMode)
}

// EnableFocusEvents enables focus/blur receiving notification events on the
// terminal.
func (t *Terminal) EnableFocusEvents() error {
	return t.EnableMode(ansi.FocusEventMode)
}

// DisableFocusEvents disables focus/blur receiving notification events on the
// terminal.
func (t *Terminal) DisableFocusEvents() error {
	return t.DisableMode(ansi.FocusEventMode)
}

// EnterAltScreen enters the alternate screen buffer. This is typically used
// for applications that want to take over the entire terminal screen.
//
// Note that [Terminal] manages the alternate screen buffer for you based on
// the [Viewport] used during [Terminal.Display]. This means that you don't
// need to call this unless you know what you're doing.
func (t *Terminal) EnterAltScreen() error {
	t.altScreen = true
	return t.EnableMode(ansi.AltScreenSaveCursorMode)
}

// ExitAltScreen exits the alternate screen buffer and returns to the normal
// screen buffer.
//
// Note that [Terminal] manages the alternate screen buffer for you based on
// the [Viewport] used during [Terminal.Display]. This means that you don't
// need to call this unless you know what you're doing.
func (t *Terminal) ExitAltScreen() error {
	t.altScreen = false
	return t.DisableMode(ansi.AltScreenSaveCursorMode)
}

// ShowCursor shows the terminal cursor.
//
// Note that [Terminal] manages the visibility of the cursor for you based on
// the [Viewport] used during [Terminal.Display]. This means that you don't
// need to call this unless you know what you're doing.
func (t *Terminal) ShowCursor() error {
	t.showCursor = true
	return t.EnableMode(ansi.TextCursorEnableMode)
}

// HideCursor hides the terminal cursor.
//
// Note that [Terminal] manages the visibility of the cursor for you based on
// the [Viewport] used during [Terminal.Display]. This means that you don't
// need to call this unless you know what you're doing.
func (t *Terminal) HideCursor() error {
	t.showCursor = false
	return t.DisableMode(ansi.TextCursorEnableMode)
}

// SetTitle sets the title of the terminal window. This is typically used to
// set the title of the terminal window to the name of the application.
func (t *Terminal) SetTitle(title string) error {
	_, err := t.WriteString(ansi.SetWindowTitle(title))
	return err
}

// Resize resizes the terminal to the given width and height. It returns an
// error if the resize fails.
func (t *Terminal) Resize(width, height int) error {
	if width == t.size.Width && height == t.size.Height {
		// No change in size.
		return nil
	}
	t.size.Width = width
	t.size.Height = height
	if t.buf != nil && t.altScreen {
		// We only resize the screen if we're in the alt screen buffer. Inline
		// mode resizes the screen based on the frame height and terminal
		// width. See [terminalWriter.render] for more details.
		t.scr.Resize(t.buf, width, height)
	}
	return nil
}

// MakeRaw puts the terminal in raw mode, which disables line buffering and
// echoing. The terminal will automatically be restored to its original state
// on [Terminal.Close] or [Terminal.Shutdown], or by manually calling
// [Terminal.Restore].
func (t *Terminal) MakeRaw() error {
	if err := t.makeRaw(); err != nil {
		return fmt.Errorf("error entering raw mode: %w", err)
	}
	return nil
}

// Start prepares the terminal for use. It starts the input reader and
// initializes the terminal state. This should be called before using the
// terminal.
func (t *Terminal) Start() error {
	if t.inTty == nil && t.outTty == nil {
		return ErrNotTerminal
	}

	t.winchTty = t.inTty
	if t.winchTty == nil {
		t.winchTty = t.outTty
	}

	if err := t.rd.Start(); err != nil {
		return fmt.Errorf("error starting terminal: %w", err)
	}

	return nil
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

// Shutdown restores the terminal to its original state and stops the event
// channel in a graceful manner.
// This waits for any pending events to be processed or the context to be
// done before closing the event channel.
func (t *Terminal) Shutdown(ctx context.Context) (rErr error) {
	defer func() {
		err := t.Close()
		if rErr == nil {
			rErr = err
		}
	}()

	// Cancel the input reader.
	t.rd.Cancel()

	// Consume any pending events or listen for the context to be done.
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.evch:
			if len(t.evch) == 0 {
				return nil
			}
		}
	}
}

// Close close any resources used by the terminal and restore the terminal to
// its original state.
func (t *Terminal) Close() (rErr error) {
	defer t.closeChannels()

	defer func() {
		if t.buf != nil && !t.altScreen {
			// Go to the bottom of the screen.
			t.scr.MoveTo(t.buf, 0, t.buf.Height()-1)
		}
		// Restore cursor visibility and normal screen buffer. Make sure we do
		// so after turning off the alternate screen to ensure we target the
		// cursor is on the normal screen buffer.
		if t.altScreen {
			t.scr.WriteString(ansi.ResetAltScreenSaveCursorMode) //nolint:errcheck
			t.altScreen = false
		}
		if t.cursorHidden {
			t.scr.WriteString(ansi.ShowCursor) //nolint:errcheck
			t.cursorHidden = false
		}

		t.scr.SetAltScreen(t.altScreen)
		t.scr.SetRelativeCursor(!t.altScreen)

		// Flush changes to the screen.
		t.scr.Flush() //nolint:errcheck

		// Reset screen.
		t.scr = t.newScreen()
	}()

	defer func() {
		err := t.Restore()
		if rErr == nil && err != nil {
			rErr = fmt.Errorf("error restoring terminal state: %w", err)
		}
	}()

	defer func() {
		err := t.rd.Close()
		if rErr == nil && err != nil {
			rErr = fmt.Errorf("error closing terminal reader: %w", err)
		}
	}()

	return
}

// closeChannels closes the event and error channels.
func (t *Terminal) closeChannels() {
	t.once.Do(func() {
		close(t.evch)
	})
}

// SendEvent is a helper function to send an event to the event channel. It
// blocks until the event is sent or the context is done. If the context is
// done, it will not send the event and will return immediately.
// This is useful to control the terminal from outside the event loop.
func (t *Terminal) SendEvent(ctx context.Context, ev Event) {
	select {
	case <-ctx.Done():
	case t.evch <- ev:
	}
}

// Events returns an event channel that will receive events from the terminal.
// Use [Terminal.Err] to check for errors that occurred while receiving events.
// The event channel is closed when the terminal is closed or when the context
// is done.
func (t *Terminal) Events(ctx context.Context) <-chan Event {
	go func() {
		defer t.closeChannels()

		// Receive events from the terminal.
		receivers := []InputReceiver{
			&TerminalInputReceiver{t.rd},
		}
		if runtime.GOOS != "windows" {
			// SIGWINCH receiver for window size changes.
			receivers = append(receivers, &WinChReceiver{t.winchTty})
		}

		t.err = NewInputManager(receivers...).ReceiveEvents(ctx, t.evch)
	}()

	return t.evch
}

// Err returns the error that occurred while receiving events from the
// terminal.
// This is typically used to check for errors that occurred while
// receiving events.
func (t *Terminal) Err() error {
	return t.err
}

// PrependStyledString is a helper function to prepend a styled string to the
// terminal screen. It is a convenience function that creates a new
// [StyledString] and calls [PrependLines] with the buffer lines of the
// [StyledString].
func (t *Terminal) PrependStyledString(method ansi.Method, str string) error {
	ss := NewStyledString(method, str)
	return t.PrependLines(ss.Buffer.Lines...)
}

// PrependLines adds lines of cells to the top of the terminal screen. The
// added line is unmanaged and will not be cleared or updated by the
// [Terminal].
//
// Using this when the terminal is using the alternate screen or when occupying
// the whole screen may not produce any visible effects. This is because once
// the terminal writes the prepended lines, they will get overwritten by the
// next frame.
func (t *Terminal) PrependLines(lines ...Line) error {
	if t.buf == nil || t.scr == nil || len(lines) == 0 {
		return nil
	}

	// We need to scroll the screen up by the number of lines in the queue.
	// We can't use [ansi.SU] because we want the cursor to move down until
	// it reaches the bottom of the screen.
	t.scr.mu.Lock()
	defer t.scr.mu.Unlock()

	t.scr.move(t.buf, 0, t.buf.Height()-1)
	t.scr.buf.WriteString(strings.Repeat("\n", len(lines))) //nolint:errcheck
	t.scr.cur.Y += len(lines)

	// XXX: Now go to the top of the screen, insert new lines, and write the
	// queued strings. It is important to use [terminalWriter.moveCursor]
	// instead of [terminalWriter.move] because we don't want to perform any
	// checks on the cursor position.
	t.scr.moveCursor(t.buf, 0, 0, false)
	t.scr.buf.WriteString(ansi.InsertLine(len(lines))) //nolint:errcheck
	for _, line := range lines {
		t.scr.buf.WriteString(line.Render() + "\r\n") //nolint:errcheck
	}

	return t.scr.flush()
}

// Write writes the given data to the output stream. It implements the
// [io.Writer] interface.
func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.out.Write(p)
}

// WriteString writes the given string to the output stream. It implements the
// [io.StringWriter] interface.
func (t *Terminal) WriteString(s string) (n int, err error) {
	return io.WriteString(t.out, s)
}
