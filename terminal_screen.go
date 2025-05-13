package tv

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// ErrInvalidDimensions is returned when the dimensions of a window are invalid
// for the operation.
var ErrInvalidDimensions = errors.New("invalid dimensions")

// capabilities represents a mask of supported ANSI escape sequences.
type capabilities uint

const (
	// Vertical Position Absolute [ansi.VPA].
	capVPA capabilities = 1 << iota
	// Horizontal Position Absolute [ansi.HPA].
	capHPA
	// Cursor Horizontal Tab [ansi.CHT].
	capCHT
	// Cursor Backward Tab [ansi.CBT].
	capCBT
	// Repeat Previous Character [ansi.REP].
	capREP
	// Erase Character [ansi.ECH].
	capECH
	// Insert Character [ansi.ICH].
	capICH
	// Scroll Down [ansi.SD].
	capSD
	// Scroll Up [ansi.SU].
	capSU

	noCaps  capabilities = 0
	allCaps              = capVPA | capHPA | capCHT | capCBT | capREP | capECH | capICH |
		capSD | capSU
)

// Contains returns whether the capabilities contains the given capability.
func (v capabilities) Contains(c capabilities) bool {
	return v&c == c
}

// cursor represents a terminal cursor.
type cursor struct {
	Style
	Link
	Position
}

// lineData represents the metadata for a line.
type lineData struct {
	// first and last changed cell indices
	firstCell, lastCell int
	// old index used for scrolling
	oldIndex int //nolint:unused
}

// tFlag is a bitmask of terminal flags.
type tFlag uint

// Terminal writer flags.
const (
	tHardTabs tFlag = 1 << iota
	tBackspace
	tRelativeCursor
	tAltScreen
	tMapNewline
)

// Set sets the given flags.
func (v *tFlag) Set(c tFlag) {
	*v |= c
}

// Reset resets the given flags.
func (v *tFlag) Reset(c tFlag) {
	*v &^= c
}

// Contains returns whether the terminal flags contains the given flags.
func (v tFlag) Contains(c tFlag) bool {
	return v&c == c
}

// terminalWriter is a writer that buffers the output until it is flushed. It
// handles rendering [Buffer] cells to the screen and supports various
// terminal optimizations.
// The alt-screen and cursor visibility are not managed by the writer and
// should be managed by the caller. The writer however, will handle hiding the
// cursor during rendering before flushing the buffer when the cursor is
// set to be shown.
type terminalWriter struct {
	w                io.Writer
	buf              *bytes.Buffer // buffer for writing to the screen
	curbuf           *Buffer       // the current buffer
	tabs             *TabStops
	touch            sync.Map
	oldhash, newhash []uint64  // the old and new hash values for each line
	hashtab          []hashmap // the hashmap table
	oldnum           []int     // old indices from previous hash
	cur, saved       cursor    // the current and saved cursors
	flags            tFlag     // terminal writer flags.
	term             string    // the terminal type
	profile          colorprofile.Profile
	mu               sync.Mutex
	width            int          // the width of the terminal.
	scrollHeight     int          // keeps track of how many lines we've scrolled down (inline mode)
	clear            bool         // whether to force clear the screen
	caps             capabilities // terminal control sequence capabilities
	atPhantom        bool         // whether the cursor is out of bounds and at a phantom cell
}

// SetColorProfile sets the color profile to use when writing to the screen.
func (s *terminalWriter) SetColorProfile(p colorprofile.Profile) {
	s.mu.Lock()
	s.profile = p
	s.mu.Unlock()
}

// SetBackspace sets whether to use backspace as a movement optimization.
func (s *terminalWriter) SetBackspace(v bool) {
	s.mu.Lock()
	if v {
		s.flags.Set(tBackspace)
	} else {
		s.flags.Reset(tBackspace)
	}
	s.mu.Unlock()
}

// SetHardTabs sets whether to use hard tabs as movement optimization.
func (s *terminalWriter) SetHardTabs(v bool) {
	s.mu.Lock()
	// We always disable HardTabs when termtype is "linux".
	if !strings.HasPrefix(s.term, "linux") {
		if v {
			s.flags.Set(tHardTabs)
		} else {
			s.flags.Reset(tHardTabs)
		}
	}
	s.mu.Unlock()
}

// SetRelativeCursor sets whether to use relative cursor movements.
func (s *terminalWriter) SetRelativeCursor(v bool) {
	s.mu.Lock()
	if v {
		s.flags.Set(tRelativeCursor)
	} else {
		s.flags.Reset(tRelativeCursor)
	}
	s.mu.Unlock()
}

// SetAltScreen sets whether we're using the alternate screen.
func (s *terminalWriter) SetAltScreen(v bool) {
	s.mu.Lock()
	if v {
		s.flags.Set(tAltScreen)
	} else {
		s.flags.Reset(tAltScreen)
	}
	s.mu.Unlock()
}

// populateDiff populates the diff between the two buffers. This is used to
// determine which cells have changed and need to be redrawn.
func (s *terminalWriter) populateDiff(newbuf *Buffer) {
	var wg sync.WaitGroup
	for y := 0; y < newbuf.Height(); y++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var chg lineData
			v, ok := s.touch.Load(y)
			if !ok {
				chg = lineData{firstCell: 0, lastCell: newbuf.Width() - 1}
			} else if e, ok := v.(lineData); ok {
				chg = e
			}
			for x := 0; x < newbuf.Width(); x++ {
				var oldc *Cell
				if s.curbuf != nil {
					oldc = s.curbuf.CellAt(x, y)
				}
				newc := newbuf.CellAt(x, y)
				if !cellEqual(oldc, newc) {
					if !ok {
						chg = lineData{firstCell: x, lastCell: x + newc.Width}
					} else {
						chg.firstCell = min(chg.firstCell, x)
						chg.lastCell = max(chg.lastCell, x+newc.Width)
					}
				}
			}

			s.touch.Store(y, chg)
		}()
	}
	wg.Wait()
}

