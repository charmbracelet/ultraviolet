package tv

import (
	"context"
	"sync"
	"time"
)

// TerminalInputReceiver receive input events from the terminal input stream.
// On Unix platforms, it simply reads unicode, control, and escape codes from
// the terminal input stream. On Windows, it uses the Windows Console API to
// read input events like window-size, mouse, key, focus events, etc. It also
// uses the [SequenceParser] to parse escape sequences incoming from the
// Windows Console API as key events.
type TerminalInputReceiver struct {
	rd           *TerminalReader
	readLoopDone chan struct{}
	once         sync.Once
}

var _ InputReceiver = (*TerminalInputReceiver)(nil)

// NewTerminalInputReceiver creates a new [TerminalInputReceiver] for the given
// [TerminalReader].
func NewTerminalInputReceiver(rd *TerminalReader) *TerminalInputReceiver {
	return &TerminalInputReceiver{
		rd:           rd,
		readLoopDone: make(chan struct{}),
	}
}

// ReceiveEvents implements InputReceiver.
func (t *TerminalInputReceiver) ReceiveEvents(ctx context.Context, events chan<- Event) (rErr error) {
	t.once.Do(func() {
		go func() {
			<-ctx.Done()
			t.rd.Cancel()
		}()

		defer close(t.readLoopDone)
	})

	for {
		evs, err := t.rd.ReadEvents()
		if err != nil {
			return err
		}
		for _, ev := range evs {
			select {
			case <-ctx.Done():
				return nil
			case events <- ev:
			}
		}
	}
}

// Shutdown gracefully shuts down the input receiver.
func (t *TerminalInputReceiver) Shutdown(ctx context.Context) error {
	select {
	case <-t.readLoopDone:
	case <-time.After(500 * time.Millisecond): //nolint:mnd
		// The read loop hangs, which means the input
		// cancelReader's cancel function has returned true even
		// though it was not able to cancel the read.
	}
	return nil
}
