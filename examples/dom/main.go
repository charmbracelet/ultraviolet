package main

import (
	"context"
	"log"
	"os"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/dom"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	// Create terminal
	t := uv.DefaultTerminal()

	// Start the terminal
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	// Use alternate screen
	t.EnterAltScreen()

	// Create a scrollable box with multiple items
	items := dom.VBox(
		dom.Text("Item 1 - Scroll with j/k or arrows"),
		dom.Text("Item 2 - Box is a block-level element"),
		dom.Text("Item 3 - Text is an inline element"),
		dom.Text("Item 4 - Simple and composable"),
		dom.Text("Item 5 - No wrappers needed"),
		dom.Text("Item 6 - Clean architecture"),
		dom.Text("Item 7 - More items..."),
		dom.Text("Item 8 - Keep scrolling!"),
		dom.Text("Item 9 - Almost there..."),
		dom.Text("Item 10 - You found it!"),
	)

	// Wrap in a scrollable box with border
	scrollableBox := dom.NewBox(items).
		WithBorder(dom.BorderStyleRounded()).
		WithPadding(1).
		WithFocus(true)

	// Create a DOM-based UI showcasing box model
	ui := dom.NewBox(
		dom.VBox(
			dom.Padding(
				dom.VBox(
					dom.Styled("Ultraviolet DOM - Block & Inline Elements", uv.Style{Attrs: uv.AttrBold}),
					dom.Spacer(0, 1),
					dom.Text("Box is a block-level element (like HTML <div>)."),
					dom.Text("Text is an inline element (like HTML <span>)."),
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					// Scrollable list box
					scrollableBox,
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					dom.HBox(
						dom.VBox(
							dom.Styled("Controls:", uv.Style{Attrs: uv.AttrBold}),
							dom.Text("↑/k: Scroll up"),
							dom.Text("↓/j: Scroll down"),
							dom.Text("q: Quit"),
						),
						dom.Spacer(2, 0),
						dom.SeparatorVertical(),
						dom.Spacer(2, 0),
						dom.VBox(
							dom.Styled("Features:", uv.Style{Attrs: uv.AttrBold}),
							dom.Text("✓ Box model"),
							dom.Text("✓ Scrolling"),
							dom.Text("✓ Focus"),
							dom.Text("✓ Borders"),
						),
					),
					dom.Spacer(0, 1),
					dom.Center(
						dom.Styled("Press 'q' or Ctrl+C to quit", uv.Style{
							Fg:    ansi.BrightYellow,
							Attrs: uv.AttrItalic,
						}),
					),
				),
				1, 1, 1, 1,
			),
		),
	).WithBorder(dom.BorderStyleRounded())

	// Display function
	display := func() {
		t.Erase()
		bounds := t.Bounds()
		ui.Render(t, bounds)
		if err := t.Display(); err != nil {
			log.Printf("display error: %v", err)
		}
	}

	// Initial display
	display()

	// Event loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			goto cleanup
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				t.Resize(ev.Width, ev.Height)
				display()
			case uv.KeyPressEvent:
				switch {
				case ev.MatchString("q", "ctrl+c", "esc"):
					cancel()
				case ev.MatchString("up", "k"):
					scrollableBox.ScrollUp(1)
					display()
				case ev.MatchString("down", "j"):
					scrollableBox.ScrollDown(1)
					display()
				default:
					display()
				}
			default:
				display()
			}
		}
	}

cleanup:
	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

func init() {
	f, err := os.OpenFile("dom_example.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(f)
}
