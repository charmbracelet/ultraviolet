package uv

import (
	"testing"
)

// BenchmarkBufferAllocation measures the cost of creating a new buffer
func BenchmarkBufferAllocation(b *testing.B) {
	b.Run("80x24", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewBuffer(80, 24)
		}
	})

	b.Run("120x40", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewBuffer(120, 40)
		}
	})

	b.Run("200x60", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewBuffer(200, 60)
		}
	})
}

// BenchmarkBufferSparseWrite measures the cost of writing to a few cells
func BenchmarkBufferSparseWrite(b *testing.B) {
	buf := NewBuffer(80, 24)
	cell := Cell{Content: "A", Width: 1}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.SetCell(i%80, (i/80)%24, &cell)
	}
}

// BenchmarkBufferRender measures the cost of rendering a sparse buffer
func BenchmarkBufferRender(b *testing.B) {
	b.Run("empty", func(b *testing.B) {
		buf := NewBuffer(80, 24)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buf.Render()
		}
	})

	b.Run("sparse-10%", func(b *testing.B) {
		buf := NewBuffer(80, 24)
		cell := Cell{Content: "X", Width: 1}
		// Fill 10% of the buffer
		for i := 0; i < 80*24/10; i++ {
			buf.SetCell(i%80, i/80, &cell)
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buf.Render()
		}
	})

	b.Run("full", func(b *testing.B) {
		buf := NewBuffer(80, 24)
		cell := Cell{Content: "X", Width: 1}
		// Fill entire buffer
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				buf.SetCell(x, y, &cell)
			}
		}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = buf.Render()
		}
	})
}

func BenchmarkBufferFrameUpdates(b *testing.B) {
	buf := NewScreenBuffer(80, 24)
	frames := []*StyledString{
		NewStyledString("Hello, World!\nThis is frame one."),
		NewStyledString("Hello, Universe!\n\x1b[31;1mThis is frame two.\x1b[0m"),
		NewStyledString(" \x1b[42;2mGoodbye!\x1b[0m\nFinal frame here."),
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		frame := frames[i%len(frames)]
		buf.Clear()
		frame.Draw(buf, buf.Bounds())
		_ = buf.Render()
	}
}
