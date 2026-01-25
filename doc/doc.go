// Package doc provides HTML document parsing and DOM representation for TUI applications.
//
// The document package parses HTML using Go's standard library and creates a DOM structure
// that can be used with Ultraviolet to build terminal user interfaces.
package doc

import (
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/net/html"
)

// Document represents a DOM structure for TUI applications.
// It's just a wrapper around the root node.
type Document struct {
	*node
	opts          Options
	activeElement Node
	renderer      *Renderer
	dirty         bool
}

var _ Node = (*Document)(nil)

// Options configures document behavior.
type Options struct {
	// BaseURL is the base URL for resolving relative URLs in the document
	BaseURL string
	// Stylesheets contains CSS stylesheets to apply to the document
	Stylesheets []string
	// Terminal is the terminal to use for rendering. If nil, DefaultTerminal is used.
	Terminal *uv.Terminal
}

// NewDocument creates a new Document from an HTML node and optional configuration.
// If opts is nil, default options are used.
func NewDocument(htmlRoot *html.Node, opts *Options) *Document {
	// Create the document node (a node without a parent)
	root := &node{
		n:         htmlRoot,
		parent:    nil, // Document has no parent
		nodeCache: make(map[*html.Node]*node),
	}

	doc := &Document{
		node:     root,
		renderer: NewRenderer(root),
		dirty:    true, // Initial render needed
	}

	// Set the document backlink in the root node
	root.document = doc

	if opts != nil {
		doc.opts = *opts
	}

	// Set the document root as the active node initially
	doc.activeElement = root

	return doc
}

// Invalidate marks the document as needing re-render.
func (d *Document) Invalidate() {
	d.dirty = true
	if d.node != nil && d.node.layout != nil {
		d.node.layout.Invalidate()
	}
}

// ActiveElement returns the currently active/focused node.
func (d *Document) ActiveElement() Node {
	if d.activeElement == nil {
		return d.node
	}
	return d.activeElement
}

// Parse parses HTML from a reader and creates a new Document.
// It returns an error if the HTML cannot be parsed.
func Parse(r io.Reader, opts *Options) (*Document, error) {
	htmlNode, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return NewDocument(htmlNode, opts), nil
}

// ParseFragment parses an HTML fragment and creates a new Document.
// It returns an error if the HTML cannot be parsed.
func ParseFragment(r io.Reader, context *html.Node, opts *Options) (*Document, error) {
	nodes, err := html.ParseFragment(r, context)
	if err != nil {
		return nil, err
	}

	// Create a container node for the fragments
	container := &html.Node{
		Type: html.ElementNode,
		Data: "div",
	}

	for _, n := range nodes {
		container.AppendChild(n)
	}

	return NewDocument(container, opts), nil
}

// GetElementByID returns the first element with the specified id attribute.
// Returns nil if no element is found.
func (d *Document) GetElementByID(id string) Node {
	return d.getElementByID(d.node.n, id)
}

// GetElementsByTagName returns all elements with the specified tag name.
func (d *Document) GetElementsByTagName(tagName string) []Node {
	var result []Node
	d.getElementsByTagName(d.node.n, tagName, &result)
	return result
}

// GetElementsByClassName returns all elements with the specified class name.
func (d *Document) GetElementsByClassName(className string) []Node {
	var result []Node
	d.getElementsByClassName(d.node.n, className, &result)
	return result
}

// QuerySelector returns the first element that matches the specified selector.
// Currently supports: tag names, #id, .class
// Returns nil if no element is found.
func (d *Document) QuerySelector(selector string) Node {
	return d.querySelector(d.node.n, selector)
}

// QuerySelectorAll returns all elements that match the specified selector.
// Currently supports: tag names, #id, .class
func (d *Document) QuerySelectorAll(selector string) []Node {
	var result []Node
	d.querySelectorAll(d.node.n, selector, &result)
	return result
}

// Helper functions

func (d *Document) getElementByID(n *html.Node, id string) Node {
	if n == nil {
		return nil
	}

	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				return d.node.wrapNode(n, d.findParentNode(n))
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := d.getElementByID(c, id); found != nil {
			return found
		}
	}

	return nil
}

func (d *Document) getElementsByTagName(n *html.Node, tagName string, result *[]Node) {
	if n == nil {
		return
	}

	if n.Type == html.ElementNode && n.Data == tagName {
		*result = append(*result, d.node.wrapNode(n, d.findParentNode(n)))
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		d.getElementsByTagName(c, tagName, result)
	}
}

