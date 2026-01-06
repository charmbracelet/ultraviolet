package dom

import (
uv "github.com/charmbracelet/ultraviolet"
"github.com/charmbracelet/x/ansi"
"github.com/lucasb-eyer/go-colorful"
)

// BorderStyle holds the characters used for each part of the border.
type BorderStyle struct {
Top         string
Bottom      string
Left        string
Right       string
TopLeft     string
TopRight    string
BottomLeft  string
BottomRight string
}

// BorderStyleNormal returns a standard box-drawing border style.
func BorderStyleNormal() BorderStyle {
return BorderStyle{"─", "─", "│", "│", "┌", "┐", "└", "┘"}
}

// BorderStyleRounded returns a border style with rounded corners.
func BorderStyleRounded() BorderStyle {
return BorderStyle{"─", "─", "│", "│", "╭", "╮", "╰", "╯"}
}

// BorderStyleDouble returns a border style with double lines.
func BorderStyleDouble() BorderStyle {
return BorderStyle{"═", "═", "║", "║", "╔", "╗", "╚", "╝"}
}

// BorderStyleThick returns a border style with thick lines.
func BorderStyleThick() BorderStyle {
return BorderStyle{"━", "━", "┃", "┃", "┏", "┓", "┗", "┛"}
}

// BorderStyleHidden returns a border style with no visible characters.
func BorderStyleHidden() BorderStyle {
return BorderStyle{" ", " ", " ", " ", " ", " ", " ", " "}
}

// BorderStyleBlock returns a border style using block characters.
func BorderStyleBlock() BorderStyle {
return BorderStyle{"█", "█", "█", "█", "█", "█", "█", "█"}
}

// Border represents a configurable border with styling options.
type Border struct {
Style BorderStyle

// Side visibility
Top    bool
Bottom bool
Left   bool
Right  bool

// Styling per side
TopStyle    uv.Style
BottomStyle uv.Style
LeftStyle   uv.Style
RightStyle  uv.Style

// Corner styling
TopLeftStyle     uv.Style
TopRightStyle    uv.Style
BottomLeftStyle  uv.Style
BottomRightStyle uv.Style
}

// NewBorder creates a new border with the given style and all sides visible.
func NewBorder(style BorderStyle) *Border {
defaultStyle := uv.Style{}
return &Border{
Style:            style,
Top:              true,
Bottom:           true,
Left:             true,
Right:            true,
TopStyle:         defaultStyle,
BottomStyle:      defaultStyle,
LeftStyle:        defaultStyle,
RightStyle:       defaultStyle,
TopLeftStyle:     defaultStyle,
TopRightStyle:    defaultStyle,
BottomLeftStyle:  defaultStyle,
BottomRightStyle: defaultStyle,
}
}

// NormalBorder creates a border with normal (square) style.
func NormalBorder() *Border {
return NewBorder(BorderStyleNormal())
}

// RoundedBorder creates a border with rounded corners.
func RoundedBorder() *Border {
return NewBorder(BorderStyleRounded())
}

// DoubleBorder creates a border with double lines.
func DoubleBorder() *Border {
return NewBorder(BorderStyleDouble())
}

// ThickBorder creates a border with thick lines.
func ThickBorder() *Border {
return NewBorder(BorderStyleThick())
}

// HiddenBorder creates a border with no visible characters.
func HiddenBorder() *Border {
return NewBorder(BorderStyleHidden())
}

// BlockBorder creates a border using block characters.
func BlockBorder() *Border {
return NewBorder(BorderStyleBlock())
}

// WithTopSide enables or disables the top side.
func (b *Border) WithTopSide(show bool) *Border {
b.Top = show
return b
}

// WithBottomSide enables or disables the bottom side.
func (b *Border) WithBottomSide(show bool) *Border {
b.Bottom = show
return b
}

// WithLeftSide enables or disables the left side.
func (b *Border) WithLeftSide(show bool) *Border {
b.Left = show
return b
}

// WithRightSide enables or disables the right side.
func (b *Border) WithRightSide(show bool) *Border {
b.Right = show
return b
}

