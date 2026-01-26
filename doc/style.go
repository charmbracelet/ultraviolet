package doc

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
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

// WhiteSpace represents the CSS white-space property.
type WhiteSpace string

const (
	WhiteSpaceNormal      WhiteSpace = "normal"       // Collapse whitespace, wrap text
	WhiteSpaceNowrap      WhiteSpace = "nowrap"       // Collapse whitespace, no wrap
	WhiteSpacePre         WhiteSpace = "pre"          // Preserve whitespace, no wrap
	WhiteSpacePreLine     WhiteSpace = "pre-line"     // Collapse whitespace except newlines, wrap
	WhiteSpacePreWrap     WhiteSpace = "pre-wrap"     // Preserve whitespace, wrap
	WhiteSpaceBreakSpaces WhiteSpace = "break-spaces" // Like pre-wrap but breaks at any space
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
	TabSize             int        // Number of spaces per tab (CSS tab-size)
	WhiteSpace          WhiteSpace // How to handle whitespace (CSS white-space)

	// Opacity/Faint
	Faint bool
}

// NewComputedStyle returns a new ComputedStyle with default values.
func NewComputedStyle() *ComputedStyle {
	return &ComputedStyle{
		Display:             DisplayBlock,
		Width:               0,   // auto
		Height:              0,   // auto
		Color:               nil, // inherit from parent
		BackgroundColor:     nil, // inherit from parent (terminal-specific; web CSS default is transparent)
		FontWeight:          FontWeightNormal,
		FontStyle:           FontStyleNormal,
		TextDecoration:      nil, // none
		TextDecorationStyle: "",  // Will inherit from parent
		TextDecorationColor: nil, // inherit or same as Color
		TextAlign:           TextAlignLeft,
		TabSize:             0,  // Will inherit from parent or default to 8
		WhiteSpace:          "", // Will inherit from parent or default to normal
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
	// BackgroundColor inherits in terminals (unlike web CSS)
	// This ensures text remains readable on colored backgrounds
	if s.BackgroundColor == nil {
		s.BackgroundColor = parent.BackgroundColor
	}
	// FontWeight always inherits (overwrite default)
	s.FontWeight = parent.FontWeight
	// FontStyle always inherits (overwrite default)
	s.FontStyle = parent.FontStyle
	// TextDecoration inherits
	if len(s.TextDecoration) == 0 {
		s.TextDecoration = parent.TextDecoration
	}
	// TextDecorationStyle inherits
	if s.TextDecorationStyle == "" {
		if parent.TextDecorationStyle != "" {
			s.TextDecorationStyle = parent.TextDecorationStyle
		} else {
			s.TextDecorationStyle = UnderlineStyleNone
		}
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
	if parent.TabSize > 0 {
		s.TabSize = parent.TabSize
	}
	// If still 0 (no parent or parent has 0), use default
	if s.TabSize == 0 {
		s.TabSize = 8
	}
	// WhiteSpace inherits
	if parent.WhiteSpace != "" {
		s.WhiteSpace = parent.WhiteSpace
	}
	// If still empty (no parent or parent has empty), use default
	if s.WhiteSpace == "" {
		s.WhiteSpace = WhiteSpaceNormal
	}
	// Faint can inherit
	if !s.Faint {
		s.Faint = parent.Faint
	}

	// These properties do NOT inherit (non-inheritable):
	// - Display
	// - Width, Height
}

// applyTagDefaults applies tag-specific styling (display, bold, italic, etc.)
// to an existing computed style.
func applyTagDefaults(tagName string, style *ComputedStyle) {
	switch tagName {
	case "div", "p", "main", "header", "footer", "section", "article":
		style.Display = DisplayBlock

	case "span", "a", "strong", "b", "em", "i", "u", "s", "strike", "del", "code":
		style.Display = DisplayInline

	case "pre":
		style.Display = DisplayBlock
		style.TabSize = 4                // Common default for code
		style.WhiteSpace = WhiteSpacePre // Preserve whitespace and newlines

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
		style.TextDecoration = append(style.TextDecoration, TextDecorationUnderline)
		style.TextDecorationStyle = UnderlineStyleSingle
	case "a":
		style.Color = ansi.BrightBlue
		style.TextDecoration = append(style.TextDecoration, TextDecorationUnderline)
		style.TextDecorationStyle = UnderlineStyleSingle
	case "s", "strike", "del":
		style.TextDecoration = append(style.TextDecoration, TextDecorationLineThrough)
	}
}

// computeStyle computes the final style for a node by merging:
// 1. Base defaults
// 2. Parent inherited properties
// 3. Tag-specific defaults (strong=bold, em=italic, etc.)
// 4. Inline style attribute
// 5. Class styles (future)
func (n *node) computeStyle() *ComputedStyle {
	// Start with base defaults
	style := NewComputedStyle()

	// Inherit from parent
	if n.parent != nil {
		if parentNode, ok := n.parent.(*node); ok {
			if parentNode.computedStyle != nil {
				style.Inherit(parentNode.computedStyle)
			}
		}
	}

	// Apply tag-specific styles (AFTER inheritance so they override)
	if n.Type() == html.ElementNode {
		tagName := n.Data()
		applyTagDefaults(tagName, style)
	}

	// Parse inline style="" attribute
	if n.Type() == html.ElementNode {
		styleAttr := getAttr(n, "style")
		if styleAttr != "" {
			parseInlineStyle(styleAttr, style)
		}
	}

	// TODO: Apply class styles

	return style
}

// parseInlineStyle parses a CSS style attribute and applies it to the given ComputedStyle.
// Example: style="color: red; font-weight: bold"
func parseInlineStyle(styleAttr string, style *ComputedStyle) {
	parser := css.NewParser(parse.NewInputString(styleAttr), true) // true = inline mode

	for {
		gt, _, data := parser.Next()
		if gt == css.ErrorGrammar {
			break
		}

		switch gt {
		case css.DeclarationGrammar:
			// data contains the property name
			property := string(data)

			// Get the values
			var values []string
			for _, val := range parser.Values() {
				if val.TokenType != css.WhitespaceToken {
					values = append(values, string(val.Data))
				}
			}

			if len(values) > 0 {
				value := strings.Join(values, " ")
				applyStyleProperty(property, value, style)
			}

		case css.AtRuleGrammar, css.BeginRulesetGrammar, css.QualifiedRuleGrammar:
			// Skip nested structures (shouldn't appear in inline styles but handle gracefully)
			parser.Next()

		case css.BeginAtRuleGrammar:
			// Skip @rules (shouldn't appear in inline styles)
			for {
				gt, _, _ := parser.Next()
				if gt == css.ErrorGrammar || gt == css.EndAtRuleGrammar {
					break
				}
			}
		}
	}
}

// applyStyleProperty applies a single CSS property to a ComputedStyle.
func applyStyleProperty(property, value string, style *ComputedStyle) {
	property = strings.ToLower(strings.TrimSpace(property))
	value = strings.ToLower(strings.TrimSpace(value))

	switch property {
	case "color":
		if c := parseColor(value); c != nil {
			style.Color = c
		}

	case "background-color":
		if c := parseColor(value); c != nil {
			style.BackgroundColor = c
		}

	case "font-weight":
		switch value {
		case "bold", "700", "800", "900":
			style.FontWeight = FontWeightBold
		case "normal", "400":
			style.FontWeight = FontWeightNormal
		}

	case "font-style":
		switch value {
		case "italic", "oblique":
			style.FontStyle = FontStyleItalic
		case "normal":
			style.FontStyle = FontStyleNormal
		}

	case "text-decoration":
		// Can be multiple values: "underline line-through"
		parts := strings.Fields(value)
		style.TextDecoration = make([]TextDecorationType, 0)
		for _, part := range parts {
			switch part {
			case "underline":
				style.TextDecoration = append(style.TextDecoration, TextDecorationUnderline)
			case "line-through":
				style.TextDecoration = append(style.TextDecoration, TextDecorationLineThrough)
			case "none":
				style.TextDecoration = []TextDecorationType{}
				return
			}
		}

	case "text-decoration-style":
		switch value {
		case "solid", "single":
			style.TextDecorationStyle = UnderlineStyleSingle
		case "double":
			style.TextDecorationStyle = UnderlineStyleDouble
		case "curly", "wavy":
			style.TextDecorationStyle = UnderlineStyleCurly
		case "dotted":
			style.TextDecorationStyle = UnderlineStyleDotted
		case "dashed":
			style.TextDecorationStyle = UnderlineStyleDashed
		}

	case "text-decoration-color":
		if c := parseColor(value); c != nil {
			style.TextDecorationColor = c
		}

	case "display":
		switch value {
		case "block":
			style.Display = DisplayBlock
		case "inline":
			style.Display = DisplayInline
		case "none":
			style.Display = DisplayNone
		case "flex":
			style.Display = DisplayFlex
		}

	case "tab-size":
		if size, err := strconv.Atoi(value); err == nil && size > 0 {
			style.TabSize = size
		}

	case "white-space":
		switch value {
		case "normal":
			style.WhiteSpace = WhiteSpaceNormal
		case "nowrap":
			style.WhiteSpace = WhiteSpaceNowrap
		case "pre":
			style.WhiteSpace = WhiteSpacePre
		case "pre-line":
			style.WhiteSpace = WhiteSpacePreLine
		case "pre-wrap":
			style.WhiteSpace = WhiteSpacePreWrap
		case "break-spaces":
			style.WhiteSpace = WhiteSpaceBreakSpaces
		}
	}
}

// parseColor parses a CSS color value into a color.Color.
// Supports: named colors, indexed colors (0-255), hex (#RGB, #RRGGBB), rgb(r,g,b)
func parseColor(value string) color.Color {
	value = strings.TrimSpace(value)

	// Try parsing as a number (indexed color 0-255)
	if num, err := strconv.Atoi(value); err == nil && num >= 0 && num <= 255 {
		return ansi.IndexedColor(uint8(num))
	}

	// Named colors
	switch value {
	case "black":
		return ansi.Black
	case "red":
		return ansi.Red
	case "green":
		return ansi.Green
	case "yellow":
		return ansi.Yellow
	case "blue":
		return ansi.Blue
	case "magenta":
		return ansi.Magenta
	case "cyan":
		return ansi.Cyan
	case "white":
		return ansi.White
	case "gray", "grey":
		return ansi.BrightBlack
	case "brightred":
		return ansi.BrightRed
	case "brightgreen":
		return ansi.BrightGreen
	case "brightyellow":
		return ansi.BrightYellow
	case "brightblue":
		return ansi.BrightBlue
	case "brightmagenta":
		return ansi.BrightMagenta
	case "brightcyan":
		return ansi.BrightCyan
	case "brightwhite":
		return ansi.BrightWhite
	}

	// Hex colors: #RGB or #RRGGBB - use colorful library
	if strings.HasPrefix(value, "#") {
		if c, err := colorful.Hex(value); err == nil {
			// Convert to RGBA for color.Color interface
			r, g, b := c.RGB255()
			return color.RGBA{R: r, G: g, B: b, A: 255}
		}
	}

	// TODO: rgb(r, g, b) parsing if needed

	return nil
}
