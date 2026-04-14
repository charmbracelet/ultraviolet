// Example: kitty text-sizing protocol.
//
// Demonstrates rendering text at 2x, 3x, and 4x scale using OSC 66, the
// kitty text-sizing protocol:
//
//	https://sw.kovidgoyal.net/kitty/text-sizing-protocol/
//
// The string passed to the screen contains the OSC 66 escape verbatim;
// ultraviolet recognises it, reserves the correct number of cells in the
// cell buffer, and forwards the bytes to the terminal so the glyph is
// rendered at the requested scale. On terminals that do not implement the
// protocol the sequence is ignored and the glyph renders at its natural
// size, so the example is safe to run anywhere.
//
// Run it in kitty to see the scaled glyphs. Press any key to exit.
package main

import (
	"fmt"
	"log"

	uv "github.com/charmbracelet/ultraviolet"
)

// scale wraps s in an OSC 66 escape requesting an sx cell scale factor.
// BEL is used as the terminator (rather than ESC+\) because it is a single
// unambiguous byte. See styled.go for why that matters when an SGR sequence
// follows the OSC.
func scale(s string, sx int) string {
	return fmt.Sprintf("\x1b]66;s=%d;%s\x07", sx, s)
}

func main() {
	t := uv.DefaultTerminal()
	scr := t.Screen()
	scr.EnterAltScreen()

	if err := t.Start(); err != nil {
		log.Fatalf("start: %v", err)
	}
	defer t.Stop()

	header := "ultraviolet + kitty text-sizing protocol"
	footer := "press any key to exit"

	lines := []string{
		"1x " + "Hello, world!",
		"2x " + scale("H", 2) + scale("i", 2) + "!",
		"3x " + scale("3", 3) + scale("x", 3),
		"4x " + scale("B", 4) + scale("I", 4) + scale("G", 4),
		"mixed: " + scale("A", 2) + " " + scale("B", 3) + " " + scale("C", 4),
	}

	draw := func() {
		b := scr.Bounds()
		for y := 0; y < b.Dy(); y++ {
			for x := 0; x < b.Dx(); x++ {
				scr.SetCell(x, y, nil)
			}
		}

		uv.NewStyledString(header).Draw(scr, uv.Rect(2, 1, len(header), 1))

		y := 4
		for _, l := range lines {
			// Reserve enough vertical space for the tallest glyph on the
			// line. The protocol scales height along with width, so a 4x
			// line occupies 4 rows.
			uv.NewStyledString(l).Draw(scr, uv.Rect(2, y, b.Dx()-4, 5))
			y += 5
		}

		uv.NewStyledString(footer).Draw(scr, uv.Rect(2, b.Dy()-2, len(footer), 1))

		scr.Render()
		scr.Flush()
	}

	draw()
	defer draw()

	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			scr.Resize(ev.Width, ev.Height)
			draw()
		case uv.KeyPressEvent:
			_ = ev
			return
		}
	}
}
