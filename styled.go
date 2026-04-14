package uv

import (
	"bytes"
	"image/color"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// StyledString is a string that can be decomposed into a series of styled
// lines and cells. It is used to disassemble a rendered string with ANSI
// escape codes into a series of cells that can be used in a [Buffer].
// A StyledString supports reading [ansi.SGR] and [ansi.Hyperlink] escape
// codes.
type StyledString struct {
	// Text is the original string that was used to create the styled string.
	Text string
	// Wrap determines whether the styled string should wrap to the next line.
	Wrap bool
	// Tail is the string that will be appended to the end of the line when the
	// string is truncated i.e. when [StyledString.Wrap] is false.
	Tail string
}

var _ Drawable = (*StyledString)(nil)

// NewStyledString creates a new [StyledString] for the given method and styled
// string. The method is used to calculate the width of each line.
func NewStyledString(str string) *StyledString {
	ss := new(StyledString)
	ss.Text = str
	return ss
}

// String returns the text of the styled string.
//
// It implements the [fmt.Stringer] interface.
func (s *StyledString) String() string {
	return s.Text
}

// Lines returns the styled string decomposed into a slice of [Line]s.
func (s *StyledString) Lines(m ansi.Method) []Line {
	return printString(nil, m, 0, 0, Rectangle{}, s.Text, false, "")
}

// Draw renders the styled string to the given buffer at the
// specified area.
func (s *StyledString) Draw(buf Screen, area Rectangle) {
	// Clear the area before drawing.
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			buf.SetCell(x, y, nil)
		}
	}
	str := s.Text
	// We need to normalize newlines "\n" to "\r\n" to emulate a raw terminal
	// output.
	str = strings.ReplaceAll(str, "\r\n", "\n")
	printString(buf, buf.WidthMethod(), area.Min.X, area.Min.Y, area, str, !s.Wrap, s.Tail)
}

// Height returns the number of lines in the styled string. This is the number
// of lines that the styled string will occupy when rendered to the screen.
func (s *StyledString) Height() int {
	return strings.Count(s.Text, "\n") + 1
}

// UnicodeWidth returns the cells width of the widest line in the styled string
// using the [ansi.GraphemeWidth] method.
func (s *StyledString) UnicodeWidth() int {
	w, _ := s.widthHeight(ansi.GraphemeWidth)
	return w
}

// WcWidth returns the cells width of the widest line in the styled string
// using the [ansi.WcWidth] method.
func (s *StyledString) WcWidth() int {
	w, _ := s.widthHeight(ansi.WcWidth)
	return w
}

func (s *StyledString) widthHeight(m ansi.Method) (w, h int) {
	lines := strings.Split(s.Text, "\n")
	h = len(lines)
	for _, l := range lines {
		w = max(w, m.StringWidth(l))
	}
	return
}

// Bounds returns the minimum area that can contain the whole styled string.
func (s *StyledString) Bounds() Rectangle {
	w, h := s.widthHeight(ansi.GraphemeWidth)
	return Rect(0, 0, w, h)
}

// multicellRect describes a single-row span of cells that have been claimed
// by a previously-emitted scaled multicell glyph in this printString call.
// Subsequent printable writes that intersect any rect are skipped so the
// glyph's spill rows are not overwritten and erased per the kitty
// text-sizing protocol.
type multicellRect struct{ x, y, w int }

// hasStringPrefix reports whether b begins with any of the five
// string-state introducers (OSC, DCS, APC, SOS, PM) in either the 7-bit
// (ESC + X) or 8-bit (single C1 byte) form. All five share the same
// terminator semantics.
func hasStringPrefix[T []byte | string](b T) bool {
	return ansi.HasOscPrefix(b) ||
		ansi.HasDcsPrefix(b) ||
		ansi.HasApcPrefix(b) ||
		ansi.HasSosPrefix(b) ||
		ansi.HasPmPrefix(b)
}

