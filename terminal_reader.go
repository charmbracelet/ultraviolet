package uv

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/muesli/cancelreader"
	"github.com/rivo/uniseg"
)

// Logger is a simple logger interface.
type Logger interface {
	Printf(format string, v ...interface{})
}

// win32InputState is a state machine for parsing key events from the Windows
// Console API into escape sequences and utf8 runes, and keeps track of the last
// control key state to determine modifier key changes. It also keeps track of
// the last mouse button state and window size changes to determine which mouse
// buttons were released and to prevent multiple size events from firing.
//
//nolint:all
type win32InputState struct {
	ansiBuf                    [256]byte
	ansiIdx                    int
	utf16Buf                   [2]rune
	utf16Half                  bool
	lastCks                    uint32 // the last control key state for the previous event
	lastMouseBtns              uint32 // the last mouse button state for the previous event
	lastWinsizeX, lastWinsizeY int16  // the last window size for the previous event to prevent multiple size events from firing
}

// TerminalReader represents an input event reader. It reads input events and
// parses escape sequences from the terminal input buffer and translates them
// into human-readable events.
type TerminalReader struct {
	SequenceParser

	// MouseMode determines whether mouse events are enabled or not. This is a
	// platform-specific feature and is only available on Windows. When this is
	// true, the reader will be initialized to read mouse events using the
	// Windows Console API.
	MouseMode *MouseMode

	r     io.Reader
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	parser    *ansi.Parser // parser is the ANSI escape sequence parser.
	lastState parser.State // lastState is the last parser state.

	// seq is the current escape sequence type being parsed. With `0xff` used
	// for printable characters and UTF8, and the other CC values used for
	// their corresponding escape sequences.
	seq        byte
	x10mouse   bool // indicates we're waiting for X10 mouse bytes.
	pastemode  bool // indicates whether we are in bracketed paste mode.
	esc        bool
	escTimeout time.Time
	grapheme   []rune // grapheme is used to store the current grapheme cluster being parsed.

	buf    [256]byte // do we need a larger buffer?
	events []Event   // slice of pending events to be returned by ReadEvents.

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState //nolint:all

	// This indicates whether the reader is closed or not. It is used to
	// prevent	multiple calls to the Close() method.
	closed bool

	logger Logger // The logger to use for debugging.
}

// NewTerminalReader returns a new input event reader. The reader reads input
// events from the terminal and parses escape sequences into human-readable
// events. It supports reading Terminfo databases.
//
// Use [TerminalReader.UseTerminfo] to use Terminfo defined key sequences.
// Use [TerminalReader.Legacy] to control legacy key encoding behavior.
//
// Example:
//
//	r, _ := input.NewTerminalReader(os.Stdin, os.Getenv("TERM"))
//	defer r.Close()
//	events, _ := r.ReadEvents()
//	for _, ev := range events {
//	  log.Printf("%v", ev)
//	}
func NewTerminalReader(r io.Reader, termType string) *TerminalReader {
	d := new(TerminalReader)
	p := ansi.NewParser()
	p.SetDataSize(4 * 1024 * 1024) // 4 MB buffer size for string data
	p.SetParamsSize(32)            // Accept up to 32 parameters in escape sequences
	p.SetHandler(ansi.Handler{
		Print:     d.handleUtf8,
		Execute:   d.handleCc,
		HandleCsi: d.handleCsi,
		HandleEsc: d.handleEsc,
		HandleDcs: d.handleDcs,
		HandleOsc: d.handleOsc,
		HandleSos: d.handleSos,
		HandleApc: d.handleApc,
		HandlePm:  d.handlePm,
	})
	d.r = r
	d.term = termType
	d.parser = p
	return d
}

// SetLogger sets the logger to use for debugging. If nil, no logging will be
// performed.
func (d *TerminalReader) SetLogger(logger Logger) {
	d.logger = logger
}

// Start initializes the reader and prepares it for reading input events. It
// sets up the cancel reader and the key sequence parser. It also sets up the
// lookup table for key sequences if it is not already set. This function
// should be called before reading input events.
func (r *TerminalReader) Start() (err error) {
	if r.rd == nil {
		r.rd, err = newCancelreader(r.r)
		if err != nil {
			return err
		}
	}
	if r.table == nil {
		r.table = buildKeysTable(r.Legacy, r.term, r.UseTerminfo)
	}
	r.closed = false
	return nil
}

