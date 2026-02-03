package main

import (
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
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

	view := uv.NewStyledString("    Hello, World!\nPress any key to exit.")
	viewWidth := view.UnicodeWidth()
	viewHeight := view.Height()

	display := func() {
		screen.Clear(scr)
		view.Draw(scr, uv.CenterRect(scr.Bounds(), viewWidth, viewHeight))
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
				scr.Resize(physicalWidth, viewHeight)
			}
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("space"):
				if scr.AltScreen() {
					scr.ExitAltScreen()
					scr.Resize(physicalWidth, viewHeight)
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