// findStringEnd scans src from offset `start` for a proper string-state
// terminator, BEL (0x07) or 7-bit ST (ESC 0x1B followed by 0x5C), and
// returns the index *after* the terminator, or `start` if the sequence
// is abandoned by a bare ESC.
//
// Critically, it does *not* treat bare 8-bit ST (0x9C) as a terminator.
// Of every byte the ansi decoder's StringState recognises as "end of
// string" (BEL, CAN, SUB, ESC, and ST), only 0x9C falls in the UTF-8
// continuation-byte range (0x80-0xBF). Nerd-Font icons like U+E738
// (encoded as 0xEE 0x9C 0xB8) therefore carry a byte the decoder
// misinterprets as ST, truncating the payload mid-UTF-8. The workaround
// is to treat 0x9C as data for payloads we know to be UTF-8 and rely on
// the unambiguous 7-bit terminators (BEL, ESC+\) instead.
func findStringEnd[T []byte | string](src T, start int) int {
	for i := start; i < len(src); i++ {
		switch src[i] {
		case 0x07: // BEL
			return i + 1
		case 0x1b: // ESC
			if i+1 < len(src) && src[i+1] == 0x5c {
				return i + 2
			}
			// Bare ESC inside a string-state sequence; the ansi
			// decoder aborts here, so do the same.
			return i
		}
	}
	return start
}

func multicellOwned(rects []multicellRect, x, y int) bool {
	for _, r := range rects {
		if r.y == y && x >= r.x && x < r.x+r.w {
			return true
		}
	}
	return false
}

