package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

func main() {
	term := uv.DefaultTerminal()
	if err := term.MakeRaw(); err != nil {
		log.Fatalf("failed to make terminal raw: %v", err)
	}
	if err := term.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	term.EnterAltScreen()

	defer func() {
		if r := recover(); r != nil {
			_ = term.Restore()
			log.Printf("recovered from panic: %v", r)
			debug.PrintStack()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := 5
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				term.SendEvent(ctx, t)
			}
		}
	}()

OUT:
	for ev := range term.Events(ctx) {
		switch e := ev.(type) {
		case uv.WindowSizeEvent:
			term.Resize(e.Width, e.Height)
		case uv.KeyPressEvent:
			switch {
			case e.MatchStrings("q", "ctrl+c"):
				cancel()
				break OUT
			}
		case time.Time: // ticker event
			counter--
			if counter < 0 {
				panic("Time's up!")
			}
		}

		view := fmt.Sprintf("Panicing after %d seconds...\nPress 'q' or 'Ctrl+C' to exit.", counter)
		uv.NewStyledString(view).Draw(term, term.Bounds())
		if err := term.Display(); err != nil {
			log.Fatalf("failed to display terminal: %v", err)
		}
	}

	ctx = context.Background()
	if err := term.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown terminal: %v", err)
	}
}
