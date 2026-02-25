package uv

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/displaywidth"
)

// NBSP is the non-breaking space character. It is used to create padding
// around text content without allowing the spaces to collapse or break onto a
// new line.
const NBSP = "\u00A0"

// DrawLines draws the given lines to the buffer at the specified area.
//
// It assumes that the lines are already wrapped to fit within the width of the
// area. Lines that exceed the height of the area will be truncated.
//
// Use [Wrapper] to wrap lines to a specific width before drawing them with
// this function.
func DrawLines(scr Screen, area Rectangle, lines ...Line) {
	for y := 0; y < len(lines) && area.Min.Y+y < area.Max.Y; y++ {
		line := lines[y]
		for x := 0; x < len(line) && area.Min.X+x < area.Max.X; x++ {
			cell := line[x]
			if cell.IsZero() {
				continue
			}
			scr.SetCell(area.Min.X+x, area.Min.Y+y, &cell)
		}
	}
}

// Characters returns a slice of [Cell]s representing the graphemes in the input string.
//
// It ignores any newline characters, so the resulting slice will not contain
// any line breaks. If you want to preserve line breaks, use [Lines] instead or
// spilt the input string into lines first and then call [Characters] on each
// line.
//
// Any tab characters will be converted into 8 spaces. To convert tabs into a
// different number of spaces, you can pre-process the input string using
// [strings.ReplaceAll] or [bytes.ReplaceAll] before calling this function.
func Characters[T ~string | ~[]byte](input T, wm WidthMethod) []Cell {
	if len(input) == 0 {
		return Line{}
	}

	const tabReplacement = "        " // 8 spaces
	switch v := any(input).(type) {
	case string:
		input = any(strings.ReplaceAll(v, "\t", tabReplacement)).(T)
	case []byte:
		input = any(bytes.ReplaceAll(v, []byte("\t"), []byte(tabReplacement))).(T)
	}

	var cells []Cell
	var grs displaywidth.Graphemes[T]
	switch v := any(input).(type) {
	case string:
		grs = any(displaywidth.StringGraphemes(v)).(displaywidth.Graphemes[T])
	case []byte:
		grs = any(displaywidth.BytesGraphemes(v)).(displaywidth.Graphemes[T])
	}

	for grs.Next() {
		gr := string(grs.Value())

		var w int
		if wm == ansi.GraphemeWidth {
			w = grs.Width()
		} else {
			w = wm.StringWidth(gr)
		}

		if w == 0 {
			// Skip any zero-width graphemes, such as standalone carriage
			// returns or zero-width joiners. This is important to prevent
			// zero-width graphemes from taking up space in the output, which
			// can cause layout issues.
			continue
		}

		cells = append(cells, Cell{
			Content: gr,
			Width:   w,
		})
		// Add padding cells for wide characters
		for i := 1; i < w; i++ {
			cells = append(cells, Cell{})
		}
	}

	return cells
}

// Lines returns a slice of [Line]s representing the lines in the input string,
// where each line is a slice of [Cell]s representing the graphemes in that
// line.
//
// It treats newline characters as line breaks, so the resulting slice will
// contain one [Line] per line in the input string. If you want to ignore line
// breaks, use [Characters] instead.
//
// Any tab characters will be converted into 8 spaces. To convert tabs into a
// different number of spaces, you can pre-process the input string using
// [strings.ReplaceAll] or [bytes.ReplaceAll] before calling this function.
func Lines[T ~string | ~[]byte](input T, wm WidthMethod) []Line {
	if len(input) == 0 {
		return nil
	}

	switch v := any(input).(type) {
	case string:
		input = any(strings.ReplaceAll(v, "\r\n", "\n")).(T)
	case []byte:
		input = any(bytes.ReplaceAll(v, []byte("\r\n"), []byte("\n"))).(T)
	}

	var inputLines []T
	switch v := any(input).(type) {
	case string:
		inputLines = any(strings.Split(v, "\n")).([]T)
	case []byte:
		inputLines = any(bytes.Split(v, []byte("\n"))).([]T)
	}

	lines := make([]Line, len(inputLines))
	for i, line := range inputLines {
		lines[i] = Characters(line, wm)
	}

	return lines
}

// Wrapper represents [Line]s wrapper that can be used to wrap lines of [Cell]s
// to a specified width.
type Wrapper struct {
	// Breakpoints is a slice of breakpoint graphemes that can be used to break
	// lines. Spaces and dashes are always treated as a breakpoint, even if it
	// is not included in this slice.
	Breakpoints []string

	// PreserveSpaces controls whether trailing whitespace is kept on wrapped
	// lines. A whitespace cell is an empty or NBSP cell with no attributes.
	// When false, trailing whitespace is trimmed from each wrapped line.
	PreserveSpaces bool
}

// NewWrapper creates a new [Wrapper] with the specified breakpoints.
func NewWrapper(breakpoints ...string) *Wrapper {
	w := &Wrapper{
		Breakpoints: breakpoints,
	}
	return w
}

