package dom

import (
	uv "github.com/charmbracelet/ultraviolet"
)

// NodeType represents the type of a Node.
// These values follow the Web DOM Node.nodeType constants.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Node/nodeType
type NodeType uint16

const (
	// ElementNode represents an Element node (e.g., <div>, <p>).
	ElementNode NodeType = 1

	// TextNode represents a Text node containing text content.
	TextNode NodeType = 3

	// DocumentNode represents the Document node (root of the DOM tree).
	DocumentNode NodeType = 9
)

// Node is the primary data type for the entire DOM.
// It represents a single node in the document tree.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Node
type Node interface {
	// NodeType returns the type of the node (Element, Text, Document, etc.).
	NodeType() NodeType

	// NodeName returns the name of the node.
	// For elements: the tag name (uppercase).
	// For text nodes: "#text".
	// For documents: "#document".
	NodeName() string

	// ParentNode returns the parent of this node, or nil if there is no parent.
	ParentNode() Node

	// ChildNodes returns a list of child nodes.
	ChildNodes() []Node

	// FirstChild returns the first child node, or nil if there are no children.
	FirstChild() Node

	// LastChild returns the last child node, or nil if there are no children.
	LastChild() Node

	// PreviousSibling returns the previous sibling node, or nil if there is none.
	PreviousSibling() Node

	// NextSibling returns the next sibling node, or nil if there is none.
	NextSibling() Node

	// AppendChild adds a node to the end of the list of children.
	// Returns the appended node.
	AppendChild(child Node) Node

	// InsertBefore inserts a node before a reference node as a child of the current node.
	// If referenceNode is nil, the node is inserted at the end of the list of children.
	// Returns the inserted node.
	InsertBefore(newNode, referenceNode Node) Node

	// RemoveChild removes a child node from the DOM.
	// Returns the removed node.
	RemoveChild(child Node) Node

	// ReplaceChild replaces a child node with a new node.
	// Returns the replaced (old) node.
	ReplaceChild(newChild, oldChild Node) Node

	// CloneNode creates a copy of the node.
	// If deep is true, the clone includes all descendants.
	CloneNode(deep bool) Node

	// TextContent returns the text content of the node and its descendants.
	TextContent() string

	// SetTextContent sets the text content of the node.
	SetTextContent(text string)

	// Render renders this node and its children to the screen within the given area.
	// This is specific to this TUI implementation and not part of Web DOM.
	Render(scr uv.Screen, area uv.Rectangle)

	// MinSize returns the minimum width and height needed to render this node.
	// This is specific to this TUI implementation and not part of Web DOM.
	MinSize(scr uv.Screen) (width, height int)

	// Internal methods for tree management
	setParentNode(parent Node)
	setPreviousSibling(node Node)
	setNextSibling(node Node)
}

// baseNode provides a base implementation of the Node interface.
// It contains common fields and methods used by all node types.
type baseNode struct {
	nodeType        NodeType
	nodeName        string
	parentNode      Node
	childNodes      []Node
	previousSibling Node
	nextSibling     Node
}

func (n *baseNode) NodeType() NodeType {
	return n.nodeType
}

func (n *baseNode) NodeName() string {
	return n.nodeName
}

func (n *baseNode) ParentNode() Node {
	return n.parentNode
}

func (n *baseNode) ChildNodes() []Node {
	return n.childNodes
}

func (n *baseNode) FirstChild() Node {
	if len(n.childNodes) > 0 {
		return n.childNodes[0]
	}
	return nil
}

func (n *baseNode) LastChild() Node {
	if len(n.childNodes) > 0 {
		return n.childNodes[len(n.childNodes)-1]
	}
	return nil
}

func (n *baseNode) PreviousSibling() Node {
	return n.previousSibling
}

func (n *baseNode) NextSibling() Node {
	return n.nextSibling
}

func (n *baseNode) AppendChild(child Node) Node {
	if child == nil {
		return nil
	}

	// Remove from old parent if it has one
	if oldParent := child.ParentNode(); oldParent != nil {
		oldParent.RemoveChild(child)
	}

	// Update sibling relationships
	if len(n.childNodes) > 0 {
		lastChild := n.childNodes[len(n.childNodes)-1]
		lastChild.setNextSibling(child)
		child.setPreviousSibling(lastChild)
	} else {
		child.setPreviousSibling(nil)
	}
	child.setNextSibling(nil)

	// Add to children
	n.childNodes = append(n.childNodes, child)
	// Note: setParentNode is called by the child, we pass the parent node
	// but baseNode can't convert itself to Node. This will be handled
	// by concrete implementations

	return child
}

