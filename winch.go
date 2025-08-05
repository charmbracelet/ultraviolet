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
	f     term.File
	sig   chan os.Signal
	donec chan struct{}
	wg    sync.WaitGroup
	m     sync.Mutex
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

// Notify starts a goroutine that listens for window size changes and notifies
// the given channel when a change occurs.
func (n *WindowSizeNotifier) Notify(ch chan<- struct{}) {
	n.wg.Add(1)
	go func(donec chan struct{}) {
		defer n.wg.Done()
		for {
			select {
			case <-donec:
				return
			case <-n.sig:
				n.m.Lock()
				select {
				case <-donec:
					return
				case ch <- struct{}{}:
				}
				n.m.Unlock()
			}
		}
	}(n.donec)
}

// Start starts the notifier by registering for the SIGWINCH signal.
func (n *WindowSizeNotifier) Start() error {
	return n.start()
}

// Close closes the notifier and stops listening for window size changes.
func (n *WindowSizeNotifier) Close() error {
	return n.close()
}

// Shutdown stops the notifier and cleans up resources.
func (n *WindowSizeNotifier) Shutdown(ctx context.Context) error {
	return n.shutdown(ctx)
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
