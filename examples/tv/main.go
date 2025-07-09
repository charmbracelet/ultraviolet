package main

import (
	"context"
	"image/color"
	"log"

	uv "github.com/charmbracelet/ultraviolet"
)

var (
	gray       = color.RGBA{R: 104, G: 104, B: 104, A: 255} // #686868
	white      = color.RGBA{R: 180, G: 180, B: 180, A: 255} // #b4b4b4
	yellow     = color.RGBA{R: 180, G: 180, B: 16, A: 255}  // #b4b416
	cyan       = color.RGBA{R: 16, G: 180, B: 180, A: 255}  // #10b4b4
	green      = color.RGBA{R: 16, G: 180, B: 16, A: 255}   // #10b410
	magenta    = color.RGBA{R: 180, G: 16, B: 180, A: 255}  // #b410b4
	red        = color.RGBA{R: 180, G: 16, B: 16, A: 255}   // #b41010
	blue       = color.RGBA{R: 16, G: 16, B: 180, A: 255}   // #1010b4
	black      = color.RGBA{R: 16, G: 16, B: 16, A: 255}    // #101010
	fullWhite  = color.RGBA{R: 235, G: 235, B: 235, A: 255} // #ebebeb
	fullBlack  = color.RGBA{R: 0, G: 0, B: 0, A: 255}       // #000000
	lightBlack = color.RGBA{R: 26, G: 26, B: 26, A: 255}    // #1a1a1a
	purple     = color.RGBA{R: 72, G: 16, B: 116, A: 255}   // #481074
	brown      = color.RGBA{R: 106, G: 52, B: 16, A: 255}   // #6a3410
	navy       = color.RGBA{R: 16, G: 70, B: 106, A: 255}   // #10466a
)

func main() {
	t := uv.DefaultTerminal()
	if err := t.MakeRaw(); err != nil {
		log.Fatalf("Error making terminal raw: %v", err)
	}
	if err := t.Start(); err != nil {
		log.Fatalf("Error starting terminal: %v", err)
	}

	var err error
	var area uv.Rectangle
	area.Max.X, area.Max.Y, err = t.GetSize()
	if err != nil {
		log.Fatalf("Error getting terminal size: %v", err)
	}

	t.EnterAltScreen()

	ctx, cancel := context.WithCancel(context.Background())

	const barCount = 7
	const botBarCount = 6
	rowColors := [3][]color.RGBA{
		0: {
			0: white,
			1: yellow,
			2: cyan,
			3: green,
			4: magenta,
			5: red,
			6: blue,
		},
		1: {
			0: blue,
			1: black,
			2: magenta,
			3: black,
			4: cyan,
			5: black,
			6: white,
		},
		2: {
			0: navy,
			1: fullWhite,
			2: purple,
			3: black,
			4: black,
			5: black,
		},
	}

	display := func() {
		t.Clear()
		topRow := uv.Rect(0, 0, area.Max.X, (area.Max.Y*66)/100)
		midRow := uv.Rect(0, topRow.Max.Y, area.Max.X, (area.Max.Y*8)/100)
		botRow := uv.Rect(0, midRow.Max.Y, area.Max.X, (area.Max.Y*26)/100)

		barWidth := topRow.Max.X / barCount
		for i, row := range []uv.Rectangle{
			topRow, midRow,
		} {
			for j := 0; j < barCount; j++ {
				bar := uv.Rect(j*barWidth, row.Min.Y, (j+1)*barWidth, row.Max.Y)
				cell := uv.EmptyCell
				cell.Style.Bg = rowColors[i][j%len(rowColors[i])]
				t.FillArea(&cell, bar)
			}
		}

		botBarWidth := botRow.Max.X / botBarCount
		for i := 0; i < botBarCount; i++ {
			bar := uv.Rect(i*botBarWidth, botRow.Min.Y, (i+1)*botBarWidth, botRow.Max.Y)
			cell := uv.EmptyCell
			cell.Style.Bg = rowColors[2][i%len(rowColors[2])]
			t.FillArea(&cell, bar)
		}

		// Special case for the before last bar
		const specialRow = 5
		subBarWidth := barWidth / 3
		for i := 0; i < 3; i++ {
			bar := uv.Rect(specialRow*barWidth+i*subBarWidth, botRow.Min.Y, subBarWidth, botRow.Max.Y)
			cell := uv.EmptyCell
			switch i {
			case 0:
				cell.Style.Bg = fullBlack
			case 1:
				continue
			case 2:
				cell.Style.Bg = lightBlack
			}
			t.FillArea(&cell, bar)
		}

		if err := t.Display(); err != nil {
			log.Fatalf("Error displaying terminal: %v", err)
		}
	}

	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case uv.WindowSizeEvent:
			area.Max.X, area.Max.Y = ev.Width, ev.Height
			t.Resize(area.Max.X, area.Max.Y)
			t.Erase()
			display()
		case uv.KeyPressEvent:
			switch {
			case ev.MatchStrings("q", "ctrl+c"):
				cancel()
			}
		}
	}

	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down terminal: %v", err)
	}
}
