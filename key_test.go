package uv

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"runtime"
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
	var p SequenceParser
	_, e := p.parseSequence([]byte("\x1b[I"))
	switch e.(type) {
	case FocusEvent:
		// ok
	default:
		t.Error("invalid sequence")
	}
}

func TestBlur(t *testing.T) {
	var p SequenceParser
	_, e := p.parseSequence([]byte("\x1b[O"))
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

		// Xterm report window text area size.
		seqTest{
			[]byte("\x1b[4;24;80t"),
			[]Event{
				WindowSizeEvent{Width: 80, Height: 24},
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

	if runtime.GOOS != isWindows {
		// Sadly, utf8.DecodeRune([]byte(0xfe)) returns a valid rune on windows.
		// This is incorrect, but it makes our test fail if we try it out.
		td = append(td, seqTest{
			[]byte{'\xfe'},
			[]Event{
				UnknownEvent(rune(0xfe)),
			},
		})
	}

	var p SequenceParser
	for _, tc := range td {
		t.Run(fmt.Sprintf("%q", string(tc.seq)), func(t *testing.T) {
			var events []Event
			buf := tc.seq
			for len(buf) > 0 {
				width, Event := p.parseSequence(buf)
				switch Event := Event.(type) {
				case MultiEvent:
					events = append(events, Event...)
				default:
					events = append(events, Event)
				}
				buf = buf[width:]
			}
			if !reflect.DeepEqual(tc.Events, events) {
				t.Errorf("\nexpected event for %q:\n    %#v\ngot:\n    %#v", tc.seq, tc.Events, events)
			}
		})
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
	}

	drv := NewTerminalReader(NewStringSliceReader(t, inputs), "dumb")
	drv.SetLogger(TLogger{t})
	if err := drv.Start(); err != nil {
		t.Fatalf("unexpected error starting terminal reader: %v", err)
	}

	var err error
	var evs []Event
	events := make(chan Event)
	go func() {
		err = drv.ReceiveEvents(context.Background(), events)
		close(events)
	}()

	for ev := range events {
		evs = append(evs, ev)
	}

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error receiving events: %v", err)
	}

	if !reflect.DeepEqual(expect, evs) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, evs)
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
	if err := drv.Start(); err != nil {
		t.Fatalf("unexpected error starting terminal reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var err error
	var evs []Event
	events := make(chan Event)
	go func() {
		err = drv.ReceiveEvents(ctx, events)
		close(events)
	}()

	for ev := range events {
		evs = append(evs, ev)
	}

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error receiving events: %v", err)
	}

	if !reflect.DeepEqual(expect, evs) {
		t.Errorf("unexpected messages, expected:\n    %+v\ngot:\n    %+v", expect, evs)
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
			"ignored osc esc",
			[]byte("\x1b]11;#123456\x1b"),
			[]Event(nil),
		},
		{
			"ignored osc can",
			[]byte("\x1b]11;#123456\x18"),
			[]Event(nil),
		},
		{
			"ignored osc sub",
			[]byte("\x1b]11;#123456\x1a"),
			[]Event(nil),
		},
		{
			"ignored apc esc",
			[]byte("\x1b_hello\x1b\x1b_abc\x1b\\\x1ba"),
			[]Event{
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
	// We'll check that the input reader finishes at the end
	// without error.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	dr := NewTerminalReader(input, "dumb")
	dr.SetLogger(TLogger{t})
	if err := dr.Start(); err != nil {
		t.Fatalf("unexpected error starting terminal reader: %v", err)
	}

	var err error
	var events []Event
	eventsc := make(chan Event)

	// Start the reader in the background.
	go func() {
		err = dr.ReceiveEvents(ctx, eventsc)
		close(eventsc)
	}()

	for ev := range eventsc {
		if ev != nil {
			events = append(events, ev)
		}
	}

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("unexpected error receiving events: %v", err)
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
	var p SequenceParser
	for seq := range sequences {
		f.Add(seq)
	}
	f.Add("\x1b]52;?\x07")                      // OSC 52
	f.Add("\x1b]11;rgb:0000/0000/0000\x1b\\")   // OSC 11
	f.Add("\x1bP>|charm terminal(0.1.2)\x1b\\") // DCS (XTVERSION)
	f.Add("\x1b_Gi=123\x1b\\")                  // APC
	f.Fuzz(func(t *testing.T, seq string) {
		n, _ := p.parseSequence([]byte(seq))
		if n == 0 && seq != "" {
			t.Errorf("expected a non-zero width for %q", seq)
		}
	})
}

// BenchmarkDetectSequenceMap benchmarks the map-based sequence
// detector.
func BenchmarkDetectSequenceMap(b *testing.B) {
	var p SequenceParser
	td := genRandomDataWithSeed(123, 10000)
	for i := 0; i < b.N; i++ {
		for j, w := 0, 0; j < len(td.data); j += w {
			w, _ = p.parseSequence(td.data[j:])
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

type stringSliceReader struct {
	t testing.TB
	s []string
	i int
}

func NewStringSliceReader(t testing.TB, s []string) io.Reader {
	return &stringSliceReader{t: t, s: s}
}

func (r *stringSliceReader) Read(p []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	// Simulate a read from terminal input.
	n = copy(p, r.s[r.i])
	r.i++
	if n < len(r.s[r.i-1]) {
		return n, nil
	}
	// time.Sleep(time.Duration(rand.Intn(99)) * time.Millisecond)
	return n, nil
}

type TLogger struct{ testing.TB }

func (t TLogger) Printf(format string, args ...interface{}) {
	t.Helper()
	if t.TB == nil {
		return
	}
	t.Logf(format, args...)
}
