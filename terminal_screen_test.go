package uv

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestRenderBufferFillAreaMarksTouched(t *testing.T) {
	b := NewRenderBuffer(5, 3)
	cell := &Cell{Content: "X", Width: 1}

	b.FillArea(cell, Rect(1, 1, 2, 1))

	if got := b.CellAt(1, 1); got == nil || got.Content != "X" {
		t.Fatalf("expected filled cell at 1,1, got %#v", got)
	}
	if got := b.CellAt(2, 1); got == nil || got.Content != "X" {
		t.Fatalf("expected filled cell at 2,1, got %#v", got)
	}
	line := b.Touched[1]
	if line == nil || line.FirstCell != 1 || line.LastCell != 3 {
		t.Fatalf("expected touched span [1,3) on line 1, got %#v", line)
	}
}

func TestTerminalScreenSetCellWritesThroughRenderBuffer(t *testing.T) {
	var out bytes.Buffer
	s := NewTerminalScreen(&out, Environ{"TERM=xterm-256color"})
	if err := s.Resize(4, 2); err != nil {
		t.Fatalf("resize: %v", err)
	}

	cell := &Cell{Content: "A", Width: 1}
	s.SetCell(2, 1, cell)

	if got := s.rbuf.CellAt(2, 1); got == nil || got.Content != "A" {
		t.Fatalf("expected render buffer cell A at 2,1, got %#v", got)
	}
	line := s.rbuf.Touched[1]
	if line == nil || line.FirstCell != 2 || line.LastCell != 3 {
		t.Fatalf("expected touched span [2,3) on line 1, got %#v", line)
	}
}

func TestTerminalScreenRenderUsesRenderBufferDirectly(t *testing.T) {
	var out bytes.Buffer
	s := NewTerminalScreen(&out, Environ{"TERM=xterm-256color"})
	if err := s.Resize(3, 1); err != nil {
		t.Fatalf("resize: %v", err)
	}

	s.rbuf.SetCell(0, 0, &Cell{Content: "A", Width: 1})
	s.win.SetCell(0, 0, &Cell{Content: "B", Width: 1})

	if err := s.Render(); err != nil {
		t.Fatalf("render: %v", err)
	}

	if got := s.rend.curbuf.CellAt(0, 0); got == nil || got.Content != "A" {
		t.Fatalf("expected renderer current buffer to use render buffer cell A, got %#v", got)
	}
	if bytes.Contains(out.Bytes(), []byte("B")) {
		t.Fatalf("expected output to avoid stale window cell B, got %q", out.String())
	}
}

func TestTerminalScreenDisplayClearsRenderBuffer(t *testing.T) {
	var out bytes.Buffer
	s := NewTerminalScreen(&out, Environ{"TERM=xterm-256color"})
	if err := s.Resize(3, 1); err != nil {
		t.Fatalf("resize: %v", err)
	}

	first := DrawableFunc(func(scr Screen, _ Rectangle) {
		scr.SetCell(0, 0, &Cell{Content: "A", Width: 1})
		scr.SetCell(1, 0, &Cell{Content: "B", Width: 1})
	})
	second := DrawableFunc(func(scr Screen, _ Rectangle) {
		scr.SetCell(0, 0, &Cell{Content: "X", Width: 1})
	})

	if err := s.Display(first); err != nil {
		t.Fatalf("first display: %v", err)
	}
	if err := s.Display(second); err != nil {
		t.Fatalf("second display: %v", err)
	}

	if got := s.rbuf.CellAt(1, 0); got == nil || got.Content != " " {
		t.Fatalf("expected second display clear to blank stale cell at 1,0, got %#v", got)
	}
	if got := s.rbuf.CellAt(0, 0); got == nil || got.Content != "X" {
		t.Fatalf("expected second display cell X at 0,0, got %#v", got)
	}
}

func TestStyledStringDrawUpdatesTerminalScreenBuffers(t *testing.T) {
	var out bytes.Buffer
	s := NewTerminalScreen(&out, Environ{"TERM=xterm-256color"})
	if err := s.Resize(4, 1); err != nil {
		t.Fatalf("resize: %v", err)
	}

	s.SetCell(3, 0, &Cell{Content: "Z", Width: 1})
	NewStyledString("AB").Draw(s, s.Bounds())

	for x, want := range []string{"A", "B", " ", " "} {
		if got := s.win.CellAt(x, 0); got == nil || got.Content != want {
			t.Fatalf("expected window cell %q at %d,0, got %#v", want, x, got)
		}
		if got := s.rbuf.CellAt(x, 0); got == nil || got.Content != want {
			t.Fatalf("expected render buffer cell %q at %d,0, got %#v", want, x, got)
		}
	}
}

func BenchmarkTerminalScreenRenderSparseUpdates(b *testing.B) {
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})
	if err := s.Resize(80, 24); err != nil {
		b.Fatalf("resize: %v", err)
	}

	cell := &Cell{Content: "X", Width: 1}
	activeX, activeY := 0, 0
	s.SetCell(activeX, activeY, cell)
	if err := s.Render(); err != nil {
		b.Fatalf("prime render: %v", err)
	}
	if err := s.Flush(); err != nil {
		b.Fatalf("prime flush: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nextX := (i + 1) % s.Bounds().Dx()
		nextY := ((i + 1) / s.Bounds().Dx()) % s.Bounds().Dy()

		s.SetCell(activeX, activeY, nil)
		s.SetCell(nextX, nextY, cell)

		if err := s.Render(); err != nil {
			b.Fatalf("render: %v", err)
		}
		if err := s.Flush(); err != nil {
			b.Fatalf("flush: %v", err)
		}

		activeX, activeY = nextX, nextY
	}
}

func BenchmarkStyledStringDrawScreenBuffer(b *testing.B) {
	const width, height = 80, 24

	ss := NewStyledString(benchmarkStyledFrame(width, height))
	scr := NewScreenBuffer(width, height)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ss.Draw(scr, scr.Bounds())
	}
}

func benchmarkStyledFrame(width, height int) string {
	var out strings.Builder
	out.Grow(height * (width + 16))

	const pattern = "abcdefghijklmnopqrstuvwxyz0123456789<>[]{}+-=*/"
	for y := 0; y < height; y++ {
		out.WriteString("\x1b[3")
		out.WriteString(strconv.Itoa((y % 6) + 1))
		out.WriteByte('m')

		var line strings.Builder
		line.Grow(width)
		line.WriteString("row ")
		if y < 10 {
			line.WriteByte('0')
		}
		line.WriteString(strconv.Itoa(y))
		line.WriteByte(' ')
		for line.Len() < width {
			line.WriteString(pattern)
		}

		out.WriteString(line.String()[:width])
		out.WriteString("\x1b[m")
		if y < height-1 {
			out.WriteByte('\n')
		}
	}

	return out.String()
}