// printString draws a string starting at the given position. If s is nil, it
// will build and return a slice of [Line]s instead (unwrapped, ignoring bounds).
func printString[T []byte | string](
	s Screen,
	m WidthMethod,
	x, y int,
	bounds Rectangle, str T,
	truncate bool, tail string,
) (lines []Line) {
	var multicells []multicellRect
	p := ansi.GetParser()
	defer ansi.PutParser(p)

	var tailc Cell
	if truncate && len(tail) > 0 {
		tailc = *NewCell(m, tail)
	}

	decoder := ansi.DecodeSequenceWc[T]
	if m == ansi.GraphemeWidth {
		decoder = ansi.DecodeSequence[T]
	}

	if s == nil {
		lines = []Line{}
	}

	var cell Cell
	var style Style
	var link Link
	var state byte
	for len(str) > 0 {
		seq, width, n, newState := decoder(str, state, p)

		// Workaround for a shared-ansi-decoder bug affecting every
		// string-state sequence (OSC, DCS, APC, SOS, PM): the decoder
		// treats bare C1 ST (byte 0x9C) as a terminator, but 0x9C is
		// also a valid UTF-8 continuation byte (the whole 0x80-0xBF
		// range is). Nerd-Font icons like U+E738 (0xEE 0x9C 0xB8)
		// carry 0x9C as a middle byte, so an OSC 66 text-sizing
		// payload gets sliced mid-UTF-8; downstream terminals then see
		// malformed bytes and render the tail (SGRs, the next OSC,
		// etc.) as literal text.
		//
		// Of the decoder's five recognised terminators, BEL (0x07),
		// CAN (0x18), SUB (0x1A), ESC (0x1B), ST (0x9C), only 0x9C
		// overlaps with UTF-8 continuation bytes; the other four live
		// in 7-bit ASCII, which UTF-8 reserves for single-byte chars.
		//
		// When the decoder returns a string-state sequence that ended
		// on bare 0x9C and there's more input available, rescan from
		// that point for the unambiguous 7-bit terminator (BEL or
		// ESC+\) and extend the sequence. We don't re-run the parser
		// over the extended bytes: p.Command() and the metadata
		// portion of p.Data() were already resolved from bytes before
		// the UTF-8 payload began.
		if n > 0 && n < len(str) && hasStringPrefix(seq) && seq[len(seq)-1] == 0x9C {
			if ext := findStringEnd(str, n); ext > n {
				seq = str[:ext]
				n = ext
			}
		}
		switch width {
		case 1, 2, 3, 4: // wide cells can go up to 4 cells wide
			cell.Width = width
			cell.Content = string(seq)
			cell.Style = style
			cell.Link = link

			if s == nil {
				// Building lines: unwrapped, no bounds
				if y >= len(lines) {
					lines = append(lines, Line{})
				}
				lines[y] = append(lines[y], cell)
				x += width
			} else {
				// Drawing to screen: handle wrapping, truncation, and bounds
				if !truncate && x+cell.Width > bounds.Max.X && y+1 < bounds.Max.Y {
					// Wrap the string to the width of the window
					x = bounds.Min.X
					y++
				}

				pos := Pos(x, y)
				if pos.In(bounds) {
					switch {
					case multicellOwned(multicells, x, y):
						// Cell is owned by an earlier scaled multicell glyph
						// from this same string; writing here would erase it
						// per the kitty text-sizing rules. Step over.
						x += width
					case truncate && tailc.Width > 0 && x+cell.Width > bounds.Max.X-tailc.Width:
						// Truncate the string and append the tail if any.
						cell = tailc
						cell.Style = style
						cell.Link = link
						s.SetCell(x, y, &cell)
						x += tailc.Width
					default:
						// Print the cell to the screen
						s.SetCell(x, y, &cell)
						x += width
					}
				}
			}

			// Reset cell for next iteration
			cell = Cell{}
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
			case ansi.HasCsiPrefix(seq) && p.Command() == 'C':
				// CUF (cursor forward). Treat the escape as a width-N
				// placeholder cell that preserves the literal bytes so
				// they make it back out to the terminal verbatim. This
				// matches what we emit for the spill rows of a kitty
				// scaled glyph (see the OSC 66 case below); without
				// this round-trip handling, the canvas's CUF placeholders
				// would be silently dropped during re-parse and the cells
				// after them would shift left into the multicell's spill,
				// erasing the glyph.
				n := 1
				if v, _, ok := p.Params().Param(0, 1); ok && v > 0 {
					n = v
				}
				if n > 4 {
					n = 4
				}
				cufCell := Cell{
					Content: string(seq),
					Width:   n,
					Style:   style,
				}
				if s == nil {
					if y >= len(lines) {
						lines = append(lines, Line{})
					}
					lines[y] = append(lines[y], cufCell)
					x += n
				} else {
					// If this position is already claimed by an OSC 66's
					// spill tracking from the same string, the cell is
					// already correctly seeded, so skip the re-write to
					// avoid a second SetCell (with a different style
					// captured later in parsing) and a duplicate
					// multicell rect. We still advance x so the rest of
					// the line lines up.
					if !multicellOwned(multicells, x, y) {
						if pos := Pos(x, y); pos.In(bounds) {
							if !truncate || x+n <= bounds.Max.X {
								s.SetCell(x, y, &cufCell)
							}
						}
						multicells = append(multicells, multicellRect{x: x, y: y, w: n})
					}
					x += n
				}
			case ansi.HasOscPrefix(seq) && p.Command() == 66:
				// Kitty text-sizing protocol. See:
				// https://sw.kovidgoyal.net/kitty/text-sizing-protocol/
				//
				// We forward the entire escape verbatim through a single
				// wide cell so the terminal can render the scaled glyph,
				// while reserving the right number of cells for layout.
				//
				// When the glyph also spans multiple rows (s>1 or h>1),
				// we additionally seed CUF "cursor forward" placeholder
				// cells in the spill rows. The kitty docs explicitly
				// recommend moving the cursor past a multicell when
				// writing the row below it; without this, any subsequent
				// write into the spill cells erases the entire glyph.
				w, h := readTextSizingDims(p.Data())
				if w < 1 {
					w = 1
				}
				if h < 1 {
					h = 1
				}
				if w > 4 {
					w = 4
				}
				if h > 4 {
					h = 4
				}
				tsCell := Cell{
					Content: string(seq),
					Width:   w,
					Style:   style,
					Link:    link,
				}
				if s == nil {
					if y >= len(lines) {
						lines = append(lines, Line{})
					}
					lines[y] = append(lines[y], tsCell)
					x += w
				} else {
					placed := false
					if pos := Pos(x, y); pos.In(bounds) {
						if !truncate || x+w <= bounds.Max.X {
							s.SetCell(x, y, &tsCell)
							placed = true
						}
					}
					if placed && h > 1 {
						// Skip cells deliberately carry NO style. When rendered
						// back out they become `CUF(w)` with a preceding style
						// reset, so no fg/bg/bold changes land at a position
						// that is a multicell extension in the terminal. Some
						// kitty builds mis-render SGR bytes emitted
						// inside/adjacent to a multicell extension as literal
						// text; keeping the span around the CUF stylistically
						// neutral avoids that whole class of parser state
						// confusion. The visible result is identical because
						// CUF doesn't paint cells.
						skip := Cell{
							Content: ansi.CursorForward(w),
							Width:   w,
						}
						for dy := 1; dy < h; dy++ {
							ny := y + dy
							if pos := Pos(x, ny); !pos.In(bounds) {
								break
							}
							s.SetCell(x, ny, &skip)
							multicells = append(multicells, multicellRect{x: x, y: ny, w: w})
						}
					}
					if placed {
						x += w
					}
				}
			case ansi.Equal(seq, T("\n")):
				if s == nil {
					// When building lines, we need to ensure empty lines are represented.
					if y >= len(lines) {
						lines = append(lines, Line{})
					}
				}
				y++
				// Always treat a NL as CR-LF similar to Termios ONLCR.
				fallthrough
			case ansi.Equal(seq, T("\r")):
				if s == nil {
					x = 0
				} else {
					x = bounds.Min.X
				}
			default:
				cell.Content += string(seq)
			}
		}

		// Advance the state and data
		state = newState
		str = str[n:]

		if y >= bounds.Max.Y {
			// We've reached the bottom of the bounds, stop processing further
			// lines.
			break
		}
	}

	// Make sure to set the last cell if it's not empty.
	if !cell.IsZero() && s != nil {
		s.SetCell(x, y, &cell)
	}

	return lines
}

