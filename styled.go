package tv

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// StyledString is a string that can be decomposed into a series of styled
// lines and cells. It is used to disassemble a rendered string with ANSI
// escape codes into a series of cells that can be used in a [Buffer].
// A StyledString supports reading [ansi.SGR] and [ansi.Hyperlink] escape
// codes.
type StyledString struct {
	// The width method used to calculate the width of each unicode character
	// and grapheme cluster.
	Method ansi.Method
	// Text is the original string that was used to create the styled string.
	Text string
	// Wrap determines whether the styled string should wrap to the next line.
	Wrap bool
	// Tail is the string that will be appended to the end of the line when the
	// string is truncated i.e. when [StyledString.Wrap] is false.
	Tail string
}

// NewStyledString creates a new [StyledString] for the given method and styled
// string. The method is used to calculate the width of each line.
func NewStyledString(method ansi.Method, str string) *StyledString {
	ss := new(StyledString)
	ss.Method = method
	ss.Text = str
	return ss
}

// Display renders the styled string to the given buffer at the specified area.
func (s *StyledString) Display(buf *Buffer, area Rectangle) error {
	if buf == nil {
		return fmt.Errorf("buffer cannot be nil")
	}
	// Clear the area before drawing.
	buf.FillArea(nil, area)
	str := s.Text
	// We need to normalize newlines "\n" to "\r\n" to emulate a raw terminal
	// output.
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\n", "\r\n")
	printString(buf, s.Method, area.Min.X, area.Min.Y, area, str, !s.Wrap, s.Tail)
	return nil
}

// Bounds returns the bounds of the styled string. This is the rectangle
// that contains the entire styled string, including all lines and cells.
func (s *StyledString) Bounds() Rectangle {
	var w, h int
	lines := strings.Split(s.Text, "\n")
	h = len(lines)
	for _, l := range lines {
		width := s.Method.StringWidth(l)
		if width > w {
			w = width
		}
	}
	return Rect(0, 0, w, h)
}

func newWcCell(s string, style *Style, link *Link) Cell {
	var c Cell
	for i, r := range s {
		if i == 0 {
			c.Rune = r
			// We only care about the first rune's width
			c.Width = runewidth.RuneWidth(r)
		} else {
			if runewidth.RuneWidth(r) > 0 {
				break
			}
			c.Comb = append(c.Comb, r)
		}
	}
	if style != nil {
		c.Style = *style
	}
	if link != nil {
		c.Link = *link
	}
	return c
}

func newGCell(s string, style *Style, link *Link) Cell {
	var c Cell
	g, _, w, _ := uniseg.FirstGraphemeClusterInString(s, -1)
	c.Width = w
	for i, r := range g {
		if i == 0 {
			c.Rune = r
		} else {
			c.Comb = append(c.Comb, r)
		}
	}
	if style != nil {
		c.Style = *style
	}
	if link != nil {
		c.Link = *link
	}
	return c
}

// printString draws a string starting at the given position.
func printString[T []byte | string](
	s *Buffer,
	m ansi.Method,
	x, y int,
	bounds Rectangle, str T,
	truncate bool, tail string,
) {
	p := ansi.NewParser()

	var tailc Cell
	if truncate && len(tail) > 0 {
		if m == ansi.WcWidth {
			tailc = newWcCell(tail, nil, nil)
		} else {
			tailc = newGCell(tail, nil, nil)
		}
	}

	decoder := ansi.DecodeSequenceWc[T]
	if m == ansi.GraphemeWidth {
		decoder = ansi.DecodeSequence[T]
	}

	var cell Cell
	var style Style
	var link Link
	var state byte
	for len(str) > 0 {
		seq, width, n, newState := decoder(str, state, p)
		switch width {
		case 1, 2, 3, 4: // wide cells can go up to 4 cells wide
			cell.Width = width
			cell.SetString(string(seq))

			if !truncate && x+cell.Width > bounds.Max.X && y+1 < bounds.Max.Y {
				// Wrap the string to the width of the window
				x = bounds.Min.X
				y++
			}

			pos := Pos(x, y)
			if pos.In(bounds) {
				if truncate && tailc.Width > 0 && x+cell.Width > bounds.Max.X-tailc.Width {
					// Truncate the string and append the tail if any.
					cell = tailc
					cell.Style = style
					cell.Link = link
					s.SetCell(x, y, &cell)
					x += tailc.Width
				} else {
					// Print the cell to the screen
					cell.Style = style
					cell.Link = link
					s.SetCell(x, y, &cell) //nolint:errcheck
					x += width
				}
			}

			// String is too long for the line, truncate it.
			// Make sure we reset the cell for the next iteration.
			cell.Reset()
		default:
			// Valid sequences always have a non-zero Cmd.
			// TODO: Handle cursor movement and other sequences
			switch {
			case ansi.HasCsiPrefix(seq) && p.Command() == 'm':
				// SGR - Select Graphic Rendition
				ReadStyle(p.Params(), &style)
			case ansi.HasOscPrefix(seq) && p.Command() == 8:
				// Hyperlinks
				ReadLink(p.Data(), &link)
			case ansi.Equal(seq, T("\n")):
				y++
			case ansi.Equal(seq, T("\r")):
				x = bounds.Min.X
			default:
				cell.Append([]rune(string(seq))...)
			}
		}

		// Advance the state and data
		state = newState
		str = str[n:]
	}

	// Make sure to set the last cell if it's not empty.
	if !cell.Empty() {
		s.SetCell(x, y, &cell) //nolint:errcheck
		cell.Reset()
	}
}

