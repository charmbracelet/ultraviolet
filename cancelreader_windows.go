//go:build windows
// +build windows

package uv

import (
	"fmt"
	"io"
	"sync"

	xwindows "github.com/charmbracelet/x/windows"
	"github.com/muesli/cancelreader"
	"golang.org/x/sys/windows"
)

type conInputReader struct {
	cancelMixin
	conin        windows.Handle
	originalMode uint32
	newMode      uint32
}

var _ cancelreader.CancelReader = &conInputReader{}

// NewCancelReader creates a new [cancelreader.CancelReader] that provides a
// cancelable reader interface that can be used to cancel reads.
func NewCancelReader(r io.Reader) (cancelreader.CancelReader, error) {
	fallback := func(io.Reader) (cancelreader.CancelReader, error) {
		return cancelreader.NewReader(r)
	}

	f, ok := r.(cancelreader.File)
	if !ok {
		return fallback(r)
	}

	// Read from the handle we were handed rather than assuming standard input. A
	// caller whose stdio is redirected opens CONIN$ itself and passes it here: that
	// handle is a console, even though it is not os.Stdin.
	conin := windows.Handle(f.Fd())

	// If data was piped in, the handle does not emit console events. GetConsoleMode
	// fails on a handle that is not a console, so fall back to the default
	// cancelreader implementation in that case.
	var dummy uint32
	if windows.GetConsoleMode(conin, &dummy) != nil {
		return fallback(r)
	}

	// Discard any pending input events.
	if err := xwindows.FlushConsoleInputBuffer(conin); err != nil {
		return fallback(r)
	}

	modes := []uint32{
		windows.ENABLE_VIRTUAL_TERMINAL_INPUT,
		windows.ENABLE_WINDOW_INPUT,
		windows.ENABLE_EXTENDED_FLAGS,
	}

	originalMode, newMode, err := prepareConsole(conin, modes...)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare console input: %w", err)
	}

	return &conInputReader{
		conin:        conin,
		originalMode: originalMode,
		newMode:      newMode,
	}, nil
}

// Cancel implements cancelreader.CancelReader.
func (r *conInputReader) Cancel() bool {
	r.setCanceled()

	return windows.CancelIoEx(r.conin, nil) == nil || windows.CancelIo(r.conin) == nil
}

// Close implements cancelreader.CancelReader.
func (r *conInputReader) Close() error {
	if r.originalMode != 0 {
		err := windows.SetConsoleMode(r.conin, r.originalMode)
		if err != nil {
			return fmt.Errorf("reset console mode: %w", err)
		}
	}

	return nil
}

// Read implements cancelreader.CancelReader.
func (r *conInputReader) Read(data []byte) (int, error) {
	if r.isCanceled() {
		return 0, cancelreader.ErrCanceled
	}

	var n uint32
	if err := windows.ReadFile(r.conin, data, &n, nil); err != nil {
		return int(n), fmt.Errorf("read console input: %w", err)
	}

	return int(n), nil
}

func prepareConsole(input windows.Handle, modes ...uint32) (originalMode, newMode uint32, err error) {
	err = windows.GetConsoleMode(input, &originalMode)
	if err != nil {
		return 0, 0, fmt.Errorf("get console mode: %w", err)
	}

	for _, mode := range modes {
		newMode |= mode
	}

	err = windows.SetConsoleMode(input, newMode)
	if err != nil {
		return 0, 0, fmt.Errorf("set console mode: %w", err)
	}

	return originalMode, newMode, nil
}

// cancelMixin represents a goroutine-safe cancelation status.
type cancelMixin struct {
	unsafeCanceled bool
	lock           sync.Mutex
}

func (c *cancelMixin) setCanceled() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.unsafeCanceled = true
}

func (c *cancelMixin) isCanceled() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.unsafeCanceled
}
