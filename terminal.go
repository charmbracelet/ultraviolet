package tv

import (
	"bytes"
	"context"
	"errors"
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
	"github.com/muesli/cancelreader"
)

var (
	// ErrNotTerminal is returned when one of the I/O streams is not a terminal.
	ErrNotTerminal = fmt.Errorf("not a terminal")

	// ErrPlatformNotSupported is returned when the platform is not supported.
	ErrPlatformNotSupported = fmt.Errorf("platform not supported")

	// ErrStarted is returned when the terminal has already been started.
	ErrStarted = fmt.Errorf("terminal already started")
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
	started     bool      // Indicates if the terminal has been started.

	// Terminal type, screen and buffer.
	mu           sync.Mutex        // Mutex to protect the screen and buffer.
	termtype     string            // The $TERM type.
	environ      Environ           // The environment variables.
	buf          *Buffer           // Reference to the last buffer used.
	scr          *TerminalRenderer // The actual screen to be drawn to.
	size         Size              // The last known size of the terminal.
	profile      colorprofile.Profile
	method       ansi.Method // The width method used to calculate the width of each unicode character
	modes        ansi.Modes  // Keep track of terminal modes.
	useTabs      bool        // Whether to use hard tabs or not.
	useBspace    bool        // Whether to use backspace or not.
	cursorHidden bool        // The current cursor visibility state.

	// Terminal input stream.
	rd           *TerminalReader
	rdr          *TerminalInputReceiver
	wrdr         *WinChReceiver
	im           *InputManager
	err          error
	evch         chan Event
	evOnce       sync.Once
	once         sync.Once
	mouseMode    MouseMode // The mouse mode for the terminal.
	readLoopDone chan struct{}

	logger Logger // The debug logger for I/O.
}

// DefaultTerminal returns a new default terminal instance that uses
// [os.Stdin], [os.Stdout], and [os.Environ].
func DefaultTerminal() *Terminal {
	return NewTerminal(os.Stdin, os.Stdout, os.Environ())
}

