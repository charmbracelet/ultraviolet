package tv

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Event represents an input event that can be received from an input source.
type Event interface{}

// Size represents the size of the terminal window.
type Size struct {
	Width  int
	Height int
}

// WindowSizeEvent represents the window size in cells.
type WindowSizeEvent Size

// WindowPixelSizeEvent represents the window size in pixels.
type WindowPixelSizeEvent Size

// InputReceiver is an interface for receiving input events from an input source.
type InputReceiver interface {
	// ReceiveEvents read input events and channel them to the given event
	// channel. The listener stops when either the context is done or an error
	// occurs. Caller is responsible for closing the channels.
	ReceiveEvents(ctx context.Context, events chan<- Event) error
}

// InputManager manages input events from multiple input sources. It listens
// for input events from the registered input sources and combines them into a
// single event channel. It also handles errors from the input sources and
// sends them to the error channel.
type InputManager struct {
	receivers []InputReceiver
}

// NewInputManager creates a new InputManager with the input receivers.
func NewInputManager(receivers ...InputReceiver) *InputManager {
	im := &InputManager{
		receivers: receivers,
	}
	return im
}

// RegisterReceiver registers a new input receiver with the input manager.
func (im *InputManager) RegisterReceiver(r InputReceiver) {
	im.receivers = append(im.receivers, r)
}

// ReceiveEvents starts receiving events from the registered input
// receivers. It sends the events to the given event and error channels.
func (im *InputManager) ReceiveEvents(ctx context.Context, events chan<- Event) error {
	errg, ctx := errgroup.WithContext(ctx)
	for _, r := range im.receivers {
		errg.Go(func() error {
			return r.ReceiveEvents(ctx, events)
		})
	}

	// Wait for all receivers to finish
	return errg.Wait()
}
