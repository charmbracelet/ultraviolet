package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/component/styledstring"
	"github.com/charmbracelet/x/ansi"
)

type FlexDemo struct {
	currentDemo int
	demos       []FlexDemoConfig
}

type FlexDemoConfig struct {
	title       string
	description string
	layout      *tv.Layout
}

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

	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#FF6B6B")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#4ECDC4")).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)

	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	// Flex item styles with different colors
	itemStyles := []lipgloss.Style{
		lipgloss.NewStyle().Background(lipgloss.Color("#FF9F43")).Foreground(lipgloss.Color("#000000")).Padding(1).Bold(true),
		lipgloss.NewStyle().Background(lipgloss.Color("#10AC84")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1).Bold(true),
		lipgloss.NewStyle().Background(lipgloss.Color("#5F27CD")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1).Bold(true),
		lipgloss.NewStyle().Background(lipgloss.Color("#EE5A24")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1).Bold(true),
		lipgloss.NewStyle().Background(lipgloss.Color("#0984E3")).Foreground(lipgloss.Color("#FFFFFF")).Padding(1).Bold(true),
	}

	// Create flex demo configurations
	flexDemo := &FlexDemo{
		currentDemo: 0,
		demos: []FlexDemoConfig{
			{
				title:       "Basic Horizontal Flex",
				description: "Three items with equal flex grow (1:1:1)",
				layout: &tv.Layout{
					Direction: tv.Horizontal,
					Items: []tv.LayoutItem{
						tv.Item(tv.FlexSize{Grow: 1}, nil),
						tv.Item(tv.FlexSize{Grow: 1}, nil),
						tv.Item(tv.FlexSize{Grow: 1}, nil),
					},
					Gap: 5,
				},
			},
			{
				title:       "Proportional Flex Growth",
				description: "Items with different grow factors (1:2:3)",
				layout: &tv.Layout{
					Direction: tv.Horizontal,
					Items: []tv.LayoutItem{
						tv.Item(tv.FlexSize{Grow: 1}, nil),
						tv.Item(tv.FlexSize{Grow: 2}, nil),
						tv.Item(tv.FlexSize{Grow: 3}, nil),
					},
					Gap: 5,
				},
			},
			{
				title:       "Fixed + Flex Layout",
				description: "Fixed sidebar (200px) + flexible content + fixed sidebar (150px)",
				layout: &tv.Layout{
					Direction: tv.Horizontal,
					Items: []tv.LayoutItem{
						tv.Item(tv.FixedSize(200), nil),
						tv.Item(tv.FlexSize{Grow: 1}, nil),
						tv.Item(tv.FixedSize(150), nil),
					},
					Gap: 10,
				},
			},
			{
				title:       "Flex with Basis",
				description: "Items with initial size (basis) that then grow",
				layout: &tv.Layout{
					Direction: tv.Horizontal,
					Items: []tv.LayoutItem{
						tv.Item(tv.FlexSize{Grow: 1, Basis: tv.FixedSize(100)}, nil),
						tv.Item(tv.FlexSize{Grow: 2, Basis: tv.FixedSize(50)}, nil),
						tv.Item(tv.FlexSize{Grow: 1, Basis: tv.FixedSize(200)}, nil),
					},
					Gap: 5,
				},
			},
			{
				title:       "Vertical Flex Layout",
				description: "Vertical flexbox with different grow factors",
				layout: &tv.Layout{
					Direction: tv.Vertical,
					Items: []tv.LayoutItem{
						tv.Item(nil, tv.FixedSize(60)),
						tv.Item(nil, tv.FlexSize{Grow: 2}),
						tv.Item(nil, tv.FlexSize{Grow: 1}),
						tv.Item(nil, tv.FixedSize(40)),
					},
					Gap: 5,
				},
			},
			{
				title:       "Cross-Axis Alignment",
				description: "Horizontal flex with center cross-axis alignment",
				layout: &tv.Layout{
					Direction:          tv.Horizontal,
					CrossAxisAlignment: tv.CrossAxisCenter,
					Items: []tv.LayoutItem{
						tv.Item(tv.FlexSize{Grow: 1}, tv.FixedSize(60)),
						tv.Item(tv.FlexSize{Grow: 1}, tv.FixedSize(100)),
						tv.Item(tv.FlexSize{Grow: 1}, tv.FixedSize(80)),
					},
					Gap: 5,
				},
			},
			{
				title:       "Main-Axis Alignment",
				description: "Fixed-size items with space-between alignment",
				layout: &tv.Layout{
					Direction:         tv.Horizontal,
					MainAxisAlignment: tv.MainAxisSpaceBetween,
					Items: []tv.LayoutItem{
						tv.Item(tv.FixedSize(120), nil),
						tv.Item(tv.FixedSize(120), nil),
						tv.Item(tv.FixedSize(120), nil),
					},
				},
			},
			{
				title:       "Space Around Alignment",
				description: "Fixed-size items with space-around alignment",
				layout: &tv.Layout{
					Direction:         tv.Horizontal,
					MainAxisAlignment: tv.MainAxisSpaceAround,
					Items: []tv.LayoutItem{
						tv.Item(tv.FixedSize(100), nil),
						tv.Item(tv.FixedSize(100), nil),
						tv.Item(tv.FixedSize(100), nil),
						tv.Item(tv.FixedSize(100), nil),
					},
				},
			},
			{
				title:       "Complex Nested Layout",
				description: "Vertical layout with horizontal flex containers",
				layout: &tv.Layout{
					Direction: tv.Vertical,
					Items: []tv.LayoutItem{
						tv.Item(nil, tv.FixedSize(80)),
						tv.Item(nil, tv.FlexSize{Grow: 1}),
						tv.Item(nil, tv.FixedSize(60)),
					},
					Gap: 5,
				},
			},
			{
				title:       "Margin and Padding Demo",
				description: "Flex items with margins and padding",
				layout: &tv.Layout{
					Direction: tv.Horizontal,
					Items: []tv.LayoutItem{
						{
							Width:   tv.FlexSize{Grow: 1},
							Margin:  tv.Uniform(5),
							Padding: tv.Uniform(3),
						},
						{
							Width:   tv.FlexSize{Grow: 2},
							Margin:  tv.HorizontalVertical(10, 5),
							Padding: tv.HorizontalVertical(5, 2),
						},
						{
							Width:   tv.FlexSize{Grow: 1},
							Margin:  tv.Spacing{Top: 2, Right: 8, Bottom: 2, Left: 8},
							Padding: tv.Uniform(4),
						},
					},
					Gap: 0, // No gap since we're using margins
				},
			},
		},
	}

	// Create main layout areas
	mainArea := tv.Rect(0, 0, physicalWidth, physicalHeight)
	mainLayout := tv.NewVerticalLayout(
		tv.Item(nil, tv.FixedSize(3)),      // Header
		tv.Item(nil, tv.FixedSize(5)),      // Title and description
		tv.Item(nil, tv.FlexSize{Grow: 1}), // Demo area
		tv.Item(nil, tv.FixedSize(3)),      // Footer
	)

	// Create frame
	f := &tv.Frame{
		Buffer:   tv.NewBuffer(physicalWidth, physicalHeight),
		Viewport: tv.FullViewport{},
		Area:     mainArea,
	}

	display := func() {
		f.Buffer.Clear()

		// Calculate main layout
		mainResult := mainLayout.Calculate(mainArea)
		headerArea := mainResult.Areas[0]
		titleArea := mainResult.Areas[1]
		demoArea := mainResult.Areas[2]
		footerArea := mainResult.Areas[3]

		// Render header
		headerText := headerStyle.Width(headerArea.Dx()).Render("Flexbox Layout System Demo")
		headerSs := styledstring.New(ansi.WcWidth, headerText)
		f.RenderComponent(headerSs, headerArea) //nolint:errcheck

		// Render title and description
		currentConfig := flexDemo.demos[flexDemo.currentDemo]
		titleText := titleStyle.Width(titleArea.Dx()).Render(
			fmt.Sprintf("[%d/%d] %s", flexDemo.currentDemo+1, len(flexDemo.demos), currentConfig.title))
		titleSs := styledstring.New(ansi.WcWidth, titleText)
		f.RenderComponent(titleSs, tv.Rect(titleArea.Min.X, titleArea.Min.Y, titleArea.Dx(), 1)) //nolint:errcheck

		descText := descStyle.Width(titleArea.Dx()).Render(currentConfig.description)
		descSs := styledstring.New(ansi.WcWidth, descText)
		f.RenderComponent(descSs, tv.Rect(titleArea.Min.X, titleArea.Min.Y+1, titleArea.Dx(), titleArea.Dy()-1)) //nolint:errcheck

		// Calculate and render the current demo layout
		demoResult := currentConfig.layout.Calculate(demoArea)

		// Handle nested layout for demo 9 (Complex Nested Layout)
		if flexDemo.currentDemo == 8 {
			// Top section
			topLayout := &tv.Layout{
				Direction: tv.Horizontal,
				Items: []tv.LayoutItem{
					tv.Item(tv.FlexSize{Grow: 1}, nil),
					tv.Item(tv.FlexSize{Grow: 2}, nil),
				},
				Gap: 5,
			}
			topResult := topLayout.Calculate(demoResult.Areas[0])

			// Middle section (main flex demo)
			middleLayout := &tv.Layout{
				Direction: tv.Horizontal,
				Items: []tv.LayoutItem{
					tv.Item(tv.FixedSize(150), nil),
					tv.Item(tv.FlexSize{Grow: 1}, nil),
					tv.Item(tv.FlexSize{Grow: 2}, nil),
					tv.Item(tv.FixedSize(100), nil),
				},
				Gap: 5,
			}
			middleResult := middleLayout.Calculate(demoResult.Areas[1])

			// Bottom section
			bottomLayout := &tv.Layout{
				Direction: tv.Horizontal,
				Items: []tv.LayoutItem{
					tv.Item(tv.FlexSize{Grow: 3}, nil),
					tv.Item(tv.FlexSize{Grow: 1}, nil),
					tv.Item(tv.FlexSize{Grow: 1}, nil),
				},
				Gap: 5,
			}
			bottomResult := bottomLayout.Calculate(demoResult.Areas[2])

			// Render all sections
			allAreas := append(topResult.Areas, append(middleResult.Areas, bottomResult.Areas...)...)
			for i, area := range allAreas {
				if i < len(itemStyles) {
					content := fmt.Sprintf("Item %d\n%dx%d", i+1, area.Dx(), area.Dy())
					itemText := itemStyles[i%len(itemStyles)].
						Width(area.Dx() - 2).
						Height(area.Dy() - 2).
						Render(content)
					itemSs := styledstring.New(ansi.WcWidth, itemText)
					f.RenderComponent(itemSs, area) //nolint:errcheck
				}
			}
		} else {
			// Regular demo rendering
			for i, area := range demoResult.Areas {
				if i < len(itemStyles) {
					var content string
					if flexDemo.currentDemo == 9 { // Margin and padding demo
						content = fmt.Sprintf("Item %d\nMargin+Padding\n%dx%d", i+1, area.Dx(), area.Dy())
					} else {
						content = fmt.Sprintf("Item %d\n%dx%d", i+1, area.Dx(), area.Dy())
					}

					itemText := itemStyles[i].
						Width(area.Dx() - 2).
						Height(area.Dy() - 2).
						Render(content)
					itemSs := styledstring.New(ansi.WcWidth, itemText)
					f.RenderComponent(itemSs, area) //nolint:errcheck
				}
			}
		}

		// Render footer
		footerText := footerStyle.Width(footerArea.Dx()).Render(
			"← → Navigate demos | q Quit | r Reset to demo 1")
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

		case tv.KeyPressEvent:
			switch {
			case ev.MatchStrings("ctrl+c", "q"):
				cancel()
			case ev.MatchStrings("right", "l", "space"):
				flexDemo.currentDemo = (flexDemo.currentDemo + 1) % len(flexDemo.demos)
			case ev.MatchStrings("left", "h"):
				flexDemo.currentDemo = (flexDemo.currentDemo - 1 + len(flexDemo.demos)) % len(flexDemo.demos)
			case ev.MatchStrings("r"):
				flexDemo.currentDemo = 0
			case ev.MatchStrings("1", "2", "3", "4", "5", "6", "7", "8", "9"):
				if demoNum := int(ev.Code - '1'); demoNum < len(flexDemo.demos) {
					flexDemo.currentDemo = demoNum
				}
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
	f, err := os.OpenFile("flexbox_demo.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}
