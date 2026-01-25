package doc

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
	"golang.org/x/net/html"
)

// DisplayType represents the CSS display property.
type DisplayType string

const (
	DisplayBlock  DisplayType = "block"
	DisplayInline DisplayType = "inline"
	DisplayNone   DisplayType = "none"
	DisplayFlex   DisplayType = "flex"
)

// TextAlign represents the CSS text-align property.
type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

// FontWeight represents the CSS font-weight property.
type FontWeight string

const (
	FontWeightNormal FontWeight = "normal"
	FontWeightBold   FontWeight = "bold"
)

// FontStyle represents the CSS font-style property.
type FontStyle string

const (
	FontStyleNormal FontStyle = "normal"
	FontStyleItalic FontStyle = "italic"
)

// TextDecorationType represents individual text decoration values.
type TextDecorationType string

const (
	TextDecorationNone        TextDecorationType = "none"
	TextDecorationUnderline   TextDecorationType = "underline"
	TextDecorationLineThrough TextDecorationType = "line-through"
)

// TextDecoration represents the CSS text-decoration property (can have multiple values).
type TextDecoration []TextDecorationType

// Has checks if the decorations include a specific decoration type.
func (td TextDecoration) Has(decoration TextDecorationType) bool {
	for _, d := range td {
		if d == decoration {
			return true
		}
	}
	return false
}

// IsNone returns true if there are no decorations or only "none".
func (td TextDecoration) IsNone() bool {
	return len(td) == 0 || (len(td) == 1 && td[0] == TextDecorationNone)
}

// UnderlineStyle represents the CSS text-decoration-style property for underlines.
type UnderlineStyle string

const (
	UnderlineStyleNone   UnderlineStyle = "none"
	UnderlineStyleSingle UnderlineStyle = "single"
	UnderlineStyleDouble UnderlineStyle = "double"
	UnderlineStyleCurly  UnderlineStyle = "curly"
	UnderlineStyleDotted UnderlineStyle = "dotted"
	UnderlineStyleDashed UnderlineStyle = "dashed"
)

// ComputedStyle represents the final computed style for a node after
// applying tag defaults, classes, inline styles, and inheritance.
type ComputedStyle struct {
	// Display type
	Display DisplayType

	// Dimensions (in cells)
	Width  int // 0 = auto
	Height int // 0 = auto

	// Text properties
	Color               color.Color
	BackgroundColor     color.Color
	FontWeight          FontWeight
	FontStyle           FontStyle
	TextDecoration      TextDecoration // Can have multiple: underline, line-through
	TextDecorationStyle UnderlineStyle
	TextDecorationColor color.Color
	TextAlign           TextAlign
	TabSize             int // Number of spaces per tab (CSS tab-size)

	// Opacity/Faint
	Faint bool
}

// NewComputedStyle returns a new ComputedStyle with default values.
func NewComputedStyle() *ComputedStyle {
	return &ComputedStyle{
		Display:             DisplayBlock,
		Width:               0,   // auto
		Height:              0,   // auto
		Color:               nil, // inherit
		BackgroundColor:     nil, // transparent
		FontWeight:          FontWeightNormal,
		FontStyle:           FontStyleNormal,
		TextDecoration:      nil, // none
		TextDecorationStyle: UnderlineStyleNone,
		TextDecorationColor: nil, // inherit or same as Color
		TextAlign:           TextAlignLeft,
		TabSize:             8, // Default tab size (CSS default)
		Faint:               false,
	}
}

// Inherit copies inheritable properties from the parent style.
func (s *ComputedStyle) Inherit(parent *ComputedStyle) {
	if parent == nil {
		return
	}

	// These properties inherit from parent
	if s.Color == nil {
		s.Color = parent.Color
	}
	// FontWeight inherits
	if s.FontWeight == "" {
		s.FontWeight = parent.FontWeight
	}
	// FontStyle inherits
	if s.FontStyle == "" {
		s.FontStyle = parent.FontStyle
	}
	// TextDecoration inherits
	if len(s.TextDecoration) == 0 {
		s.TextDecoration = parent.TextDecoration
	}
	// TextDecorationStyle inherits
	if s.TextDecorationStyle == "" {
		s.TextDecorationStyle = parent.TextDecorationStyle
	}
	// TextDecorationColor inherits (or defaults to Color if not set)
	if s.TextDecorationColor == nil {
		if parent.TextDecorationColor != nil {
			s.TextDecorationColor = parent.TextDecorationColor
		} else {
			s.TextDecorationColor = s.Color
		}
	}
	// TextAlign inherits
	if s.TextAlign == "" {
		s.TextAlign = parent.TextAlign
	}
	// TabSize inherits
	if s.TabSize == 0 {
		s.TabSize = parent.TabSize
	}
	// Faint can inherit
	if !s.Faint {
		s.Faint = parent.Faint
	}

	// These properties do NOT inherit (non-inheritable):
	// - Display
	// - Width, Height
	// - BackgroundColor
}

// getTagDefaultStyle returns the default style for HTML tags.
func getTagDefaultStyle(tagName string) *ComputedStyle {
	style := NewComputedStyle()

	switch tagName {
	case "div", "p", "main", "header", "footer", "section", "article":
		style.Display = DisplayBlock

	case "span", "a", "strong", "b", "em", "i", "u", "code":
		style.Display = DisplayInline

	case "pre":
		style.Display = DisplayBlock
		style.TabSize = 4 // Common default for code

	case "body", "html":
		style.Display = DisplayBlock
	}

	// Tag-specific styling
	switch tagName {
	case "strong", "b":
		style.FontWeight = FontWeightBold
	case "em", "i":
		style.FontStyle = FontStyleItalic
	case "u":
		style.TextDecoration = TextDecoration{TextDecorationUnderline}
		style.TextDecorationStyle = UnderlineStyleSingle
	case "a":
		style.Color = ansi.BrightBlue
		style.TextDecoration = TextDecoration{TextDecorationUnderline}
		style.TextDecorationStyle = UnderlineStyleSingle
	case "s", "strike", "del":
		style.TextDecoration = TextDecoration{TextDecorationLineThrough}
	}

	return style
}

// computeStyle computes the final style for a node by merging:
// 1. Tag defaults
// 2. Parent inherited properties
// 3. Inline style attribute (future)
// 4. Class styles (future)
func (n *node) computeStyle() *ComputedStyle {
	// Start with tag defaults (only for element nodes)
	var style *ComputedStyle
	if n.Type() == html.ElementNode {
		tagName := n.Data()
		style = getTagDefaultStyle(tagName)
	} else {
		style = NewComputedStyle()
	}

	// Inherit from parent
	if n.parent != nil {
		if parentNode, ok := n.parent.(*node); ok {
			if parentNode.computedStyle != nil {
				style.Inherit(parentNode.computedStyle)
			}
		}
	}

	// TODO: Parse inline style="" attribute
	// TODO: Apply class styles

	return style
}
