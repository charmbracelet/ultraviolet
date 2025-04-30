//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tv

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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
			s, err := checkSize(l.out)
			if err != nil {
				errch <- err
				return
			}

			evch <- s
		}
	}
}