// Read implements [io.Reader].
func (d *TerminalReader) Read(p []byte) (int, error) {
	if err := d.Start(); err != nil {
		return 0, err
	}
	return d.rd.Read(p) //nolint:wrapcheck
}

// Cancel cancels the underlying reader.
func (d *TerminalReader) Cancel() bool {
	if d.rd == nil {
		return false
	}
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *TerminalReader) Close() (rErr error) {
	if d.rd == nil {
		return fmt.Errorf("reader was not initialized")
	}
	if d.closed {
		return nil
	}
	defer func() {
		if rErr != nil {
			d.closed = true
		}
	}()
	return d.rd.Close() //nolint:wrapcheck
}

func (d *TerminalReader) logf(format string, v ...interface{}) {
	if d.logger == nil {
		return
	}
	d.logger.Printf(format, v...)
}

func (d *TerminalReader) readEvents(events []Event) (int, error) {
	if err := d.Start(); err != nil {
		return 0, err
	}

	var readBuf [256]byte
	nb, err := d.rd.Read(readBuf[:])
	if err != nil {
		return 0, err //nolint:wrapcheck
	}

	buf := readBuf[:nb]

	// Lookup table first
	if len(buf) > 0 && buf[0] == ansi.ESC {
		if k, ok := d.table[string(buf)]; ok {
			d.logf("input: %q", buf)
			return copy(events, []Event{KeyPressEvent(k)}), nil
		}
	}

	d.logf("input: %q", buf)

	for i := 0; i < len(buf); i++ {
		switch {
		case d.pastemode:
			d.paste = append(d.paste, buf[i])
			continue

		case d.seq == ansi.CSI && d.x10mouse:
			d.x10mouse = false // reset the flag
			if i+3 > len(buf) {
				d.logf("invalid X10 mouse sequence: %q", buf[i:])
				d.seq = 0 // reset the sequence
				continue
			}

			cb, cx, cy := buf[i], buf[i+1], buf[i+2]
			if e := parseX10MouseEvent(cb, cx, cy); e != nil {
				d.events = append(d.events, e)
			}

			i += 2 // One is added by the loop increment
			continue
		case d.seq == ansi.SS3:
			// Scan SS3 number from 0-9
			var mod int
			for ; i < len(buf) && buf[i] >= '0' && buf[i] <= '9'; i++ {
				mod *= 10
				mod += int(buf[i] - '0')
			}

			// Scan the final GL character
			if i >= len(buf) || buf[i] < 0x21 || buf[i] > 0x7e {
				d.logf("invalid SS3 sequence: %q", buf[i:])
				d.seq = 0 // reset the sequence
				continue
			}

			gl := buf[i]

			if i+1 >= len(buf) {
				d.logf("invalid SS3 sequence: %q", buf[i:])
				d.seq = 0 // reset the sequence
				continue
			}

			i++

			var e Event
			switch gl {
			case 'a', 'b', 'c', 'd':
				e = KeyPressEvent{Code: KeyUp + rune(gl-'a'), Mod: ModCtrl}
			case 'A', 'B', 'C', 'D':
				e = KeyPressEvent{Code: KeyUp + rune(gl-'A')}
			case 'E':
				e = KeyPressEvent{Code: KeyBegin}
			case 'F':
				e = KeyPressEvent{Code: KeyEnd}
			case 'H':
				e = KeyPressEvent{Code: KeyHome}
			case 'P', 'Q', 'R', 'S':
				e = KeyPressEvent{Code: KeyF1 + rune(gl-'P')}
			case 'M':
				e = KeyPressEvent{Code: KeyKpEnter}
			case 'X':
				e = KeyPressEvent{Code: KeyKpEqual}
			case 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y':
				e = KeyPressEvent{Code: KeyKpMultiply + rune(gl-'j')}
			default:
				e = UnknownEvent("\x1bO" + string(buf[:i]))
			}

			if e != nil {
				// Handle weird SS3 <modifier> Func
				if k, ok := e.(KeyPressEvent); ok && mod > 0 {
					k.Mod |= KeyMod(mod - 1)
					e = k
				}
				d.events = append(d.events, e)
			}
		}

		action := d.parser.Advance(buf[i])
		if action == parser.ExecuteAction && buf[i] == ansi.ESC {
			d.esc = true
			d.escTimeout = time.Now()
		}
		state := d.parser.State()
		// flush grapheme if we transitioned to a non-utf8 state or we have
		// written the whole byte slice.
		if len(d.grapheme) > 0 {
			if (d.lastState == parser.GroundState && state != parser.Utf8State) || i == len(buf)-1 {
				d.flushGrapheme()
			}
		}
		d.lastState = state
	}

	n := copy(events, d.events)
	d.events = d.events[:0] // reset the events slice
	return n, nil

	// var i int
	// for i < len(buf) {
	// 	nb, ev := d.parseSequence(buf[i:])
	// 	d.logf("input: %q", buf[i:i+nb])
	//
	// 	// Handle bracketed-paste
	// 	if d.paste != nil {
	// 		if _, ok := ev.(PasteEndEvent); !ok {
	// 			d.paste = append(d.paste, buf[i])
	// 			i++
	// 			continue
	// 		}
	// 	}
	//
	// 	switch ev.(type) {
	// 	case UnknownEvent:
	// 		// If the sequence is not recognized by the parser, try looking it up.
	// 		if k, ok := d.table[string(buf[i:i+nb])]; ok {
	// 			ev = KeyPressEvent(k)
	// 		}
	// 	case PasteStartEvent:
	// 		d.paste = []byte{}
	// 	case PasteEndEvent:
	// 		// Decode the captured data into runes.
	// 		var paste []rune
	// 		for len(d.paste) > 0 {
	// 			r, w := utf8.DecodeRune(d.paste)
	// 			if r != utf8.RuneError {
	// 				paste = append(paste, r)
	// 			}
	// 			d.paste = d.paste[w:]
	// 		}
	// 		d.paste = nil // reset the buffer
	// 		events = append(events, PasteEvent(paste))
	// 	case nil:
	// 		i++
	// 		continue
	// 	}
	//
	// 	if mevs, ok := ev.(MultiEvent); ok {
	// 		events = append(events, []Event(mevs)...)
	// 	} else {
	// 		events = append(events, ev)
	// 	}
	// 	i += nb
	// }
	//
	// return events, nil
}

