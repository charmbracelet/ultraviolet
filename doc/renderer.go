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
	boxTree  *Box // The box tree for layout
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

	// Build box tree from DOM tree
	r.boxTree = buildBoxTree(r.root)

	// Layout the box tree
	if r.boxTree != nil {
		r.layoutBox(r.boxTree, viewport)
	}

	// Paint the box tree
	r.paintBox(scr, r.boxTree, uv.Pos(0, 0))
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
	
	// Process whitespace according to CSS white-space property
	// Block-level text: trim leading/trailing whitespace
	text = processWhitespace(text, n.computedStyle.WhiteSpace, false)
	
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
// For inline children, it flows them horizontally with wrapping.
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
	children := n.Children()
	for i := 0; i < len(children); i++ {
		childNode, ok := children[i].(*node)
		if !ok {
			continue
		}

		// Check if we're exceeding viewport - stop laying out if so
		if currentY-y >= maxHeight && maxHeight > 0 {
			break
		}

		// Check if this child is inline
		isInline := childNode.computedStyle.Display == DisplayInline || childNode.Type() == html.TextNode

		if isInline {
			// Collect consecutive inline children (including whitespace-only text nodes)
			inlineChildren := []*node{childNode}
			j := i + 1
			for j < len(children) {
				nextChild, ok := children[j].(*node)
				if !ok {
					break
				}
				nextIsInline := nextChild.computedStyle.Display == DisplayInline || nextChild.Type() == html.TextNode
				if !nextIsInline {
					break
				}
				inlineChildren = append(inlineChildren, nextChild)
				j++
			}
			i = j - 1 // Skip the children we collected

			// Layout inline children horizontally (layoutInline will skip whitespace-only nodes)
			childRect := uv.Rect(x, currentY, width, maxHeight-(currentY-y))
			inlineHeight := r.layoutInline(inlineChildren, childRect)
			currentY += inlineHeight
		} else {
			// Layout block child normally
			childRect := uv.Rect(x, currentY, width, maxHeight-(currentY-y))
			layoutRect := r.layout(childNode, childRect)

			// Move down for next child
			if layoutRect.Dy() > 0 {
				currentY += layoutRect.Dy()
			}
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

// layoutInline handles inline layout (horizontal flow with wrapping).
// It collects text from all inline children and wraps them as a single flow.
func (r *Renderer) layoutInline(children []*node, availableRect uv.Rectangle) int {
	x := availableRect.Min.X
	y := availableRect.Min.Y
	width := availableRect.Dx()

	// Collect all text cells from inline children (recursively)
	var allCells uv.Line
	var allTextNodes []*node

	// Helper function to recursively collect text from inline elements
	var collectInlineText func(*node)
	collectInlineText = func(n *node) {
		if n.Type() == html.TextNode {
			// Get text content and process whitespace
			// For whitespace-only text nodes between elements, use block context (trim them)
			// For text nodes with actual content, use inline context (preserve spaces)
			rawText := n.Data()
			isWhitespaceOnly := strings.TrimSpace(rawText) == ""
			
			// Check if computedStyle is set
			if n.computedStyle == nil {
				return // Skip if no style computed
			}
			
			text := processWhitespace(rawText, n.computedStyle.WhiteSpace, !isWhitespaceOnly)
			if text == "" {
				return
			}

			// Render text to cells with styling
			lines := r.renderTextToLines(text, n)
			for _, line := range lines {
				allCells = append(allCells, line...)
			}
			allTextNodes = append(allTextNodes, n)
		} else if n.computedStyle != nil && n.computedStyle.Display == DisplayInline {
			// Recursively process children of inline elements
			for _, child := range n.Children() {
				if childNode, ok := child.(*node); ok {
					collectInlineText(childNode)
				}
			}
		}
	}

	// Process all children
	for _, child := range children {
		collectInlineText(child)
	}

	// Wrap the combined line
	wrappedLines := r.wrapLines([]uv.Line{allCells}, width)

	// Store wrapped lines in the first text node for painting
	// (only one node needs to paint the combined line)
	if len(allTextNodes) > 0 {
		allTextNodes[0].wrappedLines = wrappedLines
		allTextNodes[0].layout = NewLayoutBox(uv.Rect(x, y, width, len(wrappedLines)))
		allTextNodes[0].layout.Dirty = false

		// Set empty layout for other text nodes so they don't paint
		for i := 1; i < len(allTextNodes); i++ {
			allTextNodes[i].wrappedLines = nil
			allTextNodes[i].layout = NewLayoutBox(uv.Rect(x, y, 0, 0))
			allTextNodes[i].layout.Dirty = false
		}
	}

	// Set layout for inline element containers (they don't paint themselves)
	for _, child := range children {
		if child.Type() != html.TextNode {
			child.layout = NewLayoutBox(uv.Rect(x, y, width, len(wrappedLines)))
			child.layout.Dirty = false
		}
	}

	return len(wrappedLines)
}

// processWhitespace processes text according to CSS white-space property.
// See: https://developer.mozilla.org/en-US/docs/Web/CSS/white-space
// isInlineContext: if true, preserve leading/trailing spaces (inline elements).
// If false, trim leading/trailing whitespace (block-level text).
func processWhitespace(text string, whiteSpace WhiteSpace, isInlineContext bool) string {
	switch whiteSpace {
	case WhiteSpacePre, WhiteSpacePreWrap, WhiteSpaceBreakSpaces:
		// Preserve all whitespace including newlines and spaces
		return text
		
	case WhiteSpacePreLine:
		// Collapse whitespace but preserve newlines
		// Replace sequences of spaces/tabs with single space
		// But keep newlines
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			// Collapse whitespace in each line
			lines[i] = collapseWhitespace(line)
		}
		return strings.Join(lines, "\n")
		
	case WhiteSpaceNormal, WhiteSpaceNowrap:
		fallthrough
	default:
		// Collapse all whitespace (including newlines) into single spaces
		collapsed := collapseWhitespace(text)
		// Only trim leading/trailing whitespace for block-level text
		// In inline context, preserve leading/trailing spaces (they're significant)
		if !isInlineContext {
			collapsed = strings.TrimSpace(collapsed)
		}
		return collapsed
	}
}

// collapseWhitespace replaces sequences of whitespace with a single space.
func collapseWhitespace(text string) string {
	// Replace all whitespace (spaces, tabs, newlines) with spaces
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, text)
	
	// Collapse multiple spaces into one
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	
	return text
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
			break
		}

		// If node has an ID (but not an anchor with href), use empty URL with id in params
		if id != "" {
			cellLink = uv.Link{
				URL:    "",
				Params: "id=" + id,
			}
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

// layoutBox performs layout for a box and its children using the CSS box model.
// Returns the final rectangle occupied by the box.
func (r *Renderer) layoutBox(box *Box, availableRect uv.Rectangle) uv.Rectangle {
	if box == nil {
		return uv.Rectangle{}
	}

	if box.IsBlock() {
		return r.layoutBlockBox(box, availableRect)
	}
	
	// Inline boxes are laid out within their parent's inline formatting context
	return uv.Rectangle{}
}

// layoutBlockBox lays out a block-level box.
// This establishes a block formatting context (BFC).
func (r *Renderer) layoutBlockBox(box *Box, availableRect uv.Rectangle) uv.Rectangle {
	x := availableRect.Min.X
	y := availableRect.Min.Y
	width := availableRect.Dx()
	maxHeight := availableRect.Dy()

	// Apply width from style if set
	if box.Style.Width > 0 {
		width = box.Style.Width
	}

	currentY := y
	i := 0

	// Layout children
	for i < len(box.Children) {
		child := box.Children[i]
		
		if currentY-y >= maxHeight && maxHeight > 0 {
			break // Exceeded viewport
		}

		if child.IsBlock() {
			// Block child: layout and stack vertically
			childRect := uv.Rect(x, currentY, width, maxHeight-(currentY-y))
			layoutRect := r.layoutBox(child, childRect)
			child.Rect = layoutRect
			
			if layoutRect.Dy() > 0 {
				currentY += layoutRect.Dy()
			}
			i++
		} else {
			// Inline children: collect consecutive inline boxes and establish IFC
			var inlineBoxes []*Box
			for i < len(box.Children) && box.Children[i].IsInline() {
				inlineBoxes = append(inlineBoxes, box.Children[i])
				i++
			}
			
			if len(inlineBoxes) > 0 {
				inlineRect := uv.Rect(x, currentY, width, maxHeight-(currentY-y))
				height := r.layoutInlineFormattingContext(inlineBoxes, inlineRect)
				currentY += height
			}
		}
	}

	// Calculate total height
	totalHeight := currentY - y
	if box.Style.Height > 0 {
		totalHeight = box.Style.Height
	}

	box.Rect = uv.Rect(x, y, width, totalHeight)
	return box.Rect
}

// layoutInlineFormattingContext lays out a sequence of inline boxes.
// This establishes an inline formatting context (IFC).
// Returns the height consumed by the inline content.
func (r *Renderer) layoutInlineFormattingContext(boxes []*Box, availableRect uv.Rectangle) int {
	x := availableRect.Min.X
	y := availableRect.Min.Y
	width := availableRect.Dx()

	// Collect all text content from inline boxes (recursively)
	var allCells uv.Line
	var boxesWithText []*Box // Track which boxes contributed text

	var collectText func(*Box)
	collectText = func(box *Box) {
		if box.Text != "" {
			// Process whitespace
			rawText := box.Text
			isWhitespaceOnly := strings.TrimSpace(rawText) == ""
			
			text := processWhitespace(rawText, box.Style.WhiteSpace, !isWhitespaceOnly)
			if text == "" {
				return // Skip empty text
			}

			// Render text to cells with styling
			lines := r.renderTextToLines(text, box.Node)
			for _, line := range lines {
				allCells = append(allCells, line...)
			}
			boxesWithText = append(boxesWithText, box)
		}

		// Recursively process inline children
		for _, child := range box.Children {
			if child.IsInline() {
				collectText(child)
			}
		}
	}

	// Collect text from all inline boxes
	for _, box := range boxes {
		collectText(box)
	}

	// Wrap the combined text
	wrappedLines := r.wrapLines([]uv.Line{allCells}, width)

	// Store wrapped lines in the first box that contributed text
	if len(boxesWithText) > 0 {
		boxesWithText[0].WrappedLines = wrappedLines
		boxesWithText[0].Rect = uv.Rect(x, y, width, len(wrappedLines))

		// Clear wrapped lines for other boxes (only first one paints)
		for i := 1; i < len(boxesWithText); i++ {
			boxesWithText[i].WrappedLines = nil
			boxesWithText[i].Rect = uv.Rect(x, y, 0, 0)
		}
	}

	// Set rectangles for all inline boxes
	for _, box := range boxes {
		if box.Rect.Dx() == 0 && box.Rect.Dy() == 0 {
			// Box wasn't set yet (no text content)
			box.Rect = uv.Rect(x, y, width, len(wrappedLines))
		}
	}

	return len(wrappedLines)
}

// paintBox renders a box and its children to the screen.
func (r *Renderer) paintBox(scr uv.Screen, box *Box, offset uv.Position) {
	if box == nil {
		return
	}

	// Calculate absolute position
	absRect := box.Rect.Add(offset)

	// Skip if not visible in viewport
	if !absRect.Overlaps(r.viewport) {
		return
	}

	// Paint background if set
	if box.Style != nil && box.Style.BackgroundColor != nil {
		r.paintBackground(scr, absRect, box.Style.BackgroundColor)
	}

	// Paint text content
	if len(box.WrappedLines) > 0 {
		r.paintText(scr, absRect, box.WrappedLines)
	}

	// Paint children
	for _, child := range box.Children {
		r.paintBox(scr, child, offset)
	}
}

