package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// Document represents the entire document - the root of the DOM tree.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Document
type Document struct {
	baseNode
	documentElement *Element
}

// NewDocument creates a new document.
func NewDocument() *Document {
	doc := &Document{
		baseNode: baseNode{
			nodeType:   DocumentNode,
			nodeName:   "#document",
			childNodes: make([]Node, 0),
		},
	}
	return doc
}

// DocumentElement returns the root element of the document.
// This is typically the top-level element (like <html> in HTML).
func (d *Document) DocumentElement() *Element {
	return d.documentElement
}

// CreateElement creates a new element with the given tag name.
func (d *Document) CreateElement(tagName string) *Element {
	return NewElement(tagName)
}

// CreateTextNode creates a new text node with the given data.
func (d *Document) CreateTextNode(data string) *Text {
	return NewText(data)
}

// GetElementsByTagName returns a list of elements with the given tag name.
func (d *Document) GetElementsByTagName(tagName string) []*Element {
	tagName = strings.ToUpper(tagName)
	var results []*Element
	
	for _, child := range d.childNodes {
		if elem, ok := child.(*Element); ok {
			if elem.tagName == tagName {
				results = append(results, elem)
			}
			// Recursively search children
			results = append(results, elem.GetElementsByTagName(tagName)...)
		}
	}
	
	return results
}

// AppendChild adds a node to the document.
// If the node is an element and it's the first element, it becomes the documentElement.
// Overrides baseNode.AppendChild to properly set parent.
func (d *Document) AppendChild(child Node) Node {
	result := d.baseNode.AppendChild(child)
	if result != nil {
		result.setParentNode(d)
	}
	
	// Set documentElement if this is the first element
	if d.documentElement == nil {
		if elem, ok := child.(*Element); ok {
			d.documentElement = elem
		}
	}
	
	return result
}

// InsertBefore inserts a node before a reference node as a child.
func (d *Document) InsertBefore(newNode, referenceNode Node) Node {
	result := d.baseNode.InsertBefore(newNode, referenceNode)
	if result != nil {
		result.setParentNode(d)
	}
	return result
}

// ReplaceChild replaces a child node with a new node.
func (d *Document) ReplaceChild(newChild, oldChild Node) Node {
	result := d.baseNode.ReplaceChild(newChild, oldChild)
	if newChild != nil {
		newChild.setParentNode(d)
	}
	return result
}

// RemoveChild removes a child node from the document.
// If the removed node is the documentElement, it's set to nil.
func (d *Document) RemoveChild(child Node) Node {
	result := d.baseNode.RemoveChild(child)
	
	// Clear documentElement if it was removed
	if elem, ok := child.(*Element); ok && elem == d.documentElement {
		d.documentElement = nil
	}
	
	return result
}

// TextContent returns the text content of the entire document.
func (d *Document) TextContent() string {
	var builder strings.Builder
	for _, child := range d.childNodes {
		builder.WriteString(child.TextContent())
	}
	return builder.String()
}

// SetTextContent sets the text content, removing all children.
func (d *Document) SetTextContent(text string) {
	d.childNodes = make([]Node, 0)
	d.documentElement = nil
	
	if text != "" {
		textNode := NewText(text)
		d.AppendChild(textNode)
	}
}

// CloneNode creates a copy of the document.
// If deep is true, all descendants are also cloned.
func (d *Document) CloneNode(deep bool) Node {
	clone := &Document{
		baseNode: baseNode{
			nodeType:   DocumentNode,
			nodeName:   "#document",
			childNodes: make([]Node, 0),
		},
	}
	
	if deep {
		for _, child := range d.childNodes {
			clonedChild := child.CloneNode(true)
			clone.AppendChild(clonedChild)
		}
	}
	
	return clone
}

// Render renders the document to the screen.
func (d *Document) Render(scr uv.Screen, area uv.Rectangle) {
	if d.documentElement != nil {
		d.documentElement.Render(scr, area)
	} else {
		// Render all children if no documentElement
		y := area.Min.Y
		for _, child := range d.childNodes {
			if y >= area.Max.Y {
				break
			}

			_, h := child.MinSize(scr)
			if h == 0 {
				h = area.Max.Y - y
			}

			childArea := uv.Rect(
				area.Min.X,
				y,
				area.Dx(),
				min(h, area.Max.Y-y),
			)

			child.Render(scr, childArea)
			y += h
		}
	}
}

// MinSize returns the minimum size needed to render the document.
func (d *Document) MinSize(scr uv.Screen) (width, height int) {
	if d.documentElement != nil {
		return d.documentElement.MinSize(scr)
	}
	
	// Otherwise, calculate based on all children
	for _, child := range d.childNodes {
		w, h := child.MinSize(scr)
		if w > width {
			width = w
		}
		height += h
	}
	
	return width, height
}
