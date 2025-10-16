package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	// Use the main screen
	t.ExitAltScreen()

	defer func() {
		if r := recover(); r != nil {
			_ = t.Teardown()
			fmt.Fprintf(os.Stderr, "\r\nrecovered from panic: %v", r)
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
			case <-ticker.C:
				t.SendEvent(tickEvent{})
			}
		}
	}()

	view := func(c int) string {
		return fmt.Sprintf("Panicing after %d seconds...\nPress 'q' or 'Ctrl+C' to exit.", c)
	}

OUT:
	for {
		select {
		case <-ctx.Done():
			break OUT
		case ev := <-t.Events():
			switch e := ev.(type) {
			case uv.KeyPressEvent:
				switch {
				case e.MatchString("q", "ctrl+c"):
					cancel()
					break OUT
				}
			case tickEvent: // ticker event
				counter--
				if counter < 0 {
					panic("Time's up!")
				}
			}
		}

		t.Draw(uv.NewStyledString(view(counter)))
		if err := t.Display(); err != nil {
			log.Fatalf("failed to display terminal: %v", err)
		}
	}

	t.Draw(uv.NewStyledString(view(counter) + "\n"))
	if err := t.Display(); err != nil {
		log.Fatalf("failed to display terminal: %v", err)
	}

	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown terminal: %v", err)
	}
}
