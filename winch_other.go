//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !aix && !zos
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!aix,!zos

package tv

import "context"

func (*WinchReceiver) receiveEvents(context.Context, chan<- Event, chan<- error) {
	// No-op on non-Unix systems
}
