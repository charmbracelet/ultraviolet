// Package dom provides a Document Object Model (DOM) implementation for building
// terminal user interfaces with Ultraviolet.
//
// This package follows the Web DOM API specification from MDN:
//   - https://developer.mozilla.org/en-US/docs/Web/API/Document_Object_Model
//   - https://developer.mozilla.org/en-US/docs/Web/API/Node
//   - https://developer.mozilla.org/en-US/docs/Web/API/Document
//   - https://developer.mozilla.org/en-US/docs/Web/API/Element
//   - https://developer.mozilla.org/en-US/docs/Web/API/Attr
//
// # Core Interfaces
//
// Node - Base interface for all nodes in the DOM tree
//   - Provides NodeType, NodeName, ParentNode, ChildNodes
//   - Methods: AppendChild, RemoveChild, ReplaceChild, etc.
//
// Document - Represents the entire document (root of DOM tree)
//   - Methods: CreateElement, CreateTextNode, GetElementsByTagName
//   - Properties: DocumentElement (root element)
//
// Element - Represents an element in the document
//   - Extends Node with element-specific functionality
//   - Properties: TagName, Attributes, Children
//   - Methods: GetAttribute, SetAttribute, GetElementsByTagName
//
// Text - Represents a text node
//   - Extends Node for text content
//   - Property: Data (the text content)
//
// Attr - Represents an attribute on an element
//   - Properties: Name, Value
//
// # Rendering
//
// Unlike Web DOM which is rendered by browsers, this DOM is explicitly rendered
// to an Ultraviolet screen using the Render method on the Document or any Element.
//
// # Example Usage
//
//	doc := dom.NewDocument()
//	
//	// Create elements
//	div := doc.CreateElement("div")
//	div.SetAttribute("border", "rounded")
//	
//	text := doc.CreateTextNode("Hello, World!")
//	div.AppendChild(text)
//	
//	doc.AppendChild(div)
//	
//	// Render to screen
//	doc.Render(screen, area)
//
// The package emphasizes following Web DOM standards to provide a familiar API
// for developers with web development experience.
package dom
