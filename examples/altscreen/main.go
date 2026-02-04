package main

import (
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
)

func main() {
	t := uv.DefaultTerminal(nil)
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	defer t.Stop()

	help := "Press space to toggle screen mode or any other key to exit."
	var altScreen bool

	display := func() {
		var str string
		if altScreen {
			str = "This is using alternate screen mode.\n" + help
		} else {
			str = "This is using inline mode.\n" + help
		}

		ss := uv.NewStyledString(str)
		screen.Clear(scr)
		bounds := scr.Bounds()
		bounds.Min.X = (bounds.Dx() - ss.UnicodeWidth()) / 2
		bounds.Min.Y = (bounds.Dy() - ss.Height()) / 2
		ss.Draw(scr, bounds)
		scr.Render()
		scr.Flush()
	}

	// initial render
	display()

	var physicalWidth, physicalHeight int
	for ev := range t.Events() {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			physicalWidth = ev.Width
			physicalHeight = ev.Height
			if altScreen {
				scr.Resize(physicalWidth, physicalHeight)
			} else {
				scr.Resize(physicalWidth, 2)
			}
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchString("space"):
				if altScreen {
					scr.ExitAltScreen()
					scr.Resize(physicalWidth, 2)
				} else {
					scr.EnterAltScreen()
					scr.Resize(physicalWidth, physicalHeight)
				}
				altScreen = !altScreen
				display()
			default:
				return
			}
		}
	}
}
