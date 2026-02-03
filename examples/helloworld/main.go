package main

import (
	"log"
	"os"

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

		bounds := scr.Bounds()
		bounds.Min.X = (bounds.Dx() - viewWidth) / 2
		bounds.Min.Y = (bounds.Dy() - viewHeight) / 2

		view.Draw(scr, bounds)
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

func init() {
	f, err := os.OpenFile("uv_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open debug log file: %v", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