// ReadStyle reads a Select Graphic Rendition (SGR) escape sequences from a
// list of parameters into pen.
func ReadStyle(params ansi.Params, pen *Style) {
	if len(params) == 0 {
		pen.Reset()
		return
	}

	for i := 0; i < len(params); i++ {
		param, hasMore, _ := params.Param(i, 0)
		switch param {
		case 0: // Reset
			pen.Reset()
		case 1: // Bold
			pen.Bold(true)
		case 2: // Dim/Faint
			pen.Faint(true)
		case 3: // Italic
			pen.Italic(true)
		case 4: // Underline
			nextParam, _, ok := params.Param(i+1, 0)
			if hasMore && ok { // Only accept subparameters i.e. separated by ":"
				switch nextParam {
				case 0, 1, 2, 3, 4, 5:
					i++
					switch nextParam {
					case 0: // No Underline
						pen.UnderlineStyle(NoUnderline)
					case 1: // Single Underline
						pen.UnderlineStyle(SingleUnderline)
					case 2: // Double Underline
						pen.UnderlineStyle(DoubleUnderline)
					case 3: // Curly Underline
						pen.UnderlineStyle(CurlyUnderline)
					case 4: // Dotted Underline
						pen.UnderlineStyle(DottedUnderline)
					case 5: // Dashed Underline
						pen.UnderlineStyle(DashedUnderline)
					}
				}
			} else {
				// Single Underline
				pen.UnderlineStyle(SingleUnderline)
			}
		case 5: // Slow Blink
			pen.SlowBlink(true)
		case 6: // Rapid Blink
			pen.RapidBlink(true)
		case 7: // Reverse
			pen.Reverse(true)
		case 8: // Conceal
			pen.Conceal(true)
		case 9: // Crossed-out/Strikethrough
			pen.Strikethrough(true)
		case 22: // Normal Intensity (not bold or faint)
			pen.Bold(false).Faint(false)
		case 23: // Not italic, not Fraktur
			pen.Italic(false)
		case 24: // Not underlined
			pen.UnderlineStyle(NoUnderline)
		case 25: // Blink off
			pen.SlowBlink(false).RapidBlink(false)
		case 27: // Positive (not reverse)
			pen.Reverse(false)
		case 28: // Reveal
			pen.Conceal(false)
		case 29: // Not crossed out
			pen.Strikethrough(false)
		case 30, 31, 32, 33, 34, 35, 36, 37: // Set foreground
			pen.Foreground(ansi.Black + ansi.BasicColor(param-30)) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Foreground(c)
				i += n - 1
			}
		case 39: // Default foreground
			pen.Foreground(nil)
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			pen.Background(ansi.Black + ansi.BasicColor(param-40)) //nolint:gosec
		case 48: // Set background 256 or truecolor
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Background(c)
				i += n - 1
			}
		case 49: // Default Background
			pen.Background(nil)
		case 58: // Set underline color
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Underline(c)
				i += n - 1
			}
		case 59: // Default underline color
			pen.Underline(nil)
		case 90, 91, 92, 93, 94, 95, 96, 97: // Set bright foreground
			pen.Foreground(ansi.BrightBlack + ansi.BasicColor(param-90)) //nolint:gosec
		case 100, 101, 102, 103, 104, 105, 106, 107: // Set bright background
			pen.Background(ansi.BrightBlack + ansi.BasicColor(param-100)) //nolint:gosec
		}
	}
}

// ReadLink reads a hyperlink escape sequence from a data buffer into link.
func ReadLink(p []byte, link *Link) {
	params := bytes.Split(p, []byte{';'})
	if len(params) != 3 {
		return
	}
	link.Params = string(params[1])
	link.URL = string(params[2])
}
