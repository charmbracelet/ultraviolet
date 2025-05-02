package tv

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"
)

// WinchReceiver listens for window size changes and sends the new size to the
// event channel.
type WinchReceiver struct {
	out term.File
}

// NewWinchReceiver creates a new WinchListener that listens for window size
// changes on the given output file.
func NewWinchReceiver(out io.Writer) (*WinchReceiver, error) {
	f, ok := out.(*os.File)
	if !ok {
		return nil, fmt.Errorf("output is not a file: %T", out)
	}

	return &WinchReceiver{
		out: f,
	}, nil
}

// ReceiveEvents listens for window size changes and sends the new size to the
// event channel. It stops when the context is done or an error occurs.
func (l *WinchReceiver) ReceiveEvents(ctx context.Context, evch chan<- Event, errch chan<- error) {
	l.receiveEvents(ctx, evch, errch)
}
