package tv

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

// Event represents an input event that can be received from an input source.
type Event interface{}

// UnknownEvent represents an unknown event.
type UnknownEvent string

// String returns a string representation of the unknown event.
func (e UnknownEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// MultiEvent represents multiple messages event.
type MultiEvent []Event

// String returns a string representation of the multiple messages event.
func (e MultiEvent) String() string {
	var sb strings.Builder
	for _, ev := range e {
		sb.WriteString(fmt.Sprintf("%v\n", ev))
	}
	return sb.String()
}

// WindowSizeEvent represents the window size in cells.
type WindowSizeEvent Size

// Bounds returns the bounds corresponding to the size.
func (s WindowSizeEvent) Bounds() Rectangle {
	return Size(s).Bounds()
}

// WindowPixelSizeEvent represents the window size in pixels.
type WindowPixelSizeEvent Size

// Bounds returns the bounds corresponding to the size.
func (s WindowPixelSizeEvent) Bounds() Rectangle {
	return Size(s).Bounds()
}

// KeyPressEvent represents a key press event.
type KeyPressEvent Key

// MatchString returns true if the [Key] matches the given string. The string
// can be a key name like "enter", "tab", "a", or a printable character like
// "1" or " ". It can also have combinations of modifiers like "ctrl+a",
// "shift+enter", "alt+tab", "ctrl+shift+enter", etc.
func (k KeyPressEvent) MatchString(s string) bool {
	return Key(k).MatchString(s)
}

// MatchStrings returns true if the [Key] matches any of the given strings. The
// strings can be key names like "enter", "tab", "a", or a printable character
// like "1" or " ". It can also have combinations of modifiers like "ctrl+a",
// "shift+enter", "alt+tab", "ctrl+shift+enter", etc.
// See [Key.MatchString] for more details.
func (k KeyPressEvent) MatchStrings(ss ...string) bool {
	return Key(k).MatchStrings(ss...)
}

// String implements [fmt.Stringer] and is quite useful for matching key
// events. For details, on what this returns see [Key.String].
func (k KeyPressEvent) String() string {
	return Key(k).String()
}

// Key returns the underlying key event. This is a syntactic sugar for casting
// the key event to a [Key].
func (k KeyPressEvent) Key() Key {
	return Key(k)
}

// KeyReleaseEvent represents a key release event.
type KeyReleaseEvent Key

// MatchString returns true if the [Key] matches the given string. The string
// can be a key name like "enter", "tab", "a", or a printable character like
// "1" or " ". It can also have combinations of modifiers like "ctrl+a",
// "shift+enter", "alt+tab", "ctrl+shift+enter", etc.
func (k KeyReleaseEvent) MatchString(s string) bool {
	return Key(k).MatchString(s)
}

// MatchStrings returns true if the [Key] matches any of the given strings. The
// strings can be key names like "enter", "tab", "a", or a printable character
// like "1" or " ". It can also have combinations of modifiers like "ctrl+a",
// "shift+enter", "alt+tab", "ctrl+shift+enter", etc.
// See [Key.MatchString] for more details.
func (k KeyReleaseEvent) MatchStrings(ss ...string) bool {
	return Key(k).MatchStrings(ss...)
}

// String implements [fmt.Stringer] and is quite useful for matching key
// events. For details, on what this returns see [Key.String].
func (k KeyReleaseEvent) String() string {
	return Key(k).String()
}

// Key returns the underlying key event. This is a convenience method and
// syntactic sugar to satisfy the [KeyEvent] interface, and cast the key event to
// [Key].
func (k KeyReleaseEvent) Key() Key {
	return Key(k)
}

// KeyEvent represents a key event. This can be either a key press or a key
// release event.
type KeyEvent interface {
	fmt.Stringer

	// Text returns the text representation of the key event. This is useful
	// for matching key events along with [Key.String].
	// TODO: Use this instead of storing Text in the [Key] struct.
	// Text() string

	// Key returns the underlying key event.
	Key() Key
}

// MouseEvent represents a mouse message. This is a generic mouse message that
// can represent any kind of mouse event.
type MouseEvent interface {
	fmt.Stringer

	// Mouse returns the underlying mouse event.
	Mouse() Mouse
}

// MouseClickEvent represents a mouse button click event.
type MouseClickEvent Mouse

// String returns a string representation of the mouse click event.
func (e MouseClickEvent) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseEvent] interface, and cast the mouse
// event to [Mouse].
func (e MouseClickEvent) Mouse() Mouse {
	return Mouse(e)
}

// MouseReleaseEvent represents a mouse button release event.
type MouseReleaseEvent Mouse

// String returns a string representation of the mouse release event.
func (e MouseReleaseEvent) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseEvent] interface, and cast the mouse
// event to [Mouse].
func (e MouseReleaseEvent) Mouse() Mouse {
	return Mouse(e)
}

// MouseWheelEvent represents a mouse wheel message event.
type MouseWheelEvent Mouse