// moveCursor moves the cursor to the specified position.
func (s *terminalWriter) moveCursor(newbuf *Buffer, x, y int, overwrite bool) {
	if !s.flags.Contains(tAltScreen) && s.flags.Contains(tRelativeCursor) &&
		s.cur.X == -1 && s.cur.Y == -1 {
		// First cursor movement in inline mode, move the cursor to the first
		// column before moving to the target position.
		s.buf.WriteByte('\r') //nolint:errcheck
		s.cur.X, s.cur.Y = 0, 0
	}
	s.buf.WriteString(moveCursor(s, newbuf, x, y, overwrite)) //nolint:errcheck
	s.cur.X, s.cur.Y = x, y
}

func (s *terminalWriter) move(newbuf *Buffer, x, y int) {
	// XXX: Make sure we use the max height and width of the buffer in case
	// we're in the middle of a resize operation.
	width := max(newbuf.Width(), s.curbuf.Width())
	height := max(newbuf.Height(), s.curbuf.Height())

	if width > 0 && x >= width {
		// Handle autowrap
		y += (x / width)
		x %= width
	}

	// XXX: Disable styles if there's any
	// Some move operations such as [ansi.LF] can apply styles to the new
	// cursor position, thus, we need to reset the styles before moving the
	// cursor.
	blank := s.clearBlank()
	resetPen := y != s.cur.Y && !blank.Equal(&BlankCell)
	if resetPen {
		s.updatePen(nil)
	}

	// Reset wrap around (phantom cursor) state
	if s.atPhantom {
		s.cur.X = 0
		s.buf.WriteByte('\r') //nolint:errcheck
		s.atPhantom = false   // reset phantom cell state
	}

	// TODO: Investigate if we need to handle this case and/or if we need the
	// following code.
	//
	// if width > 0 && s.cur.X >= width {
	// 	l := (s.cur.X + 1) / width
	//
	// 	s.cur.Y += l
	// 	if height > 0 && s.cur.Y >= height {
	// 		l -= s.cur.Y - height - 1
	// 	}
	//
	// 	if l > 0 {
	// 		s.cur.X = 0
	// 		s.buf.WriteString("\r" + strings.Repeat("\n", l)) //nolint:errcheck
	// 	}
	// }

	if height > 0 {
		if s.cur.Y > height-1 {
			s.cur.Y = height - 1
		}
		if y > height-1 {
			y = height - 1
		}
	}

	if x == s.cur.X && y == s.cur.Y {
		// We give up later because we need to run checks for the phantom cell
		// and others before we can determine if we can give up.
		return
	}

	// We set the new cursor in tscreen.moveCursor].
	s.moveCursor(newbuf, x, y, true) // Overwrite cells if possible
}

// newTerminalWriter creates a new [terminalWriter].
func newTerminalWriter(w io.Writer, termtype string, width int) (s *terminalWriter) {
	s = new(terminalWriter)
	s.w = w
	s.profile = colorprofile.TrueColor
	s.width = width
	s.buf = new(bytes.Buffer)
	s.term = termtype
	s.caps = xtermCaps(termtype)
	s.cur = cursor{Position: Pos(-1, -1)} // start at -1 to force a move
	s.saved = s.cur
	s.scrollHeight = 0
	s.touch = sync.Map{}
	s.tabs = DefaultTabStops(s.width)
	s.buf.Reset()
	s.oldhash, s.newhash = nil, nil
	s.scrollHeight = 0
	s.touch = sync.Map{}
	s.tabs = DefaultTabStops(width)
	s.buf.Reset()
	s.oldhash, s.newhash = nil, nil
	return
}

// cellEqual returns whether the two cells are equal. A nil cell is considered
// a [BlankCell].
func cellEqual(a, b *Cell) bool {
	if a == b {
		return true
	}
	if a == nil {
		a = &BlankCell
	}
	if b == nil {
		b = &BlankCell
	}
	return a.Equal(b)
}

// putCell draws a cell at the current cursor position.
func (s *terminalWriter) putCell(newbuf *Buffer, cell *Cell) {
	width, height := newbuf.Width(), newbuf.Height()
	if s.flags.Contains(tAltScreen) && s.cur.X == width-1 && s.cur.Y == height-1 {
		s.putCellLR(newbuf, cell)
	} else {
		s.putAttrCell(newbuf, cell)
	}
}

// wrapCursor wraps the cursor to the next line.
//
//nolint:unused
func (s *terminalWriter) wrapCursor() {
	const autoRightMargin = true
	if autoRightMargin {
		// Assume we have auto wrap mode enabled.
		s.cur.X = 0
		s.cur.Y++
	} else {
		s.cur.X--
	}
}

func (s *terminalWriter) putAttrCell(newbuf *Buffer, cell *Cell) {
	if cell != nil && cell.Empty() {
		// XXX: Zero width cells are special and should not be written to the
		// screen no matter what other attributes they have.
		// Zero width cells are used for wide characters that are split into
		// multiple cells.
		return
	}

	if cell == nil {
		cell = s.clearBlank()
	}

	// We're at pending wrap state (phantom cell), incoming cell should
	// wrap.
	if s.atPhantom {
		s.wrapCursor()
		s.atPhantom = false
	}

	s.updatePen(cell)
	s.buf.WriteRune(cell.Rune) //nolint:errcheck
	for _, c := range cell.Comb {
		s.buf.WriteRune(c) //nolint:errcheck
	}

	s.cur.X += cell.Width
	if s.cur.X >= newbuf.Width() {
		s.atPhantom = true
	}
}