// ReadStyle reads a Select Graphic Rendition (SGR) escape sequences from a
// list of parameters into pen.
func ReadStyle(params ansi.Params, pen *Style) {
	if len(params) == 0 {
		*pen = Style{}
		return
	}

	for i := 0; i < len(params); i++ {
		param, hasMore, _ := params.Param(i, 0)
		switch param {
		case 0: // Reset
			*pen = Style{}
		case 1: // Bold
			pen.Attrs |= AttrBold
		case 2: // Dim/Faint
			pen.Attrs |= AttrFaint
		case 3: // Italic
			pen.Attrs |= AttrItalic
		case 4: // Underline
			nextParam, _, ok := params.Param(i+1, 0)
			if hasMore && ok { // Only accept subparameters i.e. separated by ":"
				switch nextParam {
				case 0, 1, 2, 3, 4, 5:
					i++
					switch nextParam {
					case 0: // No Underline
						pen.Underline = UnderlineStyleNone
					case 1: // Single Underline
						pen.Underline = UnderlineStyleSingle
					case 2: // Double Underline
						pen.Underline = UnderlineStyleDouble
					case 3: // Curly Underline
						pen.Underline = UnderlineStyleCurly
					case 4: // Dotted Underline
						pen.Underline = UnderlineStyleDotted
					case 5: // Dashed Underline
						pen.Underline = UnderlineStyleDashed
					}
				}
			} else {
				// Single Underline
				pen.Underline = UnderlineStyleSingle
			}
		case 5: // Slow Blink
			pen.Attrs |= AttrBlink
		case 6: // Rapid Blink
			pen.Attrs |= AttrRapidBlink
		case 7: // Reverse
			pen.Attrs |= AttrReverse
		case 8: // Conceal
			pen.Attrs |= AttrConceal
		case 9: // Crossed-out/Strikethrough
			pen.Attrs |= AttrStrikethrough
		case 22: // Normal Intensity (not bold or faint)
			pen.Attrs &^= (AttrBold | AttrFaint)
		case 23: // Not italic, not Fraktur
			pen.Attrs &^= AttrItalic
		case 24: // Not underlined
			pen.Underline = UnderlineStyleNone
		case 25: // Blink off
			pen.Attrs &^= (AttrBlink | AttrRapidBlink)
		case 27: // Positive (not reverse)
			pen.Attrs &^= AttrReverse
		case 28: // Reveal
			pen.Attrs &^= AttrConceal
		case 29: // Not crossed out
			pen.Attrs &^= AttrStrikethrough
		case 30, 31, 32, 33, 34, 35, 36, 37: // Set foreground
			pen.Fg = ansi.Black + ansi.BasicColor(param-30) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Fg = c
				i += n - 1
			}
		case 39: // Default foreground
			pen.Fg = nil
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			pen.Bg = ansi.Black + ansi.BasicColor(param-40) //nolint:gosec
		case 48: // Set background 256 or truecolor
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.Bg = c
				i += n - 1
			}
		case 49: // Default Background
			pen.Bg = nil
		case 58: // Set underline color
			var c color.Color
			n := ansi.ReadStyleColor(params[i:], &c)
			if n > 0 {
				pen.UnderlineColor = c
				i += n - 1
			}
		case 59: // Default underline color
			pen.UnderlineColor = nil
		case 90, 91, 92, 93, 94, 95, 96, 97: // Set bright foreground
			pen.Fg = ansi.BrightBlack + ansi.BasicColor(param-90) //nolint:gosec
		case 100, 101, 102, 103, 104, 105, 106, 107: // Set bright background
			pen.Bg = ansi.BrightBlack + ansi.BasicColor(param-100) //nolint:gosec
		}
	}
}

