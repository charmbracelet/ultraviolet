package dom

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// Element represents an element in the DOM tree.
// See: https://developer.mozilla.org/en-US/docs/Web/API/Element
type Element struct {
	baseNode
	tagName    string
	attributes *NamedNodeMap
	style      *Style
}

// NewElement creates a new element with the given tag name.
func NewElement(tagName string) *Element {
	return &Element{
		baseNode: baseNode{
			nodeType:   ElementNode,
			nodeName:   strings.ToUpper(tagName),
			childNodes: make([]Node, 0),
		},
		tagName:    strings.ToUpper(tagName),
		attributes: NewNamedNodeMap(),
		style:      NewStyle(),
	}
}

// TagName returns the tag name of the element (uppercase).
func (e *Element) TagName() string {
	return e.tagName
}

// Attributes returns the NamedNodeMap of attributes.
func (e *Element) Attributes() *NamedNodeMap {
	return e.attributes
}

// GetAttribute returns the value of a named attribute, or empty string if not found.
func (e *Element) GetAttribute(name string) string {
	attr := e.attributes.GetNamedItem(name)
	if attr == nil {
		return ""
	}
	return attr.Value
}

// SetAttribute sets the value of a named attribute.
func (e *Element) SetAttribute(name, value string) {
	e.attributes.SetNamedItem(NewAttr(name, value))
}

// RemoveAttribute removes a named attribute.
func (e *Element) RemoveAttribute(name string) {
	e.attributes.RemoveNamedItem(name)
}

// HasAttribute returns true if the element has the named attribute.
func (e *Element) HasAttribute(name string) bool {
	return e.attributes.GetNamedItem(name) != nil
}

// AppendChild adds a node to the end of the list of children.
// Overrides baseNode.AppendChild to properly set parent.
func (e *Element) AppendChild(child Node) Node {
	result := e.baseNode.AppendChild(child)
	if result != nil {
		result.setParentNode(e)
	}
	return result
}

// InsertBefore inserts a node before a reference node as a child.
func (e *Element) InsertBefore(newNode, referenceNode Node) Node {
	result := e.baseNode.InsertBefore(newNode, referenceNode)
	if result != nil {
		result.setParentNode(e)
	}
	return result
}

// ReplaceChild replaces a child node with a new node.
func (e *Element) ReplaceChild(newChild, oldChild Node) Node {
	result := e.baseNode.ReplaceChild(newChild, oldChild)
	if newChild != nil {
		newChild.setParentNode(e)
	}
	return result
}

