package main

import (
	"context"
	"image/color"
	"log"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"
)

const (
	width       = 96
	columnWidth = 30
)

var (
	hasDarkBG bool
	lightDark lipgloss.LightDarkFunc
)

func init() {
	hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark = lipgloss.LightDark(hasDarkBG)
}

func main() {
	// Style definitions.
	var (
		subtle    = lightDark(lipgloss.Color("#D9DCCF"), lipgloss.Color("#383838"))
		highlight = lightDark(lipgloss.Color("#874BFD"), lipgloss.Color("#7D56F4"))
		special   = lightDark(lipgloss.Color("#43BF6D"), lipgloss.Color("#73F59F"))

		divider = lipgloss.NewStyle().
			SetString("•").
			Padding(0, 1).
			Foreground(subtle).
			String()

		url = lipgloss.NewStyle().Foreground(special).Render

		activeTabBorder = lipgloss.Border{
			Top:         "─",
			Bottom:      " ",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┘",
			BottomRight: "└",
		}

		tabBorder = lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "┴",
			BottomRight: "┴",
		}

		tab = lipgloss.NewStyle().
			Border(tabBorder, true).
			BorderForeground(highlight).
			Padding(0, 1)

		activeTab = tab.Border(activeTabBorder, true)

		tabGap = tab.
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false)

		titleStyle = lipgloss.NewStyle().
				MarginLeft(1).
				MarginRight(5).
				Padding(0, 1).
				Italic(true).
				Foreground(lipgloss.Color("#FFF7DB")).
				SetString("Lip Gloss")

		descStyle = lipgloss.NewStyle().MarginTop(1)

		infoStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(subtle)

		dialogBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(1, 0).
				BorderTop(true).
				BorderLeft(true).
				BorderRight(true).
				BorderBottom(true)

		buttonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#888B7E")).
				Padding(0, 3).
				MarginTop(1)

		activeButtonStyle = buttonStyle.
					Foreground(lipgloss.Color("#FFF7DB")).
					Background(lipgloss.Color("#F25D94")).
					MarginRight(2).
					Underline(true)

		list = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(subtle).
			MarginRight(2).
			Height(8).
			Width(columnWidth + 1)

		listHeader = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(subtle).
				MarginRight(2).
				Render

		listItem = lipgloss.NewStyle().PaddingLeft(2).Render

		checkMark = lipgloss.NewStyle().SetString("✓").
				Foreground(special).
				PaddingRight(1).
				String()

		listDone = func(s string) string {
			return checkMark + lipgloss.NewStyle().
				Strikethrough(true).
				Foreground(lightDark(lipgloss.Color("#969B86"), lipgloss.Color("#696969"))).
				Render(s)
		}

		historyStyle = lipgloss.NewStyle().
				Align(lipgloss.Left).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(highlight).
				Margin(1, 3, 0, 0).
				Padding(1, 2).
				Height(19).
				Width(columnWidth)

		statusNugget = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Padding(0, 1)

		statusBarStyle = lipgloss.NewStyle().
				Foreground(lightDark(lipgloss.Color("#343433"), lipgloss.Color("#C1C6B2"))).
				Background(lightDark(lipgloss.Color("#D9DCCF"), lipgloss.Color("#353533")))

		statusStyle = lipgloss.NewStyle().
				Inherit(statusBarStyle).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#FF5F87")).
				Padding(0, 1).
				MarginRight(1)

		encodingStyle = statusNugget.
				Background(lipgloss.Color("#A550DF")).
				Align(lipgloss.Right)

		statusText = lipgloss.NewStyle().Inherit(statusBarStyle)

		fishCakeStyle = statusNugget.Background(lipgloss.Color("#6124DF"))

		docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	)

	doc := strings.Builder{}

	{
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			activeTab.Render("Lip Gloss"),
			tab.Render("Blush"),
			tab.Render("Eye Shadow"),
			tab.Render("Mascara"),
			tab.Render("Foundation"),
		)
		gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2)))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
		doc.WriteString(row + "\n\n")
	}

	{
		var (
			colors = colorGrid(1, 5)
			title  strings.Builder
		)

		for i, v := range colors {
			const offset = 2
			c := lipgloss.Color(v[0])
			_, _ = title.WriteString(titleStyle.MarginLeft(i * offset).Background(c).String())
			if i < len(colors)-1 {
				title.WriteRune('\n')
			}
		}

		desc := lipgloss.JoinVertical(lipgloss.Left,
			descStyle.Render("Style Definitions for Nice Terminal Layouts"),
			infoStyle.Render("From Charm"+divider+url("https://github.com/charmbracelet/lipgloss")),
		)

		row := lipgloss.JoinHorizontal(lipgloss.Top, title.String(), desc)
		doc.WriteString(row + "\n\n")
	}

	okButton := activeButtonStyle.Render("Yes")
	cancelButton := buttonStyle.Render("Maybe")

	grad := applyGradient(
		lipgloss.NewStyle(),
		"Are you sure you want to eat marmalade?",
		lipgloss.Color("#EDFF82"),
		lipgloss.Color("#F25D94"),
	)

	question := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Render(grad)

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	dialogUI := lipgloss.JoinVertical(lipgloss.Center, question, buttons)

	dialog := lipgloss.Place(width, 9,
		lipgloss.Center, lipgloss.Center,
		"",
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(subtle)),
	)

	doc.WriteString(dialog + "\n\n")

	colors := func() string {
		colors := colorGrid(14, 8)

		b := strings.Builder{}
		for _, x := range colors {
			for _, y := range x {
				s := lipgloss.NewStyle().SetString("  ").Background(lipgloss.Color(y))
				b.WriteString(s.String())
			}
			b.WriteRune('\n')
		}

		return b.String()
	}()

	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		list.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				listHeader("Citrus Fruits to Try"),
				listDone("Grapefruit"),
				listDone("Yuzu"),
				listItem("Citron"),
				listItem("Kumquat"),
				listItem("Pomelo"),
			),
		),
		list.Width(columnWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				listHeader("Actual Lip Gloss Vendors"),
				listItem("Glossier"),
				listItem("Claire's Boutique"),
				listDone("Nyx"),
				listItem("Mac"),
				listDone("Milk"),
			),
		),
	)

	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, lists, colors))

	{
		const (
			historyA = "The Romans learned from the Greeks that quinces slowly cooked with honey would \"set\" when cool. The Apicius gives a recipe for preserving whole quinces, stems and leaves attached, in a bath of honey diluted with defrutum: Roman marmalade. Preserves of quince and lemon appear (along with rose, apple, plum and pear) in the Book of ceremonies of the Byzantine Emperor Constantine VII Porphyrogennetos."
			historyB = "Medieval quince preserves, which went by the French name cotignac, produced in a clear version and a fruit pulp version, began to lose their medieval seasoning of spices in the 16th century. In the 17th century, La Varenne provided recipes for both thick and clear cotignac."
			historyC = "In 1524, Henry VIII, King of England, received a \"box of marmalade\" from Mr. Hull of Exeter. This was probably marmelada, a solid quince paste from Portugal, still made and sold in southern Europe today. It became a favourite treat of Anne Boleyn and her ladies in waiting."
		)

		doc.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			historyStyle.Align(lipgloss.Right).Render(historyA),
			historyStyle.Align(lipgloss.Center).Render(historyB),
			historyStyle.MarginRight(0).Render(historyC),
		))

		doc.WriteString("\n\n")
	}

	{
		w := lipgloss.Width

		lightDarkState := "Light"
		if hasDarkBG {
			lightDarkState = "Dark"
		}

		statusKey := statusStyle.Render("STATUS")
		encoding := encodingStyle.Render("UTF-8")
		fishCake := fishCakeStyle.Render("🍥 Fish Cake")
		statusVal := statusText.
			Width(width - w(statusKey) - w(encoding) - w(fishCake)).
			Render("Ravishingly " + lightDarkState + "!")

		bar := lipgloss.JoinHorizontal(lipgloss.Top,
			statusKey,
			statusVal,
			encoding,
			fishCake,
		)

		doc.WriteString(statusBarStyle.Width(width).Render(bar))
	}

	t := uv.DefaultTerminal(nil)
	scr := t.Screen()

	if err := t.Start(); err != nil {
		log.Fatalf("starting program: %v", err)
	}

	defer t.Stop()

	t.SetLogger(log.Default())

	physicalWidth := scr.Bounds().Dx()

	if physicalWidth > 0 {
		docStyle = docStyle.MaxWidth(physicalWidth)
	}

	scr.EnterAltScreen()
	scr.WriteString(ansi.SetMode(ansi.ModeMouseButtonEvent, ansi.ModeMouseExtSgr))

	dialogWidth := lipgloss.Width(dialogUI) + dialogBoxStyle.GetHorizontalFrameSize()
	dialogHeight := lipgloss.Height(dialogUI) + dialogBoxStyle.GetVerticalFrameSize()
	dialogX, dialogY := physicalWidth/2-dialogWidth/2+docStyle.GetVerticalFrameSize()-1, 12
	mainDoc := docStyle.Render(doc.String())

	display := func() {
		screen.Clear(scr)
		mainSs := uv.NewStyledString(mainDoc)
		mainSs.Draw(scr, scr.Bounds())
		boxArea := uv.Rect(dialogX, dialogY, dialogWidth, dialogHeight)
		box := uv.NewStyledString(dialogBoxStyle.Render(dialogUI))
		box.Draw(scr, boxArea)
		scr.Render()
		scr.Flush()
	}

	display()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case ev := <-t.Events():
			log.Printf("event: %T", ev)
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				scr.Resize(ev.Width, ev.Height)
			case uv.MouseClickEvent:
				dialogX, dialogY = ev.X-dialogWidth/2, ev.Y-dialogHeight/2
			case uv.KeyPressEvent:
				log.Printf("%T %q %q", ev, ev.String(), ev.Keystroke())
				switch {
				case ev.MatchString("ctrl+c", "q"):
					cancel()
				case ev.MatchString("left", "h"):
					dialogX--
				case ev.MatchString("down", "j"):
					dialogY++
				case ev.MatchString("up", "k"):
					dialogY--
				case ev.MatchString("right", "l"):
					dialogX++
				}
			}

			display()
		}
	}

	scr.WriteString(ansi.ResetMode(ansi.ModeMouseButtonEvent, ansi.ModeMouseExtSgr))
}

