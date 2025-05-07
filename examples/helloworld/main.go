package main

import (
	"context"
	"log"
	"os"

	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/widget/styledstring"
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
	defer t.LeaveAltScreen()

	// Create a new program
	p := tv.NewProgram(t)
	// Start the program
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	// We want to be able to stop the terminal input loop
	// whenever we call cancel().
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This will block until we close the events
	// channel or cancel the context.
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			p.Resize(ev.Width, ev.Height)
		case tv.KeyPressEvent:
			if ev.MatchStrings("q", "ctrl+c") {
				cancel() // This will stop the loop
			}
		}

		// Display the frame with the styled string
		if err := p.Display(func(f *tv.Frame) error {
			// We will use the StyledString widget to simplify
			// displaying text on the screen.
			// Using [ansi.WcWidth] will ensure that the text is
			// displayed correctly on the screen using traditional
			// terminal width calculations.
			ss := styledstring.New(ansi.WcWidth, "Hello, World!")
			// We want the widget to occupy the given area which
			// is the entire screen because we're using the alternate
			// screen buffer.
			area := f.Area
			area.Min.X = (area.Max.X / 2) - 6
			area.Min.Y = (area.Max.Y / 2) - 1
			return f.RenderWidget(ss, area)
		}); err != nil {
			log.Fatal(err)
		}
	}

	if err := p.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
