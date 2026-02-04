package main

import (
	"log"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

type layout struct {
	main    uv.Rectangle
	footer  uv.Rectangle
	sidebar uv.Rectangle
}

func makeLayout(r uv.Rectangle) layout {
	m, s := uv.SplitHorizontal(r, uv.Percent(80))
	m, f := uv.SplitVertical(m, uv.Fixed(r.Dy()-7))
	return layout{
		main:    m,
		footer:  f,
		sidebar: s,
	}
}

func main() {
	t := uv.DefaultTerminal(nil)
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalln("failed to start terminal:", err)
	}

	defer t.Stop()

	var area uv.Rectangle
	blue := uv.EmptyCell
	blue.Style.Bg = ansi.Blue
	red := uv.EmptyCell
	red.Style.Bg = ansi.Red
	green := uv.EmptyCell
	green.Style.Bg = ansi.Green

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ticker.C:
			l := makeLayout(area)
			screen.FillArea(scr, &blue, l.main)
			screen.FillArea(scr, &red, l.footer)
			screen.FillArea(scr, &green, l.sidebar)
			scr.Render()
			scr.Flush()
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				area = ev.Bounds()
				scr.Resize(area.Dx(), area.Dy())
			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("ctrl+c", "q"):
					break LOOP
				}
			}
		}
	}
}
