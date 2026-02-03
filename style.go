package uv

import (
	"image/color"

	"github.com/charmbracelet/x/ansi"
)

// StyleParser is a state-based parser for SGR style sequences. It builds
// [Style] objects based on incoming SGR parameters and command.
//
// # Color Specification (ITU T.416 ODC)
//
// Parameters 38 (foreground), 48 (background), and 58 (underline) are followed
// by a parameter substring to specify colors:
//
//   - Mode 0: implementation defined (foreground only)
//   - Mode 1: transparent
//   - Mode 2: direct RGB color (params: [colorspace_id;] r; g; b)
//   - Mode 3: direct CMY color (params: [colorspace_id;] c; m; y)
//   - Mode 4: direct CMYK color (params: [colorspace_id;] c; m; y; k)
//   - Mode 5: indexed color (params: index)
//   - Mode 6: direct RGBA color (WezTerm extension, params: [colorspace_id;] r; g; b; a [; tolerance; tolerance_colorspace])
//
// The colorspace_id parameter is optional for modes 2, 3, 4, and 6. When present,
// it is ignored. Tolerance values (parameters 7 and 8) are also not supported.
//
// This implementation uses a parameter counter approach to handle variable-length
// color sequences efficiently, consuming up to 8 sub-parameters before finalizing.
type StyleParser struct {
	style    Style
	state    styleState
	colSel   styleColSel
	colMode  int      // tracks which color mode (2=RGB, 3=CMY, 4=CMYK, etc.)
	colParam int      // current parameter index within color sequence (0-7)
	colBuf   [8]uint8 // buffer for color component values
}

