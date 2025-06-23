package main

import (
	"context"
	"errors"
	"log"

	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/component/styledstring"
	"github.com/charmbracelet/uv/screen"
	"github.com/muesli/cancelreader"
)

func main() {
	t := uv.DefaultTerminal()

	if err := t.MakeRaw(); err != nil {
		log.Fatalf("failed to make terminal raw: %v", err)
	}

	width, height, err := t.GetSize()
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	var altScreen bool
	frameHeight := 2
	updateViewport := func(altscreen bool) {
		if altscreen {
			frameHeight = height
			t.EnterAltScreen()
		} else {
			frameHeight = 2
			t.ExitAltScreen()
		}
		t.Resize(width, frameHeight)
	}

	help := "Press space to toggle screen mode or ctrl+c to exit."
	display := func() {
		updateViewport(altScreen)
		var str string
		if altScreen {
			str = "This is using alternate screen mode.\n" + help
		} else {
			str = "This is using inline mode.\n" + help
		}

		ss := styledstring.New(str)
		screen.Clear(t)
		ss.Draw(t, uv.Rect(0, 0, width, frameHeight))
		t.Display()
	}

	var cursorHidden bool
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			width, height = ev.Width, ev.Height
			t.Erase()
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchStrings("ctrl+c", "q"):
				altScreen = false
				stop()
			case ev.MatchString("space"):
				altScreen = !altScreen
			default:
				if cursorHidden {
					t.HideCursor()
				} else {
					t.ShowCursor()
				}
				cursorHidden = !cursorHidden
			}
		}
		display()
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), 5)
	defer cancel()

	if altScreen {
		t.ExitAltScreen()
	}

	if err := t.Err(); err != nil && !errors.Is(err, cancelreader.ErrCanceled) {
		log.Printf("error: %v", err)
	}

	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown program: %v", err)
	}
}