// String returns a string representation of the mouse wheel event.
func (e MouseWheelEvent) String() string {
	return Mouse(e).String()
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseEvent] interface, and cast the mouse
// event to [Mouse].
func (e MouseWheelEvent) Mouse() Mouse {
	return Mouse(e)
}

// MouseMotionEvent represents a mouse motion event.
type MouseMotionEvent Mouse

// String returns a string representation of the mouse motion event.
func (e MouseMotionEvent) String() string {
	m := Mouse(e)
	if m.Button != 0 {
		return m.String() + "+motion"
	}
	return m.String() + "motion"
}

// Mouse returns the underlying mouse event. This is a convenience method and
// syntactic sugar to satisfy the [MouseEvent] interface, and cast the mouse
// event to [Mouse].
func (e MouseMotionEvent) Mouse() Mouse {
	return Mouse(e)
}

// CursorPositionEvent represents a cursor position event. Where X is the
// zero-based column and Y is the zero-based row.
type CursorPositionEvent image.Point

// FocusEvent represents a terminal focus event.
// This occurs when the terminal gains focus.
type FocusEvent struct{}

// BlurEvent represents a terminal blur event.
// This occurs when the terminal loses focus.
type BlurEvent struct{}

// PasteEvent is an message that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteEvent string

// PasteStartEvent is an message that is emitted when the terminal starts the
// bracketed-paste text.
type PasteStartEvent struct{}

// PasteEndEvent is an message that is emitted when the terminal ends the
// bracketed-paste text.
type PasteEndEvent struct{}

// TerminalVersionEvent is a message that represents the terminal version.
type TerminalVersionEvent string

// ModifyOtherKeysEvent represents a modifyOtherKeys event.
//
//	0: disable
//	1: enable mode 1
//	2: enable mode 2
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
type ModifyOtherKeysEvent uint8

// KittyGraphicsEvent represents a Kitty Graphics response event.
//
// See https://sw.kovidgoyal.net/kitty/graphics-protocol/
type KittyGraphicsEvent struct {
	Options kitty.Options
	Payload []byte
}

// KittyEnhancementsEvent represents a Kitty enhancements event.
type KittyEnhancementsEvent int

// Contains reports whether m contains the given enhancements.
func (e KittyEnhancementsEvent) Contains(enhancements int) bool {
	return int(e)&enhancements == enhancements
}

// PrimaryDeviceAttributesEvent is an event that represents the terminal
// primary device attributes.
type PrimaryDeviceAttributesEvent []int

// ModeReportEvent is a message that represents a mode report event (DECRPM).
//
// See: https://vt100.net/docs/vt510-rm/DECRPM.html
type ModeReportEvent struct {
	// Mode is the mode number.
	Mode ansi.Mode

	// Value is the mode value.
	Value ansi.ModeSetting
}

// ForegroundColorEvent represents a foreground color event. This event is
// emitted when the terminal requests the terminal foreground color using
// [ansi.RequestForegroundColor].
type ForegroundColorEvent struct{ color.Color }

// String returns the hex representation of the color.
func (e ForegroundColorEvent) String() string {
	return colorToHex(e.Color)
}

// IsDark returns whether the color is dark.
func (e ForegroundColorEvent) IsDark() bool {
	return isDarkColor(e.Color)
}

// BackgroundColorEvent represents a background color event. This event is
// emitted when the terminal requests the terminal background color using
// [ansi.RequestBackgroundColor].
type BackgroundColorEvent struct{ color.Color }

// String returns the hex representation of the color.
func (e BackgroundColorEvent) String() string {
	return colorToHex(e)
}

// IsDark returns whether the color is dark.
func (e BackgroundColorEvent) IsDark() bool {
	return isDarkColor(e.Color)
}

// CursorColorEvent represents a cursor color change event. This event is
// emitted when the program requests the terminal cursor color using
// [ansi.RequestCursorColor].
type CursorColorEvent struct{ color.Color }

// String returns the hex representation of the color.
func (e CursorColorEvent) String() string {
	return colorToHex(e)
}

// IsDark returns whether the color is dark.
func (e CursorColorEvent) IsDark() bool {
	return isDarkColor(e)
}

// WindowOpEvent is a window operation (XTWINOPS) report event. This is used to
// report various window operations such as reporting the window size or cell
// size.
type WindowOpEvent struct {
	Op   int
	Args []int
}

// CapabilityEvent represents a Termcap/Terminfo response event. Termcap
// responses are generated by the terminal in response to RequestTermcap
// (XTGETTCAP) requests.
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
type CapabilityEvent string

// ClipboardSelection represents a clipboard selection. The most common
// clipboard selections are "system" and "primary" and selections.
type ClipboardSelection = byte

// Clipboard selections.
const (
	SystemClipboard  ClipboardSelection = ansi.SystemClipboard
	PrimaryClipboard ClipboardSelection = ansi.PrimaryClipboard
)

// ClipboardEvent is a clipboard read message event. This message is emitted when
// a terminal receives an OSC52 clipboard read message event.
type ClipboardEvent struct {
	Content   string
	Selection ClipboardSelection
}

// String returns the string representation of the clipboard message.
func (e ClipboardEvent) String() string {
	return e.Content
}