func (d *Document) getElementsByClassName(n *html.Node, className string, result *[]Node) {
	if n == nil {
		return
	}

	if n.Type == html.ElementNode {
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				// Check if className is in the class list
				classes := strings.Fields(attr.Val)
				for _, class := range classes {
					if class == className {
						*result = append(*result, d.node.wrapNode(n, d.findParentNode(n)))
						break
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		d.getElementsByClassName(c, className, result)
	}
}

func (d *Document) querySelector(n *html.Node, selector string) Node {
	if n == nil {
		return nil
	}

	// Parse selector
	if strings.HasPrefix(selector, "#") {
		// ID selector
		id := strings.TrimPrefix(selector, "#")
		return d.getElementByID(n, id)
	} else if strings.HasPrefix(selector, ".") {
		// Class selector
		className := strings.TrimPrefix(selector, ".")
		var result []Node
		d.getElementsByClassName(n, className, &result)
		if len(result) > 0 {
			return result[0]
		}
		return nil
	}

	// Tag name selector
	var result []Node
	d.getElementsByTagName(n, selector, &result)
	if len(result) > 0 {
		return result[0]
	}
	return nil
}

func (d *Document) querySelectorAll(n *html.Node, selector string, result *[]Node) {
	if n == nil {
		return
	}

	// Parse selector
	if strings.HasPrefix(selector, "#") {
		// ID selector
		id := strings.TrimPrefix(selector, "#")
		if found := d.getElementByID(n, id); found != nil {
			*result = append(*result, found)
		}
	} else if strings.HasPrefix(selector, ".") {
		// Class selector
		className := strings.TrimPrefix(selector, ".")
		d.getElementsByClassName(n, className, result)
	} else {
		// Tag name selector
		d.getElementsByTagName(n, selector, result)
	}
}

// findParentNode finds the parent node for a given html.Node.
func (d *Document) findParentNode(n *html.Node) Node {
	if n.Parent == nil {
		return d.node
	}
	// Check if already wrapped
	if wrapped, ok := d.node.nodeCache[n.Parent]; ok {
		return wrapped
	}
	// Wrap it
	return d.node.wrapNode(n.Parent, d.findParentNode(n.Parent))
}

// Serve starts the event loop and renders the document.
// It puts the terminal in raw mode, enters altscreen, and listens to events.
// The function blocks until an error occurs or the document is closed.
func (d *Document) Serve() error {
	// Get the terminal instance
	t := d.opts.Terminal
	if t == nil {
		t = uv.DefaultTerminal()
	}

	// Create window size notifier using signal-based approach
	winch := make(chan os.Signal, 16)
	uv.NotifyWinch(winch)
	defer signal.Stop(winch)

	// Create terminal screen
	scr := uv.NewTerminalScreen(t.Writer(), t.Environ())

	// Create terminal events reader
	evs := uv.NewTerminalEvents(t.Reader())
	defer evs.Close()

	// Cleanup on exit
	defer t.Close()
	defer t.Restore()

	// Put terminal in raw mode
	if _, err := t.MakeRaw(); err != nil {
		return err
	}

	// Get initial terminal size
	w, h, err := t.GetSize()
	if err != nil {
		return err
	}

	// Resize the screen to match terminal size
	scr.Resize(w, h)

	// Enter alternate screen mode
	scr.EnterAltScreen()
	scr.HideCursor()

	// Event loop
	const pollDuration = 10 * time.Millisecond

	for {
		// Poll for terminal events
		ready, err := evs.PollEvents(pollDuration)
		if err != nil {
			return err
		}

		// Check for window resize
		select {
		case <-winch:
			// Drain the channel
			for len(winch) > 0 {
				<-winch
			}
			// Get new terminal size
			w, h, err := t.GetSize()
			if err != nil {
				return err
			}
			// Resize the screen
			scr.Resize(w, h)
			// Mark document as dirty for re-layout
			d.Invalidate()
		default:
			// No resize event
		}

		// Process terminal events if ready
		if ready {
			events, err := evs.ReadEvents(false)
			if err != nil {
				return err
			}

			// Dispatch events to listeners with bubbling
			for _, ev := range events {
				handled := d.dispatchEvent(ev)

				// If no listener handled it, perform default action
				if !handled {
					switch ev.(type) {
					case uv.KeyPressEvent:
						// Default: exit on any key press
						scr.Render()
						scr.Reset()
						scr.Flush()
						return nil
					}
				}
			}
		}

		// Render the screen if dirty
		if d.dirty {
			viewport := scr.Bounds()
			d.renderer.Render(scr, viewport)
			d.dirty = false
		}
		
		scr.Render()
		scr.Flush()
	}
}
