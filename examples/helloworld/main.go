package main

import (
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	// Create a new terminal screen
	t := uv.DefaultTerminal(nil)

	if err := run(t); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(t *uv.Terminal) error {
	scr := t.Screen()

	// Start in alternate screen mode
	scr.EnterAltScreen()

	if err := t.Start(); err != nil {
		return err
	}

	defer t.Stop()

	ctx := screen.NewContext(scr)
	view := []string{
		"Hello, World!",
		"Press any key to exit.",
	}
	viewWidths := []int{
		ansi.StringWidth(view[0]),
		ansi.StringWidth(view[1]),
	}

	display := func() {
		screen.Clear(scr)
		bounds := scr.Bounds()
		for i, line := range view {
			x := (bounds.Dx() - viewWidths[i]) / 2
			y := (bounds.Dy()-len(view))/2 + i
			ctx.DrawString(line, x, y)
		}
		scr.Render()
		scr.Flush()
	}

	// initial render
	display()

	// last render
	defer display()

	var physicalWidth, physicalHeight int
	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			physicalWidth = ev.Width
			physicalHeight = ev.Height
			if scr.AltScreen() {
				scr.Resize(physicalWidth, physicalHeight)
			} else {
				scr.Resize(physicalWidth, len(view))
			}
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("space"):
				if scr.AltScreen() {
					scr.ExitAltScreen()
					scr.Resize(physicalWidth, len(view))
				} else {
					scr.EnterAltScreen()
					scr.Resize(physicalWidth, physicalHeight)
				}
				display()
			case ev.MatchString("ctrl+z"):
				_ = t.Stop()

				uv.Suspend()

				_ = t.Start()
			default:
				return nil
			}
		}
	}

	return nil
}