// putCellLR draws a cell at the lower right corner of the screen.
func (s *terminalWriter) putCellLR(newbuf *Buffer, cell *Cell) {
	// Optimize for the lower right corner cell.
	curX := s.cur.X
	if cell == nil || !cell.Empty() {
		s.buf.WriteString(ansi.ResetAutoWrapMode) //nolint:errcheck
		s.putAttrCell(newbuf, cell)
		// Writing to lower-right corner cell should not wrap.
		s.atPhantom = false
		s.cur.X = curX
		s.buf.WriteString(ansi.SetAutoWrapMode) //nolint:errcheck
	}
}

// updatePen updates the cursor pen styles.
func (s *terminalWriter) updatePen(cell *Cell) {
	if cell == nil {
		cell = &BlankCell
	}

	if s.profile != 0 {
		// Downsample colors to the given color profile.
		cell.Style = ConvertStyle(cell.Style, s.profile)
		cell.Link = ConvertLink(cell.Link, s.profile)
	}

	if !cell.Style.Equal(&s.cur.Style) {
		seq := cell.Style.DiffSequence(s.cur.Style)
		if cell.Style.Empty() && len(seq) > len(ansi.ResetStyle) {
			seq = ansi.ResetStyle
		}
		s.buf.WriteString(seq) //nolint:errcheck
		s.cur.Style = cell.Style
	}
	if !cell.Link.Equal(&s.cur.Link) {
		s.buf.WriteString(ansi.SetHyperlink(cell.Link.URL, cell.Link.Params)) //nolint:errcheck
		s.cur.Link = cell.Link
	}
}

// emitRange emits a range of cells to the buffer. It it equivalent to calling
// tscreen.putCell] for each cell in the range. This is optimized to use
// [ansi.ECH] and [ansi.REP].
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *terminalWriter) emitRange(newbuf *Buffer, line Line, n int) (eoi bool) {
	for n > 0 {
		var count int
		for n > 1 && !cellEqual(line.At(0), line.At(1)) {
			s.putCell(newbuf, line.At(0))
			line = line[1:]
			n--
		}

		cell0 := line[0]
		if n == 1 {
			s.putCell(newbuf, cell0)
			return false
		}

		count = 2
		for count < n && cellEqual(line.At(count), cell0) {
			count++
		}

		ech := ansi.EraseCharacter(count)
		cup := ansi.CursorPosition(s.cur.X+count, s.cur.Y)
		rep := ansi.RepeatPreviousCharacter(count)
		if s.caps.Contains(capECH) && count > len(ech)+len(cup) && cell0 != nil && cell0.Clear() {
			s.updatePen(cell0)
			s.buf.WriteString(ech) //nolint:errcheck

			// If this is the last cell, we don't need to move the cursor.
			if count < n {
				s.move(newbuf, s.cur.X+count, s.cur.Y)
			} else {
				return true // cursor in the middle
			}
		} else if s.caps.Contains(capREP) && count > len(rep) &&
			(cell0 == nil || (len(cell0.Comb) == 0 && cell0.Rune >= ansi.US && cell0.Rune < ansi.DEL)) {
			// We only support ASCII characters. Most terminals will handle
			// non-ASCII characters correctly, but some might not, ahem xterm.
			//
			// NOTE: [ansi.REP] only repeats the last rune and won't work
			// if the last cell contains multiple runes.

			wrapPossible := s.cur.X+count >= newbuf.Width()
			repCount := count
			if wrapPossible {
				repCount--
			}

			s.updatePen(cell0)
			s.putCell(newbuf, cell0)
			repCount-- // cell0 is a single width cell ASCII character

			s.buf.WriteString(ansi.RepeatPreviousCharacter(repCount)) //nolint:errcheck
			s.cur.X += repCount
			if wrapPossible {
				s.putCell(newbuf, cell0)
			}
		} else {
			for i := 0; i < count; i++ {
				s.putCell(newbuf, line.At(i))
			}
		}

		line = line[clamp(count, 0, len(line)):]
		n -= count
	}

	return
}

// putRange puts a range of cells from the old line to the new line.
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *terminalWriter) putRange(newbuf *Buffer, oldLine, newLine Line, y, start, end int) (eoi bool) {
	inline := min(len(ansi.CursorPosition(start+1, y+1)),
		min(len(ansi.HorizontalPositionAbsolute(start+1)),
			len(ansi.CursorForward(start+1))))
	if (end - start + 1) > inline {
		var j, same int
		for j, same = start, 0; j <= end; j++ {
			oldCell, newCell := oldLine.At(j), newLine.At(j)
			if same == 0 && oldCell != nil && oldCell.Empty() {
				continue
			}
			if cellEqual(oldCell, newCell) {
				same++
			} else {
				if same > end-start {
					s.emitRange(newbuf, newLine[start:], j-same-start)
					s.move(newbuf, j, y)
					start = j
				}
				same = 0
			}
		}

		i := s.emitRange(newbuf, newLine[start:], j-same-start)

		// Always return 1 for the next [tScreen.move] after a
		// [tScreen.putRange] if we found identical characters at end of
		// interval.
		if same == 0 {
			return i
		}
		return true
	}

	return s.emitRange(newbuf, newLine[start:], end-start+1)
}

// clearToEnd clears the screen from the current cursor position to the end of
// line.
func (s *terminalWriter) clearToEnd(newbuf *Buffer, blank *Cell, force bool) { //nolint:unparam
	if s.cur.Y >= 0 {
		curline := s.curbuf.Line(s.cur.Y)
		for j := s.cur.X; j < s.curbuf.Width(); j++ {
			if j >= 0 {
				c := curline.At(j)
				if !cellEqual(c, blank) {
					curline.Set(j, blank)
					force = true
				}
			}
		}
	}

	if force {
		s.updatePen(blank)
		count := newbuf.Width() - s.cur.X
		if s.el0Cost() <= count {
			s.buf.WriteString(ansi.EraseLineRight) //nolint:errcheck
		} else {
			for i := 0; i < count; i++ {
				s.putCell(newbuf, blank)
			}
		}
	}
}