// GetElementsByTagName returns a list of descendant elements with the given tag name.
func (e *Element) GetElementsByTagName(tagName string) []*Element {
	tagName = strings.ToUpper(tagName)
	var results []*Element
	
	for _, child := range e.childNodes {
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

// Children returns only the child elements (not text nodes).
func (e *Element) Children() []*Element {
	var children []*Element
	for _, child := range e.childNodes {
		if elem, ok := child.(*Element); ok {
			children = append(children, elem)
		}
	}
	return children
}

// Style returns the inline style of the element.
func (e *Element) Style() *Style {
	return e.style
}

// TextContent returns the text content of the element and all descendants.
func (e *Element) TextContent() string {
	var builder strings.Builder
	e.collectText(&builder)
	return builder.String()
}

func (e *Element) collectText(builder *strings.Builder) {
	for _, child := range e.childNodes {
		if textNode, ok := child.(*Text); ok {
			builder.WriteString(textNode.data)
		} else if elem, ok := child.(*Element); ok {
			elem.collectText(builder)
		}
	}
}

// SetTextContent sets the text content, removing all children and creating a single text node.
func (e *Element) SetTextContent(text string) {
	// Remove all children
	e.childNodes = make([]Node, 0)
	
	// Add single text node
	if text != "" {
		textNode := NewText(text)
		e.AppendChild(textNode)
	}
}

// CloneNode creates a copy of the element.
// If deep is true, all descendants are also cloned.
func (e *Element) CloneNode(deep bool) Node {
	clone := &Element{
		baseNode: baseNode{
			nodeType:   ElementNode,
			nodeName:   e.nodeName,
			childNodes: make([]Node, 0),
		},
		tagName:    e.tagName,
		attributes: e.attributes.Clone(),
		style:      e.style.Clone(),
	}
	
	if deep {
		for _, child := range e.childNodes {
			clonedChild := child.CloneNode(true)
			clone.AppendChild(clonedChild)
		}
	}
	
	return clone
}

// Render renders the element and its children to the screen.
func (e *Element) Render(scr uv.Screen, area uv.Rectangle) {
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return
	}

	// Render based on tag name
	switch e.tagName {
	case "DIV":
		e.renderDiv(scr, area)
	case "VBOX":
		e.renderVBox(scr, area)
	case "HBOX":
		e.renderHBox(scr, area)
	default:
		// Default: render children in vertical layout
		e.renderVBox(scr, area)
	}
}

// renderDiv renders a div element (block container with optional border/padding).
func (e *Element) renderDiv(scr uv.Screen, area uv.Rectangle) {
	contentArea := area

	// Apply border if specified
	if e.HasAttribute("border") {
		e.renderBorder(scr, area)
		contentArea = uv.Rect(
			area.Min.X+1,
			area.Min.Y+1,
			max(0, area.Dx()-2),
			max(0, area.Dy()-2),
		)
	}

	// Apply padding if specified
	if e.HasAttribute("padding") {
		// Simple padding (could parse value for different amounts)
		contentArea = uv.Rect(
			contentArea.Min.X+1,
			contentArea.Min.Y+1,
			max(0, contentArea.Dx()-2),
			max(0, contentArea.Dy()-2),
		)
	}

	// Render children vertically by default
	y := contentArea.Min.Y
	for _, child := range e.childNodes {
		if y >= contentArea.Max.Y {
			break
		}

		_, childHeight := child.MinSize(scr)
		if childHeight == 0 {
			childHeight = contentArea.Max.Y - y
		}

		childArea := uv.Rect(
			contentArea.Min.X,
			y,
			contentArea.Dx(),
			min(childHeight, contentArea.Max.Y-y),
		)

		child.Render(scr, childArea)
		y += childHeight
	}
}

// renderVBox renders children vertically (flexbox column).
func (e *Element) renderVBox(scr uv.Screen, area uv.Rectangle) {
	if len(e.childNodes) == 0 {
		return
	}

	// Calculate total fixed height and count flexible children
	totalFixed := 0
	flexCount := 0
	
	for _, child := range e.childNodes {
		_, h := child.MinSize(scr)
		if h == 0 {
			flexCount++
		} else {
			totalFixed += h
		}
	}

	// Calculate flexible height
	remaining := area.Dy() - totalFixed
	flexHeight := 0
	if flexCount > 0 && remaining > 0 {
		flexHeight = remaining / flexCount
	}

	// Render children
	y := area.Min.Y
	for _, child := range e.childNodes {
		if y >= area.Max.Y {
			break
		}

		_, h := child.MinSize(scr)
		if h == 0 {
			h = flexHeight
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

// renderHBox renders children horizontally (flexbox row).
func (e *Element) renderHBox(scr uv.Screen, area uv.Rectangle) {
	if len(e.childNodes) == 0 {
		return
	}

	// Calculate total fixed width and count flexible children
	totalFixed := 0
	flexCount := 0
	
	for _, child := range e.childNodes {
		w, _ := child.MinSize(scr)
		if w == 0 {
			flexCount++
		} else {
			totalFixed += w
		}
	}

	// Calculate flexible width
	remaining := area.Dx() - totalFixed
	flexWidth := 0
	if flexCount > 0 && remaining > 0 {
		flexWidth = remaining / flexCount
	}

	// Render children
	x := area.Min.X
	for _, child := range e.childNodes {
		if x >= area.Max.X {
			break
		}

		w, _ := child.MinSize(scr)
		if w == 0 {
			w = flexWidth
		}

		childArea := uv.Rect(
			x,
			area.Min.Y,
			min(w, area.Max.X-x),
			area.Dy(),
		)

		child.Render(scr, childArea)
		x += w
	}
}

// renderBorder renders a border around the element.
func (e *Element) renderBorder(scr uv.Screen, area uv.Rectangle) {
	borderStyle := e.GetAttribute("border")
	
	var top, bottom, left, right, tl, tr, bl, br string
	
	switch borderStyle {
	case "rounded":
		top, bottom, left, right = "─", "─", "│", "│"
		tl, tr, bl, br = "╭", "╮", "╰", "╯"
	case "double":
		top, bottom, left, right = "═", "═", "║", "║"
		tl, tr, bl, br = "╔", "╗", "╚", "╝"
	case "thick":
		top, bottom, left, right = "━", "━", "┃", "┃"
		tl, tr, bl, br = "┏", "┓", "┗", "┛"
	default: // "normal" or any other value
		top, bottom, left, right = "─", "─", "│", "│"
		tl, tr, bl, br = "┌", "┐", "└", "┘"
	}

	style := uv.Style{
		Fg: e.style.Foreground,
		Bg: e.style.Background,
	}

	// Corners
	cell := uv.NewCell(scr.WidthMethod(), tl)
	cell.Style = style
	scr.SetCell(area.Min.X, area.Min.Y, cell)
	
	cell = uv.NewCell(scr.WidthMethod(), tr)
	cell.Style = style
	scr.SetCell(area.Max.X-1, area.Min.Y, cell)
	
	cell = uv.NewCell(scr.WidthMethod(), bl)
	cell.Style = style
	scr.SetCell(area.Min.X, area.Max.Y-1, cell)
	
	cell = uv.NewCell(scr.WidthMethod(), br)
	cell.Style = style
	scr.SetCell(area.Max.X-1, area.Max.Y-1, cell)

	// Top and bottom edges
	for x := area.Min.X + 1; x < area.Max.X-1; x++ {
		cell = uv.NewCell(scr.WidthMethod(), top)
		cell.Style = style
		scr.SetCell(x, area.Min.Y, cell)
		
		cell = uv.NewCell(scr.WidthMethod(), bottom)
		cell.Style = style
		scr.SetCell(x, area.Max.Y-1, cell)
	}

	// Left and right edges
	for y := area.Min.Y + 1; y < area.Max.Y-1; y++ {
		cell = uv.NewCell(scr.WidthMethod(), left)
		cell.Style = style
		scr.SetCell(area.Min.X, y, cell)
		
		cell = uv.NewCell(scr.WidthMethod(), right)
		cell.Style = style
		scr.SetCell(area.Max.X-1, y, cell)
	}
}

// MinSize returns the minimum size needed to render the element.
func (e *Element) MinSize(scr uv.Screen) (width, height int) {
	// Calculate minimum size based on children
	switch e.tagName {
	case "HBOX":
		// Horizontal: sum widths, max height
		for _, child := range e.childNodes {
			w, h := child.MinSize(scr)
			width += w
			if h > height {
				height = h
			}
		}
	default:
		// Vertical (DIV, VBOX, or other): max width, sum heights
		for _, child := range e.childNodes {
			w, h := child.MinSize(scr)
			if w > width {
				width = w
			}
			height += h
		}
	}

	// Add border
	if e.HasAttribute("border") {
		width += 2
		height += 2
	}

	// Add padding
	if e.HasAttribute("padding") {
		width += 2
		height += 2
	}

	return width, height
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