func (d *TerminalReader) handleCc(b byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = b

	var e Event
	switch b {
	case ansi.NUL:
		if d.Legacy&flagCtrlAt != 0 {
			e = KeyPressEvent{Code: '@', Mod: ModCtrl}
		} else {
			e = KeyPressEvent{Code: KeySpace, Mod: ModCtrl}
		}
	case ansi.BS:
		e = KeyPressEvent{Code: 'h', Mod: ModCtrl}
	case ansi.HT:
		if d.Legacy&flagCtrlI != 0 {
			e = KeyPressEvent{Code: 'i', Mod: ModCtrl}
		} else {
			e = KeyPressEvent{Code: KeyTab}
		}
	case ansi.CR:
		if d.Legacy&flagCtrlM != 0 {
			e = KeyPressEvent{Code: 'm', Mod: ModCtrl}
		} else {
			e = KeyPressEvent{Code: KeyEnter}
		}
	case ansi.ESC:
		if d.Legacy&flagCtrlOpenBracket != 0 {
			e = KeyPressEvent{Code: '[', Mod: ModCtrl}
		} else {
			e = KeyPressEvent{Code: KeyEscape}
		}
	case ansi.DEL:
		if d.Legacy&flagBackspace != 0 {
			e = KeyPressEvent{Code: KeyDelete}
		} else {
			e = KeyPressEvent{Code: KeyBackspace}
		}
	case ansi.SP:
		e = KeyPressEvent{Code: KeySpace, Text: " "}
	default:
		if b >= ansi.SOH && b <= ansi.SUB {
			// Use lower case letters for control codes
			code := rune(b + 0x60)
			e = KeyPressEvent{Code: code, Mod: ModCtrl}
		} else if b >= ansi.FS && b <= ansi.US {
			code := rune(b + 0x40)
			e = KeyPressEvent{Code: code, Mod: ModCtrl}
		}
	}
	if e != nil {
		d.events = append(d.events, e)
	}
}

const utf8Seq = 0xff // Special sequence type for UTF-8 characters

func (d *TerminalReader) handleUtf8(r rune) {
	d.seq = utf8Seq

	// TODO: Handle multirune and grapheme clusters.
	if r <= ansi.US || r == ansi.DEL {
		d.handleCc(byte(r))
		return
	} else if r > ansi.US && r < ansi.DEL {
		// ASCII printable characters
		k := KeyPressEvent{Code: r, Text: string(r)}
		if unicode.IsUpper(r) {
			// Convert uppercase letters to lowercase + shift modifier
			k.Code = unicode.ToLower(r)
			k.ShiftedCode = r
			k.Mod |= ModShift
		}
		d.events = append(d.events, k)
		return
	} else {
		d.grapheme = append(d.grapheme, r)
	}
}

