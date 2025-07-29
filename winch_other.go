//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package uv

import (
	"context"
	"fmt"

	"github.com/charmbracelet/x/term"
)

func (*WinChReceiver) receiveEvents(context.Context, term.File, chan<- Event) error {
	return fmt.Errorf("SIGWINCH not supported on this platform")
}