// Advance processes the next SGR parameter, updating the internal style state.
func (sp *StyleParser) Advance(param int, hasMore bool) {
	switch sp.state {
	case styleStateAttr:
		switch param {
		case 0: // Reset
			sp.style = Style{}
		case 1: // Bold
			sp.style.Attrs |= AttrBold
		case 2: // Dim/Faint
			sp.style.Attrs |= AttrFaint
		case 3: // Italic
			sp.style.Attrs |= AttrItalic
		case 4: // Underline
			if hasMore {
				sp.state = styleStateUnderline
			} else {
				sp.style.Underline = UnderlineSingle
			}
		case 5: // Slow Blink
			sp.style.Attrs |= AttrBlink
		case 6: // Rapid Blink
			sp.style.Attrs |= AttrRapidBlink
		case 7: // Reverse
			sp.style.Attrs |= AttrReverse
		case 8: // Conceal
			sp.style.Attrs |= AttrConceal
		case 9: // Crossed-out/Strikethrough
			sp.style.Attrs |= AttrStrikethrough
		case 22: // Normal Intensity (not bold or faint)
			sp.style.Attrs &^= (AttrBold | AttrFaint)
		case 23: // Not italic, not Fraktur
			sp.style.Attrs &^= AttrItalic
		case 24: // Not underlined
			sp.style.Underline = UnderlineStyleNone
		case 25: // Blink off
			sp.style.Attrs &^= (AttrBlink | AttrRapidBlink)
		case 27: // Positive (not reverse)
			sp.style.Attrs &^= AttrReverse
		case 28: // Reveal
			sp.style.Attrs &^= AttrConceal
		case 29: // Not crossed out
			sp.style.Attrs &^= AttrStrikethrough
		case 30, 31, 32, 33, 34, 35, 36, 37: // Set foreground
			sp.style.Fg = ansi.Black + ansi.BasicColor(param-30) //nolint:gosec
		case 38: // Set foreground 256 or truecolor
			sp.colSel = styleColSelFg
			sp.state = styleStateColMode
		case 39: // Default foreground
			sp.style.Fg = nil
		case 40, 41, 42, 43, 44, 45, 46, 47: // Set background
			sp.style.Bg = ansi.Black + ansi.BasicColor(param-40) //nolint:gosec
		case 48: // Set background 256 or truecolor
			sp.colSel = styleColSelBg
			sp.state = styleStateColMode
		case 49: // Default Background
			sp.style.Bg = nil
		case 58: // Set underline color
			sp.colSel = styleColSelUnderline
			sp.state = styleStateColMode
		case 59: // Default underline color
			sp.style.UnderlineColor = nil
		case 90, 91, 92, 93, 94, 95, 96, 97: // Set bright foreground
			sp.style.Fg = ansi.BrightBlack + ansi.BasicColor(param-90) //nolint:gosec
		case 100, 101, 102, 103, 104, 105, 106, 107: // Set bright background
			sp.style.Bg = ansi.BrightBlack + ansi.BasicColor(param-100) //nolint:gosec
		}
	case styleStateUnderline:
		switch param {
		case 0:
			sp.style.Underline = UnderlineStyleNone
		case 1:
			sp.style.Underline = UnderlineSingle
		case 2:
			sp.style.Underline = UnderlineDouble
		case 3:
			sp.style.Underline = UnderlineCurly
		case 4:
			sp.style.Underline = UnderlineDotted
		case 5:
			sp.style.Underline = UnderlineDashed
		}
		sp.state = styleStateAttr
	case styleStateColMode:
		sp.colMode = param
		sp.colParam = 0
		switch param {
		case 0: // 0 - implementation defined
			// Do nothing
			sp.colSel = styleColSelNone
			sp.state = styleStateAttr
		case 1: // 1 - transparent
			switch sp.colSel { //nolint:exhaustive
			case styleColSelFg:
				sp.style.Fg = color.Transparent
			case styleColSelBg:
				sp.style.Bg = color.Transparent
			case styleColSelUnderline:
				sp.style.UnderlineColor = color.Transparent
			}
			sp.colSel = styleColSelNone
			sp.state = styleStateAttr
		case 2: // 2 - RGB (requires color space ID)
			sp.state = styleStateCol
		case 3: // 3 - CMY (requires color space ID)
			sp.state = styleStateCol
		case 4: // 4 - CMYK (requires color space ID)
			sp.state = styleStateCol
		case 5: // 5 - Indexed
			sp.state = styleStateCol
		case 6: // 6 - RGBA (WezTerm extension, no color space ID)
			sp.state = styleStateCol
		default:
			sp.colSel = styleColSelNone
			sp.state = styleStateAttr
		}
	case styleStateCol:
		sp.colBuf[sp.colParam] = uint8(param)
		sp.colParam++

		// Determine if we should finalize based on mode and param count
		// For colon-separated (hasMore=true): collect up to 8 params for optional colorspace/tolerance
		// For semicolon-separated (hasMore=false): collect fixed count based on mode
		var shouldFinalize bool
		if hasMore {
			// Colon-separated: continue until hasMore=false or 8 params
			shouldFinalize = sp.colParam >= 8
		} else {
			// Semicolon-separated: finalize after fixed count based on mode
			switch sp.colMode {
			case 2, 3: // RGB, CMY: need 3 params
				shouldFinalize = sp.colParam >= 3
			case 4: // CMYK: need 4 params
				shouldFinalize = sp.colParam >= 4
			case 5: // Indexed: need 1 param
				shouldFinalize = sp.colParam >= 1
			case 6: // RGBA: need 4 params
				shouldFinalize = sp.colParam >= 4
			default:
				shouldFinalize = true
			}
		}

		if shouldFinalize {
			// Done collecting, now parse based on what we have
			done := false
			switch sp.colMode {
			case 2: // RGB: [colorspace_id;] r; g; b [; tolerance; tolerance_colorspace]
				// Support both with and without colorspace ID
				// We determine based on param count:
				// - 3 params: r, g, b (semicolon-separated or no colorspace)
				// - 4+ params: colorspace, r, g, b [, tolerance...] (colon-separated)
				if sp.colParam == 3 {
					// RGB without colorspace: r, g, b
					sp.setColor(color.RGBA{
						R: sp.colBuf[0],
						G: sp.colBuf[1],
						B: sp.colBuf[2],
						A: 255,
					})
					done = true
				} else if sp.colParam >= 4 {
					// RGB with colorspace: colorspace, r, g, b [, tolerance...]
					sp.setColor(color.RGBA{
						R: sp.colBuf[1],
						G: sp.colBuf[2],
						B: sp.colBuf[3],
						A: 255,
					})
					done = true
				}
			case 3: // CMY: [colorspace_id;] c; m; y [; tolerance; tolerance_colorspace]
				if sp.colParam == 3 {
					// CMY without colorspace
					sp.setColor(color.CMYK{
						C: sp.colBuf[0],
						M: sp.colBuf[1],
						Y: sp.colBuf[2],
						K: 0,
					})
					done = true
				} else if sp.colParam >= 4 {
					// CMY with colorspace
					sp.setColor(color.CMYK{
						C: sp.colBuf[1],
						M: sp.colBuf[2],
						Y: sp.colBuf[3],
						K: 0,
					})
					done = true
				}
			case 4: // CMYK: [colorspace_id;] c; m; y; k [; tolerance; tolerance_colorspace]
				if sp.colParam == 4 {
					// CMYK without colorspace
					sp.setColor(color.CMYK{
						C: sp.colBuf[0],
						M: sp.colBuf[1],
						Y: sp.colBuf[2],
						K: sp.colBuf[3],
					})
					done = true
				} else if sp.colParam >= 5 {
					// CMYK with colorspace
					sp.setColor(color.CMYK{
						C: sp.colBuf[1],
						M: sp.colBuf[2],
						Y: sp.colBuf[3],
						K: sp.colBuf[4],
					})
					done = true
				}
			case 5: // Indexed: index
				if sp.colParam >= 1 {
					sp.setColor(ansi.IndexedColor(sp.colBuf[0]))
					done = true
				}
			case 6: // RGBA: [colorspace_id;] r; g; b; a [; tolerance; tolerance_colorspace] (WezTerm extension)
				if sp.colParam == 4 {
					// RGBA without colorspace: r, g, b, a
					sp.setColor(color.RGBA{
						R: sp.colBuf[0],
						G: sp.colBuf[1],
						B: sp.colBuf[2],
						A: sp.colBuf[3],
					})
					done = true
				} else if sp.colParam >= 5 {
					// RGBA with colorspace: colorspace, r, g, b, a [, tolerance...]
					sp.setColor(color.RGBA{
						R: sp.colBuf[1],
						G: sp.colBuf[2],
						B: sp.colBuf[3],
						A: sp.colBuf[4],
					})
					done = true
				}
			}

			if done {
				sp.colSel = styleColSelNone
				sp.colMode = 0
				sp.colParam = 0
				sp.state = styleStateAttr
			}
		}
	}
}