var defaultModes = ansi.Modes{
	// These are modes we care about. We only register the ones we care
	// about and ignore the rest.
	ansi.TextCursorEnableMode:    ansi.ModeSet,
	ansi.AutoWrapMode:            ansi.ModeSet,
	ansi.AltScreenSaveCursorMode: ansi.ModeReset,
	ansi.ButtonEventMouseMode:    ansi.ModeReset,
	ansi.AnyEventMouseMode:       ansi.ModeReset,
	ansi.SgrExtMouseMode:         ansi.ModeReset,
	ansi.BracketedPasteMode:      ansi.ModeReset,
	ansi.FocusEventMode:          ansi.ModeReset,
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
	t.modes = ansi.Modes{}
	t.environ = env
	t.termtype = t.environ.Getenv("TERM")
	// We need to call [Terminal.optimizeMovements] before creating the screen
	// to populate [Terminal.useBspace] and [Terminal.useTabs].
	t.optimizeMovements()
	t.scr = t.newScreen()
	t.SetColorProfile(colorprofile.Detect(out, env))
	t.rd = NewTerminalReader(t.in, t.termtype)
	t.rd.MouseMode = &t.mouseMode
	t.rdr = NewTerminalInputReceiver(t.rd)
	t.readLoopDone = make(chan struct{})
	t.evch = make(chan Event, 1)
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

// SetMethod sets the width method used to calculate the width of each
// unicode character in the terminal.
//
// By default, this is set to [ansi.WcWidth] which provides the most
// backwards-compatible behavior for calculating the width of unicode
// characters in the terminal but not necessarily the most accurate and can
// lead to unexpected results in some cases.
//
// To use the actual Unicode mono-width calculation, use [ansi.GraphemeWidth]
// instead. You can use [Terminal.EnableMode] with
// [ansi.GraphemeClusteringMode] to try to enable grapheme clustering support
// in the terminal. Then use [Terminal.RequestMode] and listen for the
// [ModeReportEvent] to check if grapheme clustering is actually supported by
// the terminal.
func (t *Terminal) SetMethod(method ansi.Method) {
	t.method = method
}

// Method returns the currently used width method used to calculate the width
// of each unicode character in the terminal.
func (t *Terminal) Method() ansi.Method {
	return t.method
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

func (t *Terminal) newScreen() *TerminalRenderer {
	s := NewTerminalRenderer(t.out, t.environ)
	s.SetColorProfile(t.profile)
	s.SetTabStops(t.size.Width)
	s.SetBackspace(t.useBspace)
	s.SetRelativeCursor(true) // Initial state is relative cursor movements.
	if t.scr != nil {
		if t.scr.AltScreen() {
			s.EnterAltScreen()
		} else {
			s.ExitAltScreen()
		}
	}
	s.SetLogger(t.logger)
	return s
}

// ClearScreen clears the terminal screen and moves the cursor to the top left
// corner on the next render.
func (t *Terminal) ClearScreen() {
	t.scr.Clear()
	if t.buf != nil {
		t.scr.Render(t.buf)
		t.buf.Clear()
	}
}

// Display displays the given frame on the terminal screen. It returns an
// error if the display fails.
func (t *Terminal) Display(f *Frame) error {
	if f == nil {
		return fmt.Errorf("cannot display nil frame")
	}
	if f.Buffer == nil {
		return fmt.Errorf("cannot display frame with nil buffer")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Cache the last buffer used.
	buf := f.Buffer
	t.buf = buf

	// Are we using the alternate screen?
	switch vp := f.Viewport.(type) {
	case nil, FullViewport:
		t.scr.SetRelativeCursor(false)
	case FixedViewport:
		// Make sure we clear areas outside the viewport.
		area := vp.ComputeArea(t.size.Width, t.size.Height)
		for _, a := range []Rectangle{
			{Min: Pos(0, 0), Max: Pos(t.size.Width, area.Min.Y)},             // top
			{Min: Pos(area.Max.X, 0), Max: Pos(t.size.Width, t.size.Height)}, // right
			{Min: Pos(0, area.Max.Y), Max: Pos(t.size.Width, t.size.Height)}, // bottom
			{Min: Pos(0, 0), Max: Pos(area.Min.X, t.size.Height)},            // left
		} {
			buf.ClearArea(a)
		}
	case InlineViewport:
		t.scr.SetRelativeCursor(true)
	}

	if t.cursorHidden != t.scr.CursorHidden() {
		if t.cursorHidden {
			t.showCursor() //nolint:errcheck
		} else {
			t.hideCursor() //nolint:errcheck
		}
		t.cursorHidden = t.scr.CursorHidden()
	}

	// XXX: We want to render the changes before moving the cursor to ensure
	// the cursor is at the position specified in the frame.
	t.scr.Render(buf)

	if f.Position != nil && f.Position.X >= 0 && f.Position.Y >= 0 {
		t.showCursor() //nolint:errcheck
		t.scr.move(buf, f.Position.X, f.Position.Y)
	}

	return t.scr.Flush()
}

// Flush flushes any pending renders to the terminal screen. This is typically
// used to flush the underlying screen buffer to the terminal.
func (t *Terminal) Flush() error {
	return t.scr.Flush()
}

// EnableMode enables the given modes on the terminal. This is typically used
// to enable mouse support, bracketed paste mode, and other terminal features.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) EnableMode(modes ...ansi.Mode) error {
	if len(modes) == 0 {
		return nil
	}
	for _, m := range modes {
		t.modes[m] = ansi.ModeSet
	}
	_, err := t.WriteString(ansi.SetMode(modes...))
	return err
}

// DisableMode disables the given modes on the terminal. This is typically
// used to disable mouse support, bracketed paste mode, and other terminal
// features.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) DisableMode(modes ...ansi.Mode) error {
	if len(modes) == 0 {
		return nil
	}
	for _, m := range modes {
		t.modes[m] = ansi.ModeReset
	}
	_, err := t.WriteString(ansi.ResetMode(modes...))
	return err
}

// RequestMode requests the current state of the given modes from the terminal.
// This is typically used to check if a specific mode is recognized, enabled,
// or disabled on the terminal.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) RequestMode(mode ansi.Mode) error {
	_, err := t.WriteString(ansi.RequestMode(mode))
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
		return nil
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
// The [Terminal] manages the alternate screen buffer for you based on the
// [Viewport] used during [Terminal.Display]. This means that you don't need to
// call this unless you know what you're doing.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) EnterAltScreen() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.enterAltScreen(true)
}

// cursor indicates whether we want to set the cursor visibility state after
// entering the alt screen.
// We do this because some terminals maintain a separate cursor visibility
// state for the alt screen and the normal screen.
func (t *Terminal) enterAltScreen(cursor bool) error {
	t.scr.EnterAltScreen()
	if cursor {
		if t.scr.CursorHidden() {
			t.scr.WriteString(ansi.HideCursor) //nolint:errcheck
		} else {
			t.scr.WriteString(ansi.ShowCursor) //nolint:errcheck
		}
	}
	t.scr.SetRelativeCursor(false)
	t.modes[ansi.AltScreenSaveCursorMode] = ansi.ModeSet
	return nil
}

