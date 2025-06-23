package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/uv/screen"
)

func setupColors(width, height int) [][]color.Color {
	height = height * 2 // double height for half blocks
	colors := make([][]color.Color, height)

	for y := 0; y < height; y++ {
		colors[y] = make([]color.Color, width)
		randomnessFactor := float64(height-y) / float64(height)

		for x := 0; x < width; x++ {
			baseValue := randomnessFactor * (float64(height-y) / float64(height))
			randomOffset := (rand.Float64() * 0.2) - 0.1
			value := clamp(baseValue+randomOffset, 0, 1)

			// Convert value to grayscale color (0-255)
			gray := uint8(value * 255)
			colors[y][x] = color.RGBA{
				R: gray, G: gray, B: gray, A: 255,
			}
		}
	}
	return colors
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

type tickEvent struct{}

func main() {
	t := uv.DefaultTerminal()
	if err := t.MakeRaw(); err != nil {
		log.Fatalf("failed to make terminal raw: %v", err)
	}

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	var err error
	var area uv.Rectangle
	area.Max.X, area.Max.Y, err = t.GetSize()
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

	t.EnterAltScreen()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var frameCount int
	var elapsed time.Duration
	now := time.Now()
	fps := 60.0
	fpsFrameCount := 0

	var lastWidth, lastHeight int
	colors := setupColors(area.Dx(), area.Dy())

	go t.SendEvent(ctx, tickEvent{})

LOOP:
	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case uv.KeyPressEvent:
			switch ev.String() {
			case "q", "ctrl+c":
				cancel()
				break LOOP
			}

		case uv.WindowSizeEvent:
			area.Max.X = ev.Width
			area.Max.Y = ev.Height
			if width, height := area.Dx(), area.Dy(); width != lastWidth || height != lastHeight {
				colors = setupColors(width, height)
				lastWidth = width
				lastHeight = height
				t.Resize(area.Dx(), area.Dy())
				t.Erase()
			}
		case tickEvent:
			if len(colors) == 0 {
				continue
			}

			frameCount++
			fpsFrameCount++
			screen.Clear(t)

			// Title
			uv.NewStyledString(fmt.Sprintf("\x1b[1mSpace / FPS: %.1f\x1b[m", fps)).
				Draw(t, uv.Rect(0, 0, area.Dx(), 1))

			// Color display
			width, height := area.Dx(), area.Dy() // Reserve one line for the title
			for y := 1; y < height; y++ {
				for x := 0; x < width; x++ {
					xi := (x + frameCount) % width
					fg := colors[y*2][xi]
					bg := colors[y*2+1][xi]
					st := uv.Style{
						Fg: fg,
						Bg: bg,
					}
					t.SetCell(x, y, &uv.Cell{
						Content: "▀",
						Style:   st,
						Width:   1,
					})
				}
			}

			t.Display()
			elapsed = time.Since(now)
			if elapsed > time.Second && fpsFrameCount > 2 {
				fps = float64(fpsFrameCount) / elapsed.Seconds()
				now = time.Now()
				fpsFrameCount = 0
			}

			go t.SendEvent(ctx, tickEvent{})
		}
	}

	if err := t.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown terminal: %v", err)
	}
}
