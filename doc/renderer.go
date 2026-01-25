package doc

import (
	"image/color"
	"strings"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/clipperhouse/displaywidth"
	"golang.org/x/net/html"
)

// Renderer handles rendering the DOM tree to a screen.
type Renderer struct {
	root     *node
	viewport uv.Rectangle
}

// NewRenderer creates a new renderer for the given document root.
func NewRenderer(root *node) *Renderer {
	return &Renderer{
		root: root,
	}
}

// Render renders the document to the screen.
func (r *Renderer) Render(scr uv.Screen, viewport uv.Rectangle) {
	r.viewport = viewport

	// Compute styles if needed
	r.computeStyles(r.root)

	// Layout the document
	r.layout(r.root, viewport)

	// Paint the document
	r.paint(scr, r.root, uv.Pos(0, 0))
}

// computeStyles recursively computes styles for all nodes.
func (r *Renderer) computeStyles(n *node) {
	if n == nil {
		return
	}

	// Compute style for this node if not cached
	if n.computedStyle == nil {
		n.computedStyle = n.computeStyle()
	}

	// Recursively compute for children
	for _, child := range n.Children() {
		if childNode, ok := child.(*node); ok {
			r.computeStyles(childNode)
		}
	}
}

// layout calculates positions and sizes for all visible nodes.
func (r *Renderer) layout(n *node, availableRect uv.Rectangle) uv.Rectangle {
	if n == nil || n.computedStyle == nil {
		return uv.Rectangle{}
	}

	// Skip if display:none
	if n.computedStyle.Display == DisplayNone {
		return uv.Rectangle{}
	}

	// Initialize layout box if needed
	if n.layout == nil {
		n.layout = NewLayoutBox(availableRect)
	}

	// Skip layout if not dirty and available space hasn't changed
	if !n.layout.Dirty && n.layout.Rect == availableRect {
		return n.layout.Rect
	}

	// Handle text nodes specially
	if n.Type() == html.TextNode {
		return r.layoutText(n, availableRect)
	}

	// Block layout: stack children vertically
	if n.computedStyle.Display == DisplayBlock {
		return r.layoutBlock(n, availableRect)
	}

	// For now, treat everything else as block
	return r.layoutBlock(n, availableRect)
}

// layoutText handles layout for text nodes.
func (r *Renderer) layoutText(n *node, availableRect uv.Rectangle) uv.Rectangle {
	if n.Type() != html.TextNode {
		return uv.Rectangle{}
	}

	text := n.Data()
	text = strings.TrimSpace(text)

	if text == "" {
		n.layout = NewLayoutBox(uv.Rect(availableRect.Min.X, availableRect.Min.Y, 0, 0))
		n.layout.Dirty = false
		return n.layout.Rect
	}

	// Render text to lines with styling and link information
	n.lines = r.renderTextToLines(text, n)

	// Wrap lines to available width
	width := availableRect.Dx()
	if width <= 0 {
		width = 80 // Default width
	}

	n.wrappedLines = r.wrapLines(n.lines, width)

	// Calculate size: width of longest line, height = number of lines
	textWidth := width
	textHeight := len(n.wrappedLines)

	n.layout = NewLayoutBox(uv.Rect(
		availableRect.Min.X,
		availableRect.Min.Y,
		textWidth,
		textHeight,
	))
	n.layout.Dirty = false

	return n.layout.Rect
}

// layoutBlock handles block-level layout (vertical stacking).
func (r *Renderer) layoutBlock(n *node, availableRect uv.Rectangle) uv.Rectangle {
	x := availableRect.Min.X
	y := availableRect.Min.Y
	width := availableRect.Dx()
	currentY := y
	maxHeight := availableRect.Dy()

	// Apply width from style if set
	if n.computedStyle.Width > 0 {
		width = n.computedStyle.Width
	}

	// Layout children
	for _, child := range n.Children() {
		childNode, ok := child.(*node)
		if !ok {
			continue
		}

		// Check if we're exceeding viewport - stop laying out if so
		if currentY-y >= maxHeight && maxHeight > 0 {
			break
		}

		// Available space for this child
		childRect := uv.Rect(x, currentY, width, maxHeight-(currentY-y))

		// Layout the child
		layoutRect := r.layout(childNode, childRect)

		// Move down for next child
		if layoutRect.Dy() > 0 {
			currentY += layoutRect.Dy()
		}
	}

	// Calculate our total height
	totalHeight := currentY - y
	if n.computedStyle.Height > 0 {
		totalHeight = n.computedStyle.Height
	}

	n.layout = NewLayoutBox(uv.Rect(x, y, width, totalHeight))
	n.layout.Dirty = false

	return n.layout.Rect
}

