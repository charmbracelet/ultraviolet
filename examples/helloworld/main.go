package main

import (
	"context"
	"log"
	"os"

	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/component/styledstring"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	// Create a new terminal screen
	t := tv.NewTerminal(os.Stdin, os.Stdout, os.Environ())
	// Or simply use...
	// t := tv.DefaultTerminal()

	// Set the terminal to raw mode.
	if err := t.MakeRaw(); err != nil {
		log.Fatal(err)
	}

	// Make sure we restore the terminal to its original state
	// before we exit. We don't care about errors here, but you
	// can handle them if you want.
	defer t.Restore() //nolint:errcheck

	// Enter the alternate screen buffer
	t.EnterAltScreen()
	// Make sure we leave the alternate screen buffer
	// when we are done with our program.
	defer t.ExitAltScreen()

	// Create a new program
	// Start the program
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	// We want to be able to stop the terminal input loop
	// whenever we call cancel().
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fixed := tv.Rect(10, 10, 40, 20)

	// This will block until we close the events
	// channel or cancel the context.
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			t.Resize(ev.Width, ev.Height)
			t.Clear()
		case tv.KeyPressEvent:
			if ev.MatchStrings("q", "ctrl+c") {
				cancel() // This will stop the loop
			}
		}

		// Display the frame with the styled string
		// We want the component to occupy the given area which is the
		// entire screen because we're using the alternate screen buffer.
		tv.FillArea(t, &tv.Cell{
			Content: " ",
			Style:   tv.Style{Bg: ansi.Red},
		}, fixed)
		// We will use the StyledString component to simplify displaying
		// text on the screen.
		ss := styledstring.New("Hello, World!")
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

	if err := t.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func init() {
	f, err := os.OpenFile("tv.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(f)
}
