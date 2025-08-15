package uv

import (
	"context"
	"os"
	"sync"

	"github.com/charmbracelet/x/term"
)

// WindowSizeNotifier represents a notifier that listens for window size
// changes using the SIGWINCH signal and notifies the given channel.
type WindowSizeNotifier struct {
	f   term.File
	sig chan os.Signal
	m   sync.Mutex
}

// NewWindowSizeNotifier creates a new WindowSizeNotifier with the given file.
func NewWindowSizeNotifier(f term.File) *WindowSizeNotifier {
	if f == nil {
		panic("no file set")
	}
	return &WindowSizeNotifier{
		f:   f,
		sig: make(chan os.Signal),
	}
}

// StreamEvents reads the terminal size change events and sends them to the
// given channel. It stops when the context is done.
func (n *WindowSizeNotifier) StreamEvents(ctx context.Context, ch chan<- Event) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-n.sig:
			cells, pixels, err := n.GetWindowSize()
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return nil
			case ch <- WindowSizeEvent(cells):
			}
			if pixels.Width > 0 && pixels.Height > 0 {
				select {
				case <-ctx.Done():
					return nil
				case ch <- WindowPixelSizeEvent(pixels):
				}
			}
		}
	}
}

// Start starts the notifier by registering for the SIGWINCH signal. It must be
// called before using [WindowSizeNotifier.Notify] or
// [WindowSizeNotifier.StreamEvents].
func (n *WindowSizeNotifier) Start() error {
	return n.start()
}

// Stop stops the notifier and cleans up resources.
func (n *WindowSizeNotifier) Stop() error {
	return n.stop()
}

// GetWindowSize returns the current size of the terminal window.
func (n *WindowSizeNotifier) GetWindowSize() (cells Size, pixels Size, err error) {
	return n.getWindowSize()
}

// GetSize returns the current cell size of the terminal window.
func (n *WindowSizeNotifier) GetSize() (width, height int, err error) {
	n.m.Lock()
	defer n.m.Unlock()

	width, height, err = term.GetSize(n.f.Fd())
	if err != nil {
		return 0, 0, err //nolint:wrapcheck
	}

	return width, height, nil
}
