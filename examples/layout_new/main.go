package main

import (
	"context"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/component/styledstring"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	t := tv.DefaultTerminal()
	if err := t.Start(); err != nil {
		log.Fatalf("starting program: %v", err)
	}

	physicalWidth, physicalHeight, err := t.GetSize()
	if err != nil {
		log.Fatalf("getting size: %v", err)
	}

	// Use altscreen mode
	t.EnterAltScreen()
	defer t.ExitAltScreen()

	// Enable mouse events
	modes := []ansi.Mode{
		ansi.ButtonEventMouseMode,
		ansi.SgrExtMouseMode,
	}

	os.Stdout.WriteString(ansi.SetMode(modes...))         //nolint:errcheck
	defer os.Stdout.WriteString(ansi.ResetMode(modes...)) //nolint:errcheck

	// Create styles
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	sidebarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#383838")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(1)

	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#F0F0F0")).
		Foreground(lipgloss.Color("#000000")).
		Padding(1)

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	cardStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FFFFFF")).
		Foreground(lipgloss.Color("#000000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#CCCCCC")).
		Padding(1).
		Margin(1)

	// Create the main layout using the new layout system
	mainArea := tv.Rect(0, 0, physicalWidth, physicalHeight)

	// Main vertical layout: header, body, footer
	mainLayout := tv.NewVerticalLayout(
		tv.Item(nil, tv.FixedSize(3)),      // Header
		tv.Item(nil, tv.FlexSize{Grow: 1}), // Body (flexible)
		tv.Item(nil, tv.FixedSize(3)),      // Footer
	)
	mainResult := mainLayout.Calculate(mainArea)

	headerArea := mainResult.Areas[0]
	bodyArea := mainResult.Areas[1]
	footerArea := mainResult.Areas[2]

	// Body layout: sidebar and content
	bodyLayout := tv.NewHorizontalLayout(
		tv.Item(tv.FixedSize(20), nil),     // Sidebar
		tv.Item(tv.FlexSize{Grow: 1}, nil), // Content (flexible)
	)
	bodyLayout.Gap = 1
	bodyResult := bodyLayout.Calculate(bodyArea)

	sidebarArea := bodyResult.Areas[0]
	contentArea := bodyResult.Areas[1]

	// Content area with grid layout for cards
	cardGrid := tv.NewGrid(3,
		tv.GridItemAt(1, 1, tv.Item(nil, tv.FixedSize(8))),
		tv.GridItemAt(2, 1, tv.Item(nil, tv.FixedSize(8))),
		tv.GridItemAt(3, 1, tv.Item(nil, tv.FixedSize(8))),
		tv.GridItemSpan(1, 2, 2, 1, tv.Item(nil, tv.FixedSize(8))), // Spans 2 columns
		tv.GridItemAt(3, 2, tv.Item(nil, tv.FixedSize(8))),
	)
	cardGrid.ColumnGap = 2
	cardGrid.RowGap = 1
	gridResult := cardGrid.Calculate(contentArea)

	// Create frame
	f := &tv.Frame{
		Buffer:   tv.NewBuffer(physicalWidth, physicalHeight),
		Viewport: tv.FullViewport{},
		Area:     mainArea,
	}

	display := func() {
		f.Buffer.Clear()

		// Render header
		headerText := headerStyle.Width(headerArea.Dx()).Render("New Layout System Demo")
		headerSs := styledstring.New(ansi.WcWidth, headerText)
		f.RenderComponent(headerSs, headerArea) //nolint:errcheck

		// Render sidebar
		sidebarContent := sidebarStyle.
			Width(sidebarArea.Dx() - 2).
			Height(sidebarArea.Dy() - 2).
			Render("Navigation\n\n• Home\n• About\n• Contact\n• Settings")
		sidebarSs := styledstring.New(ansi.WcWidth, sidebarContent)
		f.RenderComponent(sidebarSs, sidebarArea) //nolint:errcheck

		// Render main content area background
		contentText := contentStyle.
			Width(contentArea.Dx() - 2).
			Height(contentArea.Dy() - 2).
			Render("Main Content Area")
		contentSs := styledstring.New(ansi.WcWidth, contentText)
		f.RenderComponent(contentSs, contentArea) //nolint:errcheck

		// Render content cards using grid
		cardContents := []string{
			"Card 1\n\nThis is the first card with some content.",
			"Card 2\n\nThis is the second card with different content.",
			"Card 3\n\nThis is the third card in the top row.",
			"Large Card\n\nThis card spans two columns and shows how grid spanning works in the new layout system.",
			"Card 5\n\nThis is the last card in the grid.",
		}

		for i, cardArea := range gridResult.Areas {
			if i < len(cardContents) {
				cardText := cardStyle.
					Width(cardArea.Dx() - 4).
					Height(cardArea.Dy() - 4).
					Render(cardContents[i])
				cardSs := styledstring.New(ansi.WcWidth, cardText)
				f.RenderComponent(cardSs, cardArea) //nolint:errcheck
			}
		}

		// Render footer
		footerText := footerStyle.Width(footerArea.Dx()).Render("Press 'q' to quit | Arrow keys to navigate")
		footerSs := styledstring.New(ansi.WcWidth, footerText)
		f.RenderComponent(footerSs, footerArea) //nolint:errcheck

		t.Display(f)
	}

	// Set terminal to raw mode
	if err := t.MakeRaw(); err != nil {
		log.Fatalf("making raw: %v", err)
	}
	defer t.Restore() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First display
	display()

	// Event loop
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			// Mark screen to be redrawn.
			t.ClearScreen()

			// Recalculate layout on resize
			mainArea = tv.Rect(0, 0, ev.Width, ev.Height)
			f.Area = mainArea
			f.Buffer.Resize(ev.Width, ev.Height)
			t.Resize(ev.Width, ev.Height)

			// Recalculate all layouts
			mainResult = mainLayout.Calculate(mainArea)
			headerArea = mainResult.Areas[0]
			bodyArea = mainResult.Areas[1]
			footerArea = mainResult.Areas[2]

			bodyResult = bodyLayout.Calculate(bodyArea)
			sidebarArea = bodyResult.Areas[0]
			contentArea = bodyResult.Areas[1]

			gridResult = cardGrid.Calculate(contentArea)

		case tv.KeyPressEvent:
			switch {
			case ev.MatchStrings("ctrl+c", "q"):
				cancel()
			}
		}

		display()
	}

	// Shutdown gracefully
	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("shutting down program: %v", err)
	}
}

func init() {
	f, err := os.OpenFile("layout_demo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}
