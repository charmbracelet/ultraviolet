package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
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
	scr := t.Screen()
	scr.EnterAltScreen()

	if err := t.Start(); err != nil {
		log.Fatalf("failed to start terminal: %v", err)
	}

	defer t.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var frameCount int
	var elapsed time.Duration
	now := time.Now()
	fps := 60.0
	fpsFrameCount := 0

	bounds := scr.Bounds()
	colors := setupColors(bounds.Dx(), bounds.Dy())

	go t.SendEvent(tickEvent{})

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.KeyPressEvent:
				switch ev.String() {
				case "q", "ctrl+c":
					cancel()
					break LOOP
				}

			case uv.WindowSizeEvent:
				width, height := ev.Width, ev.Height
				colors = setupColors(width, height)
				scr.Resize(width, height)
			case tickEvent:
				if len(colors) == 0 {
					continue
				}

				frameCount++
				fpsFrameCount++
				screen.Clear(scr)

				bounds := scr.Bounds()

				// Title
				uv.NewStyledString(fmt.Sprintf("\x1b[1mSpace / FPS: %.1f\x1b[m", fps)).
					Draw(scr, uv.Rect(0, 0, bounds.Dx(), 1))

				// Color display
				width, height := bounds.Dx(), bounds.Dy()
				for y := 1; y < height; y++ {
					for x := 0; x < width; x++ {
						xi := (x + frameCount) % width
						fg := colors[y*2][xi]
						bg := colors[y*2+1][xi]
						st := uv.Style{
							Fg: fg,
							Bg: bg,
						}
						scr.SetCell(x, y, &uv.Cell{
							Content: "▀",
							Style:   st,
							Width:   1,
						})
					}
				}

				scr.Render()
				scr.Flush()

				elapsed = time.Since(now)
				if elapsed > time.Second && fpsFrameCount > 2 {
					fps = float64(fpsFrameCount) / elapsed.Seconds()
					now = time.Now()
					fpsFrameCount = 0
				}

				go t.SendEvent(tickEvent{})
			}
		}
	}
}
