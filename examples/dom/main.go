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

	// Create selectable and focusable items for the scrollable list
	items := []dom.Element{
		dom.MakeSelectableAndFocusable(dom.Text("Item 1 - Press Space to select")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 2 - Use Tab to navigate")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 3 - Selection is character-level")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 4 - Scroll with j/k or arrows")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 5 - Focus shown with reverse video")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 6 - Hard-wrap demo below")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 7 - More items...")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 8 - Keep scrolling!")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 9 - Almost there...")),
		dom.MakeSelectableAndFocusable(dom.Text("Item 10 - You found it!")),
	}

	// Create a scrollable container
	scrollable := dom.ScrollableVBox(items...)
	scrollable.SetFocus(0) // Focus first item

	// Create a DOM-based UI showcasing new features
	ui := dom.Window("DOM Example - Scrolling, Focus & Selection",
		dom.VBox(
			dom.PaddingAll(
				dom.VBox(
					dom.Styled("Welcome to Ultraviolet DOM!", uv.Style{Attrs: uv.AttrBold}),
					dom.Spacer(0, 1),
					// Hard-wrap demonstration
					dom.TextHardWrap("This text demonstrates hard-wrapping: when you have a very long line of text that exceeds the available width, it will automatically wrap to the next line at character boundaries instead of being truncated. This is useful for displaying large amounts of text in terminal UIs."),
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					// Scrollable list with focus and selection
					dom.Border(
						dom.PaddingAll(
							scrollable,
							1,
						),
					),
					dom.Spacer(0, 1),
					dom.Separator(),
					dom.Spacer(0, 1),
					dom.HBox(
						dom.VBox(
							dom.Styled("Controls:", uv.Style{Attrs: uv.AttrBold}),
							dom.Text("↑/k: Scroll up"),
							dom.Text("↓/j: Scroll down"),
							dom.Text("Tab: Next focus"),
							dom.Text("Space: Select"),
						),
						dom.Spacer(2, 0),
						dom.SeparatorVertical(),
						dom.Spacer(2, 0),
						dom.VBox(
							dom.Styled("Features:", uv.Style{Attrs: uv.AttrBold}),
							dom.Checkbox("Hard-wrap text", true),
							dom.Checkbox("Scrollback", true),
							dom.Checkbox("Focus tracking", true),
							dom.Checkbox("Character selection", true),
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
				switch {
				case ev.MatchString("q", "ctrl+c", "esc"):
					cancel()
				case ev.MatchString("up", "k"):
					scrollable.ScrollUp(1)
					display()
				case ev.MatchString("down", "j"):
					scrollable.ScrollDown(1)
					display()
				case ev.MatchString("tab"):
					scrollable.FocusNext()
					display()
				case ev.MatchString("shift+tab"):
					scrollable.FocusPrevious()
					display()
				case ev.MatchString(" "):
					// Toggle selection on focused element
					if focused := scrollable.GetFocusedElement(); focused != nil {
						if selectable, ok := focused.(dom.Selectable); ok {
							_, hasSelection := selectable.GetSelection()
							if hasSelection {
								selectable.ClearSelection()
							} else {
								// Select the entire first line
								selectable.SetSelection(dom.SelectionRange{
									StartLine: 0,
									StartCol:  0,
									EndLine:   0,
									EndCol:    50, // Select first 50 chars
								})
							}
						}
					}
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
