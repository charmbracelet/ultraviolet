// Command widths renders scripts and emoji whose column width the two width
// models disagree about, so drift between what ultraviolet thinks it drew and
// what the terminal actually painted is visible on screen.
//
// Every sample sits in a box of the same inner width, padded out using
// ultraviolet's own measurement. If the width model matches the terminal, the
// right edges form a straight column. Any edge that sits left or right of the
// rest is the renderer and the terminal disagreeing about that sample, which is
// the bug this example exists to show.
//
// Press w to force wcwidth, g to force grapheme width, q to quit. Terminals
// that implement DEC mode 2027 line up under grapheme width; everything else
// lines up under wcwidth.
package main

import (
	"fmt"
	"log"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

var samples = []struct {
	name string
	text string
}{
	{"ascii", "hello world"},
	{"hindi", "नमस्ते दुनिया"},
	{"hindi conjunct", "स्ते क्ष त्र ज्ञ"},
	{"combining", "café résumé"},
	{"cjk", "世界你好"},
	{"emoji", "🌈 🔥 🦄"},
	{"zwj family", "👨‍👩‍👧‍👦"},
	{"zwj work", "👨‍💻 👩‍🚀"},
	{"zwj flag", "🏳️‍🌈"},
	{"skin tone", "👍🏽 👋🏿"},
	{"vs16", "⚠️ ❤️ ✈️"},
	{"keycap", "1️⃣ 2️⃣ 3️⃣"},
	{"flags", "🇺🇸 🇯🇵"},
}

// boxWidth is the inner width every sample is padded to.
const boxWidth = 24

func main() {
	t := uv.DefaultTerminal()
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start: %v", err)
	}
	defer t.Stop() //nolint:errcheck

	scr.EnterAltScreen()

	// forced reports whether the width method was pinned from the keyboard,
	// which is what makes the two models comparable on the same terminal.
	var forced string

	display := func() {
		var b strings.Builder

		method := "wcwidth"
		if scr.WidthMethod() == ansi.GraphemeWidth {
			method = "grapheme (DEC mode 2027)"
		}
		fmt.Fprintf(&b, "width method: %s", method)
		if forced != "" {
			fmt.Fprintf(&b, "  [forced with %q]", forced)
		}
		b.WriteString("\n")
		b.WriteString("the right edges below should form one straight column\n\n")

		// A ruler, so a drifting edge can be read off in columns.
		ruler := strings.Repeat("....|....+", (boxWidth+12)/10+1)
		b.WriteString(ruler[:boxWidth+2] + "\n")

		for _, s := range samples {
			w := scr.StringWidth(s.text)
			pad := boxWidth - w
			if pad < 0 {
				pad = 0
			}
			fmt.Fprintf(&b, "│%s%s│ uv=%-2d ansi=%-2d  %s\n",
				s.text, strings.Repeat(" ", pad), w,
				ansi.StringWidth(s.text), s.name)
		}

		b.WriteString("\nw: force wcwidth   g: force grapheme width   q: quit")

		uv.NewStyledString(b.String()).Draw(scr, scr.Bounds())
		scr.Render() //nolint:errcheck
		scr.Flush()  //nolint:errcheck
	}

	display()

	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			scr.Resize(ev.Width, ev.Height)
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("w"):
				scr.SetWidthMethod(ansi.WcWidth)
				forced = "w"
				display()
			case ev.MatchString("g"):
				scr.SetWidthMethod(ansi.GraphemeWidth)
				forced = "g"
				display()
			default:
				return
			}
		}
	}
}
