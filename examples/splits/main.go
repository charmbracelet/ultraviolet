package main

import (
	"context"
	"log"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

type layout struct {
	main    uv.Rectangle
	footer  uv.Rectangle
	sidebar uv.Rectangle
}

func makeLayout(r uv.Rectangle) layout {
	m, s := uv.SplitHorizontal(r, uv.Percent(80))
	// m, s := uv.SplitHorizontal(r, uv.Fixed(r.Dx()-40))
	m, f := uv.SplitVertical(m, uv.Fixed(r.Dy()-7))
	return layout{
		main:    m,
		footer:  f,
		sidebar: s,
	}
}

func main() {
	t := uv.DefaultTerminal()

	if err := t.Start(); err != nil {
		log.Fatalln("failed to start terminal:", err)
	}

	var area uv.Rectangle
	blue := uv.EmptyCell
	blue.Style.Bg = ansi.Blue
	red := uv.EmptyCell
	red.Style.Bg = ansi.Red
	green := uv.EmptyCell
	green.Style.Bg = ansi.Green

	ticker := time.NewTicker(time.Second / 60)

LOOP:
	for {
		select {
		case <-ticker.C:
			l := makeLayout(area)
			t.FillArea(&blue, l.main)
			t.FillArea(&red, l.footer)
			t.FillArea(&green, l.sidebar)
			_ = t.Display()
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				area = ev.Bounds()
				t.Resize(area.Dx(), area.Dy())
				t.Erase()
			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("ctrl+c", "q"):
					break LOOP
				}
			}
		}
	}

	if err := t.Shutdown(context.Background()); err != nil {
		log.Println("failed to shutdown terminal:", err)
	}
}
