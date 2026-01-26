package doc

import (
	"testing"

	"golang.org/x/net/html"
)

func TestFocus(t *testing.T) {
	// Create a document with child elements
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "id", Val: "parent"},
		},
	}

	child := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child"},
		},
	}

	parent.AppendChild(child)
	doc := NewDocument(parent, nil)

	// Initially, the document itself should be the active element
	if doc.ActiveElement() != doc.node {
		t.Error("expected document to be initially active")
	}

	// Get the child node
	childNode := doc.GetElementByID("child")
	if childNode == nil {
		t.Fatal("expected to find child element")
	}

	// Focus the child
	childNode.Focus()

	// Child should now be the active element
	if doc.ActiveElement() != childNode {
		t.Error("expected child to be active after Focus()")
	}
}

func TestBlur(t *testing.T) {
	// Create a document with child elements
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "id", Val: "parent"},
		},
	}

	child := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child"},
		},
	}

	parent.AppendChild(child)
	doc := NewDocument(parent, nil)

	// Get the child node
	childNode := doc.GetElementByID("child")
	if childNode == nil {
		t.Fatal("expected to find child element")
	}

	// Focus the child
	childNode.Focus()
	if doc.ActiveElement() != childNode {
		t.Error("expected child to be active")
	}

	// Blur the child
	childNode.Blur()

	// Document root should now be the active element
	if doc.ActiveElement() != doc.node {
		t.Error("expected document to be active after Blur()")
	}
}

func TestBlurNonActiveElement(t *testing.T) {
	// Create a document with two child elements
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	child1 := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child1"},
		},
	}

	child2 := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child2"},
		},
	}

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	doc := NewDocument(parent, nil)

	// Get both children
	child1Node := doc.GetElementByID("child1")
	child2Node := doc.GetElementByID("child2")
	if child1Node == nil || child2Node == nil {
		t.Fatal("expected to find both child elements")
	}

	// Focus child1
	child1Node.Focus()
	if doc.ActiveElement() != child1Node {
		t.Error("expected child1 to be active")
	}

	// Blur child2 (which is not active)
	child2Node.Blur()

	// child1 should still be active (blurring a non-active element should do nothing)
	if doc.ActiveElement() != child1Node {
		t.Error("expected child1 to still be active after blurring child2")
	}
}

func TestFocusMultipleTimes(t *testing.T) {
	// Create a document with child elements
	parent := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	child1 := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child1"},
		},
	}

	child2 := &html.Node{
		Type: html.ElementNode,
		Data: "button",
		Attr: []html.Attribute{
			{Key: "id", Val: "child2"},
		},
	}

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	doc := NewDocument(parent, nil)

	// Get both children
	child1Node := doc.GetElementByID("child1")
	child2Node := doc.GetElementByID("child2")
	if child1Node == nil || child2Node == nil {
		t.Fatal("expected to find both child elements")
	}

	// Focus child1
	child1Node.Focus()
	if doc.ActiveElement() != child1Node {
		t.Error("expected child1 to be active")
	}

	// Focus child2
	child2Node.Focus()
	if doc.ActiveElement() != child2Node {
		t.Error("expected child2 to be active after second focus")
	}

	// Focus child1 again
	child1Node.Focus()
	if doc.ActiveElement() != child1Node {
		t.Error("expected child1 to be active again")
	}
}