// paint renders the node and its children to the screen.
func (r *Renderer) paint(scr uv.Screen, n *node, offset uv.Position) {
	if n == nil || n.layout == nil || n.computedStyle == nil {
		return
	}

	// Skip if display:none
	if n.computedStyle.Display == DisplayNone {
		return
	}

	// Calculate absolute position with scroll offset
	absRect := n.layout.Rect.Add(offset)

	// Skip if not visible in viewport
	if !absRect.Overlaps(r.viewport) {
		return
	}

	// Paint background if set
	if n.computedStyle.BackgroundColor != nil {
		r.paintBackground(scr, absRect, n.computedStyle.BackgroundColor)
	}

	// Paint text content
	if n.Type() == html.TextNode && len(n.wrappedLines) > 0 {
		r.paintText(scr, absRect, n.wrappedLines)
	}

	// Paint children
	for _, child := range n.Children() {
		if childNode, ok := child.(*node); ok {
			r.paint(scr, childNode, offset)
		}
	}
}

// paintBackground fills the area with the background color.
func (r *Renderer) paintBackground(scr uv.Screen, rect uv.Rectangle, bgColor color.Color) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		if y < r.viewport.Min.Y || y >= r.viewport.Max.Y {
			continue
		}

		for x := rect.Min.X; x < rect.Max.X; x++ {
			if x < r.viewport.Min.X || x >= r.viewport.Max.X {
				continue
			}

			cell := uv.Cell{
				Content: " ",
				Width:   1,
				Style: uv.Style{
					Bg: bgColor,
				},
			}
			scr.SetCell(x, y, &cell)
		}
	}
}

// paintText renders text lines to the screen.
func (r *Renderer) paintText(scr uv.Screen, rect uv.Rectangle, lines []uv.Line) {
	y := rect.Min.Y

	for _, line := range lines {
		if y >= rect.Max.Y || y >= r.viewport.Max.Y {
			break
		}

		if y >= r.viewport.Min.Y {
			x := rect.Min.X

			for _, cell := range line {
				if x >= rect.Max.X || x >= r.viewport.Max.X {
					break
				}

				if x >= r.viewport.Min.X {
					scr.SetCell(x, y, &cell)
				}

				x += cell.Width
			}
		}

		y++
	}
}

