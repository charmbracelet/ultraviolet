package uv

import (
	"reflect"
	"testing"
)

func TestSignedSGRMouse(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want Event
	}{
		{"\x1b[<0;-5;328M", 12, MouseClickEvent{X: -6, Y: 327, Button: MouseLeft}},
		{"\x1b[<0;-5;328m", 12, MouseReleaseEvent{X: -6, Y: 327, Button: MouseLeft}},
		{"\x1b[<32;50;-3M", 12, MouseMotionEvent{X: 49, Y: -4, Button: MouseLeft}},
		{"\x1b[<32;-5;-3M", 12, MouseMotionEvent{X: -6, Y: -4, Button: MouseLeft}},
		{"\x1b[<64;-5;10M", 12, MouseWheelEvent{X: -6, Y: 9, Button: MouseWheelUp}},
		{"\x1b[<52;-5;10M", 12, MouseMotionEvent{X: -6, Y: 9, Button: MouseLeft, Mod: ModShift | ModCtrl}},
		{"\x9b<0;-5;328M", 11, MouseClickEvent{X: -6, Y: 327, Button: MouseLeft}},
		// The report ends after the first one: the rest is the next event.
		{"\x1b[<0;-5;328Mx", 12, MouseClickEvent{X: -6, Y: 327, Button: MouseLeft}},
		// Incomplete: the whole buffer is unknown, so the reader waits.
		{"\x1b[<32;-", 7, UnknownEvent("\x1b[<32;-")},
		{"\x1b[<32;-5;3", 10, UnknownEvent("\x1b[<32;-5;3")},
		// Malformed.
		{"\x1b[<32;-;3M", 7, UnknownEvent("\x1b[<32;-")},
		{"\x1b[<32;--5;3M", 7, UnknownEvent("\x1b[<32;-")},
		{"\x1b[<32;-5;3;4M", 10, UnknownEvent("\x1b[<32;-5;3")},
		{"\x1b[<32;-5;3X", 10, UnknownEvent("\x1b[<32;-5;3")},
	}
	for _, c := range cases {
		var d EventDecoder
		n, ev := d.Decode([]byte(c.in))
		if n != c.n || !reflect.DeepEqual(ev, c.want) {
			t.Errorf("Decode(%q) = %d, %#v; want %d, %#v", c.in, n, ev, c.n, c.want)
		}
	}
}