// clearBlank returns a blank cell based on the current cursor background color.
func (s *terminalWriter) clearBlank() *Cell {
	c := BlankCell
	if !s.cur.Style.Empty() || !s.cur.Link.Empty() {
		c.Style = s.cur.Style
		c.Link = s.cur.Link
	}
	return &c
}

// insertCells inserts the count cells pointed by the given line at the current
// cursor position.
func (s *terminalWriter) insertCells(newbuf *Buffer, line Line, count int) {
	supportsICH := s.caps.Contains(capICH)
	if supportsICH {
		// Use [ansi.ICH] as an optimization.
		s.buf.WriteString(ansi.InsertCharacter(count)) //nolint:errcheck
	} else {
		// Otherwise, use [ansi.IRM] mode.
		s.buf.WriteString(ansi.SetInsertReplaceMode) //nolint:errcheck
	}

	for i := 0; count > 0; i++ {
		s.putAttrCell(newbuf, line[i])
		count--
	}

	if !supportsICH {
		s.buf.WriteString(ansi.ResetInsertReplaceMode) //nolint:errcheck
	}
}

// el0Cost returns the cost of using [ansi.EL] 0 i.e. [ansi.EraseLineRight]. If
// this terminal supports background color erase, it can be cheaper to use
// [ansi.EL] 0 i.e. [ansi.EraseLineRight] to clear
// trailing spaces.
func (s *terminalWriter) el0Cost() int {
	if s.caps != noCaps {
		return 0
	}
	return len(ansi.EraseLineRight)
}

// transformLine transforms the given line in the current window to the
// corresponding line in the new window. It uses [ansi.ICH] and [ansi.DCH] to
// insert or delete characters.
func (s *terminalWriter) transformLine(newbuf *Buffer, y int) {
	var firstCell, oLastCell, nLastCell int // first, old last, new last index
	oldLine := s.curbuf.Line(y)
	newLine := newbuf.Line(y)

	// Find the first changed cell in the line
	var lineChanged bool
	for i := 0; i < newbuf.Width(); i++ {
		if !cellEqual(newLine.At(i), oldLine.At(i)) {
			lineChanged = true
			break
		}
	}

	const ceolStandoutGlitch = false
	if ceolStandoutGlitch && lineChanged {
		s.move(newbuf, 0, y)
		s.clearToEnd(newbuf, nil, false)
		s.putRange(newbuf, oldLine, newLine, y, 0, newbuf.Width()-1)
	} else {
		blank := newLine.At(0)

		// It might be cheaper to clear leading spaces with [ansi.EL] 1 i.e.
		// [ansi.EraseLineLeft].
		if blank == nil || blank.Clear() {
			var oFirstCell, nFirstCell int
			for oFirstCell = 0; oFirstCell < s.curbuf.Width(); oFirstCell++ {
				if !cellEqual(oldLine.At(oFirstCell), blank) {
					break
				}
			}
			for nFirstCell = 0; nFirstCell < newbuf.Width(); nFirstCell++ {
				if !cellEqual(newLine.At(nFirstCell), blank) {
					break
				}
			}

			if nFirstCell == oFirstCell {
				firstCell = nFirstCell

				// Find the first differing cell
				for firstCell < newbuf.Width() &&
					cellEqual(oldLine.At(firstCell), newLine.At(firstCell)) {
					firstCell++
				}
			} else if oFirstCell > nFirstCell {
				firstCell = nFirstCell
			} else if oFirstCell < nFirstCell {
				firstCell = oFirstCell
				el1Cost := len(ansi.EraseLineLeft)
				if el1Cost < nFirstCell-oFirstCell {
					if nFirstCell >= newbuf.Width() {
						s.move(newbuf, 0, y)
						s.updatePen(blank)
						s.buf.WriteString(ansi.EraseLineRight) //nolint:errcheck
					} else {
						s.move(newbuf, nFirstCell-1, y)
						s.updatePen(blank)
						s.buf.WriteString(ansi.EraseLineLeft) //nolint:errcheck
					}

					for firstCell < nFirstCell {
						oldLine.Set(firstCell, blank)
						firstCell++
					}
				}
			}
		} else {
			// Find the first differing cell
			for firstCell < newbuf.Width() && cellEqual(newLine.At(firstCell), oldLine.At(firstCell)) {
				firstCell++
			}
		}

		// If we didn't find one, we're done
		if firstCell >= newbuf.Width() {
			return
		}

		blank = newLine.At(newbuf.Width() - 1)
		if blank != nil && !blank.Clear() {
			// Find the last differing cell
			nLastCell = newbuf.Width() - 1
			for nLastCell > firstCell && cellEqual(newLine.At(nLastCell), oldLine.At(nLastCell)) {
				nLastCell--
			}

			if nLastCell >= firstCell {
				s.move(newbuf, firstCell, y)
				s.putRange(newbuf, oldLine, newLine, y, firstCell, nLastCell)
				if firstCell < len(oldLine) && firstCell < len(newLine) {
					copy(oldLine[firstCell:], newLine[firstCell:])
				} else {
					copy(oldLine, newLine)
				}
			}

			return
		}

		// Find last non-blank cell in the old line.
		oLastCell = s.curbuf.Width() - 1
		for oLastCell > firstCell && cellEqual(oldLine.At(oLastCell), blank) {
			oLastCell--
		}

		// Find last non-blank cell in the new line.
		nLastCell = newbuf.Width() - 1
		for nLastCell > firstCell && cellEqual(newLine.At(nLastCell), blank) {
			nLastCell--
		}

		if nLastCell == firstCell && s.el0Cost() < oLastCell-nLastCell {
			s.move(newbuf, firstCell, y)
			if !cellEqual(newLine.At(firstCell), blank) {
				s.putCell(newbuf, newLine.At(firstCell))
			}
			s.clearToEnd(newbuf, blank, false)
		} else if nLastCell != oLastCell &&
			!cellEqual(newLine.At(nLastCell), oldLine.At(oLastCell)) {
			s.move(newbuf, firstCell, y)
			if oLastCell-nLastCell > s.el0Cost() {
				if s.putRange(newbuf, oldLine, newLine, y, firstCell, nLastCell) {
					s.move(newbuf, nLastCell+1, y)
				}
				s.clearToEnd(newbuf, blank, false)
			} else {
				n := max(nLastCell, oLastCell)
				s.putRange(newbuf, oldLine, newLine, y, firstCell, n)
			}
		} else {
			nLastNonBlank := nLastCell
			oLastNonBlank := oLastCell

			// Find the last cells that really differ.
			// Can be -1 if no cells differ.
			for cellEqual(newLine.At(nLastCell), oldLine.At(oLastCell)) {
				if !cellEqual(newLine.At(nLastCell-1), oldLine.At(oLastCell-1)) {
					break
				}
				nLastCell--
				oLastCell--
				if nLastCell == -1 || oLastCell == -1 {
					break
				}
			}

			n := min(oLastCell, nLastCell)
			if n >= firstCell {
				s.move(newbuf, firstCell, y)
				s.putRange(newbuf, oldLine, newLine, y, firstCell, n)
			}

			if oLastCell < nLastCell {
				m := max(nLastNonBlank, oLastNonBlank)
				if n != 0 {
					for n > 0 {
						wide := newLine.At(n + 1)
						if wide == nil || !wide.Empty() {
							break
						}
						n--
						oLastCell--
					}
				} else if n >= firstCell && newLine.At(n) != nil && newLine.At(n).Width > 1 {
					next := newLine.At(n + 1)
					for next != nil && next.Empty() {
						n++
						oLastCell++
					}
				}

				s.move(newbuf, n+1, y)
				ichCost := 3 + nLastCell - oLastCell
				if s.caps.Contains(capICH) && (nLastCell < nLastNonBlank || ichCost > (m-n)) {
					s.putRange(newbuf, oldLine, newLine, y, n+1, m)
				} else {
					s.insertCells(newbuf, newLine[n+1:], nLastCell-oLastCell)
				}
			} else if oLastCell > nLastCell {
				s.move(newbuf, n+1, y)
				dchCost := 3 + oLastCell - nLastCell
				if dchCost > len(ansi.EraseLineRight)+nLastNonBlank-(n+1) {
					if s.putRange(newbuf, oldLine, newLine, y, n+1, nLastNonBlank) {
						s.move(newbuf, nLastNonBlank+1, y)
					}
					s.clearToEnd(newbuf, blank, false)
				} else {
					s.updatePen(blank)
					s.deleteCells(oLastCell - nLastCell)
				}
			}
		}
	}

	// Update the old line with the new line
	if firstCell < len(oldLine) && firstCell < len(newLine) {
		copy(oldLine[firstCell:], newLine[firstCell:])
	} else {
		copy(oldLine, newLine)
	}
}

