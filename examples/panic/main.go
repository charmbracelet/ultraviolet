package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

type tickEvent struct{}

func main() {
	t := uv.DefaultTerminal()
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	t.EnterAltScreen()

	defer func() {
		if r := recover(); r != nil {
			_ = t.Restore()
			log.Printf("recovered from panic: %v", r)
			debug.PrintStack()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	evch := make(chan uv.Event)
	counter := 5
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.SendEvent(ctx, tickEvent{})
			}
		}
	}()
	go func() {
		defer close(evch)
		_ = t.StreamEvents(ctx, evch)
	}()

OUT:
	for ev := range evch {
		switch e := ev.(type) {
		case uv.WindowSizeEvent:
			t.Resize(e.Width, e.Height)
		case uv.KeyPressEvent:
			switch {
			case e.MatchStrings("q", "ctrl+c"):
				cancel()
				break OUT
			}
		case tickEvent: // ticker event
			counter--
			if counter < 0 {
				panic("Time's up!")
			}
		}

		view := fmt.Sprintf("Panicing after %d seconds...\nPress 'q' or 'Ctrl+C' to exit.", counter)
		uv.NewStyledString(view).Draw(t, t.Bounds())
		if err := t.Display(); err != nil {
			log.Fatalf("failed to display terminal: %v", err)
		}
	}

	ctx = context.Background()
	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown terminal: %v", err)
	}
}
