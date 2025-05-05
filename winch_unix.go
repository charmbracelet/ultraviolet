//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tv

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/termios"
)

func (l *WinChReceiver) receiveEvents(ctx context.Context, f term.File, evch chan<- Event) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer signal.Stop(sig)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sig:
			winsize, err := termios.GetWinsize(int(f.Fd()))
			if err != nil {
				return err
			}

			go func() {
				select {
				case <-ctx.Done():
				case evch <- WindowSizeEvent{int(winsize.Col), int(winsize.Row)}:
				}
			}()

			go func() {
				select {
				case <-ctx.Done():
				case evch <- WindowPixelSizeEvent{int(winsize.Xpixel), int(winsize.Ypixel)}:
				}
			}()
		}
	}
}