// WithSides enables or disables all sides.
func (b *Border) WithSides(top, right, bottom, left bool) *Border {
b.Top = top
b.Right = right
b.Bottom = bottom
b.Left = left
return b
}

// Foreground sets the foreground color for all sides.
func (b *Border) Foreground(color ansi.Color) *Border {
style := uv.Style{Fg: color}
b.TopStyle = style
b.BottomStyle = style
b.LeftStyle = style
b.RightStyle = style
b.TopLeftStyle = style
b.TopRightStyle = style
b.BottomLeftStyle = style
b.BottomRightStyle = style
return b
}

// Background sets the background color for all sides.
func (b *Border) Background(color ansi.Color) *Border {

b.TopStyle.Bg = color
b.BottomStyle.Bg = color
b.LeftStyle.Bg = color
b.RightStyle.Bg = color
b.TopLeftStyle.Bg = color
b.TopRightStyle.Bg = color
b.BottomLeftStyle.Bg = color
b.BottomRightStyle.Bg = color
return b
}

// Attrs sets the attributes for all sides.
func (b *Border) Attrs(attrs uint8) *Border {
b.TopStyle.Attrs = attrs
b.BottomStyle.Attrs = attrs
b.LeftStyle.Attrs = attrs
b.RightStyle.Attrs = attrs
b.TopLeftStyle.Attrs = attrs
b.TopRightStyle.Attrs = attrs
b.BottomLeftStyle.Attrs = attrs
b.BottomRightStyle.Attrs = attrs
return b
}

// TopForeground sets the foreground color for the top side.
func (b *Border) TopForeground(color ansi.Color) *Border {
b.TopStyle.Fg = color
b.TopLeftStyle.Fg = color
b.TopRightStyle.Fg = color
return b
}

// BottomForeground sets the foreground color for the bottom side.
func (b *Border) BottomForeground(color ansi.Color) *Border {
b.BottomStyle.Fg = color
b.BottomLeftStyle.Fg = color
b.BottomRightStyle.Fg = color
return b
}

// LeftForeground sets the foreground color for the left side.
func (b *Border) LeftForeground(color ansi.Color) *Border {
b.LeftStyle.Fg = color
b.TopLeftStyle.Fg = color
b.BottomLeftStyle.Fg = color
return b
}

// RightForeground sets the foreground color for the right side.
func (b *Border) RightForeground(color ansi.Color) *Border {
b.RightStyle.Fg = color
b.TopRightStyle.Fg = color
b.BottomRightStyle.Fg = color
return b
}

// TopBackground sets the background color for the top side.
func (b *Border) TopBackground(color ansi.Color) *Border {
b.TopStyle.Bg = color
b.TopLeftStyle.Bg = color
b.TopRightStyle.Bg = color
return b
}

// BottomBackground sets the background color for the bottom side.
func (b *Border) BottomBackground(color ansi.Color) *Border {
b.BottomStyle.Bg = color
b.BottomLeftStyle.Bg = color
b.BottomRightStyle.Bg = color
return b
}

// LeftBackground sets the background color for the left side.
func (b *Border) LeftBackground(color ansi.Color) *Border {
b.LeftStyle.Bg = color
b.TopLeftStyle.Bg = color
b.BottomLeftStyle.Bg = color
return b
}

// RightBackground sets the background color for the right side.
func (b *Border) RightBackground(color ansi.Color) *Border {
b.RightStyle.Bg = color
b.TopRightStyle.Bg = color
b.BottomRightStyle.Bg = color
return b
}

// TopAttrs sets the attributes for the top side.
func (b *Border) TopAttrs(attrs uint8) *Border {
b.TopStyle.Attrs = attrs
b.TopLeftStyle.Attrs = attrs
b.TopRightStyle.Attrs = attrs
return b
}

// BottomAttrs sets the attributes for the bottom side.
func (b *Border) BottomAttrs(attrs uint8) *Border {
b.BottomStyle.Attrs = attrs
b.BottomLeftStyle.Attrs = attrs
b.BottomRightStyle.Attrs = attrs
return b
}

// LeftAttrs sets the attributes for the left side.
func (b *Border) LeftAttrs(attrs uint8) *Border {
b.LeftStyle.Attrs = attrs
b.TopLeftStyle.Attrs = attrs
b.BottomLeftStyle.Attrs = attrs
return b
}

