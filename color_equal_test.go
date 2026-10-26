package uv

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

func TestColorEqualIndexedColors(t *testing.T) {
	rgbBlack := color.RGBA{0, 0, 0, 0xff}
	tests := []struct {
		name string
		a, b color.Color
		want bool
	}{
		{"same basic color", ansi.BasicColor(1), ansi.BasicColor(1), true},
		{"same indexed color", ansi.IndexedColor(200), ansi.IndexedColor(200), true},
		{"basic and indexed with the same palette entry", ansi.BasicColor(9), ansi.IndexedColor(9), true},
		{"different palette entries with the same default RGB", ansi.IndexedColor(16), ansi.BasicColor(0), false},
		{"basic color and its default RGB", ansi.BasicColor(0), rgbBlack, false},
		{"default RGB and indexed color", rgbBlack, ansi.IndexedColor(16), false},
		{"same RGB", rgbBlack, ansi.TrueColor(0x000000), true},
		{"different RGB", rgbBlack, ansi.TrueColor(0x010101), false},
		{"nil and nil", nil, nil, true},
		{"nil and indexed color", nil, ansi.BasicColor(0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := colorEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("colorEqual(%#v, %#v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Reported in https://github.com/charmbracelet/ultraviolet/issues/205.
func TestRendererIndexedColorAgainstMatchingRGB(t *testing.T) {
	rgbBlack := color.RGBA{0, 0, 0, 0xff}
	cell := func(content string, fg color.Color) Cell {
		return Cell{Content: content, Width: 1, Style: Style{Fg: fg}}
	}
	render := func(profile colorprofile.Profile, frames ...[]Cell) []string {
		var out bytes.Buffer
		r := NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
		r.SetColorProfile(profile)
		r.Resize(len(frames[0]), 1)
		var writes []string
		for _, cells := range frames {
			buf := NewRenderBuffer(len(cells), 1)
			for x := range cells {
				buf.SetCell(x, 0, &cells[x])
			}
			r.Render(buf)
			if err := r.Flush(); err != nil {
				t.Fatal(err)
			}
			writes = append(writes, out.String())
			out.Reset()
		}
		return writes
	}

	for _, profile := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.TrueColor} {
		t.Run(profile.String(), func(t *testing.T) {
			w := render(profile, []Cell{cell("a", rgbBlack), cell("b", ansi.BasicColor(0))})
			if !bytes.Contains([]byte(w[0]), []byte("\x1b[30mb")) {
				t.Errorf("BasicColor(0) after RGB black reused the RGB SGR: %q", w[0])
			}
			w = render(profile, []Cell{cell("a", ansi.BasicColor(0))}, []Cell{cell("a", rgbBlack)})
			if w[1] == "" {
				t.Errorf("BasicColor(0) changed to RGB black was not repainted")
			}
			w = render(profile, []Cell{cell("a", ansi.IndexedColor(16))}, []Cell{cell("a", ansi.BasicColor(0))})
			if w[1] == "" {
				t.Errorf("IndexedColor(16) changed to BasicColor(0) was not repainted")
			}
		})
	}
}
