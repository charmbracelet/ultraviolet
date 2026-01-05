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

	// Create a DOM-based UI
	ui := dom.Window("DOM Example",
		dom.VBox(
			dom.PaddingAll(
				dom.VBox(
					dom.Text("Welcome to Ultraviolet DOM!"),
					dom.Spacer(0, 1),
					dom.Paragraph("This is a demonstration of the new DOM package that allows you to build TUIs using a declarative, DOM-inspired approach similar to HTML and FTXUI."),
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					dom.HBox(
						dom.VBox(
							dom.Styled("Left Panel", uv.Style{Attrs: uv.AttrBold}),
							dom.Spacer(0, 1),
							dom.Checkbox("Option 1", true),
							dom.Checkbox("Option 2", false),
							dom.Checkbox("Option 3", true),
						),
						dom.Spacer(2, 0),
						dom.SeparatorVertical(),
						dom.Spacer(2, 0),
						dom.VBox(
							dom.Styled("Right Panel", uv.Style{Attrs: uv.AttrBold}),
							dom.Spacer(0, 1),
							dom.HBox(
								dom.Button("OK"),
								dom.Spacer(2, 0),
								dom.Button("Cancel"),
							),
							dom.Spacer(0, 1),
							dom.Input(20, "Enter text..."),
						),
					),
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					dom.Center(
						dom.Border(
							dom.PaddingAll(
								dom.Styled("Centered and Bordered", uv.Style{
									Fg: ansi.BrightCyan,
								}),
								1,
							),
						),
					),
				),
				1,
			),
		),
	)

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
				if ev.MatchString("q", "ctrl+c", "esc") {
					cancel()
				} else {
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
