package doc

import (
	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Node represents a node in the document tree.
// Both Document and element nodes implement this interface.
type Node interface {
	// Type returns the type of the node (element, text, document, etc.)
	Type() html.NodeType

	// Data returns the node's data.
	// For element nodes, this is the tag name.
	// For text nodes, this is the text content.
	// For Document nodes, this is "document".
	Data() string

	// DataAtom returns the atom for the node's data, or zero if not a known tag.
	DataAtom() atom.Atom

	// Namespace returns the namespace of the node.
	Namespace() string

	// Attr returns the attributes of the node (for element nodes).
	Attr() []html.Attribute

	// Parent returns the parent node, or nil if this is the root (Document).
	Parent() Node

	// FirstChild returns the first child node, or nil if there are no children.
	FirstChild() Node

	// LastChild returns the last child node, or nil if there are no children.
	LastChild() Node

	// PrevSibling returns the previous sibling node, or nil if this is the first child.
	PrevSibling() Node

	// NextSibling returns the next sibling node, or nil if this is the last child.
	NextSibling() Node

	// Children returns all child nodes.
	Children() []Node

	// AddEventListener adds an event listener to this node for the specified event type.
	AddEventListener(eventType EventType, listener EventListener)

	// RemoveEventListener removes all event listeners for the specified event type from this node.
	RemoveEventListener(eventType EventType)

	// Focus makes this node the active element in its document.
	Focus()

	// Blur removes focus from this node if it is the active element.
	Blur()
}

// node represents a node in the document tree.
// The Document is just a node without a parent.
type node struct {
	n         *html.Node
	parent    Node
	listeners map[string][]EventListener
	nodeCache map[*html.Node]*node
	document  *Document // Backlink to containing document

	// Rendering state
	computedStyle *ComputedStyle
	layout        *LayoutBox
	lines         []uv.Line // Original text lines (for text nodes)
	wrappedLines  []uv.Line // Wrapped text lines (cached)
}

var _ Node = (*node)(nil)

// Type returns the type of the node.
func (n *node) Type() html.NodeType {
	if n.n == nil {
		return html.DocumentNode
	}
	return n.n.Type
}

// Data returns the node's data.
func (n *node) Data() string {
	if n.n == nil {
		return "document"
	}
	return n.n.Data
}

// DataAtom returns the atom for the node's data.
func (n *node) DataAtom() atom.Atom {
	if n.n == nil {
		return 0
	}
	return n.n.DataAtom
}

// Namespace returns the namespace of the node.
func (n *node) Namespace() string {
	if n.n == nil {
		return ""
	}
	return n.n.Namespace
}

// Attr returns the attributes of the node.
func (n *node) Attr() []html.Attribute {
	if n.n == nil {
		return nil
	}
	return n.n.Attr
}

// Parent returns the parent node.
func (n *node) Parent() Node {
	return n.parent
}

// FirstChild returns the first child node.
func (n *node) FirstChild() Node {
	if n.n == nil || n.n.FirstChild == nil {
		return nil
	}
	return n.wrapNode(n.n.FirstChild, n)
}

// LastChild returns the last child node.
func (n *node) LastChild() Node {
	if n.n == nil || n.n.LastChild == nil {
		return nil
	}
	return n.wrapNode(n.n.LastChild, n)
}

// PrevSibling returns the previous sibling node.
func (n *node) PrevSibling() Node {
	if n.n == nil || n.n.PrevSibling == nil {
		return nil
	}
	return n.wrapNode(n.n.PrevSibling, n.parent)
}

// NextSibling returns the next sibling node.
func (n *node) NextSibling() Node {
	if n.n == nil || n.n.NextSibling == nil {
		return nil
	}
	return n.wrapNode(n.n.NextSibling, n.parent)
}

// Children returns all child nodes.
func (n *node) Children() []Node {
	if n.n == nil {
		return nil
	}
	var children []Node
	for c := n.n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, n.wrapNode(c, n))
	}
	return children
}

// AddEventListener adds an event listener to this node.
func (n *node) AddEventListener(eventType EventType, listener EventListener) {
	if n.listeners == nil {
		n.listeners = make(map[string][]EventListener)
	}
	n.listeners[string(eventType)] = append(n.listeners[string(eventType)], listener)
}

// RemoveEventListener removes all event listeners for the specified event type.
func (n *node) RemoveEventListener(eventType EventType) {
	delete(n.listeners, string(eventType))
}

// Focus makes this node the active element in its document.
func (n *node) Focus() {
	// Walk up to find the root document
	root := n
	for root.parent != nil {
		if p, ok := root.parent.(*node); ok {
			root = p
		} else {
			break
		}
	}

	// Use the document backlink to set active element
	if root.document != nil {
		root.document.activeElement = n
	}
}

// Blur removes focus from this node if it is the active element.
func (n *node) Blur() {
	// Walk up to find the root document
	root := n
	for root.parent != nil {
		if p, ok := root.parent.(*node); ok {
			root = p
		} else {
			break
		}
	}

	// Clear active element if it's this node
	if root.document != nil && root.document.activeElement == n {
		root.document.activeElement = root
	}
}

// wrapNode wraps an html.Node, using cache to maintain identity.
func (n *node) wrapNode(htmlNode *html.Node, parent Node) *node {
	// Use the root's cache (walk up to find it)
	root := n
	for root.parent != nil {
		if p, ok := root.parent.(*node); ok {
			root = p
		} else {
			break
		}
	}

	if root.nodeCache == nil {
		root.nodeCache = make(map[*html.Node]*node)
	}

	if wrapped, ok := root.nodeCache[htmlNode]; ok {
		return wrapped
	}

	wrapped := &node{
		n:         htmlNode,
		parent:    parent,
		nodeCache: root.nodeCache, // Share the cache
		document:  root.document,  // Share the document backlink
	}
	root.nodeCache[htmlNode] = wrapped
	return wrapped
}
