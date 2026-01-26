package doc

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/net/html"
)

func TestEventAPI(t *testing.T) {
	// Create a document with a button
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	button := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "submit"},
		},
	}

	parent.AppendChild(button)
	doc := NewDocument(parent, nil)

	// Get the button and focus it
	buttonNode := doc.GetElementByID("submit")
	if buttonNode == nil {
		t.Fatal("expected to find button")
	}
	buttonNode.Focus()

	// Add listener that checks Event properties
	var receivedEvent *Event
	listener := func(ev *Event) bool {
		receivedEvent = ev
		return true
	}

	buttonNode.AddEventListener(EventKeyPress, listener)

	// Dispatch a key press event
	doc.dispatchEvent(uv.KeyPressEvent{})

	// Verify the Event structure
	if receivedEvent == nil {
		t.Fatal("expected to receive event")
	}

	if receivedEvent.Type != EventKeyPress {
		t.Errorf("expected Type to be 'keypress', got %q", receivedEvent.Type)
	}

	if receivedEvent.Target != buttonNode {
		t.Error("expected Target to be the button node")
	}

	if receivedEvent.Event == nil {
		t.Error("expected Event to contain the underlying UV event")
	}

	if _, ok := receivedEvent.Event.(uv.KeyPressEvent); !ok {
		t.Error("expected Event to be a KeyPressEvent")
	}
}

func TestEventBubblingWithTarget(t *testing.T) {
	// Create nested structure: div > button
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "id", Val: "parent"},
		},
	}

	button := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child"},
		},
	}

	parent.AppendChild(button)
	doc := NewDocument(parent, nil)

	// Get nodes
	parentNode := doc.GetElementByID("parent")
	buttonNode := doc.GetElementByID("child")
	if parentNode == nil || buttonNode == nil {
		t.Fatal("expected to find both nodes")
	}

	// Focus the button
	buttonNode.Focus()

	// Track which nodes received the event
	var targets []Node
	listener := func(ev *Event) bool {
		targets = append(targets, ev.Target)
		return false // Allow bubbling
	}

	// Add listener to both button and parent
	buttonNode.AddEventListener(EventKeyPress, listener)
	parentNode.AddEventListener(EventKeyPress, listener)

	// Dispatch event
	doc.dispatchEvent(uv.KeyPressEvent{})

	// Should have received event twice during bubbling
	if len(targets) != 2 {
		t.Fatalf("expected 2 events during bubbling, got %d", len(targets))
	}

	// First event should target the button (where focus is)
	if targets[0] != buttonNode {
		t.Error("expected first event target to be button")
	}

	// Second event should target the parent (during bubbling)
	if targets[1] != parentNode {
		t.Error("expected second event target to be parent during bubbling")
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		uvEvent  uv.Event
		expected EventType
	}{
		{uv.KeyPressEvent{}, EventKeyPress},
		{uv.MouseClickEvent{}, EventMouseClick},
		{uv.WindowSizeEvent{}, EventResize},
	}

	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}
	doc := NewDocument(parent, nil)

	for _, tt := range tests {
		var receivedType EventType
		listener := func(ev *Event) bool {
			receivedType = ev.Type
			return true
		}

		doc.RemoveEventListener(EventAll) // Clear previous listeners
		doc.AddEventListener(EventAll, listener)
		doc.dispatchEvent(tt.uvEvent)

		if receivedType != tt.expected {
			t.Errorf("for %T: expected Type %q, got %q", tt.uvEvent, tt.expected, receivedType)
		}
	}
}