// readTextSizingDims parses the metadata section of a kitty OSC 66
// text-sizing payload and returns the cell dimensions (width, height) the
// rendered glyph occupies.
//
// The data buffer holds the bytes between the OSC introducer and the string
// terminator, including the leading "66" command identifier:
//
//	66;<key=value[:key=value...]>;<text>
//
// Recognised metadata keys mirror the kitty docs:
//
//	s=<n>  scale factor; multiplies both the horizontal and vertical extent.
//	w=<n>  explicit cell width override.
//	h=<n>  explicit cell height override.
//
// When neither w nor s is supplied the width defaults to 1; the same applies
// to height. Multi-grapheme text payloads are not yet width-measured here;
// callers passing wider text should set w= explicitly until that work lands.
func readTextSizingDims(data []byte) (int, int) {
	parts := bytes.SplitN(data, []byte{';'}, 3)
	if len(parts) < 2 {
		return 1, 1
	}
	meta := parts[1]
	scale := 1
	explicitW := 0
	explicitH := 0
	for _, kv := range bytes.Split(meta, []byte{':'}) {
		eq := bytes.IndexByte(kv, '=')
		if eq < 1 {
			continue
		}
		key := string(kv[:eq])
		val, err := strconv.Atoi(string(kv[eq+1:]))
		if err != nil || val < 1 {
			continue
		}
		switch key {
		case "s":
			scale = val
		case "w":
			explicitW = val
		case "h":
			explicitH = val
		}
	}
	w := explicitW
	if w == 0 {
		w = scale
	}
	h := explicitH
	if h == 0 {
		h = scale
	}
	return w, h
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