func (n *baseNode) InsertBefore(newNode, referenceNode Node) Node {
	if newNode == nil {
		return nil
	}

	// If no reference node, append at end
	if referenceNode == nil {
		return n.AppendChild(newNode)
	}

	// Find reference node index
	refIndex := -1
	for i, child := range n.childNodes {
		if child == referenceNode {
			refIndex = i
			break
		}
	}

	if refIndex == -1 {
		// Reference node not found, append at end
		return n.AppendChild(newNode)
	}

	// Remove from old parent if it has one
	if oldParent := newNode.ParentNode(); oldParent != nil {
		oldParent.RemoveChild(newNode)
	}

	// Insert at index
	n.childNodes = append(n.childNodes[:refIndex+1], n.childNodes[refIndex:]...)
	n.childNodes[refIndex] = newNode
	newNode.setParentNode(n)

	// Update sibling relationships
	if refIndex > 0 {
		prevNode := n.childNodes[refIndex-1]
		prevNode.setNextSibling(newNode)
		newNode.setPreviousSibling(prevNode)
	} else {
		newNode.setPreviousSibling(nil)
	}
	newNode.setNextSibling(referenceNode)
	referenceNode.setPreviousSibling(newNode)

	return newNode
}

func (n *baseNode) RemoveChild(child Node) Node {
	if child == nil {
		return nil
	}

	// Find child index
	childIndex := -1
	for i, c := range n.childNodes {
		if c == child {
			childIndex = i
			break
		}
	}

	if childIndex == -1 {
		return nil
	}

	// Update sibling relationships
	if childIndex > 0 {
		prevNode := n.childNodes[childIndex-1]
		prevNode.setNextSibling(child.NextSibling())
	}
	if childIndex < len(n.childNodes)-1 {
		nextNode := n.childNodes[childIndex+1]
		nextNode.setPreviousSibling(child.PreviousSibling())
	}

	// Remove from children
	n.childNodes = append(n.childNodes[:childIndex], n.childNodes[childIndex+1:]...)
	child.setParentNode(nil)
	child.setPreviousSibling(nil)
	child.setNextSibling(nil)

	return child
}

func (n *baseNode) ReplaceChild(newChild, oldChild Node) Node {
	if newChild == nil || oldChild == nil {
		return nil
	}

	// Find old child index
	oldIndex := -1
	for i, c := range n.childNodes {
		if c == oldChild {
			oldIndex = i
			break
		}
	}

	if oldIndex == -1 {
		return nil
	}

	// Remove from old parent if it has one
	if oldParent := newChild.ParentNode(); oldParent != nil {
		oldParent.RemoveChild(newChild)
	}

	// Replace
	n.childNodes[oldIndex] = newChild
	newChild.setParentNode(n)

	// Update sibling relationships
	newChild.setPreviousSibling(oldChild.PreviousSibling())
	newChild.setNextSibling(oldChild.NextSibling())
	if prev := oldChild.PreviousSibling(); prev != nil {
		prev.setNextSibling(newChild)
	}
	if next := oldChild.NextSibling(); next != nil {
		next.setPreviousSibling(newChild)
	}

	oldChild.setParentNode(nil)
	oldChild.setPreviousSibling(nil)
	oldChild.setNextSibling(nil)

	return oldChild
}

func (n *baseNode) setParentNode(parent Node) {
	n.parentNode = parent
}

func (n *baseNode) setPreviousSibling(node Node) {
	n.previousSibling = node
}

func (n *baseNode) setNextSibling(node Node) {
	n.nextSibling = node
}

// CloneNode provides a default implementation (should be overridden by concrete types)
func (n *baseNode) CloneNode(deep bool) Node {
	return nil // Must be implemented by concrete types
}

// TextContent provides a default implementation (should be overridden by concrete types)
func (n *baseNode) TextContent() string {
	return ""
}

// SetTextContent provides a default implementation (should be overridden by concrete types)
func (n *baseNode) SetTextContent(text string) {
}

// Render provides a default implementation (should be overridden by concrete types)
func (n *baseNode) Render(scr uv.Screen, area uv.Rectangle) {
}

// MinSize provides a default implementation (should be overridden by concrete types)
func (n *baseNode) MinSize(scr uv.Screen) (width, height int) {
	return 0, 0
}
