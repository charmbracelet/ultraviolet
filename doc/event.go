package doc

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// EventType represents the type of event that can be listened to.
type EventType string

const (
	// EventAll matches all events
	EventAll EventType = "*"
	// EventKeyPress matches key press events
	EventKeyPress EventType = "keypress"
	// EventKeyRelease matches key release events
	EventKeyRelease EventType = "keyrelease"
	// EventMouseClick matches mouse click events
	EventMouseClick EventType = "mouseclick"
	// EventMouseRelease matches mouse release events
	EventMouseRelease EventType = "mouserelease"
	// EventMouseWheel matches mouse wheel events
	EventMouseWheel EventType = "mousewheel"
	// EventMouseMotion matches mouse motion events
	EventMouseMotion EventType = "mousemotion"
	// EventResize matches window resize events
	EventResize EventType = "resize"
	// EventFocus matches focus events
	EventFocus EventType = "focus"
	// EventBlur matches blur events
	EventBlur EventType = "blur"
	// EventPaste matches paste events
	EventPaste EventType = "paste"
)

// Event represents a DOM event with properties similar to JavaScript events.
type Event struct {
	// Type is the type of the event (e.g., "keypress", "mouseclick")
	Type EventType
	// Target is the node that the event was dispatched to
	Target Node
	// Event is the underlying UV event
	Event uv.Event
}

// EventListener is a function that handles events on any node (including Document).
// Returns true if the event was handled (preventing it from bubbling up), or false to allow bubbling.
type EventListener func(ev *Event) bool

// newEvent creates a new Event from a UV event and target node.
func newEvent(uvEvent uv.Event, target Node) *Event {
	return &Event{
		Type:   getEventType(uvEvent),
		Target: target,
		Event:  uvEvent,
	}
}

// dispatchEvent dispatches an event to the active node and bubbles up if not handled.
// Returns true if the event was handled.
func (d *Document) dispatchEvent(uvEvent uv.Event) bool {
	current := d.ActiveElement()
	eventType := getEventType(uvEvent)

	// Bubble up from active node to document root
	for current != nil {
		// Create event with current target
		ev := newEvent(uvEvent, current)

		if n, ok := current.(*node); ok {
			// Try specific event type listeners
			if listeners, ok := n.listeners[string(eventType)]; ok {
				for _, listener := range listeners {
					if listener(ev) {
						return true // Event was handled
					}
				}
			}

			// Try "all events" listeners
			if listeners, ok := n.listeners[string(EventAll)]; ok {
				for _, listener := range listeners {
					if listener(ev) {
						return true // Event was handled
					}
				}
			}
		}

		// Move to parent
		current = current.Parent()
	}

	return false // Event was not handled
}

// getEventType returns the EventType for a uv.Event.
func getEventType(ev uv.Event) EventType {
	switch ev.(type) {
	case uv.KeyPressEvent:
		return EventKeyPress
	case uv.KeyReleaseEvent:
		return EventKeyRelease
	case uv.MouseClickEvent:
		return EventMouseClick
	case uv.MouseReleaseEvent:
		return EventMouseRelease
	case uv.MouseWheelEvent:
		return EventMouseWheel
	case uv.MouseMotionEvent:
		return EventMouseMotion
	case uv.WindowSizeEvent:
		return EventResize
	case uv.FocusEvent:
		return EventFocus
	case uv.BlurEvent:
		return EventBlur
	case uv.PasteEvent:
		return EventPaste
	default:
		return EventAll
	}
}
