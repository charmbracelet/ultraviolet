//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package uv

import (
	"context"
	"fmt"

	"github.com/charmbracelet/x/term"
)

func (n *WindowSizeNotifier) start() error {
	return ErrPlatformNotSupported
}

func (n *WindowSizeNotifier) close() error {
	return ErrPlatformNotSupported
}

func (n *WindowSizeNotifier) shutdown(context.Context) error {
	return ErrPlatformNotSupported
}

func (n *WindowSizeNotifier) getWindowSize() (cells Size, pixels Size, err error) {
	cells.Width, cells.Height, err = n.GetSize()
	return cells, pixels, err
}

func (*WinChReceiver) receiveEvents(context.Context, term.File, chan<- Event) error {
	return fmt.Errorf("SIGWINCH not supported on this platform")
}