// Apply applies the constructed style to the provided [Style] object.
func (sp *StyleParser) Apply(s *Style) {
	if s == nil {
		return
	}
	*s = sp.style
}

// Build finalizes and returns the constructed [Style] object.
func (sp *StyleParser) Build() Style {
	return sp.style
}

// Reset clears the internal state of the parser for reuse.
func (sp *StyleParser) Reset() {
	sp.style = Style{}
	sp.state = styleStateAttr
	sp.colSel = styleColSelNone
	sp.colMode = 0
	sp.colParam = 0
}

// setColor is a helper to set the color based on colSel.
func (sp *StyleParser) setColor(c color.Color) {
	switch sp.colSel { //nolint:exhaustive
	case styleColSelFg:
		sp.style.Fg = c
	case styleColSelBg:
		sp.style.Bg = c
	case styleColSelUnderline:
		sp.style.UnderlineColor = c
	}
}

type styleColSel byte

const (
	styleColSelNone styleColSel = iota
	styleColSelFg
	styleColSelBg
	styleColSelUnderline
)

type styleState int

const (
	styleStateAttr      styleState = iota // collecting attributes
	styleStateUnderline                   // collecting underline style extensions
	styleStateColMode                     // collecting color mode
	styleStateCol                         // collecting color parameters
)

// LinkParser is a state-based parser for hyperlink escape sequences. It builds
// [Link] objects based on incoming data.
type LinkParser struct {
	link  Link
	state linkState
	cmd   int
}

// Advance processes the next parameter for the hyperlink sequence.
func (lp *LinkParser) Advance(data []byte) {
	for _, b := range data {
		switch lp.state {
		case linkStateStart:
			if b >= '0' && b <= '9' {
				lp.cmd = lp.cmd*10 + int(b-'0')
			} else if b == ';' {
				lp.state = linkStateParams
			}
		case linkStateParams:
			if lp.cmd != 8 {
				// Unsupported command, ignore rest
				return
			}
			if b == ';' {
				lp.state = linkStateURL
			} else {
				lp.link.Params += string(b)
			}
		case linkStateURL:
			lp.link.URL += string(b)
		}
	}
}

// Apply applies the constructed link to the provided [Link] object.
func (lp *LinkParser) Apply(l *Link) {
	if l == nil {
		return
	}
	*l = lp.link
}

// Build finalizes and returns the constructed [Link] object.
func (lp *LinkParser) Build() Link {
	return lp.link
}

// Reset clears the internal state of the parser for reuse.
func (lp *LinkParser) Reset() {
	lp.link = Link{}
	lp.state = linkStateStart
	lp.cmd = 0
}

type linkState int

const (
	linkStateStart linkState = iota
	linkStateParams
	linkStateURL
)
