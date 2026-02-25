package uv

import (
	"bytes"
	"slices"
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
	// lines. A whitespace cell is an empty cell with no attributes (see
	// [isWhitespace]) or a NBSP cell with no attributes. When false (the
	// default), trailing whitespace is trimmed from each wrapped line.
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
	if width <= 0 {
		return lines
	}

	var result []Line

	for _, line := range lines {
		wrapped := wrapLine(line, width, lw.Breakpoints)
		result = append(result, wrapped...)
	}

	if !lw.PreserveSpaces {
		for i := range result {
			result[i] = trimTrailingWhitespace(result[i])
		}
	}

	return result
}

// wrapLine wraps a single line to the given width.
func wrapLine(line Line, width int, extraBreakpoints []string) []Line {
	if len(line) == 0 {
		return []Line{{}}
	}

	// Check if line fits without wrapping
	if len(line) <= width {
		return []Line{line}
	}

	var result []Line
	var currentLine Line

	i := 0
	for i < len(line) {
		// Find the next word (sequence of cells until a breakpoint)
		wordStart := i
		wordEnd := i

		// Consume the word (non-breakpoint characters)
		for wordEnd < len(line) && !isBreakpoint(line[wordEnd], extraBreakpoints) {
			wordEnd++
		}

		// Include trailing non-whitespace breakpoints in the word
		// (dashes and extra breakpoints stay with the word)
		for wordEnd < len(line) && isBreakpoint(line[wordEnd], extraBreakpoints) && !isWhitespace(line[wordEnd]) {
			wordEnd++
		}

		word := line[wordStart:wordEnd]

		// Check if word fits on current line
		if len(currentLine)+len(word) <= width {
			// Word fits, add it
			currentLine = append(currentLine, word...)
			i = wordEnd
		} else if len(currentLine) == 0 {
			// Word doesn't fit and line is empty - hard wrap
			currentLine, i = hardWrapWord(word, width, currentLine, wordStart)
			result, currentLine = flushIfNeeded(result, currentLine, width)
		} else {
			// Word doesn't fit - start new line
			result = append(result, currentLine)
			currentLine = nil
			// Don't advance i - retry the word on the new line
		}

		// Handle trailing whitespace after the word
		for i < len(line) && isWhitespace(line[i]) {
			ws := line[i]
			if len(currentLine)+1 <= width {
				currentLine = append(currentLine, ws)
				i++
			} else {
				// Whitespace doesn't fit - start new line
				// Skip the whitespace that caused the break
				i++
				if i < len(line) {
					result = append(result, currentLine)
					currentLine = nil
				}
				break
			}
		}
	}

	// Add the last line
	if len(currentLine) > 0 || len(result) == 0 {
		result = append(result, currentLine)
	}

	return result
}

// hardWrapWord breaks a word that's too long to fit on a single line.
func hardWrapWord(word Line, width int, currentLine Line, startIdx int) (Line, int) {
	consumed := 0
	for _, cell := range word {
		// Empty cell (part of wide char) - include with previous
		if cell.Width == 0 {
			currentLine = append(currentLine, cell)
			consumed++
			continue
		}

		// Check if this cell (plus its placeholder cells) would fit
		cellWidth := cell.Width
		if len(currentLine)+cellWidth > width {
			// Cell doesn't fit
			if len(currentLine) == 0 && cellWidth > width {
				// Single cell wider than width - include it anyway
				currentLine = append(currentLine, cell)
				consumed++
				// Include placeholder cells
				for j := consumed; j < len(word) && word[j].Width == 0; j++ {
					currentLine = append(currentLine, word[j])
					consumed++
				}
			}
			break
		}

		currentLine = append(currentLine, cell)
		consumed++
	}

	return currentLine, startIdx + consumed
}

// flushIfNeeded checks if current line is at capacity and needs flushing.
func flushIfNeeded(result []Line, currentLine Line, width int) ([]Line, Line) {
	if len(currentLine) >= width {
		result = append(result, currentLine)
		return result, nil
	}
	return result, currentLine
}

// isBreakpoint returns true if the cell is a breakpoint character.
func isBreakpoint(cell Cell, extraBreakpoints []string) bool {
	return isWhitespace(cell) || cell.Content == "-" ||
		slices.Contains(extraBreakpoints, cell.Content)
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