// getAttr gets an attribute value from a node by name.
func getAttr(n *node, name string) string {
	for _, attr := range n.Attr() {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

// renderTextToLines converts text to styled uv.Line based on computed style.
// Uses displaywidth library for proper grapheme cluster iteration and width calculation.
// Handles newlines by splitting text into multiple lines.
// Converts tabs to spaces according to tab-size style property.
// Sets cell.Link based on element ID or parent anchor element.
func (r *Renderer) renderTextToLines(text string, n *node) []uv.Line {
	style := n.computedStyle

	// Build the style with attributes
	cellStyle := uv.Style{
		Fg: style.Color,
		Bg: style.BackgroundColor,
	}

	// Build link based on element ID or parent anchor
	var cellLink uv.Link
	var hasLink bool

	// Check if this node or any parent is an anchor or has an ID
	currentNode := n
	for currentNode != nil {
		// Get ID and href for this node
		id := getAttr(currentNode, "id")
		href := getAttr(currentNode, "href")

		// If node is an anchor with href, use href as URL
		if currentNode.Type() == html.ElementNode && currentNode.Data() == "a" && href != "" {
			cellLink = uv.Link{
				URL:    href,
				Params: "",
			}
			// If anchor also has an ID, add it to params
			if id != "" {
				cellLink.Params = "id=" + id
			}
			hasLink = true
			break
		}

		// If node has an ID (but not an anchor with href), use empty URL with id in params
		if id != "" {
			cellLink = uv.Link{
				URL:    "",
				Params: "id=" + id,
			}
			hasLink = true
			break
		}

		// Move to parent
		if currentNode.parent != nil {
			if parentNode, ok := currentNode.parent.(*node); ok {
				currentNode = parentNode
			} else {
				break
			}
		} else {
			break
		}
	}

	// Only set link on cells if we found one
	_ = hasLink

	// Apply text attributes
	if style.FontWeight == FontWeightBold {
		cellStyle.Attrs |= uv.AttrBold
	}
	if style.FontStyle == FontStyleItalic {
		cellStyle.Attrs |= uv.AttrItalic
	}
	if style.Faint {
		cellStyle.Attrs |= uv.AttrFaint
	}

	// Apply text decorations (can have multiple)
	if style.TextDecoration.Has(TextDecorationUnderline) {
		// Map our underline style to UV underline constants
		switch style.TextDecorationStyle {
		case UnderlineStyleSingle:
			cellStyle.Underline = uv.UnderlineSingle
		case UnderlineStyleDouble:
			cellStyle.Underline = uv.UnderlineDouble
		case UnderlineStyleCurly:
			cellStyle.Underline = uv.UnderlineCurly
		case UnderlineStyleDotted:
			cellStyle.Underline = uv.UnderlineDotted
		case UnderlineStyleDashed:
			cellStyle.Underline = uv.UnderlineDashed
		default:
			cellStyle.Underline = uv.UnderlineSingle
		}

		// Apply underline color if set
		if style.TextDecorationColor != nil {
			cellStyle.UnderlineColor = style.TextDecorationColor
		}
	}

	if style.TextDecoration.Has(TextDecorationLineThrough) {
		// Line-through is represented by the strikethrough attribute
		cellStyle.Attrs |= uv.AttrStrikethrough
	}

	// Get tab size (default to 8 if not set)
	tabSize := style.TabSize
	if tabSize <= 0 {
		tabSize = 8
	}

	// Handle multiline text by splitting on newlines and converting tabs to spaces
	var lines []uv.Line
	currentLine := make(uv.Line, 0)
	currentColumn := 0             // Track column position for tab expansion
	var pendingControlChars string // Accumulate control characters
	grs := displaywidth.StringGraphemes(text)

	for grs.Next() {
		gr := grs.Value()

		// Handle newline - start a new line
		if gr == "\n" {
			// If we have pending control chars, attach them to the last cell
			if pendingControlChars != "" && len(currentLine) > 0 {
				lastIdx := len(currentLine) - 1
				currentLine[lastIdx].Content += pendingControlChars
				pendingControlChars = ""
			}
			lines = append(lines, currentLine)
			currentLine = make(uv.Line, 0)
			currentColumn = 0
			continue
		}

		// Handle tab - convert to spaces
		if gr == "\t" {
			// If we have pending control chars, attach them to the last cell
			if pendingControlChars != "" && len(currentLine) > 0 {
				lastIdx := len(currentLine) - 1
				currentLine[lastIdx].Content += pendingControlChars
				pendingControlChars = ""
			}

			// Calculate number of spaces to next tab stop
			spacesToNextTab := tabSize - (currentColumn % tabSize)

			// Add space cells
			for range spacesToNextTab {
				cell := uv.Cell{
					Content: " ",
					Width:   1,
					Style:   cellStyle,
					Link:    cellLink,
				}
				currentLine = append(currentLine, cell)
				currentColumn++
			}
			continue
		}

		w := grs.Width()

		// Handle zero-width characters (control characters, combining marks, etc.)
		if w == 0 {
			// Accumulate control characters to attach to next cell
			pendingControlChars += gr
			continue
		}

		// We have a visible character - create cell with any pending control chars
		cell := uv.Cell{
			Content: pendingControlChars + gr,
			Width:   w,
			Style:   cellStyle,
			Link:    cellLink,
		}
		pendingControlChars = "" // Reset

		currentLine = append(currentLine, cell)
		currentColumn += w
	}

	// If we have pending control chars at the end, attach to last cell
	if pendingControlChars != "" && len(currentLine) > 0 {
		lastIdx := len(currentLine) - 1
		currentLine[lastIdx].Content += pendingControlChars
	}

	// Add the last line if it has content or if text ended with newline
	if len(currentLine) > 0 || len(lines) > 0 {
		lines = append(lines, currentLine)
	}

	// Return at least one empty line if text is empty
	if len(lines) == 0 {
		lines = append(lines, uv.Line{})
	}

	return lines
}

// wrapLines wraps lines to fit within the given width, respecting word boundaries.
func (r *Renderer) wrapLines(lines []uv.Line, maxWidth int) []uv.Line {
	if maxWidth <= 0 {
		return lines
	}

	var wrapped []uv.Line

	for _, line := range lines {
		// Calculate total width of the line
		totalWidth := 0
		for _, cell := range line {
			totalWidth += cell.Width
		}

		// If line fits, keep it
		if totalWidth <= maxWidth {
			wrapped = append(wrapped, line)
			continue
		}

		// Wrap the line at word boundaries when possible
		currentLine := make(uv.Line, 0, maxWidth)
		currentWidth := 0
		wordStart := 0
		wordWidth := 0
		inWord := false

		for i, cell := range line {
			isSpace := unicode.IsSpace([]rune(cell.Content)[0])

			// Track word boundaries
			if !isSpace && !inWord {
				// Start of a word
				wordStart = len(currentLine)
				wordWidth = 0
				inWord = true
			} else if isSpace && inWord {
				// End of word
				inWord = false
			}

			if inWord {
				wordWidth += cell.Width
			}

			// Check if adding this cell would exceed width
			if currentWidth+cell.Width > maxWidth {
				// Try to wrap at word boundary
				if inWord && wordStart > 0 && wordWidth <= maxWidth {
					// Wrap before the current word
					wrapped = append(wrapped, currentLine[:wordStart])
					currentLine = currentLine[wordStart:]
					currentWidth = wordWidth
				} else {
					// Can't wrap at word boundary, wrap here
					if len(currentLine) > 0 {
						wrapped = append(wrapped, currentLine)
					}
					currentLine = make(uv.Line, 0, maxWidth)
					currentWidth = 0
					wordStart = 0
					wordWidth = cell.Width
				}
			}

			currentLine = append(currentLine, cell)
			currentWidth += cell.Width

			// Handle last cell
			if i == len(line)-1 && len(currentLine) > 0 {
				wrapped = append(wrapped, currentLine)
			}
		}
	}

	return wrapped
}