// ExitAltScreen exits the alternate screen buffer and returns to the normal
// screen buffer.
//
// The [Terminal] manages the alternate screen buffer for you based on the
// [Viewport] used during [Terminal.Display]. This means that you don't need to
// call this unless you know what you're doing.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) ExitAltScreen() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.exitAltScreen(true)
}

// cursor indicates whether we want to set the cursor visibility state after
// exiting the alt screen.
// We do this because some terminals maintain a separate cursor visibility
// state for the alt screen and the normal screen.
func (t *Terminal) exitAltScreen(cursor bool) error {
	t.scr.ExitAltScreen()
	if cursor {
		if t.scr.CursorHidden() {
			t.scr.WriteString(ansi.HideCursor) //nolint:errcheck
		} else {
			t.scr.WriteString(ansi.ShowCursor) //nolint:errcheck
		}
	}
	t.scr.SetRelativeCursor(true)
	t.modes[ansi.AltScreenSaveCursorMode] = ansi.ModeReset
	return nil
}

// ShowCursor shows the terminal cursor.
//
// The [Terminal] manages the visibility of the cursor for you based on the
// [Viewport] used during [Terminal.Display]. This means that you don't need to
// call this unless you know what you're doing.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) ShowCursor() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.showCursor()
}

func (t *Terminal) showCursor() error {
	t.scr.ShowCursor()
	t.modes[ansi.TextCursorEnableMode] = ansi.ModeSet
	return nil
}

// HideCursor hides the terminal cursor.
//
// The [Terminal] manages the visibility of the cursor for you based on the
// [Viewport] used during [Terminal.Display]. This means that you don't need to
// call this unless you know what you're doing.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) HideCursor() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.hideCursor()
}

func (t *Terminal) hideCursor() error {
	t.scr.HideCursor()
	t.modes[ansi.TextCursorEnableMode] = ansi.ModeReset
	return nil
}

// SetTitle sets the title of the terminal window. This is typically used to
// set the title of the terminal window to the name of the application.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) SetTitle(title string) error {
	_, err := t.WriteString(ansi.SetWindowTitle(title))
	return err
}

// Resize resizes the terminal to the given width and height. It returns an
// error if the resize fails.
func (t *Terminal) Resize(width, height int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if width == t.size.Width && height == t.size.Height {
		// No change in size.
		return nil
	}
	t.size.Width = width
	t.size.Height = height
	if t.buf != nil && t.scr.AltScreen() {
		// We only resize the screen if we're in the alt screen buffer. Inline
		// mode resizes the screen based on the frame height and terminal
		// width. See [terminalWriter.render] for more details.
		t.scr.Resize(width, height)
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
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		return ErrStarted
	}

	if t.inTty == nil && t.outTty == nil {
		return ErrNotTerminal
	}

	t.winchTty = t.inTty
	if t.winchTty == nil {
		t.winchTty = t.outTty
	}

	// We always hide the cursor when we start.
	t.hideCursor() //nolint:errcheck

	if err := t.rd.Start(); err != nil {
		return fmt.Errorf("error starting terminal: %w", err)
	}

	recvs := []InputReceiver{t.rdr}
	if runtime.GOOS != "windows" {
		t.wrdr = &WinChReceiver{t.winchTty}
		if err := t.wrdr.Start(); err != nil {
			return fmt.Errorf("error starting window size receiver: %w", err)
		}
		recvs = append(recvs, t.wrdr)
	}

	t.im = NewInputManager(recvs...)
	t.started = true

	return nil
}

// Restore restores the terminal to its original state. This should be called
// after [MakeRaw] to restore the terminal to its original state. Otherwise, it
// is a no-op.
func (t *Terminal) Restore() error {
	t.mu.Lock()
	defer t.mu.Unlock()

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
	// Display nothing.
	if t.buf != nil && !t.scr.AltScreen() {
		// Go to the bottom of the screen.
		t.scr.MoveTo(t.buf, 0, t.buf.Height()-1)
	}
	if t.scr.AltScreen() {
		t.exitAltScreen(false) //nolint:errcheck
	}
	if t.scr.CursorHidden() {
		t.showCursor() //nolint:errcheck
		t.cursorHidden = false
	}
	var buf bytes.Buffer
	for m, s := range t.modes {
		switch m {
		case ansi.TextCursorEnableMode, ansi.AltScreenSaveCursorMode:
			// These modes are handled by the renderer.
			continue
		}
		var reset bool
		ds, ok := defaultModes[m]
		if ok && s != ds {
			reset = s.IsSet() != ds.IsSet()
		} else {
			reset = s.IsSet()
		}
		if reset {
			buf.WriteString(ansi.ResetMode(m))
		}
	}
	if _, err := t.scr.WriteString(buf.String()); err != nil {
		return fmt.Errorf("error resetting terminal modes: %w", err)
	}
	return t.scr.Flush()
}