// RightAttrs sets the attributes for the right side.
func (b *Border) RightAttrs(attrs uint8) *Border {
b.RightStyle.Attrs = attrs
b.TopRightStyle.Attrs = attrs
b.BottomRightStyle.Attrs = attrs
return b
}

// WithGradient applies a horizontal gradient from left to right.
func (b *Border) WithGradient(startColor, endColor ansi.Color) *Border {
// Apply gradient to left and right sides
b.LeftStyle.Fg = startColor
b.TopLeftStyle.Fg = startColor
b.BottomLeftStyle.Fg = startColor

b.RightStyle.Fg = endColor
b.TopRightStyle.Fg = endColor
b.BottomRightStyle.Fg = endColor

// Top and bottom get interpolated colors (middle)
r1, g1, b1, _ := startColor.RGBA()
r2, g2, b2, _ := endColor.RGBA()

// Convert from 16-bit to 8-bit
c1 := colorful.Color{R: float64(r1) / 65535.0, G: float64(g1) / 65535.0, B: float64(b1) / 65535.0}
c2 := colorful.Color{R: float64(r2) / 65535.0, G: float64(g2) / 65535.0, B: float64(b2) / 65535.0}
mid := c1.BlendLab(c2, 0.5)
midColor := ansi.HexColor(mid.Hex())

b.TopStyle.Fg = midColor
b.BottomStyle.Fg = midColor

return b
}

// WithVerticalGradient applies a vertical gradient from top to bottom.
func (b *Border) WithVerticalGradient(startColor, endColor ansi.Color) *Border {
// Apply gradient to top and bottom sides
b.TopStyle.Fg = startColor
b.TopLeftStyle.Fg = startColor
b.TopRightStyle.Fg = startColor

b.BottomStyle.Fg = endColor
b.BottomLeftStyle.Fg = endColor
b.BottomRightStyle.Fg = endColor

// Left and right get interpolated colors (middle)
r1, g1, b1, _ := startColor.RGBA()
r2, g2, b2, _ := endColor.RGBA()

// Convert from 16-bit to 8-bit
c1 := colorful.Color{R: float64(r1) / 65535.0, G: float64(g1) / 65535.0, B: float64(b1) / 65535.0}
c2 := colorful.Color{R: float64(r2) / 65535.0, G: float64(g2) / 65535.0, B: float64(b2) / 65535.0}
mid := c1.BlendLab(c2, 0.5)
midColor := ansi.HexColor(mid.Hex())

b.LeftStyle.Fg = midColor
b.RightStyle.Fg = midColor

return b
}

// WithCustomStyle sets a custom BorderStyle with specific characters.
func (b *Border) WithCustomStyle(style BorderStyle) *Border {
b.Style = style
return b
}

