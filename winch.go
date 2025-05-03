package tv

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/x/term"
)

// WinChReceiver listens for window size changes using (SIGWINCH) and sends the
// new size to the event channel.
// This is a Unix-specific implementation and should be used on Unix-like
// systems and won't work on Windows.
type WinChReceiver struct{ io.Writer }

// ReceiveEvents listens for window size changes and sends the new size to the
// event channel. It stops when the context is done or an error occurs.
func (l *WinChReceiver) ReceiveEvents(ctx context.Context, evch chan<- Event) error {
	f, ok := l.Writer.(term.File)
	if !ok {
		return fmt.Errorf("output is not a terminal: %T", l.Writer)
	}
	return l.receiveEvents(ctx, f, evch)
}
