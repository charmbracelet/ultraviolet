//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris zos

package uv

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/termios"
)

func (n *WindowSizeNotifier) start() error {
	n.m.Lock()
	defer n.m.Unlock()
	if n.f == nil || !term.IsTerminal(n.f.Fd()) {
		return ErrNotTerminal
	}

	n.donec = make(chan struct{})
	signal.Notify(n.sig, syscall.SIGWINCH)
	return nil
}

func (n *WindowSizeNotifier) close() error {
	n.m.Lock()
	signal.Stop(n.sig)
	close(n.donec)
	n.m.Unlock()
	return nil
}

func (n *WindowSizeNotifier) shutdown(ctx context.Context) (err error) {
	go func(err *error) {
		*err = n.close()
		n.wg.Wait()
	}(&err)

	select {
	case <-ctx.Done():
		return nil
	case <-n.donec:
	}

	return err
}

func (n *WindowSizeNotifier) getWindowSize() (cells Size, pixels Size, err error) {
	n.m.Lock()
	defer n.m.Unlock()

	winsize, err := termios.GetWinsize(int(n.f.Fd()))
	if err != nil {
		return Size{}, Size{}, err //nolint:wrapcheck
	}

	cells = Size{
		Width:  int(winsize.Col),
		Height: int(winsize.Row),
	}
	pixels = Size{
		Width:  int(winsize.Xpixel),
		Height: int(winsize.Ypixel),
	}
	return cells, pixels, nil
}

func (l *WinChReceiver) receiveEvents(ctx context.Context, f term.File, evch chan<- Event) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer signal.Stop(sig)

	sendWinSize := func(w, h int) {
		select {
		case <-ctx.Done():
		case evch <- WindowSizeEvent{w, h}:
		}
	}

	sendPixelSize := func(w, h int) {
		select {
		case <-ctx.Done():
		case evch <- WindowPixelSizeEvent{w, h}:
		}
	}

	// Listen for window size changes.
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sig:
			winsize, err := termios.GetWinsize(int(f.Fd()))
			if err != nil {
				return err //nolint:wrapcheck
			}

			go sendWinSize(int(winsize.Col), int(winsize.Row))
			go sendPixelSize(int(winsize.Xpixel), int(winsize.Ypixel))
		}
	}
}
