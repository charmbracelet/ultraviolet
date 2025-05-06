package tv

import (
	"context"
)

// TerminalInputReceiver receive input events from the terminal input stream.
// On Unix platforms, it simply reads unicode, control, and escape codes from
// the terminal input stream. On Windows, it uses the Windows Console API to
// read input events like window-size, mouse, key, focus events, etc. It also
// uses the [SequenceParser] to parse escape sequences incoming from the
// Windows Console API as key events.
type TerminalInputReceiver struct {
	*TerminalReader
}

var _ InputReceiver = (*TerminalInputReceiver)(nil)

// ReceiveEvents implements InputReceiver.
func (t *TerminalInputReceiver) ReceiveEvents(ctx context.Context, events chan<- Event) (rErr error) {
	go func() {
		// Wait for the context to be done and cancel the reader.
		<-ctx.Done()
		t.Cancel()
	}()

	for {
		evs, err := t.ReadEvents()
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
