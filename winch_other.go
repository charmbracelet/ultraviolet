//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package uv

import (
	"context"
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