// Shutdown restores the terminal to its original state and stops the event
// channel in a graceful manner.
// This waits for any pending events to be processed or the context to be
// done before closing the event channel.
func (t *Terminal) Shutdown(ctx context.Context) (rErr error) {
	defer func() {
		err := t.close(false)
		if rErr == nil {
			rErr = err
		}
	}()

	// Cancel the input reader.
	t.rd.Cancel()
	if err := t.rdr.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down input reader: %w", err)
	}

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

// close closes any resources used by the terminal. This is typically used to
// close the terminal when it is no longer needed. When reset is true, it will
// also reset the terminal screen.
func (t *Terminal) close(reset bool) (rErr error) {
	defer t.closeChannels()

	defer func() {
		err := t.Restore()
		if rErr == nil && err != nil {
			rErr = fmt.Errorf("error restoring terminal state: %w", err)
		}
		t.buf = nil
		if reset {
			// Reset screen.
			t.scr = t.newScreen()
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

// Close close any resources used by the terminal and restore the terminal to
// its original state.
func (t *Terminal) Close() error {
	return t.close(true)
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
	// Start receiving events from the terminal if it hasn't been started yet.
	t.evOnce.Do(func() {
		go func() {
			defer t.closeChannels()

			t.err = t.im.ReceiveEvents(ctx, t.evch)
			if errors.Is(t.err, io.EOF) || errors.Is(t.err, cancelreader.ErrCanceled) {
				t.err = nil
			}
		}()
	})
	return t.evch
}

// Err returns the error that occurred while receiving events from the
// terminal.
// This is typically used to check for errors that occurred while
// receiving events.
func (t *Terminal) Err() error {
	return t.err
}

// NewStyledString is a convenience function to create a new [StyledString]
// with the string. It uses the [Terminal]'s [ansi.Method] to calculate the
// width of the string and returns a new [StyledString] with the given string.
func (t *Terminal) NewStyledString(str string) *StyledString {
	return NewStyledString(t.method, str)
}

// PrependString adds the given string to the top of the terminal screen. The
// string is split into lines and each line is added as a new line at the top
// of the screen. The added lines are not managed by the terminal and will not
// be cleared or updated by the [Terminal].
//
// This will truncate each line to the terminal width, so if the string is
// longer than the terminal width, it will be truncated to fit.
//
// Using this when the terminal is using the alternate screen or when occupying
// the whole screen may not produce any visible effects. This is because once
// the terminal writes the prepended lines, they will get overwritten by the
// next frame.
//
// Note that this won't take any effect until the next [Terminal.Display] or
// [Terminal.Flush] call.
func (t *Terminal) PrependString(str string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// We truncate the string to the terminal width.
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		if t.method.StringWidth(line) > t.size.Width {
			sb.WriteString(t.method.Truncate(line, t.size.Width, ""))
		} else {
			sb.WriteString(line)
		}
		if i < len(lines)-1 {
			sb.WriteByte('\n')
		}
	}

	t.scr.PrependString(sb.String())
	return nil
}

// PrependLines adds lines of cells to the top of the terminal screen. The
// added line is unmanaged and will not be cleared or updated by the
// [Terminal].
//
// This will truncate each line to the terminal width, so if the string is
// longer than the terminal width, it will be truncated to fit.
//
// Using this when the terminal is using the alternate screen or when occupying
// the whole screen may not produce any visible effects. This is because once
// the terminal writes the prepended lines, they will get overwritten by the
// next frame.
func (t *Terminal) PrependLines(lines ...Line) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	truncatedLines := make([]Line, 0, len(lines))
	for _, l := range lines {
		// We truncate the line to the terminal width.
		if len(l) > t.size.Width {
			truncatedLines = append(truncatedLines, l[:t.size.Width])
		} else {
			truncatedLines = append(truncatedLines, l)
		}
	}

	t.scr.PrependLines(truncatedLines...)
	return nil
}

// Write writes the given data to the underlying screen buffer. It implements
// the [io.Writer] interface. Data written won't be flushed to the terminal
// until [Terminal.Flush] is called.
func (t *Terminal) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.scr.Write(p)
}

// WriteString writes the given string to the underlying screen buffer. It
// implements the [io.StringWriter] interface. Data written won't be flushed to
// the terminal until [Terminal.Flush] is called.
func (t *Terminal) WriteString(s string) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.scr.WriteString(s)
}