func (d *TerminalReader) flushGrapheme() {
	if len(d.grapheme) == 0 {
		return
	}

	var (
		cl    string
		state = -1
		gr    = string(d.grapheme)
	)
	for len(gr) > 0 {
		cl, gr, _, state = uniseg.FirstGraphemeClusterInString(gr, state)
		code, _ := utf8.DecodeRuneInString(cl)
		if code != utf8.RuneError {
			for i := range cl {
				if i > 0 {
					code = KeyExtended
				}
			}
			d.events = append(d.events, KeyPressEvent{Code: code, Text: cl})
		}
	}
}

func (d *TerminalReader) handleCsi(cmd ansi.Cmd, pa ansi.Params) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.CSI

	var e Event = formatSeq(ansi.CSI, cmd, pa, nil)
	paramsLen := len(pa)

	switch cmd {
	case 'y' | '?'<<parser.PrefixShift | '$'<<parser.IntermedShift:
		// Report Mode (DECRPM)
		mode, _, ok := pa.Param(0, -1)
		if !ok || mode == -1 {
			break
		}
		value, _, ok := pa.Param(1, -1)
		if !ok || value == -1 {
			break
		}
		e = ModeReportEvent{Mode: ansi.DECMode(mode), Value: ansi.ModeSetting(value)}
	case 'c' | '?'<<parser.PrefixShift:
		// Primary Device Attributes
		e = parsePrimaryDevAttrs(pa)
	case 'u' | '?'<<parser.PrefixShift:
		// Kitty keyboard flags
		flags, _, ok := pa.Param(0, -1)
		if !ok || flags == -1 {
			break
		}
		e = KittyEnhancementsEvent(flags)
	case 'R' | '?'<<parser.PrefixShift:
		// This report may return a third parameter representing the page
		// number, but we don't really need it.
		row, _, ok := pa.Param(0, 1)
		if !ok {
			break
		}
		col, _, ok := pa.Param(1, 1)
		if !ok {
			break
		}
		e = CursorPositionEvent{Y: row - 1, X: col - 1}
	case 'm' | '<'<<parser.PrefixShift, 'M' | '<'<<parser.PrefixShift:
		// Handle SGR mouse
		if paramsLen == 3 {
			e = parseSGRMouseEvent(cmd, pa)
		}
	case 'm' | '>'<<parser.PrefixShift:
		// XTerm modifyOtherKeys
		mok, _, ok := pa.Param(0, 0)
		if !ok || mok != 4 {
			break
		}
		val, _, ok := pa.Param(1, -1)
		if !ok || val == -1 {
			break
		}
		e = ModifyOtherKeysEvent(val) //nolint:gosec
	case 'I':
		e = FocusEvent{}
	case 'O':
		e = BlurEvent{}
	case 'R':
		// Cursor position report OR modified F3
		row, _, rok := pa.Param(0, 1)
		col, _, cok := pa.Param(1, 1)
		if paramsLen == 2 && rok && cok {
			m := CursorPositionEvent{Y: row - 1, X: col - 1}
			if row == 1 && col-1 <= int(ModMeta|ModShift|ModAlt|ModCtrl) {
				// XXX: We cannot differentiate between cursor position report and
				// CSI 1 ; <mod> R (which is modified F3) when the cursor is at the
				// row 1. In this case, we report both messages.
				//
				// For a non ambiguous cursor position report, use
				// [ansi.RequestExtendedCursorPosition] (DECXCPR) instead.
				e = MultiEvent{KeyPressEvent{Code: KeyF3, Mod: KeyMod(col - 1)}, m}
			} else {
				e = m
			}
		}

		if paramsLen != 0 {
			break
		}

		// Unmodified key F3 (CSI R)
		fallthrough
	case 'a', 'b', 'c', 'd', 'A', 'B', 'C', 'D', 'E', 'F', 'H', 'P', 'Q', 'S', 'Z':
		var k KeyPressEvent
		switch cmd {
		case 'a', 'b', 'c', 'd':
			k = KeyPressEvent{Code: KeyUp + rune(cmd-'a'), Mod: ModShift}
		case 'A', 'B', 'C', 'D':
			k = KeyPressEvent{Code: KeyUp + rune(cmd-'A')}
		case 'E':
			k = KeyPressEvent{Code: KeyBegin}
		case 'F':
			k = KeyPressEvent{Code: KeyEnd}
		case 'H':
			k = KeyPressEvent{Code: KeyHome}
		case 'P', 'Q', 'R', 'S':
			k = KeyPressEvent{Code: KeyF1 + rune(cmd-'P')}
		case 'Z':
			k = KeyPressEvent{Code: KeyTab, Mod: ModShift}
		}
		id, _, _ := pa.Param(0, 1)
		if id == 0 {
			id = 1
		}
		mod, _, _ := pa.Param(1, 1)
		if mod == 0 {
			mod = 1
		}
		if paramsLen > 1 && id == 1 && mod != -1 {
			// CSI 1 ; <modifiers> A
			k.Mod |= KeyMod(mod - 1)
		}
		// Don't forget to handle Kitty keyboard protocol
		e = parseKittyKeyboardExt(pa, k)
	case 'M':
		// Handle X10 mouse
		d.x10mouse = true
		e = nil
	case 'y' | '$'<<parser.IntermedShift:
		// Report Mode (DECRPM)
		mode, _, ok := pa.Param(0, -1)
		if !ok || mode == -1 {
			break
		}
		val, _, ok := pa.Param(1, -1)
		if !ok || val == -1 {
			break
		}
		e = ModeReportEvent{Mode: ansi.ANSIMode(mode), Value: ansi.ModeSetting(val)}
	case 'u':
		// Kitty keyboard protocol & CSI u (fixterms)
		if paramsLen != 0 {
			e = parseKittyKeyboard(pa)
		}
	case '_':
		// Win32 Input Mode
		if paramsLen != 6 {
			break
		}

		vrc, _, _ := pa.Param(5, 0)
		rc := uint16(vrc) //nolint:gosec
		if rc == 0 {
			rc = 1
		}

		vk, _, _ := pa.Param(0, 0)
		sc, _, _ := pa.Param(1, 0)
		uc, _, _ := pa.Param(2, 0)
		kd, _, _ := pa.Param(3, 0)
		cs, _, _ := pa.Param(4, 0)
		e = d.parseWin32InputKeyEvent(
			nil,
			uint16(vk), //nolint:gosec // Vk wVirtualKeyCode
			uint16(sc), //nolint:gosec // Sc wVirtualScanCode
			rune(uc),   // Uc UnicodeChar
			kd == 1,    // Kd bKeyDown
			uint32(cs), //nolint:gosec // Cs dwControlKeyState
			rc,         // Rc wRepeatCount
			nil,
		)

	case '@', '^', '~':
		if paramsLen == 0 {
			break
		}
		param, _, _ := pa.Param(0, 0)
		switch cmd {
		case '~':
			switch param {
			case 27:
				// XTerm modifyOtherKeys 2
				if paramsLen != 3 {
					break
				}
				e = parseXTermModifyOtherKeys(pa)
			case 200:
				// bracketed-paste start
				d.events = append(d.events, PasteStartEvent{})
				d.pastemode = true
				d.paste = []byte{} // initialize the paste buffer
				return
			case 201:
				// bracketed-paste end
				var paste []rune
				for len(d.paste) > 0 {
					r, w := utf8.DecodeRune(d.paste)
					if r != utf8.RuneError {
						paste = append(paste, r)
					}
					d.paste = d.paste[w:]
				}
				d.paste = nil // reset the buffer
				d.events = append(d.events, PasteEvent(paste), PasteEndEvent{})
				d.pastemode = false

				return
			}
		}

		if e == nil {
			switch param {
			case 1, 2, 3, 4, 5, 6, 7, 8,
				11, 12, 13, 14, 15,
				17, 18, 19, 20, 21,
				23, 24, 25, 26,
				28, 29, 31, 32, 33, 34:
				var k KeyPressEvent
				switch param {
				case 1:
					if d.Legacy&flagFind != 0 {
						k = KeyPressEvent{Code: KeyFind}
					} else {
						k = KeyPressEvent{Code: KeyHome}
					}
				case 2:
					k = KeyPressEvent{Code: KeyInsert}
				case 3:
					k = KeyPressEvent{Code: KeyDelete}
				case 4:
					if d.Legacy&flagSelect != 0 {
						k = KeyPressEvent{Code: KeySelect}
					} else {
						k = KeyPressEvent{Code: KeyEnd}
					}
				case 5:
					k = KeyPressEvent{Code: KeyPgUp}
				case 6:
					k = KeyPressEvent{Code: KeyPgDown}
				case 7:
					k = KeyPressEvent{Code: KeyHome}
				case 8:
					k = KeyPressEvent{Code: KeyEnd}
				case 11, 12, 13, 14, 15:
					k = KeyPressEvent{Code: KeyF1 + rune(param-11)}
				case 17, 18, 19, 20, 21:
					k = KeyPressEvent{Code: KeyF6 + rune(param-17)}
				case 23, 24, 25, 26:
					k = KeyPressEvent{Code: KeyF11 + rune(param-23)}
				case 28, 29:
					k = KeyPressEvent{Code: KeyF15 + rune(param-28)}
				case 31, 32, 33, 34:
					k = KeyPressEvent{Code: KeyF17 + rune(param-31)}
				}

				// modifiers
				mod, _, _ := pa.Param(1, -1)
				if paramsLen > 1 && mod != -1 {
					k.Mod |= KeyMod(mod - 1)
				}

				// Handle URxvt weird keys
				switch cmd {
				case '~':
					// Don't forget to handle Kitty keyboard protocol
					e = parseKittyKeyboardExt(pa, k)
				case '^':
					k.Mod |= ModCtrl
					e = k
				case '@':
					k.Mod |= ModCtrl | ModShift
					e = k
				}
			}
		}

	case 't':
		param, _, ok := pa.Param(0, 0)
		if !ok {
			break
		}

		var winop WindowOpEvent
		winop.Op = param
		for j := 1; j < paramsLen; j++ {
			val, _, ok := pa.Param(j, 0)
			if ok {
				winop.Args = append(winop.Args, val)
			}
		}

		e = winop
	}

	if e != nil {
		d.events = append(d.events, e)
	}
}