// Draw renders the border on the screen within the given area.
func (b *Border) Draw(scr uv.Screen, area uv.Rectangle) {
if area.Dx() < 2 || area.Dy() < 2 {
return
}

// Draw corners
if b.Top && b.Left {
scr.SetCell(area.Min.X, area.Min.Y, &uv.Cell{
Content: b.Style.TopLeft,
Width:   1,
Style:   b.TopLeftStyle,
})
}
if b.Top && b.Right {
scr.SetCell(area.Max.X-1, area.Min.Y, &uv.Cell{
Content: b.Style.TopRight,
Width:   1,
Style:   b.TopRightStyle,
})
}
if b.Bottom && b.Left {
scr.SetCell(area.Min.X, area.Max.Y-1, &uv.Cell{
Content: b.Style.BottomLeft,
Width:   1,
Style:   b.BottomLeftStyle,
})
}
if b.Bottom && b.Right {
scr.SetCell(area.Max.X-1, area.Max.Y-1, &uv.Cell{
Content: b.Style.BottomRight,
Width:   1,
Style:   b.BottomRightStyle,
})
}

// Draw top and bottom edges with gradient support
if b.Top {
width := area.Max.X - area.Min.X - 2
for i, x := 0, area.Min.X+1; x < area.Max.X-1; i, x = i+1, x+1 {
style := b.TopStyle
// Apply gradient if both corners have different colors
if b.TopLeftStyle.Fg != nil && b.TopRightStyle.Fg != nil && b.TopLeftStyle.Fg != b.TopRightStyle.Fg {
style = interpolateStyle(b.TopLeftStyle, b.TopRightStyle, float64(i)/float64(width))
}
scr.SetCell(x, area.Min.Y, &uv.Cell{
Content: b.Style.Top,
Width:   1,
Style:   style,
})
}
}

if b.Bottom {
width := area.Max.X - area.Min.X - 2
for i, x := 0, area.Min.X+1; x < area.Max.X-1; i, x = i+1, x+1 {
style := b.BottomStyle
// Apply gradient if both corners have different colors
if b.BottomLeftStyle.Fg != nil && b.BottomRightStyle.Fg != nil && b.BottomLeftStyle.Fg != b.BottomRightStyle.Fg {
style = interpolateStyle(b.BottomLeftStyle, b.BottomRightStyle, float64(i)/float64(width))
}
scr.SetCell(x, area.Max.Y-1, &uv.Cell{
Content: b.Style.Bottom,
Width:   1,
Style:   style,
})
}
}

// Draw left and right edges with gradient support
if b.Left {
height := area.Max.Y - area.Min.Y - 2
for i, y := 0, area.Min.Y+1; y < area.Max.Y-1; i, y = i+1, y+1 {
style := b.LeftStyle
// Apply gradient if both corners have different colors
if b.TopLeftStyle.Fg != nil && b.BottomLeftStyle.Fg != nil && b.TopLeftStyle.Fg != b.BottomLeftStyle.Fg {
style = interpolateStyle(b.TopLeftStyle, b.BottomLeftStyle, float64(i)/float64(height))
}
scr.SetCell(area.Min.X, y, &uv.Cell{
Content: b.Style.Left,
Width:   1,
Style:   style,
})
}
}

if b.Right {
height := area.Max.Y - area.Min.Y - 2
for i, y := 0, area.Min.Y+1; y < area.Max.Y-1; i, y = i+1, y+1 {
style := b.RightStyle
// Apply gradient if both corners have different colors
if b.TopRightStyle.Fg != nil && b.BottomRightStyle.Fg != nil && b.TopRightStyle.Fg != b.BottomRightStyle.Fg {
style = interpolateStyle(b.TopRightStyle, b.BottomRightStyle, float64(i)/float64(height))
}
scr.SetCell(area.Max.X-1, y, &uv.Cell{
Content: b.Style.Right,
Width:   1,
Style:   style,
})
}
}
}

// interpolateStyle interpolates between two styles based on a ratio (0.0 to 1.0).
func interpolateStyle(start, end uv.Style, ratio float64) uv.Style {
result := start

// Interpolate foreground color if both are set
if start.Fg != nil && end.Fg != nil {
r1, g1, b1, _ := start.Fg.RGBA()
r2, g2, b2, _ := end.Fg.RGBA()

// Convert from 16-bit to 8-bit range
c1 := colorful.Color{R: float64(r1) / 65535.0, G: float64(g1) / 65535.0, B: float64(b1) / 65535.0}
c2 := colorful.Color{R: float64(r2) / 65535.0, G: float64(g2) / 65535.0, B: float64(b2) / 65535.0}
interpolated := c1.BlendLab(c2, ratio)
result.Fg = ansi.HexColor(interpolated.Hex())
}

// Interpolate background color if both are set
if start.Bg != nil && end.Bg != nil {
r1, g1, b1, _ := start.Bg.RGBA()
r2, g2, b2, _ := end.Bg.RGBA()

// Convert from 16-bit to 8-bit range
c1 := colorful.Color{R: float64(r1) / 65535.0, G: float64(g1) / 65535.0, B: float64(b1) / 65535.0}
c2 := colorful.Color{R: float64(r2) / 65535.0, G: float64(g2) / 65535.0, B: float64(b2) / 65535.0}
interpolated := c1.BlendLab(c2, ratio)
result.Bg = ansi.HexColor(interpolated.Hex())
}

return result
}
