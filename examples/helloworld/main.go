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

	width, height, err := t.GetSize()
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

	// Create a new program
	// Start the program
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	// We want to be able to stop the terminal input loop
	// whenever we call cancel().
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	f := &tv.Frame{
		Buffer:   tv.NewBuffer(width, height),
		Viewport: tv.FullViewport{},
		Area:     tv.Rect(0, 0, width, height),
	}

	// This will block until we close the events
	// channel or cancel the context.
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			width, height = ev.Width, ev.Height
			t.Resize(ev.Width, ev.Height)
			f.Area = tv.Rect(0, 0, ev.Width, ev.Height)
			f.Buffer.Resize(ev.Width, ev.Height)
		case tv.KeyPressEvent:
			if ev.MatchStrings("q", "ctrl+c") {
				cancel() // This will stop the loop
			}
		}

		// Display the frame with the styled string
		f.Buffer.Clear()
		// We will use the StyledString component to simplify displaying
		// text on the screen.
		// Using [ansi.WcWidth] will ensure that the text is
		// displayed correctly on the screen using traditional
		// terminal width calculations.
		ss := styledstring.New(ansi.WcWidth, "Hello, World!")
		// We want the component to occupy the given area which is the
		// entire screen because we're using the alternate screen buffer.
		area := f.Area
		area.Min.X = (area.Max.X / 2) - 6
		area.Min.Y = (area.Max.Y / 2) - 1
		f.RenderComponent(ss, area)
		if err := t.Display(f); err != nil {
			log.Fatal(err)
		}
	}

	if err := t.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
