// Package dom provides a Document Object Model (DOM) implementation for building
// terminal user interfaces with Ultraviolet.
//
// This package follows the core concepts of the Web DOM (https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model):
//   - Elements form a tree structure (like DOM nodes)
//   - Text nodes for inline content (analogous to DOM Text)
//   - Block-level container elements (analogous to HTMLDivElement)
//   - Declarative composition (like building HTML)
//   - CSS box model (content, padding, border)
//
// # Core Element Types
//
// TextNode - Inline text content (like DOM Text nodes and HTML <span>)
//   - Flows inline, wraps only on explicit newlines
//   - Example: Text("Hello, World!")
//
// Box - Block-level container (like DOM HTMLDivElement and HTML <div>)
//   - Follows CSS box model with padding, borders, scrolling
//   - Can contain any child elements
//   - Example: NewBox(content).WithBorder(BorderStyleRounded())
//
// # Layout Containers
//
// VBox - Vertical flex container (like CSS flexbox with flex-direction: column)
// HBox - Horizontal flex container (like CSS flexbox with flex-direction: row)
//
// # Example Usage
//
//	ui := dom.VBox(
//	    dom.Text("Header"),
//	    dom.NewBox(
//	        dom.HBox(
//	            dom.Text("Left"),
//	            dom.Text("Right"),
//	        ),
//	    ).WithBorder(dom.BorderStyleRounded()),
//	)
//	ui.Render(screen, area)
//
// The package emphasizes simplicity and follows web standards where applicable,
// making it familiar to developers with HTML/CSS/DOM experience.
package dom
