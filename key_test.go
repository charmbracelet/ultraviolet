package uv

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image/color"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/lucasb-eyer/go-colorful"
)

var sequences = buildKeysTable(LegacyKeyEncoding(0), "dumb", true)

func TestKeyString(t *testing.T) {
	t.Run("alt+space", func(t *testing.T) {
		k := KeyPressEvent{Code: KeySpace, Mod: ModAlt}
		if got := k.String(); got != "alt+space" {
			t.Fatalf(`expected a "alt+space", got %q`, got)
		}
	})

	t.Run("runes", func(t *testing.T) {
		k := KeyPressEvent{Code: 'a', Text: "a"}
		if got := k.String(); got != "a" {
			t.Fatalf(`expected an "a", got %q`, got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		k := KeyPressEvent{Code: 99999}
		if got := k.String(); got != "𘚟" {
			t.Fatalf(`expected a "unknown", got %q`, got)
		}
	})
}

type seqTest struct {
	seq    []byte
	Events []Event
}

var f3CurPosRegexp = regexp.MustCompile(`\x1b\[1;(\d+)R`)

// buildBaseSeqTests returns sequence tests that are valid for the
// detectSequence() function.
func buildBaseSeqTests() []seqTest {
	td := []seqTest{}
	for seq, key := range sequences {
		k := KeyPressEvent(key)
		st := seqTest{seq: []byte(seq), Events: []Event{k}}

		// XXX: This is a special case to handle F3 key sequence and cursor
		// position report having the same sequence. See [parseCsi] for more
		// information.
		if f3CurPosRegexp.MatchString(seq) {
			st.Events = []Event{k, CursorPositionEvent{Y: 0, X: int(key.Mod)}}
		}
		td = append(td, st)
	}

	// Additional special cases.
	td = append(td,
		// Unrecognized CSI sequence.
		seqTest{
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Event{
				UnknownCsiEvent([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'}),
			},
		},
		// A lone space character.
		seqTest{
			[]byte{' '},
			[]Event{
				KeyPressEvent{Code: KeySpace, Text: " "},
			},
		},
		// An escape character with the alt modifier.
		seqTest{
			[]byte{'\x1b', ' '},
			[]Event{
				KeyPressEvent{Code: KeySpace, Mod: ModAlt},
			},
		},
	)
	return td
}

func TestFocus(t *testing.T) {
	var p EventDecoder
	_, e := p.Decode([]byte("\x1b[I"))
	switch e.(type) {
	case FocusEvent:
		// ok
	default:
		t.Error("invalid sequence")
	}
}

func TestBlur(t *testing.T) {
	var p EventDecoder
	_, e := p.Decode([]byte("\x1b[O"))
	switch e.(type) {
	case BlurEvent:
		// ok
	default:
		t.Error("invalid sequence")
	}
}

func TestParseSequence(t *testing.T) {
	td := buildBaseSeqTests()
	td = append(td,
		// Teritary Device Attributes (DA3)
		seqTest{
			[]byte("\x1bP!|4368726d\x1b\\"),
			[]Event{
				TertiaryDeviceAttributesEvent("Chrm"),
			},
		},

		// XTGETTCAP response
		seqTest{
			[]byte("\x1bP1+r524742\x1b\\"),
			[]Event{
				CapabilityEvent("RGB"),
			},
		},

		// Unknown sequences.
		seqTest{
			[]byte("\x1b[z\x1bOz\x1bO2 \x1bP?1;2:3+zABC\x1b\\"),
			[]Event{
				UnknownCsiEvent("\x1b[z"),
				UnknownSs3Event("\x1bOz"),
				UnknownEvent("\x1bO2"),
				KeyPressEvent{Code: KeySpace, Text: " "},
				UnknownDcsEvent("\x1bP?1;2:3+zABC\x1b\\"),
			},
		},

		// OSC 52 read clipboard
		seqTest{
			[]byte("\x1b]52\x1b\\\x1b]52;c;!\x1b\\\x1b]52;c;aGk=\x1b\\"),
			[]Event{
				ClipboardEvent{},
				ClipboardEvent{Content: "!"},
				ClipboardEvent{Content: "hi", Selection: 'c'},
			},
		},

		// Invalid Xterm modifyOtherKeys key sequence
		seqTest{
			[]byte("\x1b[27;3~"),
			[]Event{UnknownCsiEvent("\x1b[27;3~")},
		},

		// Empty @ ^ ~
		seqTest{
			[]byte("\x1b[@\x1b[^\x1b[~"),
			[]Event{
				UnknownCsiEvent("\x1b[@"),
				UnknownCsiEvent("\x1b[^"),
				UnknownCsiEvent("\x1b[~"),
			},
		},

		// Win32 input mode key sequences
		seqTest{
			[]byte("\x1b[65;0;97;1;0;1_\x1b[0;0;0_"),
			[]Event{
				KeyPressEvent{Code: 'a', BaseCode: 'a', Text: "a"},
				UnknownCsiEvent("\x1b[0;0;0_"),
			},
		},

		// Report mode responses
		seqTest{
			[]byte("\x1b[2;1$y\x1b[$y\x1b[2$y\x1b[2;$y"),
			[]Event{
				ModeReportEvent{Mode: ansi.KeyboardActionMode, Value: ansi.ModeSet},
				UnknownCsiEvent("\x1b[$y"),
				UnknownCsiEvent("\x1b[2$y"),
				ModeReportEvent{Mode: ansi.KeyboardActionMode, Value: ansi.ModeNotRecognized},
			},
		},

		// Short X10 mouse input
		seqTest{
			[]byte("\x1b[M !"),
			[]Event{
				UnknownCsiEvent("\x1b[M"),
				KeyPressEvent{Code: ' ', Text: " "},
				KeyPressEvent{Code: '!', Text: "!"},
			},
		},

		// Invalid report mode responses
		seqTest{
			[]byte("\x1b[?$y\x1b[?1049$y\x1b[?1049;$y"),
			[]Event{
				UnknownCsiEvent("\x1b[?$y"),
				UnknownCsiEvent("\x1b[?1049$y"),
				ModeReportEvent{Mode: ansi.AltScreenSaveCursorMode, Value: ansi.ModeNotRecognized},
			},
		},

		// Xterm modifyOtherKeys response
		seqTest{
			[]byte("\x1b[>4;1m\x1b[>4m\x1b[>3m"),
			[]Event{
				ModifyOtherKeysEvent(1),
				UnknownCsiEvent("\x1b[>4m"),
				UnknownCsiEvent("\x1b[>3m"),
			},
		},

		// F3 with modifier key press or cursor position report
		seqTest{
			[]byte("\x1b[1;5R\x1b[1;5;7R"),
			[]Event{
				KeyPressEvent{Code: KeyF3, Mod: ModCtrl},
				CursorPositionEvent{Y: 0, X: 4},
				UnknownCsiEvent("\x1b[1;5;7R"),
			},
		},

		// Cursor position report
		seqTest{
			[]byte("\x1b[?12;34R\x1b[?14R"),
			[]Event{
				CursorPositionEvent{Y: 11, X: 33},
				UnknownCsiEvent("\x1b[?14R"),
			},
		},

		// Unknown CSI sequence
		seqTest{
			[]byte("\x1b[10;2;3c"),
			[]Event{UnknownCsiEvent([]byte("\x1b[10;2;3c"))},
		},

		// Kitty Keyboard response
		seqTest{
			[]byte("\x1b[?16u\x1b[?u"),
			[]Event{
				KittyEnhancementsEvent(16),
				KittyEnhancementsEvent(0),
			},
		},

		// Secondary Device Attributes (DA2)
		seqTest{
			[]byte("\x1b[>1;2;3c"),
			[]Event{SecondaryDeviceAttributesEvent{1, 2, 3}},
		},

		// Primary Device Attributes (DA1)
		seqTest{
			[]byte("\x1b[?1;2;3c"),
			[]Event{PrimaryDeviceAttributesEvent{1, 2, 3}},
		},

		// esc followed by non-key event sequence
		seqTest{
			[]byte("\x1b\x1b[?2004;1$y"),
			[]Event{
				KeyPressEvent{Code: KeyEscape},
				ModeReportEvent{Mode: ansi.BracketedPasteMode, Value: ansi.ModeSet},
			},
		},

		// 8-bit sequences
		seqTest{
			[]byte(
				"\x9bA" + // CSI A
					"\x8fA" + // SS3 A
					"\x90>|Ultraviolet\x1b\\" + // DCS >|Ultraviolet ST
					"\x9d11;#123456\x9c" + // OSC 11 ; #123456 ST
					"\x98hi\x9c" + // SOS hi ST
					"\x9fhello\x9c" + // APC hello ST
					"\x9ebye\x9c", // PM bye ST
			),
			[]Event{
				KeyPressEvent{Code: KeyUp},
				KeyPressEvent{Code: KeyUp},
				TerminalVersionEvent("Ultraviolet"),
				BackgroundColorEvent{ansi.XParseColor("#123456")},
				UnknownSosEvent("\x98hi\x9c"),
				UnknownApcEvent("\x9fhello\x9c"),
				UnknownPmEvent("\x9ebye\x9c"),
			},
		},

		// Empty input
		seqTest{
			[]byte(""),
			[]Event(nil),
		},

		// Broken escape sequence introducers.
		seqTest{
			[]byte("\x1b["), // CSI
			[]Event{KeyPressEvent{Code: '[', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b]"), // OSC
			[]Event{KeyPressEvent{Code: ']', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b^"), // PM
			[]Event{KeyPressEvent{Code: '^', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b_"), // APC
			[]Event{KeyPressEvent{Code: '_', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1bP"), // DCS
			[]Event{KeyPressEvent{Code: 'p', Mod: ModShift | ModAlt}},
		},
		seqTest{
			[]byte("\x1bX"), // SOS
			[]Event{KeyPressEvent{Code: 'x', Mod: ModShift | ModAlt}},
		},
		seqTest{
			[]byte("\x1bO"), // SS3
			[]Event{KeyPressEvent{Code: 'o', Mod: ModShift | ModAlt}},
		},
		seqTest{
			[]byte("\x1b"), // ESC
			[]Event{KeyPressEvent{Code: KeyEscape}},
		},

		// Kitty invalid key sequence
		seqTest{
			[]byte("\x1b[u"),
			[]Event{UnknownCsiEvent("\x1b[u")},
		},

		// Kitty printable keys with lock modifiers.
		seqTest{
			[]byte("\x1b[97;65u" + // caps lock on
				"\x1b[97;2u" + // shift pressed
				"\x1b[97;65u" + // caps lock on
				"\x1b[97;66u" + // caps lock on and shift pressed
				"\x1b[97;129u" + // num lock on
				"\x1b[97;130u" + // num lock on and shift pressed
				"\x1b[97;194u"), // num lock on and caps lock on and shift pressed
			[]Event{
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModCapsLock},
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModShift},
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModCapsLock},
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModCapsLock | ModShift},
				KeyPressEvent{Code: 'a', Text: "a", Mod: ModNumLock},
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModNumLock | ModShift},
				KeyPressEvent{Code: 'a', Text: "A", Mod: ModNumLock | ModCapsLock | ModShift},
			},
		},

		// Kitty NumPad keys.
		seqTest{
			[]byte("\x1b[57399u" +
				"\x1b[57400u" +
				"\x1b[57401u" +
				"\x1b[57402u" +
				"\x1b[57403u" +
				"\x1b[57404u" +
				"\x1b[57405u" +
				"\x1b[57406u" +
				"\x1b[57407u" +
				"\x1b[57408u" +
				"\x1b[57409u" +
				"\x1b[57410u" +
				"\x1b[57411u" +
				"\x1b[57412u" +
				"\x1b[57413u" +
				"\x1b[57414u" +
				"\x1b[57415u" +
				"\x1b[57416u" +
				"\x1b[57417u" +
				"\x1b[57418u" +
				"\x1b[57419u" +
				"\x1b[57420u" +
				"\x1b[57421u" +
				"\x1b[57422u" +
				"\x1b[57423u" +
				"\x1b[57424u" +
				"\x1b[57425u"),
			[]Event{
				KeyPressEvent{Code: KeyKp0, Text: "0"},
				KeyPressEvent{Code: KeyKp1, Text: "1"},
				KeyPressEvent{Code: KeyKp2, Text: "2"},
				KeyPressEvent{Code: KeyKp3, Text: "3"},
				KeyPressEvent{Code: KeyKp4, Text: "4"},
				KeyPressEvent{Code: KeyKp5, Text: "5"},
				KeyPressEvent{Code: KeyKp6, Text: "6"},
				KeyPressEvent{Code: KeyKp7, Text: "7"},
				KeyPressEvent{Code: KeyKp8, Text: "8"},
				KeyPressEvent{Code: KeyKp9, Text: "9"},
				KeyPressEvent{Code: KeyKpDecimal, Text: "."},
				KeyPressEvent{Code: KeyKpDivide, Text: "/"},
				KeyPressEvent{Code: KeyKpMultiply, Text: "*"},
				KeyPressEvent{Code: KeyKpMinus, Text: "-"},
				KeyPressEvent{Code: KeyKpPlus, Text: "+"},
				KeyPressEvent{Code: KeyKpEnter},
				KeyPressEvent{Code: KeyKpEqual, Text: "="},
				KeyPressEvent{Code: KeyKpSep, Text: ","},
				KeyPressEvent{Code: KeyKpLeft},
				KeyPressEvent{Code: KeyKpRight},
				KeyPressEvent{Code: KeyKpUp},
				KeyPressEvent{Code: KeyKpDown},
				KeyPressEvent{Code: KeyKpPgUp},
				KeyPressEvent{Code: KeyKpPgDown},
				KeyPressEvent{Code: KeyKpHome},
				KeyPressEvent{Code: KeyKpEnd},
				KeyPressEvent{Code: KeyKpInsert},
			},
		},

		// Invalid CSI sequence.
		seqTest{
			[]byte("\x1b[?2004;1$y"),
			[]Event{ModeReportEvent{Mode: ansi.BracketedPasteMode, Value: ansi.ModeSet}},
		},

		// Invalid CSI sequence.
		seqTest{
			[]byte("\x1b[?2004;1$"),
			[]Event{UnknownEvent("\x1b[?2004;1$")},
		},

		// Light/dark color scheme reports.
		seqTest{
			[]byte("\x1b[?997;1n"),
			[]Event{DarkColorSchemeEvent{}},
		},
		seqTest{
			[]byte("\x1b[?997;2n"),
			[]Event{LightColorSchemeEvent{}},
		},

		// ESC [ [ansi.CSI]
		seqTest{
			[]byte("\x1b["),
			[]Event{KeyPressEvent{Code: '[', Mod: ModAlt}},
		},
		// ESC ] [ansi.OSC]
		seqTest{
			[]byte("\x1b]"),
			[]Event{KeyPressEvent{Code: ']', Mod: ModAlt}},
		},
		// ESC ^ [ansi.PM]
		seqTest{
			[]byte("\x1b^"),
			[]Event{KeyPressEvent{Code: '^', Mod: ModAlt}},
		},
		// ESC _ [ansi.APC]
		seqTest{
			[]byte("\x1b_"),
			[]Event{KeyPressEvent{Code: '_', Mod: ModAlt}},
		},
		// ESC p
		seqTest{
			[]byte("\x1bp"),
			[]Event{KeyPressEvent{Code: 'p', Mod: ModAlt}},
		},
		// ESC P [ansi.DCS]
		seqTest{
			[]byte("\x1bP"),
			[]Event{KeyPressEvent{Code: 'p', Mod: ModShift | ModAlt}},
		},
		// ESC x
		seqTest{
			[]byte("\x1bx"),
			[]Event{KeyPressEvent{Code: 'x', Mod: ModAlt}},
		},
		// ESC X [ansi.SOS]
		seqTest{
			[]byte("\x1bX"),
			[]Event{KeyPressEvent{Code: 'x', Mod: ModShift | ModAlt}},
		},

		// OSC 11 with ST termination.
		seqTest{
			[]byte("\x1b]11;#123456\x1b\\"),
			[]Event{BackgroundColorEvent{
				Color: func() color.Color {
					c, _ := colorful.Hex("#123456")
					return c
				}(),
			}},
		},

		// Kitty Graphics response.
		seqTest{
			[]byte("\x1b_Ga=t;OK\x1b\\"),
			[]Event{KittyGraphicsEvent{
				Options: kitty.Options{Action: kitty.Transmit},
				Payload: []byte("OK"),
			}},
		},
		seqTest{
			[]byte("\x1b_Gi=99,I=13;OK\x1b\\"),
			[]Event{KittyGraphicsEvent{
				Options: kitty.Options{ID: 99, Number: 13},
				Payload: []byte("OK"),
			}},
		},
		seqTest{
			[]byte("\x1b_Gi=1337,q=1;EINVAL:your face\x1b\\"),
			[]Event{KittyGraphicsEvent{
				Options: kitty.Options{ID: 1337, Quite: 1},
				Payload: []byte("EINVAL:your face"),
			}},
		},

		// Xterm modifyOtherKeys CSI 27 ; <modifier> ; <code> ~
		seqTest{
			[]byte("\x1b[27;3;20320~"),
			[]Event{KeyPressEvent{Code: '你', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;65~"),
			[]Event{KeyPressEvent{Code: 'A', Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;8~"),
			[]Event{KeyPressEvent{Code: KeyBackspace, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;27~"),
			[]Event{KeyPressEvent{Code: KeyEscape, Mod: ModAlt}},
		},
		seqTest{
			[]byte("\x1b[27;3;127~"),
			[]Event{KeyPressEvent{Code: KeyBackspace, Mod: ModAlt}},
		},

		// Xterm report window text area size in pixels.
		seqTest{
			[]byte("\x1b[4;24;80t" + // window pixel size
				"\x1b[6;13;7t" + // single cell size
				"\x1b[8;24;80t" + // window cells area size
				"\x1b[48;24;80;312;560t" + // in-band resize response
				"\x1b[t" + // invalid
				"\x1b[999t" + // invalid
				"\x1b[999;1t" + // invalid
				""),
			[]Event{
				WindowPixelSizeEvent{Width: 80, Height: 24},
				CellSizeEvent{Width: 7, Height: 13},
				WindowSizeEvent{Width: 80, Height: 24},
				WindowSizeEvent{Width: 80, Height: 24},
				WindowPixelSizeEvent{Width: 560, Height: 312},
				UnknownCsiEvent("\x1b[t"),
				WindowOpEvent{Op: 999},
				WindowOpEvent{Op: 999, Args: []int{1}},
			},
		},

		// Kitty keyboard / CSI u (fixterms)
		seqTest{
			[]byte("\x1b[1B"),
			[]Event{KeyPressEvent{Code: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;B"),
			[]Event{KeyPressEvent{Code: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;4B"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;4:1B"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[1;4:2B"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyDown, IsRepeat: true}},
		},
		seqTest{
			[]byte("\x1b[1;4:3B"),
			[]Event{KeyReleaseEvent{Mod: ModShift | ModAlt, Code: KeyDown}},
		},
		seqTest{
			[]byte("\x1b[8~"),
			[]Event{KeyPressEvent{Code: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;~"),
			[]Event{KeyPressEvent{Code: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[8;10~"),
			[]Event{KeyPressEvent{Mod: ModShift | ModMeta, Code: KeyEnd}},
		},
		seqTest{
			[]byte("\x1b[27;4u"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyEscape}},
		},
		seqTest{
			[]byte("\x1b[127;4u"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyBackspace}},
		},
		seqTest{
			[]byte("\x1b[57358;4u"),
			[]Event{KeyPressEvent{Mod: ModShift | ModAlt, Code: KeyCapsLock}},
		},
		seqTest{
			[]byte("\x1b[9;2u"),
			[]Event{KeyPressEvent{Mod: ModShift, Code: KeyTab}},
		},
		seqTest{
			[]byte("\x1b[195;u"),
			[]Event{KeyPressEvent{Text: "Ã", Code: 'Ã'}},
		},
		seqTest{
			[]byte("\x1b[20320;2u"),
			[]Event{KeyPressEvent{Text: "你", Mod: ModShift, Code: '你'}},
		},
		seqTest{
			[]byte("\x1b[195;:1u"),
			[]Event{KeyPressEvent{Text: "Ã", Code: 'Ã'}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Event{KeyReleaseEvent{Code: 'Ã', Text: "Ã", Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:2u"),
			[]Event{KeyPressEvent{Code: 'Ã', Text: "Ã", IsRepeat: true, Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:1u"),
			[]Event{KeyPressEvent{Code: 'Ã', Text: "Ã", Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[195;2:3u"),
			[]Event{KeyReleaseEvent{Code: 'Ã', Text: "Ã", Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[97;2;65u"),
			[]Event{KeyPressEvent{Code: 'a', Text: "A", Mod: ModShift}},
		},
		seqTest{
			[]byte("\x1b[97;;229u"),
			[]Event{KeyPressEvent{Code: 'a', Text: "å"}},
		},

		// focus/blur
		seqTest{
			[]byte{'\x1b', '[', 'I'},
			[]Event{
				FocusEvent{},
			},
		},
		seqTest{
			[]byte{'\x1b', '[', 'O'},
			[]Event{
				BlurEvent{},
			},
		},
		// Mouse event.
		seqTest{
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Event{
				MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		// SGR Mouse event.
		seqTest{
			[]byte("\x1b[<0;33;17M"),
			[]Event{
				MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
			},
		},
		// Runes.
		seqTest{
			[]byte{'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
			},
		},
		seqTest{
			[]byte{'\x1b', 'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModAlt},
			},
		},
		seqTest{
			[]byte{'a', 'a', 'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'a', Text: "a"},
			},
		},
		// Multi-byte rune.
		seqTest{
			[]byte("☃"),
			[]Event{
				KeyPressEvent{Code: '☃', Text: "☃"},
			},
		},
		seqTest{
			[]byte("\x1b☃"),
			[]Event{
				KeyPressEvent{Code: '☃', Mod: ModAlt},
			},
		},
		// Standalone control characters.
		seqTest{
			[]byte{'\x1b'},
			[]Event{
				KeyPressEvent{Code: KeyEscape},
			},
		},
		seqTest{
			[]byte{ansi.SOH},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.SOH},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModCtrl | ModAlt},
			},
		},
		seqTest{
			[]byte{ansi.NUL},
			[]Event{
				KeyPressEvent{Code: KeySpace, Mod: ModCtrl},
			},
		},
		seqTest{
			[]byte{'\x1b', ansi.NUL},
			[]Event{
				KeyPressEvent{Code: KeySpace, Mod: ModCtrl | ModAlt},
			},
		},
		// C1 control characters.
		seqTest{
			[]byte{'\x80'},
			[]Event{
				KeyPressEvent{Code: rune(0x80 - '@'), Mod: ModCtrl | ModAlt},
			},
		},
	)

	if !isWindows {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			[]Event{
				UnknownEvent(rune(0xfe)),
			},
		})
	}

	var p EventDecoder
	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Event
			buf := tc.seq
			for len(buf) > 0 {
				width, Event := p.Decode(buf)
				switch Event := Event.(type) {
				case MultiEvent:
					events = append(events, Event...)
				default:
					events = append(events, Event)
				}
				buf = buf[width:]
			}
			if len(tc.Events) != len(events) {
				t.Fatalf("\nexpected %d events for %q:\n    %#v\ngot %d:\n    %#v", len(tc.Events), tc.seq, tc.Events, len(events), events)
			}
			for i := range tc.Events {
				if !reflect.DeepEqual(tc.Events[i], events[i]) {
					t.Errorf("\nexpected event %d for %q:\n    %#v\ngot:\n    %#v", i, tc.seq, tc.Events[i], events[i])
				}
			}
		})
	}
}

func TestEmptyBufferDecode(t *testing.T) {
	var p EventDecoder
	if n, e := p.Decode([]byte{}); n != 0 || e != nil {
		t.Errorf("expected (0, nil), got (%d, %v)", n, e)
	}
}

func TestSplitReads(t *testing.T) {
	expect := []Event{
		KeyPressEvent{Code: 'a', Text: "a"},
		KeyPressEvent{Code: 'b', Text: "b"},
		KeyPressEvent{Code: 'c', Text: "c"},
		KeyPressEvent{Code: KeyUp},
		MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		FocusEvent{},
		MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		BlurEvent{},
		MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		KeyPressEvent{Code: KeyUp},
		MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		FocusEvent{},
		UnknownEvent("\x1b[12;34;9"),
	}
	inputs := []string{
		"abc",
		"\x1b[A",
		"\x1b[<0;33",
		";17M",
		"\x1b[I",
		"\x1b",
		"[",
		"<",
		"0",
		";",
		"3",
		"3",
		";",
		"1",
		"7",
		"M",
		"\x1b[O",
		"\x1b",
		"]",
		"2",
		";",
		"a",
		"b",
		"c",
		"\x1b",
		"\x1b[",
		"<0;3",
		"3;17M",
		"\x1b[A\x1b[",
		"<0;33;17M\x1b[",
		"<0;33;17M\x1b[I",
		"\x1b[12;34;9",
	}

	r := LimitedReader(strings.NewReader(strings.Join(inputs, "")), 8)
	drv := NewTerminalReader(r, "dumb")
	drv.SetLogger(TLogger{t})

	eventc := make(chan Event)
	go func(t testing.TB) {
		defer close(eventc)
		if err := drv.StreamEvents(t.Context(), eventc); err != nil {
			t.Errorf("error streaming events: %v", err)
		}
	}(t)

	var events []Event
	for ev := range eventc {
		events = append(events, ev)
	}

	if !reflect.DeepEqual(expect, events) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, events)
	}
}

func TestReadLongInput(t *testing.T) {
	expect := make([]Event, 1000)
	for i := 0; i < 1000; i++ {
		expect[i] = KeyPressEvent{Code: 'a', Text: "a"}
	}
	input := strings.Repeat("a", 1000)
	rdr := strings.NewReader(input)
	drv := NewTerminalReader(rdr, "dumb")

	eventc := make(chan Event)
	go func(t testing.TB) {
		defer close(eventc)
		if err := drv.StreamEvents(context.TODO(), eventc); err != nil {
			t.Errorf("error streaming events: %v", err)
		}
	}(t)

	var events []Event
	for ev := range eventc {
		events = append(events, ev)
	}

	if !reflect.DeepEqual(expect, events) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, events)
	}
}

func TestReadInput(t *testing.T) {
	type test struct {
		keyname string
		in      []byte
		out     []Event
	}
	testData := []test{
		{
			"ignored osc",
			[]byte("\x1b]11;#123456\x18\x1b]11;#123456\x1a\x1b]11;#123456\x1b"),
			[]Event(nil),
		},
		{
			"ignored apc",
			[]byte("\x9f\x9c\x1b_hello\x1b\x1b_hello\x18\x1b_abc\x1b\\\x1ba"),
			[]Event{
				UnknownApcEvent("\x9f\x9c"),
				UnknownApcEvent("\x1b_abc\x1b\\"),
				KeyPressEvent{Code: 'a', Mod: ModAlt},
			},
		},
		{
			"alt+] alt+'",
			[]byte("\x1b]\x1b'"),
			[]Event{
				KeyPressEvent{Code: ']', Mod: ModAlt},
				KeyPressEvent{Code: '\'', Mod: ModAlt},
			},
		},
		{
			"alt+^ alt+&",
			[]byte("\x1b^\x1b&"),
			[]Event{
				KeyPressEvent{Code: '^', Mod: ModAlt},
				KeyPressEvent{Code: '&', Mod: ModAlt},
			},
		},
		{
			"a",
			[]byte{'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
			},
		},
		{
			"space",
			[]byte{' '},
			[]Event{
				KeyPressEvent{Code: KeySpace, Text: " "},
			},
		},
		{
			"a alt+a",
			[]byte{'a', '\x1b', 'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'a', Mod: ModAlt},
			},
		},
		{
			"a alt+a a",
			[]byte{'a', '\x1b', 'a', 'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'a', Mod: ModAlt},
				KeyPressEvent{Code: 'a', Text: "a"},
			},
		},
		{
			"ctrl+a",
			[]byte{byte(ansi.SOH)},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModCtrl},
			},
		},
		{
			"ctrl+a ctrl+b",
			[]byte{byte(ansi.SOH), byte(ansi.STX)},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModCtrl},
				KeyPressEvent{Code: 'b', Mod: ModCtrl},
			},
		},
		{
			"alt+a",
			[]byte{byte(0x1b), 'a'},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModAlt},
			},
		},
		{
			"a b c d",
			[]byte{'a', 'b', 'c', 'd'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'b', Text: "b"},
				KeyPressEvent{Code: 'c', Text: "c"},
				KeyPressEvent{Code: 'd', Text: "d"},
			},
		},
		{
			"up",
			[]byte("\x1b[A"),
			[]Event{
				KeyPressEvent{Code: KeyUp},
			},
		},
		{
			"wheel up",
			[]byte{'\x1b', '[', 'M', byte(32) + 0b0100_0000, byte(65), byte(49)},
			[]Event{
				MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
			},
		},
		{
			"left motion release",
			[]byte{
				'\x1b', '[', 'M', byte(32) + 0b0010_0000, byte(32 + 33), byte(16 + 33),
				'\x1b', '[', 'M', byte(32) + 0b0000_0011, byte(64 + 33), byte(32 + 33),
			},
			[]Event{
				MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
				MouseReleaseEvent{X: 64, Y: 32, Button: MouseNone},
			},
		},
		{
			"shift+tab",
			[]byte{'\x1b', '[', 'Z'},
			[]Event{
				KeyPressEvent{Code: KeyTab, Mod: ModShift},
			},
		},
		{
			"enter",
			[]byte{'\r'},
			[]Event{KeyPressEvent{Code: KeyEnter}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\r'},
			[]Event{
				KeyPressEvent{Code: KeyEnter, Mod: ModAlt},
			},
		},
		{
			"insert",
			[]byte{'\x1b', '[', '2', '~'},
			[]Event{
				KeyPressEvent{Code: KeyInsert},
			},
		},
		{
			"ctrl+alt+a",
			[]byte{'\x1b', byte(ansi.SOH)},
			[]Event{
				KeyPressEvent{Code: 'a', Mod: ModCtrl | ModAlt},
			},
		},
		{
			"CSI?----X?",
			[]byte{'\x1b', '[', '-', '-', '-', '-', 'X'},
			[]Event{UnknownCsiEvent([]byte{'\x1b', '[', '-', '-', '-', '-', 'X'})},
		},
		// Powershell sequences.
		{
			"up",
			[]byte{'\x1b', 'O', 'A'},
			[]Event{KeyPressEvent{Code: KeyUp}},
		},
		{
			"down",
			[]byte{'\x1b', 'O', 'B'},
			[]Event{KeyPressEvent{Code: KeyDown}},
		},
		{
			"right",
			[]byte{'\x1b', 'O', 'C'},
			[]Event{KeyPressEvent{Code: KeyRight}},
		},
		{
			"left",
			[]byte{'\x1b', 'O', 'D'},
			[]Event{KeyPressEvent{Code: KeyLeft}},
		},
		{
			"alt+enter",
			[]byte{'\x1b', '\x0d'},
			[]Event{KeyPressEvent{Code: KeyEnter, Mod: ModAlt}},
		},
		{
			"alt+backspace",
			[]byte{'\x1b', '\x7f'},
			[]Event{KeyPressEvent{Code: KeyBackspace, Mod: ModAlt}},
		},
		{
			"ctrl+space",
			[]byte{'\x00'},
			[]Event{KeyPressEvent{Code: KeySpace, Mod: ModCtrl}},
		},
		{
			"ctrl+alt+space",
			[]byte{'\x1b', '\x00'},
			[]Event{KeyPressEvent{Code: KeySpace, Mod: ModCtrl | ModAlt}},
		},
		{
			"esc",
			[]byte{'\x1b'},
			[]Event{KeyPressEvent{Code: KeyEscape}},
		},
		{
			"alt+esc",
			[]byte{'\x1b', '\x1b'},
			[]Event{KeyPressEvent{Code: KeyEscape, Mod: ModAlt}},
		},
		{
			"a b o",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', ' ', 'b',
				'\x1b', '[', '2', '0', '1', '~',
				'o',
			},
			[]Event{
				PasteStartEvent{},
				PasteEvent("a b"),
				PasteEndEvent{},
				KeyPressEvent{Code: 'o', Text: "o"},
			},
		},
		{
			"a\x03\nb",
			[]byte{
				'\x1b', '[', '2', '0', '0', '~',
				'a', '\x03', '\n', 'b',
				'\x1b', '[', '2', '0', '1', '~',
			},
			[]Event{
				PasteStartEvent{},
				PasteEvent("a\x03\nb"),
				PasteEndEvent{},
			},
		},
		{
			"?0xfe?",
			[]byte{'\xfe'},
			[]Event{
				UnknownEvent(rune(0xfe)),
			},
		},
		{
			"a ?0xfe?   b",
			[]byte{'a', '\xfe', ' ', 'b'},
			[]Event{
				KeyPressEvent{Code: 'a', Text: "a"},
				UnknownEvent(rune(0xfe)),
				KeyPressEvent{Code: KeySpace, Text: " "},
				KeyPressEvent{Code: 'b', Text: "b"},
			},
		},
	}

	for i, td := range testData {
		t.Run(fmt.Sprintf("%d: %s", i, td.keyname), func(t *testing.T) {
			events := testReadInputs(t, bytes.NewReader(td.in))
			var buf strings.Builder
			for i, event := range events {
				if i > 0 {
					buf.WriteByte(' ')
				}
				if s, ok := event.(fmt.Stringer); ok {
					buf.WriteString(s.String())
				} else {
					fmt.Fprintf(&buf, "%#v:%T", event, event)
				}
			}

			if len(events) != len(td.out) {
				t.Fatalf("unexpected message list length: got %d, expected %d\n  got: %#v\n  expected: %#v\n", len(events), len(td.out), events, td.out)
			}

			if len(td.out) != len(events) {
				t.Fatalf("expected %d events, got %d: %s", len(td.out), len(events), buf.String())
			}
			for i, e := range events {
				if !reflect.DeepEqual(td.out[i], e) {
					t.Errorf("expected event %d to be %T %v, got %T %v", i, td.out[i], td.out[i], e, e)
				}
			}
			if !reflect.DeepEqual(td.out, events) {
				t.Fatalf("expected:\n%#v\ngot:\n%#v", td.out, events)
			}
		})
	}
}

func testReadInputs(t *testing.T, input io.Reader) []Event {
	drv := NewTerminalReader(input, "dumb")
	drv.SetLogger(TLogger{t})

	eventc := make(chan Event)
	go func(t testing.TB) {
		defer close(eventc)
		if err := drv.StreamEvents(context.TODO(), eventc); err != nil {
			t.Errorf("error streaming events: %v", err)
		}
	}(t)

	var events []Event
	for ev := range eventc {
		events = append(events, ev)
	}

	return events
}

// randTest defines the test input and expected output for a sequence
// of interleaved control sequences and control characters.
type randTest struct {
	data    []byte
	lengths []int
	names   []string
}

// seed is the random seed to randomize the input. This helps check
// that all the sequences get ultimately exercised.
var seed = flag.Int64("seed", 0, "random seed (0 to autoselect)")

// genRandomData generates a randomized test, with a random seed unless
// the seed flag was set.
func genRandomData(logfn func(int64), length int) randTest {
	// We'll use a random source. However, we give the user the option
	// to override it to a specific value for reproduceability.
	s := *seed
	if s == 0 {
		s = time.Now().UnixNano()
	}
	// Inform the user so they know what to reuse to get the same data.
	logfn(s)
	return genRandomDataWithSeed(s, length)
}

// genRandomDataWithSeed generates a randomized test with a fixed seed.
func genRandomDataWithSeed(s int64, length int) randTest {
	src := rand.NewSource(s)
	r := rand.New(src)

	// allseqs contains all the sequences, in sorted order. We sort
	// to make the test deterministic (when the seed is also fixed).
	type seqpair struct {
		seq  string
		name string
	}
	var allseqs []seqpair
	for seq, key := range sequences {
		allseqs = append(allseqs, seqpair{seq, key.String()})
	}
	sort.Slice(allseqs, func(i, j int) bool { return allseqs[i].seq < allseqs[j].seq })

	// res contains the computed test.
	var res randTest

	for len(res.data) < length {
		alt := r.Intn(2)
		prefix := ""
		esclen := 0
		if alt == 1 {
			prefix = "alt+"
			esclen = 1
		}
		kind := r.Intn(3)
		switch kind {
		case 0:
			// A control character.
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res.data = append(res.data, 1)
			res.names = append(res.names, "ctrl+"+prefix+"a")
			res.lengths = append(res.lengths, 1+esclen)

		case 1, 2:
			// A sequence.
			seqi := r.Intn(len(allseqs))
			s := allseqs[seqi]
			if strings.Contains(s.name, "alt+") || strings.Contains(s.name, "meta+") {
				esclen = 0
				prefix = ""
				alt = 0
			}
			if alt == 1 {
				res.data = append(res.data, '\x1b')
			}
			res.data = append(res.data, s.seq...)
			if strings.HasPrefix(s.name, "ctrl+") {
				prefix = "ctrl+" + prefix
			}
			name := prefix + strings.TrimPrefix(s.name, "ctrl+")
			res.names = append(res.names, name)
			res.lengths = append(res.lengths, len(s.seq)+esclen)
		}
	}
	return res
}

func FuzzParseSequence(f *testing.F) {
	var p EventDecoder
	for seq := range sequences {
		f.Add(seq)
	}
	f.Add("\x1b]52;?\x07")                      // OSC 52
	f.Add("\x1b]11;rgb:0000/0000/0000\x1b\\")   // OSC 11
	f.Add("\x1bP>|charm terminal(0.1.2)\x1b\\") // DCS (XTVERSION)
	f.Add("\x1b_Gi=123\x1b\\")                  // APC
	f.Fuzz(func(t *testing.T, seq string) {
		n, _ := p.Decode([]byte(seq))
		if n == 0 && seq != "" {
			t.Errorf("expected a non-zero width for %q", seq)
		}
	})
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	var p EventDecoder
	td := genRandomDataWithSeed(123, 10000)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			w, _ = p.Decode(td.data[j:])
		}
	}
}

func TestMouseEvent_String(t *testing.T) {
	tt := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			name:     "unknown",
			event:    MouseClickEvent{Button: MouseButton(0xff)},
			expected: "unknown",
		},
		{
			name:     "left",
			event:    MouseClickEvent{Button: MouseLeft},
			expected: "left",
		},
		{
			name:     "right",
			event:    MouseClickEvent{Button: MouseRight},
			expected: "right",
		},
		{
			name:     "middle",
			event:    MouseClickEvent{Button: MouseMiddle},
			expected: "middle",
		},
		{
			name:     "release",
			event:    MouseReleaseEvent{Button: MouseNone},
			expected: "",
		},
		{
			name:     "wheelup",
			event:    MouseWheelEvent{Button: MouseWheelUp},
			expected: "wheelup",
		},
		{
			name:     "wheeldown",
			event:    MouseWheelEvent{Button: MouseWheelDown},
			expected: "wheeldown",
		},
		{
			name:     "wheelleft",
			event:    MouseWheelEvent{Button: MouseWheelLeft},
			expected: "wheelleft",
		},
		{
			name:     "wheelright",
			event:    MouseWheelEvent{Button: MouseWheelRight},
			expected: "wheelright",
		},
		{
			name:     "motion",
			event:    MouseMotionEvent{Button: MouseNone},
			expected: "motion",
		},
		{
			name:     "shift+left",
			event:    MouseReleaseEvent{Button: MouseLeft, Mod: ModShift},
			expected: "shift+left",
		},
		{
			name: "shift+left", event: MouseClickEvent{Button: MouseLeft, Mod: ModShift},
			expected: "shift+left",
		},
		{
			name:     "ctrl+shift+left",
			event:    MouseClickEvent{Button: MouseLeft, Mod: ModCtrl | ModShift},
			expected: "ctrl+shift+left",
		},
		{
			name:     "alt+left",
			event:    MouseClickEvent{Button: MouseLeft, Mod: ModAlt},
			expected: "alt+left",
		},
		{
			name:     "ctrl+left",
			event:    MouseClickEvent{Button: MouseLeft, Mod: ModCtrl},
			expected: "ctrl+left",
		},
		{
			name:     "ctrl+alt+left",
			event:    MouseClickEvent{Button: MouseLeft, Mod: ModAlt | ModCtrl},
			expected: "ctrl+alt+left",
		},
		{
			name:     "ctrl+alt+shift+left",
			event:    MouseClickEvent{Button: MouseLeft, Mod: ModAlt | ModCtrl | ModShift},
			expected: "ctrl+alt+shift+left",
		},
		{
			name:     "ignore coordinates",
			event:    MouseClickEvent{X: 100, Y: 200, Button: MouseLeft},
			expected: "left",
		},
		{
			name:     "broken type",
			event:    MouseClickEvent{Button: MouseButton(120)},
			expected: "unknown",
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := fmt.Sprint(tc.event)

			if tc.expected != actual {
				t.Fatalf("expected %q but got %q",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseX10MouseDownEvent(t *testing.T) {
	encode := func(b byte, x, y int) []byte {
		return []byte{
			'\x1b',
			'[',
			'M',
			byte(32) + b,
			byte(x + 32 + 1),
			byte(y + 32 + 1),
		}
	}

	tt := []struct {
		name     string
		buf      []byte
		expected Event
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0b0000_0000, 0, 0),
			expected: MouseClickEvent{X: 0, Y: 0, Button: MouseLeft},
		},
		{
			name:     "max position",
			buf:      encode(0b0000_0000, 222, 222), // Because 255 (max int8) - 32 - 1.
			expected: MouseClickEvent{X: 222, Y: 222, Button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0b0000_0000, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(0b0000_0001, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(0b0010_0001, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(0b0000_0010, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "right in motion",
			buf:      encode(0b0010_0010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(0b0010_0011, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(0b0100_0000, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(0b0100_0001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(0b0100_0010, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(0b0100_0011, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelRight},
		},
		{
			name:     "release",
			buf:      encode(0b0000_0011, 32, 16),
			expected: MouseReleaseEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "backward",
			buf:      encode(0b1000_0000, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(0b1000_0001, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseForward},
		},
		{
			name:     "button 10",
			buf:      encode(0b1000_0010, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseButton10},
		},
		{
			name:     "button 11",
			buf:      encode(0b1000_0011, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseButton11},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(0b0000_1010, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModAlt, Button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(0b0001_0010, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModCtrl, Button: MouseRight},
		},
		{
			name:     "left in motion",
			buf:      encode(0b0010_0000, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "alt+right in motion",
			buf:      encode(0b0010_1010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Mod: ModAlt, Button: MouseRight},
		},
		{
			name:     "ctrl+right in motion",
			buf:      encode(0b0011_0010, 32, 16),
			expected: MouseMotionEvent{X: 32, Y: 16, Mod: ModCtrl, Button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(0b0001_1010, 32, 16),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModAlt | ModCtrl, Button: MouseRight},
		},
		{
			name:     "ctrl+wheel up",
			buf:      encode(0b0101_0000, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModCtrl, Button: MouseWheelUp},
		},
		{
			name:     "alt+wheel down",
			buf:      encode(0b0100_1001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModAlt, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel down",
			buf:      encode(0b0101_1001, 32, 16),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModAlt | ModCtrl, Button: MouseWheelDown},
		},
		// Overflow position.
		{
			name:     "overflow position",
			buf:      encode(0b0010_0000, 250, 223), // Because 255 (max int8) - 32 - 1.
			expected: MouseMotionEvent{X: -6, Y: -33, Button: MouseLeft},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := parseX10MouseEvent(tc.buf)

			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}

func TestParseSGRMouseEvent(t *testing.T) {
	type csiSequence struct {
		params []ansi.Param
		cmd    ansi.Cmd
	}
	encode := func(b, x, y int, r bool) *csiSequence {
		re := 'M'
		if r {
			re = 'm'
		}
		return &csiSequence{
			params: []ansi.Param{
				ansi.Param(b),
				ansi.Param(x + 1),
				ansi.Param(y + 1),
			},
			cmd: ansi.Cmd(re) | ('<' << parser.PrefixShift),
		}
	}

	tt := []struct {
		name     string
		buf      *csiSequence
		expected Event
	}{
		// Position.
		{
			name:     "zero position",
			buf:      encode(0, 0, 0, false),
			expected: MouseClickEvent{X: 0, Y: 0, Button: MouseLeft},
		},
		{
			name:     "225 position",
			buf:      encode(0, 225, 225, false),
			expected: MouseClickEvent{X: 225, Y: 225, Button: MouseLeft},
		},
		// Simple.
		{
			name:     "left",
			buf:      encode(0, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left in motion",
			buf:      encode(32, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "left",
			buf:      encode(0, 32, 16, true),
			expected: MouseReleaseEvent{X: 32, Y: 16, Button: MouseLeft},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle in motion",
			buf:      encode(33, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "middle",
			buf:      encode(1, 32, 16, true),
			expected: MouseReleaseEvent{X: 32, Y: 16, Button: MouseMiddle},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "right",
			buf:      encode(2, 32, 16, true),
			expected: MouseReleaseEvent{X: 32, Y: 16, Button: MouseRight},
		},
		{
			name:     "motion",
			buf:      encode(35, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseNone},
		},
		{
			name:     "wheel up",
			buf:      encode(64, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelUp},
		},
		{
			name:     "wheel down",
			buf:      encode(65, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelDown},
		},
		{
			name:     "wheel left",
			buf:      encode(66, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelLeft},
		},
		{
			name:     "wheel right",
			buf:      encode(67, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Button: MouseWheelRight},
		},
		{
			name:     "backward",
			buf:      encode(128, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "backward in motion",
			buf:      encode(160, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseBackward},
		},
		{
			name:     "forward",
			buf:      encode(129, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Button: MouseForward},
		},
		{
			name:     "forward in motion",
			buf:      encode(161, 32, 16, false),
			expected: MouseMotionEvent{X: 32, Y: 16, Button: MouseForward},
		},
		// Combinations.
		{
			name:     "alt+right",
			buf:      encode(10, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModAlt, Button: MouseRight},
		},
		{
			name:     "ctrl+right",
			buf:      encode(18, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModCtrl, Button: MouseRight},
		},
		{
			name:     "ctrl+alt+right",
			buf:      encode(26, 32, 16, false),
			expected: MouseClickEvent{X: 32, Y: 16, Mod: ModAlt | ModCtrl, Button: MouseRight},
		},
		{
			name:     "alt+wheel",
			buf:      encode(73, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModAlt, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+wheel",
			buf:      encode(81, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModCtrl, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+wheel",
			buf:      encode(89, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModAlt | ModCtrl, Button: MouseWheelDown},
		},
		{
			name:     "ctrl+alt+shift+wheel",
			buf:      encode(93, 32, 16, false),
			expected: MouseWheelEvent{X: 32, Y: 16, Mod: ModAlt | ModShift | ModCtrl, Button: MouseWheelDown},
		},
	}

	for i := range tt {
		tc := tt[i]

		t.Run(tc.name, func(t *testing.T) {
			actual := parseSGRMouseEvent(tc.buf.cmd, tc.buf.params)
			if tc.expected != actual {
				t.Fatalf("expected %#v but got %#v",
					tc.expected,
					actual,
				)
			}
		})
	}
}

// TestMatchStrings tests the MatchStrings method
func TestMatchStrings(t *testing.T) {
	tests := []struct {
		name    string
		key     Key
		inputs  []string
		want    bool
	}{
		{
			name:   "matches first string",
			key:    Key{Code: 'a', Mod: ModCtrl},
			inputs: []string{"ctrl+a", "ctrl+b", "ctrl+c"},
			want:   true,
		},
		{
			name:   "matches middle string",
			key:    Key{Code: 'b', Mod: ModCtrl},
			inputs: []string{"ctrl+a", "ctrl+b", "ctrl+c"},
			want:   true,
		},
		{
			name:   "matches last string",
			key:    Key{Code: 'c', Mod: ModCtrl},
			inputs: []string{"ctrl+a", "ctrl+b", "ctrl+c"},
			want:   true,
		},
		{
			name:   "no match",
			key:    Key{Code: 'd', Mod: ModCtrl},
			inputs: []string{"ctrl+a", "ctrl+b", "ctrl+c"},
			want:   false,
		},
		{
			name:   "empty inputs",
			key:    Key{Code: 'a', Mod: ModCtrl},
			inputs: []string{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.key.MatchStrings(tt.inputs...)
			if got != tt.want {
				t.Errorf("MatchStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyMatchString(t *testing.T) {
	cases := []struct {
		name  string
		key   Key
		input string
		want  bool
	}{
		{
			name:  "ctrl+a",
			key:   Key{Code: 'a', Mod: ModCtrl},
			input: "ctrl+a",
			want:  true,
		},
		{
			name:  "ctrl+alt+a",
			key:   Key{Code: 'a', Mod: ModCtrl | ModAlt},
			input: "ctrl+alt+a",
			want:  true,
		},
		{
			name:  "ctrl+alt+shift+a",
			key:   Key{Code: 'a', Mod: ModCtrl | ModAlt | ModShift},
			input: "ctrl+alt+shift+a",
			want:  true,
		},
		{
			name:  "H",
			key:   Key{Code: 'H', Text: "H"},
			input: "H",
			want:  true,
		},
		{
			name:  "shift+h",
			key:   Key{Code: 'h', Mod: ModShift, Text: "H"},
			input: "H",
			want:  true,
		},
		{
			name:  "?",
			key:   Key{Code: '/', Mod: ModShift, Text: "?"},
			input: "?",
			want:  true,
		},
		{
			name:  "shift+/",
			key:   Key{Code: '/', Mod: ModShift, Text: "?"},
			input: "shift+/",
			want:  true,
		},
		{
			name:  "capslock+a",
			key:   Key{Code: 'a', Mod: ModCapsLock, Text: "A"},
			input: "A",
			want:  true,
		},
		{
			name:  "ctrl+capslock+a",
			key:   Key{Code: 'a', Mod: ModCtrl | ModCapsLock},
			input: "ctrl+a",
			want:  false,
		},
		{
			name:  "space",
			key:   Key{Code: KeySpace, Text: " "},
			input: "space",
			want:  true,
		},
		{
			name:  "whitespace",
			key:   Key{Code: KeySpace, Text: " "},
			input: " ",
			want:  true,
		},
		{
			name:  "ctrl+space",
			key:   Key{Code: KeySpace, Mod: ModCtrl},
			input: "ctrl+space",
			want:  true,
		},
		{
			name:  "shift+whitespace",
			key:   Key{Code: KeySpace, Mod: ModShift, Text: " "},
			input: " ",
			want:  true,
		},
		{
			name:  "shift+space",
			key:   Key{Code: KeySpace, Mod: ModShift, Text: " "},
			input: "shift+space",
			want:  true,
		},
		{
			name:  "meta modifier",
			key:   Key{Code: 'a', Mod: ModMeta},
			input: "meta+a",
			want:  true,
		},
		{
			name:  "hyper modifier",
			key:   Key{Code: 'a', Mod: ModHyper},
			input: "hyper+a",
			want:  true,
		},
		{
			name:  "super modifier",
			key:   Key{Code: 'a', Mod: ModSuper},
			input: "super+a",
			want:  true,
		},
		{
			name:  "scrolllock modifier",
			key:   Key{Code: 'a', Mod: ModScrollLock},
			input: "scrolllock+a",
			want:  true,
		},
		{
			name:  "numlock modifier",
			key:   Key{Code: 'a', Mod: ModNumLock},
			input: "numlock+a",
			want:  true,
		},
		{
			name:  "multi-rune key",
			key:   Key{Code: KeyExtended, Text: "hello"},
			input: "hello",
			want:  true,
		},
		{
			name:  "enter key",
			key:   Key{Code: KeyEnter},
			input: "enter",
			want:  true,
		},
		{
			name:  "tab key",
			key:   Key{Code: KeyTab},
			input: "tab",
			want:  true,
		},
		{
			name:  "escape key",
			key:   Key{Code: KeyEscape},
			input: "esc",
			want:  true,
		},
		{
			name:  "f1 key",
			key:   Key{Code: KeyF1},
			input: "f1",
			want:  true,
		},
		{
			name:  "backspace key",
			key:   Key{Code: KeyBackspace},
			input: "backspace",
			want:  true,
		},
		{
			name:  "delete key",
			key:   Key{Code: KeyDelete},
			input: "delete",
			want:  true,
		},
		{
			name:  "home key",
			key:   Key{Code: KeyHome},
			input: "home",
			want:  true,
		},
		{
			name:  "end key",
			key:   Key{Code: KeyEnd},
			input: "end",
			want:  true,
		},
		{
			name:  "pgup key",
			key:   Key{Code: KeyPgUp},
			input: "pgup",
			want:  true,
		},
		{
			name:  "pgdown key",
			key:   Key{Code: KeyPgDown},
			input: "pgdown",
			want:  true,
		},
		{
			name:  "up arrow",
			key:   Key{Code: KeyUp},
			input: "up",
			want:  true,
		},
		{
			name:  "down arrow",
			key:   Key{Code: KeyDown},
			input: "down",
			want:  true,
		},
		{
			name:  "left arrow",
			key:   Key{Code: KeyLeft},
			input: "left",
			want:  true,
		},
		{
			name:  "right arrow",
			key:   Key{Code: KeyRight},
			input: "right",
			want:  true,
		},
		{
			name:  "insert key",
			key:   Key{Code: KeyInsert},
			input: "insert",
			want:  true,
		},
		{
			name:  "single printable character",
			key:   Key{Code: '1', Text: "1"},
			input: "1",
			want:  true,
		},
		{
			name:  "uppercase letter without shift",
			key:   Key{Code: 'A', Text: "A"},
			input: "A",
			want:  true,
		},
		{
			name:  "no match different key",
			key:   Key{Code: 'a', Mod: ModCtrl},
			input: "ctrl+b",
			want:  false,
		},
		{
			name:  "no match different modifier",
			key:   Key{Code: 'a', Mod: ModCtrl},
			input: "alt+a",
			want:  false,
		},
		{
			name:  "unknown key name",
			key:   Key{Code: 'x'},
			input: "unknownkey",
			want:  false,
		},
		{
			name:  "multi-rune string that doesn't match",
			key:   Key{Code: 'a'},
			input: "hello",
			want:  false,
		},
		{
			name:  "printable character with ctrl modifier",
			key:   Key{Code: 'a', Mod: ModCtrl},
			input: "a",
			want:  false,
		},
		{
			name:  "lowercase letter with shift",
			key:   Key{Code: 'h', Mod: ModShift},
			input: "shift+h",
			want:  true,
		},
		{
			name:  "uppercase letter with capslock",
			key:   Key{Code: 'h', Mod: ModCapsLock},
			input: "capslock+h",
			want:  true,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			got := tc.key.MatchString(tc.input)
			if got != tc.want {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

// TestSplitSequences tests that string-terminated sequences work correctly
// when split across multiple read() calls.
func TestSplitSequences(t *testing.T) {
	tests := []struct {
		name   string
		chunks [][]byte
		want   []Event
		delay  time.Duration
		limit  int // limit the number of bytes read at once
	}{
		{
			name: "OSC 11 background color with ST terminator",
			chunks: [][]byte{
				[]byte("\x1b]11;rgb:1a1a/1b1b/2c2c"),
				[]byte("\x1b\\"),
			},
			want: []Event{
				BackgroundColorEvent{Color: ansi.XParseColor("rgb:1a1a/1b1b/2c2c")},
			},
		},
		{
			name: "OSC 11 background color with BEL terminator",
			chunks: [][]byte{
				[]byte("\x1b]11;rgb:1a1a/1b1b/2c2c"),
				[]byte("\x07"),
			},
			want: []Event{
				BackgroundColorEvent{Color: ansi.XParseColor("rgb:1a1a/1b1b/2c2c")},
			},
		},
		{
			name: "OSC 10 foreground color split",
			chunks: [][]byte{
				[]byte("\x1b]10;rgb:ffff/0000/"),
				[]byte("0000\x1b\\"),
			},
			want: []Event{
				ForegroundColorEvent{Color: ansi.XParseColor("rgb:ffff/0000/0000")},
			},
		},
		{
			name: "OSC 12 cursor color split",
			chunks: [][]byte{
				[]byte("\x1b]12;rgb:"),
				[]byte("8080/8080/8080\x07"),
			},
			want: []Event{
				CursorColorEvent{Color: ansi.XParseColor("rgb:8080/8080/8080")},
			},
		},
		{
			name: "DCS sequence split",
			chunks: [][]byte{
				[]byte("\x1bP1$r"),
				[]byte("test\x1b\\"),
			},
			want: []Event{
				UnknownDcsEvent("\x1bP1$rtest\x1b\\"),
			},
		},
		{
			name: "long DCS sequence split",
			chunks: [][]byte{
				[]byte("\x1bP1$raaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdef"),
				[]byte("test\x1b\\"),
			},
			want: []Event{
				UnknownDcsEvent("\x1bP1$raaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabcdeftest\x1b\\"),
			},
			limit: 256,
		},
		{
			name: "APC sequence split",
			chunks: [][]byte{
				[]byte("\x1b_T"),
				[]byte("test\x1b\\"),
			},
			want: []Event{
				UnknownApcEvent("\x1b_Ttest\x1b\\"),
			},
		},
		{
			name: "Multiple chunks OSC",
			chunks: [][]byte{
				[]byte("\x1b]11;"),
				[]byte("rgb:1234/"),
				[]byte("5678/9abc\x07"),
			},
			want: []Event{
				BackgroundColorEvent{Color: ansi.XParseColor("rgb:1234/5678/9abc")},
			},
		},
		{
			name: "OSC followed by regular key",
			chunks: [][]byte{
				[]byte("\x1b]11;rgb:1111/2222/3333"),
				[]byte("\x07a"),
			},
			want: []Event{
				BackgroundColorEvent{Color: ansi.XParseColor("rgb:1111/2222/3333")},
				KeyPressEvent{Code: 'a', Text: "a"},
			},
		},
		{
			name: "unknown sequence after timeout",
			chunks: [][]byte{
				[]byte("\x1b]11;rgb:1111/2222/3333"),
				[]byte("abc"),
				[]byte("x"),
				[]byte("x"),
				[]byte("x"),
				[]byte("x"),
			},
			want: []Event{
				UnknownEvent("\x1b]11;rgb:1111/2222/3333"),
				KeyPressEvent{Code: 'a', Text: "a"},
				KeyPressEvent{Code: 'b', Text: "b"},
				KeyPressEvent{Code: 'c', Text: "c"},
				KeyPressEvent{Code: 'x', Text: "x"},
				KeyPressEvent{Code: 'x', Text: "x"},
				KeyPressEvent{Code: 'x', Text: "x"},
				KeyPressEvent{Code: 'x', Text: "x"},
			},
			delay: 60 * time.Millisecond, // Ensure the timeout is triggered.
		},
		{
			name: "multiple broken down sequences",
			chunks: [][]byte{
				[]byte("\x1b[B"),
				[]byte("\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B"),
				[]byte("\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b["),
				[]byte("B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b"),
				[]byte("[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B\x1b[B"),
			},
			want: []Event{
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
				KeyPressEvent{Code: KeyDown},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rds := make([]io.Reader, len(tt.chunks))
			for i, chunk := range tt.chunks {
				rds[i] = bytes.NewReader(chunk)
			}
			var r io.Reader
			limit := 32
			if tt.limit > 0 {
				limit = tt.limit
			}
			if tt.delay > 0 {
				r = DelayedLimitedReader(io.MultiReader(rds...), limit, tt.delay)
			} else {
				r = LimitedReader(io.MultiReader(rds...), limit)
			}
			ir := NewTerminalReader(r, "xterm-256color")
			ir.SetLogger(TLogger{TB: t})

			eventc := make(chan Event)
			go func(t testing.TB) {
				defer close(eventc)
				if err := ir.StreamEvents(t.Context(), eventc); err != nil {
					t.Errorf("error streaming events: %v", err)
				}
			}(t)

			var events []Event
			for ev := range eventc {
				events = append(events, ev)
			}

			if len(events) != len(tt.want) {
				t.Fatalf("got %d events, want %d: %#v", len(events), len(tt.want), events)
			}

			for i, want := range tt.want {
				if !reflect.DeepEqual(events[i], want) {
					t.Errorf("event %d: got %#v, want %#v", i, events[i], want)
				}
			}
		})
	}
}

// limitedSleepReader is a simple io.Reader that limits the number of bytes
// read on each Read call and simulates a delay to mimic terminal input
// behavior.
type limitedSleepReader struct {
	r io.Reader
	n int
	d time.Duration
	i int
}

func DelayedLimitedReader(r io.Reader, n int, d time.Duration) io.Reader {
	return &limitedSleepReader{r: r, n: n, d: d}
}

func LimitedReader(r io.Reader, n int) io.Reader {
	return &limitedSleepReader{r: r, n: n}
}

func (r *limitedSleepReader) Read(p []byte) (n int, err error) {
	if r.i > 0 && r.d > 0 {
		time.Sleep(r.d)
	}
	if r.n <= 0 {
		return 0, io.EOF
	}
	if len(p) > r.n {
		p = p[0:r.n]
	}
	n, err = r.r.Read(p)
	r.i++
	return n, err
}

type TLogger struct{ testing.TB }

func (t TLogger) Printf(format string, args ...interface{}) {
	t.Helper()
	if t.TB == nil {
		return
	}
	t.Logf(format, args...)
}

// TestKeystroke tests the Keystroke method
func TestKeystroke(t *testing.T) {
	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "simple key",
			key:  Key{Code: 'a'},
			want: "a",
		},
		{
			name: "ctrl+a",
			key:  Key{Code: 'a', Mod: ModCtrl},
			want: "ctrl+a",
		},
		{
			name: "alt+a",
			key:  Key{Code: 'a', Mod: ModAlt},
			want: "alt+a",
		},
		{
			name: "shift+a",
			key:  Key{Code: 'a', Mod: ModShift},
			want: "shift+a",
		},
		{
			name: "meta+a",
			key:  Key{Code: 'a', Mod: ModMeta},
			want: "meta+a",
		},
		{
			name: "hyper+a",
			key:  Key{Code: 'a', Mod: ModHyper},
			want: "hyper+a",
		},
		{
			name: "super+a",
			key:  Key{Code: 'a', Mod: ModSuper},
			want: "super+a",
		},
		{
			name: "ctrl+alt+shift+a",
			key:  Key{Code: 'a', Mod: ModCtrl | ModAlt | ModShift},
			want: "ctrl+alt+shift+a",
		},
		{
			name: "all modifiers",
			key:  Key{Code: 'a', Mod: ModCtrl | ModAlt | ModShift | ModMeta | ModHyper | ModSuper},
			want: "ctrl+alt+shift+meta+hyper+super+a",
		},
		{
			name: "space key",
			key:  Key{Code: KeySpace},
			want: "space",
		},
		{
			name: "extended key with text",
			key:  Key{Code: KeyExtended, Text: "hello"},
			want: "hello",
		},
		{
			name: "enter key",
			key:  Key{Code: KeyEnter},
			want: "enter",
		},
		{
			name: "tab key",
			key:  Key{Code: KeyTab},
			want: "tab",
		},
		{
			name: "escape key",
			key:  Key{Code: KeyEscape},
			want: "esc",
		},
		{
			name: "f1 key",
			key:  Key{Code: KeyF1},
			want: "f1",
		},
		{
			name: "backspace key",
			key:  Key{Code: KeyBackspace},
			want: "backspace",
		},
		{
			name: "left ctrl key alone",
			key:  Key{Code: KeyLeftCtrl, Mod: ModCtrl},
			want: "leftctrl",
		},
		{
			name: "right ctrl key alone",
			key:  Key{Code: KeyRightCtrl, Mod: ModCtrl},
			want: "rightctrl",
		},
		{
			name: "left alt key alone",
			key:  Key{Code: KeyLeftAlt, Mod: ModAlt},
			want: "leftalt",
		},
		{
			name: "right alt key alone",
			key:  Key{Code: KeyRightAlt, Mod: ModAlt},
			want: "rightalt",
		},
		{
			name: "left shift key alone",
			key:  Key{Code: KeyLeftShift, Mod: ModShift},
			want: "leftshift",
		},
		{
			name: "right shift key alone",
			key:  Key{Code: KeyRightShift, Mod: ModShift},
			want: "rightshift",
		},
		{
			name: "left meta key alone",
			key:  Key{Code: KeyLeftMeta, Mod: ModMeta},
			want: "leftmeta",
		},
		{
			name: "right meta key alone",
			key:  Key{Code: KeyRightMeta, Mod: ModMeta},
			want: "rightmeta",
		},
		{
			name: "left hyper key alone",
			key:  Key{Code: KeyLeftHyper, Mod: ModHyper},
			want: "lefthyper",
		},
		{
			name: "right hyper key alone",
			key:  Key{Code: KeyRightHyper, Mod: ModHyper},
			want: "righthyper",
		},
		{
			name: "left super key alone",
			key:  Key{Code: KeyLeftSuper, Mod: ModSuper},
			want: "leftsuper",
		},
		{
			name: "right super key alone",
			key:  Key{Code: KeyRightSuper, Mod: ModSuper},
			want: "rightsuper",
		},
		{
			name: "key with base code",
			key:  Key{Code: 'A', BaseCode: 'a'},
			want: "a",
		},
		{
			name: "unknown key with base code",
			key:  Key{Code: 99999, BaseCode: 'x'},
			want: "x",
		},
		{
			name: "printable rune",
			key:  Key{Code: '€'},
			want: "€",
		},
		{
			name: "unknown key without base code",
			key:  Key{Code: 99999},
			want: "𘚟",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.key.Keystroke()
			if got != tt.want {
				t.Errorf("Keystroke() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestKeystrokeCoverage tests edge cases for Keystroke method
func TestKeystrokeCoverage(t *testing.T) {
	// Test a key that's not in keyTypeString and is KeySpace
	// This is actually impossible since KeySpace is in keyTypeString
	// But we can test a key that's not in keyTypeString and has BaseCode of KeySpace
	k := Key{Code: 999999, BaseCode: KeySpace}
	if got := k.Keystroke(); got != "space" {
		t.Errorf("Keystroke() = %q, want %q", got, "space")
	}
	
	// Test a key that's not in keyTypeString, has no BaseCode, and is KeySpace
	// This would require KeySpace to not be in keyTypeString, which it is
	// So this branch might be unreachable
}
func TestKeyStringMore(t *testing.T) {
	tests := []struct {
		name string
		key  Key
		want string
	}{
		{
			name: "space character",
			key:  Key{Code: KeySpace, Text: " "},
			want: "space",
		},
		{
			name: "empty text",
			key:  Key{Code: 'a', Text: ""},
			want: "a",
		},
		{
			name: "text with multiple characters",
			key:  Key{Code: KeyExtended, Text: "hello"},
			want: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.key.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
