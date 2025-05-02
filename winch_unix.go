//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tv

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/termios"
)

func (l *WinchReceiver) receiveEvents(ctx context.Context, evch chan<- Event, errch chan<- error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer signal.Stop(sig)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sig:
			winsize, err := termios.GetWinsize(int(l.out.Fd()))
			if err != nil {
				errch <- err
				return
			}

			go func() { evch <- WindowSize{int(winsize.Col), int(winsize.Row)} }()
		}
	}
}