// deleteCells deletes the count cells at the current cursor position and moves
// the rest of the line to the left. This is equivalent to [ansi.DCH].
func (s *terminalWriter) deleteCells(count int) {
	// [ansi.DCH] will shift in cells from the right margin so we need to
	// ensure that they are the right style.
	s.buf.WriteString(ansi.DeleteCharacter(count)) //nolint:errcheck
}

// clearToBottom clears the screen from the current cursor position to the end
// of the screen.
func (s *terminalWriter) clearToBottom(blank *Cell) {
	row, col := s.cur.Y, s.cur.X
	if row < 0 {
		row = 0
	}

	s.updatePen(blank)
	s.buf.WriteString(ansi.EraseScreenBelow) //nolint:errcheck
	// Clear the rest of the current line
	s.curbuf.ClearArea(Rect(col, row, s.curbuf.Width()-col, 1))
	// Clear everything below the current line
	s.curbuf.ClearArea(Rect(0, row+1, s.curbuf.Width(), s.curbuf.Height()-row-1))
}

// clearBottom tests if clearing the end of the screen would satisfy part of
// the screen update. Scan backwards through lines in the screen checking if
// each is blank and one or more are changed.
// It returns the top line.
func (s *terminalWriter) clearBottom(newbuf *Buffer, total int) (top int) {
	if total <= 0 {
		return
	}

	top = total
	last := min(s.curbuf.Width(), newbuf.Width())
	blank := s.clearBlank()
	canClearWithBlank := blank == nil || blank.Clear()

	if canClearWithBlank {
		var row int
		for row = total - 1; row >= 0; row-- {
			oldLine := s.curbuf.Line(row)
			newLine := newbuf.Line(row)

			var col int
			ok := true
			for col = 0; ok && col < last; col++ {
				ok = cellEqual(newLine.At(col), blank)
			}
			if !ok {
				break
			}

			for col = 0; ok && col < last; col++ {
				ok = cellEqual(oldLine.At(col), blank)
			}
			if !ok {
				top = row
			}
		}

		if top < total {
			s.move(newbuf, 0, max(0, top-1)) // top is 1-based
			s.clearToBottom(blank)
			if s.oldhash != nil && s.newhash != nil &&
				row < len(s.oldhash) && row < len(s.newhash) {
				for row := top; row < newbuf.Height(); row++ {
					s.oldhash[row] = s.newhash[row]
				}
			}
		}
	}

	return
}

