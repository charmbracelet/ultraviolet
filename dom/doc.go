// Package dom provides DOM-inspired primitives for building terminal user
// interfaces using Ultraviolet. The package follows a declarative approach
// similar to HTML and FTXUI, where UI components are composed as a tree of
// elements.
//
// Elements are the building blocks of a DOM-based TUI. They implement the
// Element interface and can be composed together to create complex layouts.
// The package provides containers (VBox, HBox), text elements (Text), and
// decorators (Border, Padding) among others.
//
// Example usage:
//
//	elem := dom.VBox(
//	    dom.Text("Hello, World!"),
//	    dom.HBox(
//	        dom.Text("Left"),
//	        dom.Separator(),
//	        dom.Text("Right"),
//	    ),
//	)
//	elem.Render(screen, area)
package dom
