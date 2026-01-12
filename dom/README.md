# DOM Package

A Document Object Model (DOM) implementation for building terminal user interfaces with Ultraviolet, following Web DOM API standards from MDN.

## Overview

This package implements the core concepts of the Web DOM API, providing familiar interfaces for developers with web development experience:

- **Node** - Base interface for all nodes in the DOM tree
- **Document** - Root of the DOM tree with methods to create elements and text nodes
- **Element** - Represents elements with attributes, children, and tag-based rendering
- **Text** - Text nodes containing text content
- **Attr** - Attributes on elements

See: https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model

## Core Interfaces

### Node

The primary data type for the entire DOM. All nodes implement this interface:

- `NodeType()` - Returns the type (Element, Text, Document)
- `NodeName()` - Returns the node name
- `ParentNode()` / `ChildNodes()` - Tree navigation
- `AppendChild()` / `RemoveChild()` / `ReplaceChild()` - Tree manipulation
- `CloneNode()` - Clones the node
- `TextContent()` / `SetTextContent()` - Text content access
- `Render()` / `MinSize()` - TUI-specific rendering methods

### Document

Represents the entire document and provides factory methods:

```go
doc := dom.NewDocument()
elem := doc.CreateElement("div")
text := doc.CreateTextNode("Hello, World!")
doc.AppendChild(elem)
```

### Element

Represents elements in the document tree:

```go
elem := doc.CreateElement("div")
elem.SetAttribute("border", "rounded")
elem.SetAttribute("padding", "1")
elem.AppendChild(text)
```

**Tag Names:**
- `div` - Block container with optional border/padding
- `vbox` - Vertical flexbox layout (stacks children vertically)
- `hbox` - Horizontal flexbox layout (arranges children horizontally)

**Attributes:**
- `border` - Border style: "normal", "rounded", "double", "thick"
- `padding` - Adds inner spacing (simple padding for now)

### Text

Represents text nodes:

```go
text := doc.CreateTextNode("Hello, World!")
text.SetData("Updated text")
```

Text nodes automatically:
- Handle newlines (renders each line separately)
- Use Unicode-aware width calculation with displaywidth
- Preserve grapheme clusters

### Attr

Represents attributes on elements. Attributes are managed through Element methods:

```go
elem.SetAttribute("border", "rounded")
value := elem.GetAttribute("border")  // returns "rounded"
hasIt := elem.HasAttribute("border")  // returns true
elem.RemoveAttribute("border")
```

## Example Usage

```go
package main

import (
    "github.com/charmbracelet/ultraviolet"
    "github.com/charmbracelet/ultraviolet/dom"
)

func main() {
    // Create document
    doc := dom.NewDocument()
    
    // Create root element with border
    root := doc.CreateElement("div")
    root.SetAttribute("border", "rounded")
    root.SetAttribute("padding", "1")
    
    // Add title
    title := doc.CreateTextNode("🌟 My Application 🌟")
    root.AppendChild(title)
    
    // Create horizontal layout
    hbox := doc.CreateElement("hbox")
    
    // Left panel
    left := doc.CreateElement("div")
    left.SetAttribute("border", "normal")
    left.AppendChild(doc.CreateTextNode("Left Panel"))
    hbox.AppendChild(left)
    
    // Right panel
    right := doc.CreateElement("div")
    right.SetAttribute("border", "double")
    right.AppendChild(doc.CreateTextNode("Right Panel"))
    hbox.AppendChild(right)
    
    root.AppendChild(hbox)
    doc.AppendChild(root)
    
    // Render to screen
    scr := ultraviolet.NewScreen()
    scr.Init()
    defer scr.Fini()
    
    width, height := scr.Size()
    area := ultraviolet.Rect(0, 0, width, height)
    
    scr.Clear()
    doc.Render(scr, area)
    scr.Show()
}
```

## Layout System

### VBox (Vertical Flexbox)

Stacks children vertically. Children with `MinSize` returning 0 height are flexible and share remaining space.

```go
vbox := doc.CreateElement("vbox")
vbox.AppendChild(doc.CreateTextNode("Line 1"))
vbox.AppendChild(doc.CreateTextNode("Line 2"))
```

### HBox (Horizontal Flexbox)

Arranges children horizontally. Children with `MinSize` returning 0 width are flexible and share remaining space.

```go
hbox := doc.CreateElement("hbox")
hbox.AppendChild(doc.CreateTextNode("Left"))
hbox.AppendChild(doc.CreateTextNode("Right"))
```

## Border Styles

Pre-defined border styles available via the `border` attribute:

- `normal` - Standard box-drawing characters (┌─┐│└─┘)
- `rounded` - Rounded corners (╭─╮│╰─╯)
- `double` - Double lines (╔═╗║╚═╝)
- `thick` - Thick lines (┏━┓┃┗━┛)

## DOM Tree Manipulation

The DOM API provides standard methods for tree manipulation:

```go
// Navigate tree
parent := node.ParentNode()
children := parent.ChildNodes()
first := parent.FirstChild()
next := node.NextSibling()

// Modify tree
parent.AppendChild(child)
parent.InsertBefore(newChild, referenceChild)
parent.RemoveChild(child)
parent.ReplaceChild(newChild, oldChild)

// Clone nodes
clone := elem.CloneNode(false)  // shallow
deepClone := elem.CloneNode(true)  // deep

// Query tree
elements := elem.GetElementsByTagName("div")
text := elem.TextContent()
elem.SetTextContent("New text")
```

## Differences from Web DOM

While this package follows Web DOM concepts, there are some differences:

1. **Rendering is explicit** - Call `Render(screen, area)` to draw the DOM
2. **MinSize method** - Elements report minimum required dimensions
3. **Limited tag names** - Only div, vbox, hbox are supported
4. **Simple attributes** - Attributes are string key-value pairs
5. **No CSS** - Styling is done through attributes directly

## Testing

The package includes comprehensive tests covering:

- Document and element creation
- Tree manipulation (append, remove, replace)
- Attribute management
- Text content access
- Node cloning
- Rendering with borders and layouts

Run tests with:
```bash
go test github.com/charmbracelet/ultraviolet/dom
```

## Future Enhancements

Following the requirements, future versions will support:

- **Dynamic border/padding/margin sizes** - Not always 2, depends on content
- **Per-side styling** - Set border/padding/margin for individual sides (top, right, bottom, left)
- **Custom border styles** - Define custom border characters beyond pre-defined ones
- **Scrolling** - Overflow handling for content larger than viewport
- **Highlighting** - Selection and focus states for interactive elements
- **Text wrapping** - Word wrap and hard wrap modes for text content
- **More attributes** - Additional styling and behavioral options

## References

- [MDN: Document Object Model](https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model)
- [MDN: Node](https://developer.mozilla.org/en-US/docs/Web/API/Node)
- [MDN: Document](https://developer.mozilla.org/en-US/docs/Web/API/Document)
- [MDN: Element](https://developer.mozilla.org/en-US/docs/Web/API/Element)
- [MDN: Attr](https://developer.mozilla.org/en-US/docs/Web/API/Attr)
