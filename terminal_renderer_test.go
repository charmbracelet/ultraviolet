package uv

import (
	"bytes"
	"testing"
)

func TestSimpleRendererOutput(t *testing.T) {
	const w, h = 5, 3
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color", // This will enable 256 colors for the renderer
		"COLORTERM=truecolor", // This will enable true color support for the renderer
	})

	r.EnterAltScreen()

	// r.SetTabStops(5) // Use tab character \t for cursor movements.
	// r.SetBackspace(true) // Use backspace character \b for cursor movements.
	// r.SetMapNewline(true) // Map newline characters to \r\n for proper line endings.
	r.Resize(w, h)

	cellbuf := NewBuffer(5, 3)
	// 'X', ' ', ' ', ' ', ' '
	// ' ', 'X', ' ', ' ', ' '
	// ' ', ' ', 'X', ' ', ' '

	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)
	cellbuf.SetCell(1, 1, &cell)
	cellbuf.SetCell(2, 2, &cell)

	r.Render(cellbuf)
	r.ExitAltScreen()
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	expected := "\x1b[?25l\x1b[?1049h\x1b[H\x1b[2JX\nX\nX\x1b[?1049l\x1b[?25h"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\nbut got:\n%q", expected, buf.String())
	}
}

func TestInlineRendererOutput(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color", // This will enable 256 colors for the renderer
		"COLORTERM=truecolor", // This will enable true color support for the renderer
	})

	r.SetRelativeCursor(true) // Use relative cursor movements.

	const physicalWidth, physicalHeight = 80, 24 // Terminal width
	const width, height = 80, 3                  // Application width
	r.Resize(physicalWidth, physicalHeight)
	cellbuf := NewBuffer(physicalWidth, height)

	for i, r := range "Hello, World!" {
		cell := Cell{Content: string(r), Width: 1}
		cellbuf.SetCell(i, 0, &cell)
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	expected := "\x1b[?25l\rHello, World!\r\n\n\x1b[?25h"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\nbut got:\n%q", expected, buf.String())
	}
}