// clearScreen clears the screen and put cursor at home.
func (s *terminalWriter) clearScreen(blank *Cell) {
	s.updatePen(blank)
	s.buf.WriteString(ansi.CursorHomePosition) //nolint:errcheck
	s.buf.WriteString(ansi.EraseEntireScreen)  //nolint:errcheck
	s.cur.X, s.cur.Y = 0, 0
	s.curbuf.Fill(blank)
}

// clearBelow clears everything below and including the row.
func (s *terminalWriter) clearBelow(newbuf *Buffer, blank *Cell, row int) {
	s.move(newbuf, 0, row)
	s.clearToBottom(blank)
}

// clearUpdate forces a screen redraw.
func (s *terminalWriter) clearUpdate(newbuf *Buffer) {
	blank := s.clearBlank()
	var nonEmpty int
	if s.flags.Contains(tAltScreen) {
		// XXX: We're using the maximum height of the two buffers to ensure we
		// write newly added lines to the screen in
		// [terminalWriter.transformLine].
		nonEmpty = max(s.curbuf.Height(), newbuf.Height())
		s.clearScreen(blank)
	} else {
		nonEmpty = newbuf.Height()
		// FIXME: Investigate the double [ansi.ClearScreenBelow] call.
		// Commenting the line below out seems to work but it might cause other
		// bugs.
		s.clearBelow(newbuf, blank, 0)
	}
	nonEmpty = s.clearBottom(newbuf, nonEmpty)
	for i := 0; i < nonEmpty; i++ {
		s.transformLine(newbuf, i)
	}
}

// Flush flushes the buffer to the screen.
func (s *terminalWriter) Flush() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flush()
}

func (s *terminalWriter) flush() (err error) {
	// Write the buffer
	if n := s.buffered(); n > 0 {
		logger.Printf("Flushing %d bytes to the screen %q\n", n, s.buf.String())
		nr, err := s.buf.WriteTo(s.w)
		if err != nil {
			// When we get a short write error, truncate the buffer to the
			// remaining bytes.
			s.buf.Truncate(int(nr))
		}
	}

	return
}

// Buffered returns how many bytes are currently buffered.
func (s *terminalWriter) Buffered() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buffered()
}

func (s *terminalWriter) buffered() int {
	return s.buf.Len()
}

// Touched returns the number of lines that have been touched or changed.
func (s *terminalWriter) Touched() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.touched()
}

func (t *terminalWriter) touched() int {
	return syncMapLen(&t.touch)
}

// Render renders changes of the screen to the internal buffer. Call
// [terminalWriter.Flush] to flush pending changes to the screen.
func (s *terminalWriter) Render(newbuf *Buffer) {
	s.mu.Lock()
	s.populateDiff(newbuf)
	s.render(newbuf)
	s.mu.Unlock()
}

func syncMapLen(m *sync.Map) (n int) {
	m.Range(func(_, _ any) bool {
		n++
		return true
	})
	return
}

func (s *terminalWriter) render(newbuf *Buffer) {
	// Do we have a buffer to compare to?
	if s.curbuf == nil {
		s.curbuf = NewBuffer(newbuf.Width(), newbuf.Height())
	}

	// Do we need to render anything?
	touchedLines := s.touched()
	if !s.clear && touchedLines == 0 {
		return
	}

	// TODO: Investigate whether this is necessary. Theoretically, terminals
	// can add/remove tab stops and we should be able to handle that. We could
	// use [ansi.DECTABSR] to read the tab stops, but that's not implemented in
	// most terminals :/
	// // Are we using hard tabs? If so, ensure tabs are using the
	// // default interval using [ansi.DECST8C].
	// if s.opts.HardTabs && !s.initTabs {
	// 	s.buf.WriteString(ansi.SetTabEvery8Columns)
	// 	s.initTabs = true
	// }

	var nonEmpty int

	// XXX: In inline mode, after a screen resize, we need to clear the extra
	// lines at the bottom of the screen. This is because in inline mode, we
	// don't use the full screen height and the current buffer size might be
	// larger than the new buffer size.
	partialClear := !s.flags.Contains(tAltScreen) && s.cur.X != -1 && s.cur.Y != -1 &&
		s.curbuf.Width() == newbuf.Width() &&
		s.curbuf.Height() > 0 &&
		s.curbuf.Height() > newbuf.Height()

	if !s.clear && partialClear {
		s.clearBelow(newbuf, nil, newbuf.Height()-1)
	}

	if s.clear {
		s.clearUpdate(newbuf)
		s.clear = false
	} else if s.touched() > 0 {
		if s.flags.Contains(tAltScreen) {
			// Optimize scrolling for the alternate screen buffer.
			// TODO: Should we optimize for inline mode as well? If so, we need
			// to know the actual cursor position to use [ansi.DECSTBM].
			s.scrollOptimize(newbuf)
		}

		var changedLines int
		var i int

		if s.flags.Contains(tAltScreen) {
			nonEmpty = min(s.curbuf.Height(), newbuf.Height())
		} else {
			nonEmpty = newbuf.Height()
		}

		nonEmpty = s.clearBottom(newbuf, nonEmpty)
		for i = 0; i < nonEmpty; i++ {
			_, ok := s.touch.Load(i)
			if ok {
				s.transformLine(newbuf, i)
				changedLines++
			}
		}
	}

	// Ensure we have scrolled the screen to the bottom when we're not using
	// alt screen mode.
	if !s.flags.Contains(tAltScreen) && s.scrollHeight < newbuf.Height()-1 {
		s.move(newbuf, 0, newbuf.Height()-1)
	}

	// Sync windows and screen
	s.touch = sync.Map{}

	if s.curbuf.Width() != newbuf.Width() || s.curbuf.Height() != newbuf.Height() {
		// Resize the old buffer to match the new buffer.
		_, oldh := s.curbuf.Width(), s.curbuf.Height()
		s.curbuf.Resize(newbuf.Width(), newbuf.Height())
		// Sync new lines to old lines
		for i := oldh - 1; i < newbuf.Height(); i++ {
			copy(s.curbuf.Line(i), newbuf.Line(i))
		}
	}

	s.updatePen(nil) // nil indicates a blank cell with no styles
}

