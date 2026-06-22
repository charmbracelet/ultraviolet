// Renders a dense grid of multi-byte glyphs (wide emoji, ZWJ sequences and
// variation-selector clusters) framed by '|' borders on both sides of every
// row. The borders only stay aligned if the terminal and renderer agree on how
// far the cursor advances after each multi-byte cell.
package main

import (
	"fmt"
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
)

var emojis = []string{
	"🫩", "🫆", "🫜", "🪾", "🪉", "🪏", "🫠", "🫡", "🫨", "🥹",
	"🫥", "🫢", "🫣", "🫤", "🫷", "🫸", "🫎", "🫏", "🪽", "🪿",
	"🪼", "🫛", "🫚", "🪭", "🪮", "🪇", "🩷", "🩵", "🩶", "🐦‍🔥",
	"🍋‍🟩", "🍄‍🟫", "🙂‍↔️", "🙂‍↕️", "☹️", "❤️", "✈️", "☀️", "⛓️‍💥",
}

func main() {
	t := uv.DefaultTerminal()
	if err := run(t); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(t *uv.Terminal) error {
	scr := t.Screen()
	scr.EnterAltScreen()
	scr.SetSynchronizedUpdates(true)

	if err := t.Start(); err != nil {
		return err
	}
	defer t.Stop()

	if w, h, err := t.GetSize(); err == nil {
		scr.Resize(w, h)
	}
	display(scr)
	defer display(scr)

	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			scr.Resize(ev.Width, ev.Height)
			display(scr)
		case uv.KeyPressEvent:
			if ev.MatchString("q") || ev.MatchString("ctrl+c") || ev.MatchString("esc") {
				return nil
			}
		}
	}

	return nil
}

func display(scr *uv.TerminalScreen) {
	screen.Clear(scr)
	ctx := screen.NewContext(scr)
	bounds := scr.Bounds()
	width := bounds.Dx()
	if width <= 0 {
		return
	}

	for y := range bounds.Dy() {
		drawRow(scr, ctx, y, width)
	}

	scr.Render()
	scr.Flush()
}

func drawRow(scr *uv.TerminalScreen, ctx *screen.Context, y, width int) {
	prefix := fmt.Sprintf("%2d |", y)
	ctx.DrawString(prefix, 0, y)

	x := scr.WidthMethod().StringWidth(prefix)
	for i := 0; x < width-1; i++ {
		emoji := emojis[(y*17+i)%len(emojis)]
		w := scr.WidthMethod().StringWidth(emoji)
		if x+w > width-1 {
			break
		}
		ctx.DrawString(emoji, x, y)
		x += w
	}

	ctx.DrawString("|", width-1, y)
}