func (d *TerminalReader) handleEsc(cmd ansi.Cmd) {
	d.flushGrapheme() // flush any pending grapheme clusters
	switch cmd.Final() {
	case 'O':
		// We need access to the bytes buffer to handle SS3 sequences.
		d.seq = ansi.SS3
	default:
		// Handle all other cases as alt+<key>
		d.events = append(d.events, KeyPressEvent{
			Mod:  ModAlt,
			Code: rune(cmd),
		})
	}
}

func (d *TerminalReader) handleDcs(cmd ansi.Cmd, pa ansi.Params, data []byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.DCS
}

func (d *TerminalReader) handleOsc(cmd int, data []byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.OSC
}

func (d *TerminalReader) handleSos(data []byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.SOS
}

func (d *TerminalReader) handleApc(data []byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.APC
}

func (d *TerminalReader) handlePm(data []byte) {
	d.flushGrapheme() // flush any pending grapheme clusters
	d.seq = ansi.PM
}

func formatSeq(seq byte, cmd ansi.Cmd, pa ansi.Params, data []byte) string {
	var s strings.Builder
	terminator := ""
	switch seq {
	case ansi.CSI:
		s.WriteString("\x1b[")
	case ansi.DCS:
		s.WriteString("\x1bP")
		terminator = "\x1b\\"
	case ansi.OSC:
		s.WriteString("\x1b]")
		terminator = "\x07"
	case ansi.SOS:
		s.WriteString("\x1bX")
		terminator = "\x1b\\"
	case ansi.APC:
		s.WriteString("\x1b_")
		terminator = "\x1b\\"
	case ansi.PM:
		s.WriteString("\x1b^")
		terminator = "\x1b\\"
	case ansi.SS3:
		s.WriteString("\x1bO")
	case ansi.ESC:
		s.WriteString("\x1b")
	default:
		s.WriteByte(seq)
	}
	if prefix := cmd.Prefix(); prefix != 0 {
		s.WriteByte(prefix)
	}
	pa.ForEach(-1, func(i, param int, hasMore bool) {
		if i > 0 {
			if hasMore {
				s.WriteByte(':')
			} else {
				s.WriteByte(';')
			}
		}
		if param != -1 {
			fmt.Fprintf(&s, "%d", param)
		}
	})
	if intermed := cmd.Intermediate(); intermed != 0 {
		s.WriteByte(intermed)
	}
	if cmd := cmd.Final(); cmd != 0 {
		s.WriteByte(cmd)
	}
	if len(data) > 0 {
		s.Write(data)
	}
	if terminator != "" {
		s.WriteString(terminator)
	}
	return s.String()
}