// reset resets the screen to its initial state.
func (s *terminalWriter) reset() {
}

// Clear marks the screen to be fully cleared on the next render.
func (s *terminalWriter) Clear() {
	s.mu.Lock()
	s.clear = true
	s.mu.Unlock()
}

// Resize resizes the screen.
func (s *terminalWriter) Resize(newbuf *Buffer, width, height int) bool {
	s.width = width

	oldw := newbuf.Width()
	oldh := newbuf.Height()

	altScreen := s.flags.Contains(tAltScreen)
	if altScreen || width != oldw {
		// We only clear the whole screen if the width changes. Adding/removing
		// rows is handled by the [tScreen.render] and [tScreen.transformLine]
		// methods.
		s.clear = true
	}

	// Clear new columns and lines
	if width > oldh {
		newbuf.ClearArea(Rect(max(oldw-1, 0), 0, width-oldw, height))
	} else if width < oldw {
		newbuf.ClearArea(Rect(max(width, 0), 0, oldw-width, height))
	}

	if height > oldh {
		newbuf.ClearArea(Rect(0, max(oldh-1, 0), width, height-oldh))
	} else if height < oldh {
		newbuf.ClearArea(Rect(0, max(height, 0), width, oldh-height))
	}

	s.mu.Lock()
	newbuf.Resize(width, height)
	s.tabs.Resize(width)
	s.oldhash, s.newhash = nil, nil
	s.scrollHeight = 0 // reset scroll lines
	s.mu.Unlock()

	return true
}

// Position returns the current cursor position.
func (s *terminalWriter) Position() (x, y int) {
	return s.cur.X, s.cur.Y
}

// SetPosition changes the logical cursor position. This can be used when we
// change the cursor position outside of the screen and need to update the
// screen cursor position.
func (s *terminalWriter) SetPosition(x, y int) {
	s.mu.Lock()
	s.cur.X, s.cur.Y = x, y
	s.mu.Unlock()
}

// WriteString writes the given string to the underlying buffer.
func (s *terminalWriter) WriteString(str string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.WriteString(str)
}

// Write writes the given bytes to the underlying buffer.
func (s *terminalWriter) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(b)
}

// MoveTo calculates and writes the shortest sequence to move the cursor to the
// given position. It uses the current cursor position and the new position to
// calculate the shortest amount of sequences to move the cursor.
func (s *terminalWriter) MoveTo(newbuf *Buffer, x, y int) {
	s.mu.Lock()
	s.move(newbuf, x, y)
	s.mu.Unlock()
}

// notLocal returns whether the coordinates are not considered local movement
// using the defined thresholds.
// This takes the number of columns, and the coordinates of the current and
// target positions.
func notLocal(cols, fx, fy, tx, ty int) bool {
	// The typical distance for a [ansi.CUP] sequence. Anything less than this
	// is considered local movement.
	const longDist = 8 - 1
	return (tx > longDist) &&
		(tx < cols-1-longDist) &&
		(abs(ty-fy)+abs(tx-fx) > longDist)
}

