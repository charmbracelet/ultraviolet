//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris zos

package uv

import (
	"context"
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
	wgc := make(chan struct{})
	go func() {
		defer close(wgc)
		n.wg.Wait()
	}()

	if err := n.close(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return nil
	case <-wgc:
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