func colorGrid(xSteps, ySteps int) [][]string {
	x0y0, _ := colorful.Hex("#F25D94")
	x1y0, _ := colorful.Hex("#EDFF82")
	x0y1, _ := colorful.Hex("#643AFF")
	x1y1, _ := colorful.Hex("#14F9D5")

	x0 := make([]colorful.Color, ySteps)
	for i := range x0 {
		x0[i] = x0y0.BlendLuv(x0y1, float64(i)/float64(ySteps))
	}

	x1 := make([]colorful.Color, ySteps)
	for i := range x1 {
		x1[i] = x1y0.BlendLuv(x1y1, float64(i)/float64(ySteps))
	}

	grid := make([][]string, ySteps)
	for x := 0; x < ySteps; x++ {
		y0 := x0[x]
		grid[x] = make([]string, xSteps)
		for y := 0; y < xSteps; y++ {
			grid[x][y] = y0.BlendLuv(x1[x], float64(y)/float64(xSteps)).Hex()
		}
	}

	return grid
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func applyGradient(base lipgloss.Style, input string, from, to color.Color) string {
	g := uniseg.NewGraphemes(input)
	var chars []string
	for g.Next() {
		chars = append(chars, g.Str())
	}

	a, _ := colorful.MakeColor(to)
	b, _ := colorful.MakeColor(from)
	var output strings.Builder
	var hex string
	for i := 0; i < len(chars); i++ {
		hex = a.BlendLuv(b, float64(i)/float64(len(chars)-1)).Hex()
		output.WriteString(base.Foreground(lipgloss.Color(hex)).Render(chars[i]))
	}

	return output.String()
}

func init() {
	f, err := os.OpenFile("layout.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}
