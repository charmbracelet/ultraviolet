package uv

import (
	"context"
	"fmt"
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

// WinChReceiver listens for window size changes using (SIGWINCH) and sends the
// new size to the event channel.
// This is a Unix-specific implementation and should be used on Unix-like
// systems and won't work on Windows.
type WinChReceiver struct{ term.File }

// Start starts the receiver.
func (l *WinChReceiver) Start() error {
	if l.File == nil {
		return fmt.Errorf("no file set")
	}
	_, _, err := term.GetSize(l.Fd())
	return err //nolint:wrapcheck
}

// ReceiveEvents listens for window size changes and sends the new size to the
// event channel. It stops when the context is done or an error occurs.
func (l *WinChReceiver) ReceiveEvents(ctx context.Context, evch chan<- Event) error {
	return l.receiveEvents(ctx, l.File, evch)
}
