package screen

import (
	"fmt"
	"image/color"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/clipperhouse/uax29/v2/graphemes"
)

// Context represents a drawing context for rendering operations on a screen.
type Context struct {
	scr uv.Screen

	style uv.Style
	link  uv.Link
	pos   uv.Position
}

// NewContext creates a new drawing context for the given screen.
func NewContext(scr uv.Screen) *Context {
	c := &Context{scr: scr}
	c.Reset()
	return c
}

// Reset resets the context to its default state.
func (c *Context) Reset() {
	c.style = uv.Style{}
	c.link = uv.Link{}
	c.pos = uv.Position{X: 0, Y: 0}
}

// SetStyle sets the style of the context.
func (c *Context) SetStyle(style uv.Style) {
	c.style = style
}

// WithStyle returns a copy of the context with the given style.
func (c Context) WithStyle(style uv.Style) Context {
	c.SetStyle(style)
	return c
}

// SetLink sets the link of the context.
func (c *Context) SetLink(link uv.Link) {
	c.link = link
}

// WithLink returns a copy of the context with the given link.
func (c Context) WithLink(link uv.Link) Context {
	c.SetLink(link)
	return c
}

// SetAttrs sets the attributes of the context.
func (c *Context) SetAttrs(attrs uint8) {
	c.style.Attrs = attrs
}

// WithAttrs returns a copy of the context with the given attributes.
func (c Context) WithAttrs(attrs uint8) Context {
	c.SetAttrs(attrs)
	return c
}

// SetBackground sets the background color of the context. Use nil to reset to default.
func (c *Context) SetBackground(bg color.Color) {
	c.style.Bg = bg
}

// WithBackground returns a copy of the context with the given background color.
func (c Context) WithBackground(bg color.Color) Context {
	c.SetBackground(bg)
	return c
}

// SetForeground sets the foreground color of the context. Use nil to reset to default.
func (c *Context) SetForeground(fg color.Color) {
	c.style.Fg = fg
}

// WithForeground returns a copy of the context with the given foreground color.
func (c Context) WithForeground(fg color.Color) Context {
	c.SetForeground(fg)
	return c
}

// SetBold sets whether the text in the context should be bold.
func (c *Context) SetBold(bold bool) {
	if bold {
		c.style.Attrs |= uv.AttrBold
	} else {
		c.style.Attrs &^= uv.AttrBold
	}
}

// WithBold returns a copy of the context with the given bold attribute.
func (c Context) WithBold(bold bool) Context {
	c.SetBold(bold)
	return c
}

// SetItalic sets whether the text in the context should be italic.
func (c *Context) SetItalic(italic bool) {
	if italic {
		c.style.Attrs |= uv.AttrItalic
	} else {
		c.style.Attrs &^= uv.AttrItalic
	}
}

// WithItalic returns a copy of the context with the given italic attribute.
func (c Context) WithItalic(italic bool) Context {
	c.SetItalic(italic)
	return c
}

// SetStrikethrough sets whether the text in the context should be strikethrough.
func (c *Context) SetStrikethrough(strikethrough bool) {
	if strikethrough {
		c.style.Attrs |= uv.AttrStrikethrough
	} else {
		c.style.Attrs &^= uv.AttrStrikethrough
	}
}

// WithStrikethrough returns a copy of the context with the given strikethrough
// attribute.
func (c Context) WithStrikethrough(strikethrough bool) Context {
	c.SetStrikethrough(strikethrough)
	return c
}

// SetFaint sets whether the text in the context should be faint.
func (c *Context) SetFaint(faint bool) {
	if faint {
		c.style.Attrs |= uv.AttrFaint
	} else {
		c.style.Attrs &^= uv.AttrFaint
	}
}

// WithFaint returns a copy of the context with the given faint attribute.
func (c Context) WithFaint(faint bool) Context {
	c.SetFaint(faint)
	return c
}

// SetBlink sets whether the text in the context should blink.
func (c *Context) SetBlink(blink bool) {
	if blink {
		c.style.Attrs |= uv.AttrBlink
	} else {
		c.style.Attrs &^= uv.AttrBlink
	}
}

// WithBlink returns a copy of the context with the given blink attribute.
func (c Context) WithBlink(blink bool) Context {
	c.SetBlink(blink)
	return c
}

// SetReverse sets whether the text in the context should be reversed.
func (c *Context) SetReverse(reverse bool) {
	if reverse {
		c.style.Attrs |= uv.AttrReverse
	} else {
		c.style.Attrs &^= uv.AttrReverse
	}
}

// WithReverse returns a copy of the context with the given reverse attribute.
func (c Context) WithReverse(reverse bool) Context {
	c.SetReverse(reverse)
	return c
}

// SetConceal sets whether the text in the context should be concealed.
func (c *Context) SetConceal(conceal bool) {
	if conceal {
		c.style.Attrs |= uv.AttrConceal
	} else {
		c.style.Attrs &^= uv.AttrConceal
	}
}

// WithConceal returns a copy of the context with the given conceal attribute.
func (c Context) WithConceal(conceal bool) Context {
	c.SetConceal(conceal)
	return c
}

// SetUnderlineStyle sets the underline style of the context.
func (c *Context) SetUnderlineStyle(u uv.Underline) {
	c.style.Underline = u
}

