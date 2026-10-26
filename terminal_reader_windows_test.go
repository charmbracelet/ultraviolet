//go:build windows
// +build windows

package uv

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode"

	xwindows "github.com/charmbracelet/x/windows"
)

// win32Key packs a [xwindows.KeyEventRecord] into the raw byte layout of an
// [xwindows.InputRecord], the way the Windows Console API hands key events to
// the reader. Tests build records this way so that the code under test reads
// them back through the very same accessor it uses in production.
func win32Key(k xwindows.KeyEventRecord) xwindows.InputRecord {
	var record xwindows.InputRecord
	record.EventType = xwindows.KEY_EVENT
	var keyDown uint32
	if k.KeyDown {
		keyDown = 1
	}
	binary.LittleEndian.PutUint32(record.Event[0:4], keyDown)
	binary.LittleEndian.PutUint16(record.Event[4:6], k.RepeatCount)
	binary.LittleEndian.PutUint16(record.Event[6:8], k.VirtualKeyCode)
	binary.LittleEndian.PutUint16(record.Event[8:10], k.VirtualScanCode)
	// The console reports characters as UTF-16 code units, so they always fit
	// in 16 bits here.
	binary.LittleEndian.PutUint16(record.Event[10:12], uint16(k.Char))
	binary.LittleEndian.PutUint32(record.Event[12:16], k.ControlKeyState)
	return record
}

// modifierRecord builds a standalone modifier key event, i.e. a key event
// without a character attached, the way the console reports Shift, Control,
// Alt and Windows key presses and releases.
func modifierRecord(vkc, sc uint16, cks uint32, down bool) xwindows.InputRecord {
	return win32Key(xwindows.KeyEventRecord{
		KeyDown:         down,
		RepeatCount:     1,
		VirtualKeyCode:  vkc,
		VirtualScanCode: sc,
		ControlKeyState: cks,
	})
}

// vtCharRecord builds the key press record the console produces in VT input
// mode for a character that has no virtual-key code of its own, such as the
// bytes of an escape sequence.
func vtCharRecord(r rune) xwindows.InputRecord {
	return win32Key(xwindows.KeyEventRecord{
		KeyDown:     true,
		RepeatCount: 1,
		Char:        r,
	})
}

// serializeRecords runs the records through
// [TerminalReader.serializeWin32InputRecords] and returns the bytes the reader
// would hand over to the event decoder.
func serializeRecords(t *testing.T, vtInput bool, records ...xwindows.InputRecord) string {
	t.Helper()
	d := NewTerminalReader(nil, "dumb")
	d.vtInput = vtInput
	var buf bytes.Buffer
	d.serializeWin32InputRecords(records, &buf)
	return buf.String()
}

// ConPTY can insert standalone Shift records when converting pasted uppercase
// text to console input. Those records must not insert NULs into the paste.
func TestSerializeWin32InputVTPasteWithShift(t *testing.T) {
	const payload = "TWX-040G0-044"

	shift := modifierRecord(xwindows.VK_SHIFT, 0x2a, xwindows.SHIFT_PRESSED, true)

	var records []xwindows.InputRecord
	for _, r := range "\x1b[200~" {
		records = append(records, vtCharRecord(r))
	}
	for _, r := range payload {
		if unicode.IsUpper(r) {
			// The modifier and character arrive as separate records.
			records = append(records, shift, win32Key(xwindows.KeyEventRecord{
				KeyDown:         true,
				RepeatCount:     1,
				VirtualKeyCode:  uint16(r),
				Char:            r,
				ControlKeyState: xwindows.SHIFT_PRESSED,
			}))
			continue
		}
		records = append(records, vtCharRecord(r))
	}
	records = append(records, modifierRecord(xwindows.VK_SHIFT, 0x2a, 0, false))
	for _, r := range "\x1b[201~" {
		records = append(records, vtCharRecord(r))
	}

	want := "\x1b[200~" + payload + "\x1b[201~"
	if got := serializeRecords(t, true, records...); got != want {
		t.Errorf("serialized paste = %q, want %q", got, want)
	}
}

