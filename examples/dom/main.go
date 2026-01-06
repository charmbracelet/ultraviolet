package main

import (
	"fmt"

	"github.com/charmbracelet/ultraviolet/dom"
)

func main() {
	fmt.Println("DOM Package Example - Following Web API Standards")
	fmt.Println("==================================================\n")

	// Create a new document (root of the DOM tree)
	doc := dom.NewDocument()

	// Create a root div element
	root := doc.CreateElement("div")
	root.SetAttribute("border", "rounded")
	root.SetAttribute("padding", "1")

	// Create a title
	title := doc.CreateTextNode("🌟 DOM Example - Following Web API Standards 🌟")
	root.AppendChild(title)

	// Create a separator (empty div for spacing)
	separator := doc.CreateElement("div")
	separator.AppendChild(doc.CreateTextNode(""))
	root.AppendChild(separator)

	// Create a description paragraph
	descDiv := doc.CreateElement("div")
	desc := doc.CreateTextNode("This is a DOM implementation following MDN Web API standards.\nIt includes Node, Document, Element, Text, and Attr interfaces.")
	descDiv.AppendChild(desc)
	root.AppendChild(descDiv)

	// Add another separator
	root.AppendChild(doc.CreateElement("div"))

	// Create an HBox container for horizontal layout
	hbox := doc.CreateElement("hbox")

	// Left column
	leftBox := doc.CreateElement("div")
	leftBox.SetAttribute("border", "normal")
	leftBox.AppendChild(doc.CreateTextNode("Left Panel\n\n• Node interface\n• Element interface\n• Text nodes"))
	hbox.AppendChild(leftBox)

	// Right column
	rightBox := doc.CreateElement("div")
	rightBox.SetAttribute("border", "double")
	rightBox.AppendChild(doc.CreateTextNode("Right Panel\n\n• Document API\n• Attr management\n• DOM tree"))
	hbox.AppendChild(rightBox)

	root.AppendChild(hbox)

	// Add another separator
	root.AppendChild(doc.CreateElement("div"))

	// Create a footer
	footer := doc.CreateElement("div")
	footer.AppendChild(doc.CreateTextNode("DOM tree structure created successfully!"))
	root.AppendChild(footer)

	// Append root to document
	doc.AppendChild(root)

	// Demonstrate DOM API
	fmt.Println("✓ Created Document")
	fmt.Printf("✓ Document element: %s\n", doc.DocumentElement().NodeName())
	fmt.Printf("✓ Root has %d children\n", len(root.ChildNodes()))
	fmt.Printf("✓ Root has border attribute: %s\n", root.GetAttribute("border"))
	fmt.Printf("✓ Found %d div elements\n", len(doc.GetElementsByTagName("div")))
	fmt.Println("\n✓ DOM tree structure:")
	printTree(doc, 0)
	
	fmt.Println("\n✓ Text content:")
	fmt.Println(doc.TextContent())
	
	fmt.Println("\n✓ DOM package successfully demonstrates Web DOM API standards!")
}

func printTree(node dom.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	nodeName := node.NodeName()
	fmt.Printf("%s- %s", indent, nodeName)
	
	if elem, ok := node.(*dom.Element); ok {
		if elem.Attributes().Length() > 0 {
			fmt.Print(" [")
			for i := 0; i < elem.Attributes().Length(); i++ {
				attr := elem.Attributes().Item(i)
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s=\"%s\"", attr.Name, attr.Value)
			}
			fmt.Print("]")
		}
	} else if text, ok := node.(*dom.Text); ok {
		data := text.Data()
		if len(data) > 30 {
			data = data[:30] + "..."
		}
		// Show first line only for multiline text
		for idx := 0; idx < len(data); idx++ {
			if data[idx] == '\n' {
				data = data[:idx] + "..."
				break
			}
		}
		fmt.Printf(": \"%s\"", data)
	}
	fmt.Println()
	
	for _, child := range node.ChildNodes() {
		printTree(child, depth+1)
	}
}
