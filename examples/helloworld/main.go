package main

import (
	"context"
	"log"
	"os"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	// Create a new terminal screen
	t := uv.NewTerminal(os.Stdin, os.Stdout, os.Environ())
	// Or simply use...
	// t := uv.DefaultTerminal()

	// Create a new program
	// Start the program
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	fixed := uv.Rect(10, 10, 40, 20)
	display := func() {
		// Display the frame with the styled string
		// We want the component to occupy the given area which is the
		// entire screen because we're using the alternate screen buffer.
		screen.FillArea(t, &uv.Cell{
			Content: " ",
			Style:   uv.Style{Fg: ansi.Red},
		}, fixed)
		// We will use the StyledString component to simplify displaying
		// text on the screen.
		ss := uv.NewStyledString("Hello, World!")
		carea := fixed
		carea.Min.X = (carea.Max.X / 2) - 6
		carea.Min.Y = (carea.Max.Y / 2) - 1
		carea.Max.X = carea.Min.X + 12
		carea.Max.Y = carea.Min.Y + 1
		ss.Draw(t, carea)
		if err := t.Display(); err != nil {
			log.Fatal(err)
		}
	}

	// We want to be able to stop the terminal input loop
	// whenever we call cancel().
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This will block until we close the events
	// channel or cancel the context.
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				t.Resize(ev.Width, ev.Height)
				t.Erase()
			case uv.KeyPressEvent:
				if ev.MatchString("q", "ctrl+c") {
					cancel() // This will stop the loop
				} else if ev.MatchString("ctrl+z") {
					t.Erase()
					if err := t.Display(); err != nil {
						log.Fatal(err)
					}
					if t.Pause() != nil {
						log.Fatal("failed to shutdown terminal")
					}

					uv.Suspend()

					if err := t.Resume(); err != nil {
						log.Fatal("failed to resume terminal")
					}
				}
			}
		}

		display()
	}

	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func init() {
	f, err := os.OpenFile("uv_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(f)
}
