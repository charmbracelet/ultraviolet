package uv_test

import (
	"bytes"
	"image"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/layout"
)

// FuzzBufferAPI exercises the public buffer, tab-stop, layout and renderer
// APIs with arbitrary (including out-of-range and negative) arguments. None
// of these need a terminal or the ghostty oracle, so the target runs
// everywhere the root module does, including short local runs and CI.
func FuzzBufferAPI(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00})
	f.Fuzz(func(t *testing.T, b []byte) {
		next := func(n int) int {
			v := 0
			for i := 0; i < n && len(b) > 0; i++ {
				v = v<<8 | int(b[0])
				b = b[1:]
			}
			return v
		}

		// Buffer operations with unchecked dimensions and areas.
		buf := uv.NewBuffer(next(2)%64, next(2)%32)
		buf.Resize(next(2)%64-32, next(2)%32-16)
		area := image.Rect(next(2)%96-48, next(2)%96-48, next(2)%96-48, next(2)%96-48)
		if area.Dx() < 0 {
			area.Min, area.Max = area.Max, area.Min
		}
		if area.Dy() < 0 {
			area.Min, area.Max = area.Max, area.Min
		}
		buf.InsertLineArea(next(2)%96-48, next(2)%32-16, nil, area)
		buf.DeleteLineArea(next(2)%96-48, next(2)%32-16, nil, area)
		buf.InsertCellArea(next(2)%96-48, next(2)%96-48, next(2)%16-8, nil, area)
		buf.FillArea(nil, area)
		buf.ClearArea(area)

		// Tab stops with arbitrary width and interval.
		ts := uv.NewTabStops(next(2)%128-64, next(1)%64-32)
		for col := next(1)%128 - 64; col < 128; col++ {
			ts.IsStop(col)
		}

		// Split must be deterministic: mutating one result may not change
		// the next one.
		var l layout.Layout
		l = l.WithDirection(layout.DirectionHorizontal)
		l = l.WithConstraints(layout.Len(1), layout.Len(1), layout.Len(1))
		r := image.Rect(0, 0, 1+next(2)%64, 1+next(2)%64)
		s1 := l.Split(r)
		s2 := l.Split(r)
		if len(s1) == len(s2) {
			for i := range s1 {
				if s1[i] != s2[i] {
					t.Fatalf("Split result changed between calls: %v vs %v", s1[i], s2[i])
				}
			}
		}

		// A renderer narrower than a prepended string must not divide by zero.
		var out bytes.Buffer
		rd := uv.NewTerminalRenderer(&out, []string{"TERM=xterm-256color"})
		rd.Resize(next(1)%8, 1+next(1)%8)
		cb := uv.NewRenderBuffer(next(1)%8, 1+next(1)%8)
		rd.PrependString(cb, "hi")
	})
}