// relativeCursorMove returns the relative cursor movement sequence using one or two
// of the following sequences [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB],
// [ansi.VPA], [ansi.HPA].
// When overwrite is true, this will try to optimize the sequence by using the
// screen cells values to move the cursor instead of using escape sequences.
func relativeCursorMove(s *terminalWriter, newbuf *Buffer, fx, fy, tx, ty int, overwrite, useTabs, useBackspace bool) string {
	var seq strings.Builder

	width, height := newbuf.Width(), newbuf.Height()
	if ty != fy {
		var yseq string
		if s.caps.Contains(capVPA) && !s.flags.Contains(tRelativeCursor) {
			yseq = ansi.VerticalPositionAbsolute(ty + 1)
		}

		// OPTIM: Use [ansi.LF] and [ansi.ReverseIndex] as optimizations.

		if ty > fy {
			n := ty - fy
			if cud := ansi.CursorDown(n); yseq == "" || len(cud) < len(yseq) {
				yseq = cud
			}
			shouldScroll := !s.flags.Contains(tAltScreen) && fy+n >= s.scrollHeight
			if lf := strings.Repeat("\n", n); shouldScroll || (fy+n < height && len(lf) < len(yseq)) {
				// TODO: Ensure we're not unintentionally scrolling the screen down.
				yseq = lf
				s.scrollHeight = max(s.scrollHeight, fy+n)
				if s.flags.Contains(tMapNewline) {
					fx = 0
				}
			}
		} else if ty < fy {
			n := fy - ty
			if cuu := ansi.CursorUp(n); yseq == "" || len(cuu) < len(yseq) {
				yseq = cuu
			}
			if n == 1 && fy-1 > 0 {
				// TODO: Ensure we're not unintentionally scrolling the screen up.
				yseq = ansi.ReverseIndex
			}
		}

		seq.WriteString(yseq)
	}

	if tx != fx {
		var xseq string
		if s.caps.Contains(capHPA) && !s.flags.Contains(tRelativeCursor) {
			xseq = ansi.HorizontalPositionAbsolute(tx + 1)
		}

		if tx > fx {
			n := tx - fx
			if useTabs {
				var tabs int
				var col int
				for col = fx; s.tabs.Next(col) <= tx; col = s.tabs.Next(col) {
					tabs++
					if col == s.tabs.Next(col) || col >= width-1 {
						break
					}
				}

				if tabs > 0 {
					cht := ansi.CursorHorizontalForwardTab(tabs)
					tab := strings.Repeat("\t", tabs)
					if false && s.caps.Contains(capCHT) && len(cht) < len(tab) {
						// TODO: The linux console and some terminals such as
						// Alacritty don't support [ansi.CHT]. Enable this when
						// we have a way to detect this, or after 5 years when
						// we're sure everyone has updated their terminals :P
						seq.WriteString(cht)
					} else {
						seq.WriteString(tab)
					}

					n = tx - col
					fx = col
				}
			}

			if cuf := ansi.CursorForward(n); xseq == "" || len(cuf) < len(xseq) {
				xseq = cuf
			}

			// If we have no attribute and style changes, overwrite is cheaper.
			var ovw string
			if overwrite && ty >= 0 {
				for i := 0; i < n; i++ {
					cell := newbuf.CellAt(fx+i, ty)
					if cell != nil && cell.Width > 0 {
						i += cell.Width - 1
						if !cell.Style.Equal(&s.cur.Style) || !cell.Link.Equal(&s.cur.Link) {
							overwrite = false
							break
						}
					}
				}
			}

			if overwrite && ty >= 0 {
				for i := 0; i < n; i++ {
					cell := newbuf.CellAt(fx+i, ty)
					if cell != nil && cell.Width > 0 {
						ovw += cell.String()
						i += cell.Width - 1
					} else {
						ovw += " "
					}
				}
			}

			if overwrite && len(ovw) < len(xseq) {
				xseq = ovw
			}
		} else if tx < fx {
			n := fx - tx
			if useTabs && s.caps.Contains(capCBT) {
				// VT100 does not support backward tabs [ansi.CBT].

				col := fx

				var cbt int // cursor backward tabs count
				for s.tabs.Prev(col) >= tx {
					col = s.tabs.Prev(col)
					cbt++
					if col == s.tabs.Prev(col) || col <= 0 {
						break
					}
				}

				if cbt > 0 {
					seq.WriteString(ansi.CursorBackwardTab(cbt))
					n = col - tx
				}
			}

			if cub := ansi.CursorBackward(n); xseq == "" || len(cub) < len(xseq) {
				xseq = cub
			}

			if useBackspace && n < len(xseq) {
				xseq = strings.Repeat("\b", n)
			}
		}

		seq.WriteString(xseq)
	}

	return seq.String()
}

// moveCursor moves and returns the cursor movement sequence to move the cursor
// to the specified position.
// When overwrite is true, this will try to optimize the sequence by using the
// screen cells values to move the cursor instead of using escape sequences.
func moveCursor(s *terminalWriter, newbuf *Buffer, x, y int, overwrite bool) (seq string) {
	fx, fy := s.cur.X, s.cur.Y

	if !s.flags.Contains(tRelativeCursor) {
		// Method #0: Use [ansi.CUP] if the distance is long.
		seq = ansi.CursorPosition(x+1, y+1)
		if fx == -1 || fy == -1 || notLocal(newbuf.Width(), fx, fy, x, y) {
			return
		}
	}

	// Optimize based on options.
	trials := 0
	if s.flags.Contains(tHardTabs) {
		trials |= 2 // 0b10 in binary
	}
	if s.flags.Contains(tBackspace) {
		trials |= 1 // 0b01 in binary
	}

	// Try all possible combinations of hard tabs and backspace optimizations.
	for i := 0; i <= trials; i++ {
		// Skip combinations that are not enabled.
		if i & ^trials != 0 {
			continue
		}

		useHardTabs := i&2 != 0
		useBackspace := i&1 != 0

		// Method #1: Use local movement sequences.
		nseq := relativeCursorMove(s, newbuf, fx, fy, x, y, overwrite, useHardTabs, useBackspace)
		if (i == 0 && len(seq) == 0) || len(nseq) < len(seq) {
			seq = nseq
		}

		// Method #2: Use [ansi.CR] and local movement sequences.
		nseq = "\r" + relativeCursorMove(s, newbuf, 0, fy, x, y, overwrite, useHardTabs, useBackspace)
		if len(nseq) < len(seq) {
			seq = nseq
		}

		if !s.flags.Contains(tRelativeCursor) {
			// Method #3: Use [ansi.CursorHomePosition] and local movement sequences.
			nseq = ansi.CursorHomePosition + relativeCursorMove(s, newbuf, 0, 0, x, y, overwrite, useHardTabs, useBackspace)
			if len(nseq) < len(seq) {
				seq = nseq
			}
		}
	}

	return
}

// xtermCaps returns whether the terminal is xterm-like. This means that the
// terminal supports ECMA-48 and ANSI X3.64 escape sequences.
// xtermCaps returns a list of control sequence capabilities for the given
// terminal type. This only supports a subset of sequences that can
// be different among terminals.
// NOTE: A hybrid approach would be to support Terminfo databases for a full
// set of capabilities.
func xtermCaps(termtype string) (v capabilities) {
	parts := strings.Split(termtype, "-")
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case
		"contour",
		"foot",
		"ghostty",
		"kitty",
		"rio",
		"st",
		"tmux",
		"wezterm",
		"xterm":
		v = allCaps
	case "alacritty":
		v = allCaps
		v &^= capCHT // NOTE: alacritty added support for [ansi.CHT] in 2024-12-28 #62d5b13.
	case "screen":
		// See https://www.gnu.org/software/screen/manual/screen.html#Control-Sequences-1
		v = allCaps
		v &^= capREP
	case "linux":
		// See https://man7.org/linux/man-pages/man4/console_codes.4.html
		v = capVPA | capHPA | capECH | capICH
	}

	return
}