func TestSerializeWin32InputVTModifiersProduceNoInput(t *testing.T) {
	modifiers := []uint16{
		xwindows.VK_SHIFT, xwindows.VK_LSHIFT, xwindows.VK_RSHIFT,
		xwindows.VK_CONTROL, xwindows.VK_LCONTROL, xwindows.VK_RCONTROL,
		xwindows.VK_MENU, xwindows.VK_LMENU, xwindows.VK_RMENU,
		xwindows.VK_LWIN, xwindows.VK_RWIN,
	}
	records := []xwindows.InputRecord{vtCharRecord('a')}
	for _, vkc := range modifiers {
		records = append(records, modifierRecord(vkc, 0, 0, true), modifierRecord(vkc, 0, 0, false))
	}
	records = append(records, vtCharRecord('b'))
	if got := serializeRecords(t, true, records...); got != "ab" {
		t.Errorf("text around modifiers = %q, want %q", got, "ab")
	}
}

// A nonzero character remains text even when its virtual-key code names a modifier.
func TestSerializeWin32InputVTModifierWithCharacterKeepsText(t *testing.T) {
	const altGr = xwindows.LEFT_CTRL_PRESSED | xwindows.RIGHT_ALT_PRESSED

	record := win32Key(xwindows.KeyEventRecord{
		KeyDown:         true,
		RepeatCount:     1,
		VirtualKeyCode:  xwindows.VK_MENU,
		VirtualScanCode: 0x38,
		Char:            'e',
		ControlKeyState: altGr,
	})
	if got, want := serializeRecords(t, true, record), "e"; got != want {
		t.Errorf("serialized = %q, want %q", got, want)
	}
}

// TestSerializeWin32InputVTKeepsGenuineNul keeps the fix a modifier filter
// instead of a NUL filter. Ctrl+Space is reported as a zero character on a
// real key, and VT input mode reports raw input bytes with a zero virtual-key
// code, both of which are legitimate NUL input.
func TestSerializeWin32InputVTKeepsGenuineNul(t *testing.T) {
	ctrlSpace := win32Key(xwindows.KeyEventRecord{
		KeyDown:         true,
		RepeatCount:     1,
		VirtualKeyCode:  xwindows.VK_SPACE,
		VirtualScanCode: 0x39,
		ControlKeyState: xwindows.LEFT_CTRL_PRESSED,
	})
	rawNul := win32Key(xwindows.KeyEventRecord{
		KeyDown:     true,
		RepeatCount: 1,
	})
	ctrlDown := modifierRecord(xwindows.VK_CONTROL, 0x1d, xwindows.LEFT_CTRL_PRESSED, true)
	ctrlUp := modifierRecord(xwindows.VK_CONTROL, 0x1d, 0, false)

	if got := serializeRecords(t, true, ctrlDown, ctrlSpace, rawNul, ctrlUp); got != "\x00\x00" {
		t.Errorf("NUL input = %q, want %q", got, "\x00\x00")
	}
}

// A standalone modifier must not corrupt a UTF-16 surrogate pair.
func TestSerializeWin32InputVTSurrogatePairSurvivesModifier(t *testing.T) {
	high := vtCharRecord(0xd83d)
	shift := modifierRecord(xwindows.VK_SHIFT, 0x2a, xwindows.SHIFT_PRESSED, true)
	low := vtCharRecord(0xde00)

	if got, want := serializeRecords(t, true, high, shift, low), "\U0001f600"; got != want {
		t.Errorf("serialized = %q, want %q", got, want)
	}
}

// TestSerializeWin32InputNonVTEncodesModifiers keeps the fix scoped to VT
// input mode. With Win32 input mode negotiated, modifier key records are the
// only way the decoder learns about Shift, Control and Alt, so they must still
// be encoded verbatim.
func TestSerializeWin32InputNonVTEncodesModifiers(t *testing.T) {
	press := modifierRecord(xwindows.VK_SHIFT, 0x2a, xwindows.SHIFT_PRESSED, true)
	release := modifierRecord(xwindows.VK_SHIFT, 0x2a, 0, false)

	want := "\x1b[16;42;0;1;16;1_\x1b[16;42;0;0;0;1_"
	if got := serializeRecords(t, false, press, release); got != want {
		t.Errorf("serialized = %q, want %q", got, want)
	}
}
