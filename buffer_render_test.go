package uv

import (
	"image/color"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// BenchmarkLineRender renders a line of colored words with trailing blanks,
// as a terminal's scrolled-off line is rendered for its scrollback.
func BenchmarkLineRender(b *testing.B) {
	line := make(Line, 120)
	for x := range line {
		line[x] = EmptyCell
	}

	styles := []Style{
		{Fg: ansi.BasicColor(3)},
		{Fg: ansi.BasicColor(3), Attrs: AttrBold},
		{Bg: ansi.IndexedColor(200)},
		{Fg: color.RGBA{R: 10, G: 20, B: 30, A: 0xff}},
	}
	for x := range 80 {
		line[x] = Cell{Content: string(rune('a' + x%26)), Width: 1, Style: styles[x/20]}
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = line.Render()
	}
}