// Wrap wraps the input lines to the specified width, returning a slice of
// [Line]s representing the wrapped lines.
//
// It attempts to break lines at the specified breakpoints, but will fall back
// to hard breaking if a single grapheme exceeds the width.
func (lw *Wrapper) Wrap(lines []Line, width int) []Line {
	if len(lines) == 0 {
		return nil
	}

	if width <= 0 {
		return lines
	}

	var result []Line
	for _, line := range lines {
		wrapped := lw.wrapLine(line, width)
		result = append(result, wrapped...)
	}
	return result
}

// isBreakpoint returns true if the cell is a space, dash, or one of the
// configured breakpoints.
func (lw *Wrapper) isBreakpoint(cell Cell) bool {
	if cell.Content == " " || cell.Content == NBSP || cell.Content == "-" {
		return true
	}
	for _, bp := range lw.Breakpoints {
		if cell.Content == bp {
			return true
		}
	}
	return false
}

// wrapLine wraps a single line to the given width.
func (lw *Wrapper) wrapLine(line Line, width int) []Line {
	if len(line) <= width {
		if !lw.PreserveSpaces {
			line = trimTrailingWhitespace(line)
		}
		return []Line{line}
	}

	var result []Line
	pos := 0
	n := len(line)

	for pos < n {
		end := pos + width
		if end >= n {
			cur := line[pos:n]
			if !lw.PreserveSpaces {
				cur = trimTrailingWhitespace(cur)
			}
			result = append(result, cur)
			break
		}

		// Check if the cell at `end` (the overflow position) is a
		// whitespace breakpoint. If so, everything before it fits exactly.
		if isWhitespace(line[end]) && lw.isBreakpoint(line[end]) {
			cur := line[pos:end]
			if !lw.PreserveSpaces {
				cur = trimTrailingWhitespace(cur)
			}
			result = append(result, cur)
			pos = end
			// Consume the whitespace run after the break.
			if !lw.PreserveSpaces {
				pos = lw.skipWhitespace(line, pos, n, width, &result)
			} else {
				pos++
			}
			continue
		}

		// Scan backwards from end-1 for a breakpoint within this chunk.
		breakAt := -1
		for i := end - 1; i > pos; i-- {
			cell := line[i]
			if cell.Width == 0 {
				continue
			}
			if lw.isBreakpoint(cell) {
				if isWhitespace(cell) {
					breakAt = i
				} else {
					breakAt = i + 1
				}
				break
			}
		}

		if breakAt > pos {
			if breakAt < n && isWhitespace(line[breakAt]) {
				if lw.PreserveSpaces {
					cur := line[pos : breakAt+1]
					result = append(result, cur)
					pos = breakAt + 1
				} else {
					cur := trimTrailingWhitespace(line[pos:breakAt])
					result = append(result, cur)
					pos = breakAt
					pos = lw.skipWhitespace(line, pos, n, width, &result)
				}
			} else {
				cur := line[pos:breakAt]
				if !lw.PreserveSpaces {
					cur = trimTrailingWhitespace(cur)
				}
				result = append(result, cur)
				pos = breakAt
			}
		} else {
			// Hard wrap: no breakpoint found, cut at width boundary.
			// If the cell at `end` is a wide-char placeholder, back up
			// so we don't split a wide character.
			cut := end
			for cut > pos && line[cut].Width == 0 && line[cut].Content == "" {
				cut--
			}
			if cut == pos {
				cut = end
			}
			cur := line[pos:cut]
			if !lw.PreserveSpaces {
				cur = trimTrailingWhitespace(cur)
			}
			result = append(result, cur)
			pos = cut
		}
	}

	return result
}

// skipWhitespace advances pos past whitespace cells. When PreserveSpaces is
// true, it wraps the whitespace into width-sized lines. When false, it emits
// an empty line for every `width` consecutive whitespace cells consumed, then
// skips the rest.
func (lw *Wrapper) skipWhitespace(line Line, pos, n, width int, result *[]Line) int {
	if lw.PreserveSpaces {
		for pos < n && isWhitespace(line[pos]) {
			end := pos + width
			if end > n {
				end = n
			}
			// Check if the remaining content is all whitespace.
			allWs := true
			for j := pos; j < end && j < n; j++ {
				if !isWhitespace(line[j]) {
					allWs = false
					break
				}
			}
			if !allWs {
				break
			}
			*result = append(*result, line[pos:end])
			pos = end
		}
		return pos
	}
	start := pos
	for pos < n && isWhitespace(line[pos]) {
		pos++
	}
	consumed := pos - start
	// Emit an empty line for every full width of whitespace consumed.
	for i := width; i <= consumed; i += width {
		*result = append(*result, Line{})
	}
	return pos
}

var nbspCell = Cell{
	Content: NBSP,
	Width:   1,
}

// isWhitespace returns true if the cell is a space with no attributes.
func isWhitespace(cell Cell) bool {
	return cell.Equal(&EmptyCell) || cell.Equal(&nbspCell)
}

// trimTrailingWhitespace removes trailing whitespace cells from a line.
func trimTrailingWhitespace(line Line) Line {
	i := len(line)
	for i > 0 && isWhitespace(line[i-1]) {
		i--
	}
	return line[:i]
}
