package tv

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/x/term"
)

// Size represents the row-column size of a window.
type Size struct {
	Width  int
	Height int
}

// WinchReceiver listens for window size changes and sends the new size to the
// event channel.
type WinchReceiver struct {
	out                   term.File
	lastWidth, lastHeight int
}

// NewWinchReceiver creates a new WinchListener that listens for window size
// changes on the given output file.
func NewWinchReceiver(out io.Writer) (*WinchReceiver, error) {
	f, ok := out.(*os.File)
	if !ok {
		return nil, fmt.Errorf("output is not a file: %T", out)
	}
	w, h, err := term.GetSize(f.Fd())
	if err != nil {
		return nil, fmt.Errorf("failed to get window size: %w", err)
	}

	return &WinchReceiver{
		out:        f,
		lastWidth:  w,
		lastHeight: h,
	}, nil
}

// ReceiveEvents listens for window size changes and sends the new size to the
// event channel. It stops when the context is done or an error occurs.
func (l *WinchReceiver) ReceiveEvents(ctx context.Context, evch chan<- Event, errch chan<- error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	defer signal.Stop(sig)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sig:
			s, err := l.checkSize()
			if err != nil {
				errch <- err
				return
			}

			evch <- s
		}
	}
}
