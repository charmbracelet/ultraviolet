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
	c := &Context{
		scr: scr,
	}
	return c
}

// SetStyle sets the style of the context.
func (c *Context) SetStyle(style uv.Style) {
	c.style = style
}

// SetLink sets the link of the context.
func (c *Context) SetLink(link uv.Link) {
	c.link = link
}

// SetBackground sets the background color of the context. Use nil to reset to default.
func (c *Context) SetBackground(bg color.Color) {
	c.style.Bg = bg
}

// SetForeground sets the foreground color of the context. Use nil to reset to default.
func (c *Context) SetForeground(fg color.Color) {
	c.style.Fg = fg
}

// SetBold sets whether the text in the context should be bold.
func (c *Context) SetBold(bold bool) {
	if bold {
		c.style.Attrs |= uv.AttrBold
	} else {
		c.style.Attrs &^= uv.AttrBold
	}
}

// SetItalic sets whether the text in the context should be italic.
func (c *Context) SetItalic(italic bool) {
	if italic {
		c.style.Attrs |= uv.AttrItalic
	} else {
		c.style.Attrs &^= uv.AttrItalic
	}
}

// SetStrikethrough sets whether the text in the context should be strikethrough.
func (c *Context) SetStrikethrough(strikethrough bool) {
	if strikethrough {
		c.style.Attrs |= uv.AttrStrikethrough
	} else {
		c.style.Attrs &^= uv.AttrStrikethrough
	}
}

// SetFaint sets whether the text in the context should be faint.
func (c *Context) SetFaint(faint bool) {
	if faint {
		c.style.Attrs |= uv.AttrFaint
	} else {
		c.style.Attrs &^= uv.AttrFaint
	}
}

// SetBlink sets whether the text in the context should blink.
func (c *Context) SetBlink(blink bool) {
	if blink {
		c.style.Attrs |= uv.AttrBlink
	} else {
		c.style.Attrs &^= uv.AttrBlink
	}
}

// SetReverse sets whether the text in the context should be reversed.
func (c *Context) SetReverse(reverse bool) {
	if reverse {
		c.style.Attrs |= uv.AttrReverse
	} else {
		c.style.Attrs &^= uv.AttrReverse
	}
}

// SetConceal sets whether the text in the context should be concealed.
func (c *Context) SetConceal(conceal bool) {
	if conceal {
		c.style.Attrs |= uv.AttrConceal
	} else {
		c.style.Attrs &^= uv.AttrConceal
	}
}

// SetUnderlineStyle sets the underline style of the context.
func (c *Context) SetUnderlineStyle(u uv.Underline) {
	c.style.Underline = u
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

// SetUnderlineColor sets the underline color of the context. Use nil to reset to default.
func (c *Context) SetUnderlineColor(color color.Color) {
	c.style.UnderlineColor = color
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

// Position returns the current position of the context.
func (c *Context) Position() uv.Position {
	return c.pos
}

// SetPosition moves the current position of the context to the given
// coordinates.
func (c *Context) SetPosition(x, y int) {
	c.pos.X = x
	c.pos.Y = y
}

// Print prints the given string to the screen at the current position, updating
// the position accordingly.
func (c *Context) Print(v ...any) {
	s := fmt.Sprint(v...)
	c.pos.X, c.pos.Y = drawStringAt(c.scr, s, c.pos.X, c.pos.Y, c.style, c.link, true)
}

// Println prints the given string to the screen at the current position, appending
// a newline, and updating the position accordingly.
func (c *Context) Println(v ...any) {
	s := fmt.Sprintln(v...)
	c.pos.X, c.pos.Y = drawStringAt(c.scr, s, c.pos.X, c.pos.Y, c.style, c.link, true)
}

// Printf formats according to a format specifier and writes to the screen at the
// current position, updating the position accordingly.
func (c *Context) Printf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	c.pos.X, c.pos.Y = drawStringAt(c.scr, s, c.pos.X, c.pos.Y, c.style, c.link, true)
}

// DrawString draws the given string at the given position with the current
// style and link, cropping the string when it reaches the edge of the screen.
func (c *Context) DrawString(s string, x, y int) {
	drawStringAt(c.scr, s, x, y, c.style, c.link, false)
}

// DrawStringWrapped draws the given string at the given position with the current
// style and link, wrapping the string when it reaches the edge of the screen.
func (c *Context) DrawStringWrapped(s string, x, y int) {
	drawStringAt(c.scr, s, x, y, c.style, c.link, true)
}

func drawStringAt(scr uv.Screen, s string, x, y int, style uv.Style, link uv.Link, wrap bool) (int, int) {
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
	grs := graphemes.FromString(s)
	for grs.Next() {
		gr := grs.Value()
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
