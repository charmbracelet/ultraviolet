// Mimics a chat TUI like docker-agent: a left "conversation" pane full of
// emoji, a vertical separator, and a right sidebar. The separator is drawn at a
// fixed column on every row. It only stays straight if the terminal and the
// renderer agree on how far the cursor advances after each multi-byte emoji
// cell. Without the cursor realignment fix, the separator (and the whole
// sidebar) drifts left or right on rows that contain wide or ZWJ emoji.
//
// Press 'r' to redraw with a fresh set of emoji, 'q' to quit.
package main

import (
	"log"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
)

// A mix of wide emoji, ZWJ sequences and variation-selector clusters. These are
// exactly the kind of multi-byte/multi-rune cells that desync the cursor.
var emojis = []string{
	"🫩", "🫆", "🫜", "🪾", "🪉", "🫠", "🫡", "🫨", "🥹", "🐦‍🔥",
	"🍋‍🟩", "🍄‍🟫", "🙂‍↔️", "☹️", "❤️", "✈️", "⛓️‍💥", "🩷", "🩵", "🩶",
}

// Short fake chat lines; %s slots are filled with emoji.
var chatLines = []string{
	"user: ship it %s %s",
	"agent: building the image %s",
	"agent: running tests %s %s %s",
	"agent: all green %s done %s",
	"user: nice %s",
	"agent: pushing %s to the registry %s",
}

const sidebarWidth = 24

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

	offset := 0
	display(scr, offset)
	defer display(scr, offset)

	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			scr.Resize(ev.Width, ev.Height)
			display(scr, offset)
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("q"), ev.MatchString("ctrl+c"), ev.MatchString("esc"):
				return nil
			case ev.MatchString("r"):
				offset++
				display(scr, offset)
			}
		}
	}

	return nil
}

func display(scr *uv.TerminalScreen, offset int) {
	screen.Clear(scr)
	ctx := screen.NewContext(scr)

	bounds := scr.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < sidebarWidth+10 || height < 1 {
		return
	}

	// The separator sits just left of the sidebar, at a fixed column.
	sepX := width - sidebarWidth - 1

	for y := range height {
		drawChatRow(scr, ctx, y, sepX, offset)
		ctx.DrawString("│", sepX, y)
		drawSidebarRow(ctx, y, sepX+2)
	}

	scr.Render()
	scr.Flush()
}

// drawChatRow fills the left pane with an emoji-laden chat line, cropped so it
// never reaches the separator column.
func drawChatRow(scr *uv.TerminalScreen, ctx *screen.Context, y, sepX, offset int) {
	tmpl := chatLines[(y+offset)%len(chatLines)]
	text := buildChatLine(tmpl, y+offset)
	// Crop to the available width before the separator.
	for scr.StringWidth(text) > sepX-1 && len(text) > 0 {
		text = text[:len(text)-1]
	}
	ctx.DrawString(text, 0, y)
}

// buildChatLine fills the %s slots of tmpl with emoji picked from seed.
func buildChatLine(tmpl string, seed int) string {
	parts := strings.Split(tmpl, "%s")
	var b strings.Builder
	for i, p := range parts {
		b.WriteString(p)
		if i < len(parts)-1 {
			b.WriteString(emojis[(seed+i)%len(emojis)])
		}
	}
	return b.String()
}

// drawSidebarRow draws a stable, emoji-free sidebar so any wobble is obviously
// caused by the left pane's emoji, not the sidebar content.
func drawSidebarRow(ctx *screen.Context, y, x int) {
	rows := []string{
		" Session",
		" ─────────",
		" model:  gpt",
		" tokens: 1234",
		" tools:  7",
		" status: ready",
	}
	if y < len(rows) {
		ctx.DrawString(rows[y], x, y)
	}
}
