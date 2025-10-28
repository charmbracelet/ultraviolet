package uvtest

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestHelloApp(t *testing.T) {
	tt, err := NewTestTerminal(t)
	if err != nil {
		t.Fatalf("failed to create test terminal: %v", err)
	}

	draw := uv.NewStyledString("\x1b[1mHello, World!\x1b[m")

	go func() {
		tt.SendText("x")
		tt.SendText("q")
	}()

	var quit bool
LOOP:
	for !quit {
		select {
		case <-t.Context().Done():
			break LOOP
		case ev := <-tt.Events():
			switch ev := ev.(type) {
			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("ctrl+c", "q"):
					quit = true
				}

				draw.Text += ev.String()
			}

			tt.Render(draw)
			if err := tt.Display(); err != nil {
				t.Errorf("failed to display terminal: %v", err)
			}

			// Wait a bit for the terminal to render
			time.Sleep(50 * time.Millisecond)

			// Take a snapshot after each event
			tt.Snapshot()
		}
	}
}