// WithUnderlineStyle returns a copy of the context with the given underline style.
func (c Context) WithUnderlineStyle(u uv.Underline) Context {
	c.SetUnderlineStyle(u)
	return c
}

// SetUnderline sets whether the text in the context should be underlined.
//
// This is a convenience method that sets the underline style to single or
// none. It is equivalent to calling [Context.SetUnderlineStyle] with
// [uv.UnderlineSingle] or [uv.UnderlineNone].
func (c *Context) SetUnderline(underline bool) {
	if underline {
		c.SetUnderlineStyle(uv.UnderlineSingle)
	} else {
		c.SetUnderlineStyle(uv.UnderlineNone)
	}
}

// WithUnderline returns a copy of the context with the given underline
// attribute.
func (c Context) WithUnderline(underline bool) Context {
	c.SetUnderline(underline)
	return c
}

// SetUnderlineColor sets the underline color of the context. Use nil to reset to default.
func (c *Context) SetUnderlineColor(color color.Color) {
	c.style.UnderlineColor = color
}

// WithUnderlineColor returns a copy of the context with the given underline color.
func (c Context) WithUnderlineColor(color color.Color) Context {
	c.SetUnderlineColor(color)
	return c
}

// SetURL sets the URL link for the context. Use an empty string to reset.
func (c *Context) SetURL(url string, params ...string) {
	if url == "" {
		c.link = uv.Link{}
		return
	}
	c.link = uv.Link{
		URL:    url,
		Params: strings.Join(params, ":"),
	}
}

// WithURL returns a copy of the context with the given URL link.
func (c Context) WithURL(url string, params ...string) Context {
	c.SetURL(url, params...)
	return c
}

// Position returns the current position of the context.
func (c *Context) Position() uv.Position {
	return c.pos
}

// SetPosition moves the current position of the context cursor to the given
// coordinates.
//
// This is an alias for [Context.MoveTo].
func (c *Context) SetPosition(x, y int) {
	c.MoveTo(x, y)
}

// WithPosition returns a copy of the context with the given position.
func (c Context) WithPosition(x, y int) Context {
	c.MoveTo(x, y)
	return c
}

// MoveTo moves the current position of the context cursor to the given
// coordinates.
func (c *Context) MoveTo(x, y int) {
	c.pos.X = x
	c.pos.Y = y
}

// Print prints the given string to the screen at the current position, updating
// the position accordingly.
func (c *Context) Print(v ...any) (int, error) {
	return fmt.Fprint(c, v...)
}

// Println prints the given string to the screen at the current position, appending
// a newline, and updating the position accordingly.
func (c *Context) Println(v ...any) (int, error) {
	return fmt.Fprintln(c, v...)
}

// Printf formats according to a format specifier and writes to the screen at the
// current position, updating the position accordingly.
func (c *Context) Printf(format string, a ...any) (int, error) {
	return fmt.Fprintf(c, format, a...)
}

// DrawString draws the given string at the given position with the current
// style and link, cropping the string when it reaches the edge of the screen.
func (c *Context) DrawString(s string, x, y int) {
	drawStringAt(c.scr, graphemes.FromString(s), x, y, c.style, c.link, false)
}

// DrawStringWrapped draws the given string at the given position with the current
// style and link, wrapping the string when it reaches the edge of the screen.
func (c *Context) DrawStringWrapped(s string, x, y int) {
	drawStringAt(c.scr, graphemes.FromString(s), x, y, c.style, c.link, true)
}

// Write implements the [io.Writer] interface for the context, writing the
// given byte slice to the screen at the current position, updating the
// position accordingly.
func (c *Context) Write(p []byte) (n int, err error) {
	c.pos.X, c.pos.Y = drawStringAt(c.scr, graphemes.FromBytes(p), c.pos.X, c.pos.Y, c.style, c.link, true)
	return len(p), nil
}

// WriteString implements the [io.StringWriter] interface for the context,
// writing the given string to the screen at the current position, updating the
// position accordingly.
func (c *Context) WriteString(s string) (n int, err error) {
	c.pos.X, c.pos.Y = drawStringAt(c.scr, graphemes.FromString(s), c.pos.X, c.pos.Y, c.style, c.link, true)
	return len(s), nil
}

func drawStringAt[T []byte | string](scr uv.Screen, grs *graphemes.Iterator[T], x, y int, style uv.Style, link uv.Link, wrap bool) (int, int) {
	bounds := scr.Bounds()
	bounds.Max.X -= bounds.Min.X
	bounds.Max.Y -= bounds.Min.Y
	bounds.Min.X = 0
	bounds.Min.Y = 0
	pos := uv.Pos(x, y)
	if !pos.In(bounds) {
		return x, y
	}

	wm := scr.WidthMethod()
	for grs.Next() {
		gr := string(grs.Value())
		switch gr {
		case "\n":
			x = bounds.Min.X
			y++
			continue
		}

		w := wm.StringWidth(gr)
		pos := uv.Pos(x, y)
		if x+w > bounds.Max.X {
			if wrap {
				x = bounds.Min.X
				y++
				pos = uv.Pos(x, y)
			} else {
				break
			}
		}
		if !pos.In(bounds) {
			break
		}

		scr.SetCell(x, y, &uv.Cell{
			Content: gr,
			Width:   w,
			Style:   style,
			Link:    link,
		})

		x += w
		if wrap && x >= bounds.Max.X {
			x = bounds.Min.X
			y++
		}
	}

	return x, y
}
