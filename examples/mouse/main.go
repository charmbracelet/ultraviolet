package main

import (
	"fmt"
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	t := uv.DefaultTerminal()
	ws, err := t.GetWinsize()
	if err != nil {
		log.Fatalf("failed to get window size: %v", err)
	}
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	defer t.Stop()

	scr.SetMouseMode(uv.MouseModeMotion)
	scr.SetMouseEncoding(uv.MouseEncodingSGRPixel)

	var lastBtn uv.MouseButton
	var lastX, lastY int

	var width int

	display := func() {
		label := fmt.Sprintf(" Button: %-12s Position: (%d, %d)", lastBtn, lastX, lastY)
		var st uv.Style
		st.Bg = ansi.BasicColor(4)
		st.Fg = ansi.Black
		bg := uv.EmptyCell
		bg.Style = st
		screen.FillArea(scr, &bg, uv.Rect(0, 0, width, 1))
		for i, r := range label {
			scr.SetCell(i, 0, &uv.Cell{
				Content: string(r),
				Style:   st,
				Width:   1,
			})
		}
		scr.Render()
		scr.Flush()
	}

	// initial render
	display()

	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			width = ev.Width
			ws.Col = uint16(ev.Width)
			ws.Row = uint16(ev.Height)
			scr.Resize(width, 1)

			// Query the pixel dimensions of the window in-case this platform
			// doesn't report them via Termios [uv.Terminal.GetWinsize].
			_, _ = scr.WriteString(ansi.WindowOp(4))

			display()
		case uv.PixelSizeEvent:
			ws.Xpixel = uint16(ev.Width)
			ws.Ypixel = uint16(ev.Height)
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("q", "ctrl+c"):
				return
			}
		case uv.MouseEvent:
			m := ev.Mouse()
			m = uv.MousePixelToCell(m, ws)
			lastX = m.X
			lastY = m.Y
			if m.Button != uv.MouseNone {
				lastBtn = m.Button
			}
			scr.InsertAbove(fmt.Sprintf("%-20s (%d, %d) %s", ev.String(), m.X, m.Y, m.Button))
			display()
		}
	}
}
