package uv

import (
	"bytes"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestSimpleRendererOutput(t *testing.T) {
	const w, h = 5, 3
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color", // This will enable 256 colors for the renderer
		"COLORTERM=truecolor", // This will enable true color support for the renderer
	})

	r.SetFullscreen(true)
	r.SetRelativeCursor(false)
	r.SaveCursor()
	r.Erase()

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
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	expected := "\x1b[H\x1b[2JX\nX\nX"
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

	expected := "\rHello, World!\r\n\n"
	if buf.String() != expected {
		t.Errorf("expected output:\n%q\nbut got:\n%q", expected, buf.String())
	}
}

func TestRendererColorProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  colorprofile.Profile
		env      []string
		expected colorprofile.Profile
	}{
		{
			name:     "truecolor",
			profile:  colorprofile.TrueColor,
			env:      []string{"COLORTERM=truecolor"},
			expected: colorprofile.TrueColor,
		},
		{
			name:     "256 color",
			profile:  colorprofile.ANSI256,
			env:      []string{"TERM=xterm-256color"},
			expected: colorprofile.ANSI256,
		},
		{
			name:     "16 color",
			profile:  colorprofile.ANSI,
			env:      []string{"TERM=xterm"},
			expected: colorprofile.ANSI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewTerminalRenderer(&buf, tt.env)
			r.SetColorProfile(tt.profile)

			// Test that the profile was set correctly by rendering a colored cell
			cellbuf := NewBuffer(1, 1)
			cell := Cell{
				Content: "X",
				Width:   1,
				Style:   Style{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
			}
			cellbuf.SetCell(0, 0, &cell)

			r.Render(cellbuf)
			if err := r.Flush(); err != nil {
				t.Fatalf("failed to flush renderer: %v", err)
			}

			// The output should contain color sequences appropriate for the profile
			output := buf.String()
			if !strings.Contains(output, "X") {
				t.Errorf("expected output to contain 'X', got: %q", output)
			}
		})
	}
}

func TestRendererPosition(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Initial position should be -1, -1
	x, y := r.Position()
	if x != -1 || y != -1 {
		t.Errorf("expected initial position (-1, -1), got (%d, %d)", x, y)
	}

	// Set position
	r.SetPosition(5, 10)
	x, y = r.Position()
	if x != 5 || y != 10 {
		t.Errorf("expected position (5, 10), got (%d, %d)", x, y)
	}
}

func TestRendererMoveTo(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.MoveTo(5, 3)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Should contain cursor positioning sequence
	output := buf.String()
	if !strings.Contains(output, "\x1b[") {
		t.Errorf("expected cursor positioning sequence in output: %q", output)
	}
}

func TestRendererWriteString(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	n, err := r.WriteString("Hello, World!")
	if err != nil {
		t.Fatalf("failed to write string: %v", err)
	}

	if n != 13 {
		t.Errorf("expected to write 13 bytes, wrote %d", n)
	}

	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got: %q", output)
	}
}

func TestRendererWrite(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	data := []byte("Hello, World!")
	n, err := r.Write(data)
	if err != nil {
		t.Fatalf("failed to write bytes: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
	}

	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hello, World!") {
		t.Errorf("expected output to contain 'Hello, World!', got: %q", output)
	}
}

func TestRendererRedraw(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(3, 1)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)

	// First render
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	firstOutput := buf.String()
	buf.Reset()

	// Redraw should force a full redraw
	r.Redraw(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	secondOutput := buf.String()

	// Both outputs should contain the cell content
	if !strings.Contains(firstOutput, "X") {
		t.Errorf("expected first output to contain 'X', got: %q", firstOutput)
	}
	if !strings.Contains(secondOutput, "X") {
		t.Errorf("expected second output to contain 'X', got: %q", secondOutput)
	}
}

func TestRendererErase(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(3, 1)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)

	// Mark for erase
	r.Erase()

	// Render should perform a full clear
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}
}

func TestRendererResize(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Test resize
	r.Resize(80, 24)

	// Should not crash and should handle the resize
	cellbuf := NewBuffer(80, 24)
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}
}

func TestRendererPrependString(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.Resize(10, 5)
	cellbuf := NewBuffer(10, 5)

	r.PrependString(cellbuf, "Prepended line")
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Prepended line") {
		t.Errorf("expected output to contain 'Prepended line', got: %q", output)
	}
}

func TestRendererPrependLines(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.Resize(10, 5)
	cellbuf := NewBuffer(10, 5)

	// Create a line to prepend
	line := make(Line, 5)
	for i, ch := range "Hello" {
		line[i] = Cell{Content: string(ch), Width: 1}
	}

	r.PrependString(cellbuf, line.Render())
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Hello") {
		t.Errorf("expected output to contain 'Hello', got: %q", output)
	}
}

func TestRendererCapabilities(t *testing.T) {
	tests := []struct {
		name string
		term string
		test func(*testing.T, *TerminalRenderer)
	}{
		{
			name: "xterm capabilities",
			term: "xterm-256color",
			test: func(t *testing.T, r *TerminalRenderer) {
				// xterm should support all capabilities
				if !r.caps.Contains(capCHA) {
					t.Error("expected xterm to support CHA")
				}
				// NOTE: We have disabled HPA for xterm due to some terminals
				// not supporting it correctly i.e. Konsole.
				// if !r.caps.Contains(capHPA) {
				// 	t.Error("expected xterm to support HPA")
				// }
			},
		},
		{
			name: "linux terminal capabilities",
			term: "linux",
			test: func(t *testing.T, r *TerminalRenderer) {
				// linux terminal has limited capabilities
				if !r.caps.Contains(capVPA) {
					t.Error("expected linux to support VPA")
				}
				if !r.caps.Contains(capHPA) {
					t.Error("expected linux to support HPA")
				}
				if r.caps.Contains(capREP) {
					t.Error("expected linux to not support REP")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewTerminalRenderer(&buf, []string{"TERM=" + tt.term})
			tt.test(t, r)
		})
	}
}

func TestRendererTabStops(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Enable tab stops
	r.SetTabStops(8)

	// Test that tab stops are set
	cellbuf := NewBuffer(20, 1)
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Disable tab stops
	r.SetTabStops(-1)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}
}

func TestRendererBackspace(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Enable backspace optimization
	r.SetBackspace(true)

	cellbuf := NewBuffer(10, 1)
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Disable backspace optimization
	r.SetBackspace(false)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}
}

func TestRendererMapNewline(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Enable newline mapping
	r.SetMapNewline(true)

	cellbuf := NewBuffer(10, 2)
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Disable newline mapping
	r.SetMapNewline(false)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}
}

func TestRendererTouched(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(5, 3)

	// Initially, no lines should be touched (empty buffer)
	touched := r.Touched(cellbuf)
	if touched != 0 {
		t.Errorf("expected 0 touched lines initially, got %d", touched)
	}

	// Mark some lines as touched by setting cells
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)
	cellbuf.SetCell(0, 2, &cell)

	// Should have touched lines where we set cells
	touched = r.Touched(cellbuf)
	if touched != 2 {
		t.Errorf("expected 2 touched lines after setting cells, got %d", touched)
	}

	// After rendering, the Touched method still counts all lines as touched
	// because the renderer sets all LineData to non-nil (even with FirstCell: -1, LastCell: -1)
	// This is the actual behavior of the renderer
	r.Render(cellbuf)
	touched = r.Touched(cellbuf)
	if touched != 3 {
		t.Errorf("expected 3 touched lines after render (all lines have LineData), got %d", touched)
	}

	// But if we check the actual touched state by looking at FirstCell/LastCell
	actualTouched := 0
	for _, lineData := range cellbuf.Touched {
		if lineData != nil && (lineData.FirstCell != -1 || lineData.LastCell != -1) {
			actualTouched++
		}
	}
	if actualTouched != 0 {
		t.Errorf("expected 0 actually touched lines after render, got %d", actualTouched)
	}
}

// Test wide character handling
func TestRendererWideCharacters(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 1)

	// Test wide characters (emoji, CJK characters)
	wideChars := []string{"🌟", "中", "文", "字"}
	for i, char := range wideChars {
		cell := Cell{Content: char, Width: 2} // Wide characters typically have width 2
		cellbuf.SetCell(i*2, 0, &cell)
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	for _, char := range wideChars {
		if !strings.Contains(output, char) {
			t.Errorf("expected output to contain wide character '%s', got: %q", char, output)
		}
	}
}

// Test zero-width characters
func TestRendererZeroWidthCharacters(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(5, 1)

	// Test zero-width characters (combining marks, etc.)
	cell := Cell{Content: "a\u0301", Width: 1} // 'a' with combining acute accent
	cellbuf.SetCell(0, 0, &cell)

	// Zero-width cell
	zeroCell := Cell{Content: "\u200B", Width: 0} // Zero-width space
	cellbuf.SetCell(1, 0, &zeroCell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "a\u0301") {
		t.Errorf("expected output to contain combining character, got: %q", output)
	}
}

// Test styled text rendering
func TestRendererStyledText(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 1)

	// Test various styles
	styles := []Style{
		{Attrs: AttrBold},
		{Fg: color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{Bg: color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{Attrs: AttrBold, Fg: color.RGBA{R: 0, G: 0, B: 255, A: 255}},
	}

	for i, style := range styles {
		cell := Cell{Content: "X", Width: 1, Style: style}
		cellbuf.SetCell(i, 0, &cell)
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	// Should contain ANSI escape sequences for styling
	if !strings.Contains(output, "\x1b[") {
		t.Errorf("expected output to contain ANSI escape sequences, got: %q", output)
	}
}

// Test hyperlink rendering
func TestRendererHyperlinks(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 1)

	// Test hyperlink
	link := NewLink("https://example.com")
	cell := Cell{Content: "link", Width: 4, Link: link}
	cellbuf.SetCell(0, 0, &cell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "link") {
		t.Errorf("expected output to contain 'link', got: %q", output)
	}
}

func TestRendererSwitchBuffer(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Start with small buffer
	cellbuf := NewBuffer(5, 3)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Resize to larger buffer
	largeBuf := NewBuffer(10, 6)
	largeBuf.SetCell(0, 1, &cell) // Place at visible position

	r.Render(largeBuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	expected := "\x1b[HX\r\n\n\x1b[J\x1bMX\x1b[K\r\n\n\n\n"
	if output != expected {
		t.Errorf("expected output after resize to be %q, got: %q", expected, output)
	}
}

// Test relative cursor movement
func TestRendererRelativeCursor(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.SetRelativeCursor(true)

	cellbuf := NewBuffer(10, 3)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(5, 1, &cell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}

	// Test disabling relative cursor
	r.SetRelativeCursor(false)
	buf.Reset()

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}
}

// Test logger functionality
func TestRendererLogger(t *testing.T) {
	var buf bytes.Buffer
	var logBuf bytes.Buffer

	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Create a simple logger
	logger := &testLogger{buf: &logBuf}
	r.SetLogger(logger)

	cellbuf := NewBuffer(3, 1)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Should have logged output
	if logBuf.Len() == 0 {
		t.Error("expected logger to have recorded output")
	}

	// Test removing logger
	r.SetLogger(nil)
	logBuf.Reset()

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// Should not have logged anything
	if logBuf.Len() != 0 {
		t.Error("expected no logging after removing logger")
	}
}

// Test scroll optimization
func TestRendererScrollOptimization(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.SetFullscreen(true) // Scroll optimization is enabled in alt screen mode

	cellbuf := NewBuffer(10, 5)

	// Fill buffer with content
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			cell := Cell{Content: string(rune('A' + y)), Width: 1}
			cellbuf.SetCell(x, y, &cell)
		}
	}

	// First render
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	buf.Reset()

	// Simulate scrolling by shifting content up
	newBuf := NewBuffer(10, 5)
	for y := 0; y < 4; y++ {
		for x := 0; x < 10; x++ {
			cell := Cell{Content: string(rune('A' + y + 1)), Width: 1}
			newBuf.SetCell(x, y, &cell)
		}
	}
	// Add new line at bottom
	for x := 0; x < 10; x++ {
		cell := Cell{Content: "F", Width: 1}
		newBuf.SetCell(x, 4, &cell)
	}

	r.Render(newBuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "F") {
		t.Errorf("expected output to contain new content 'F', got: %q", output)
	}
}

// Test multiple prepend operations
func TestRendererMultiplePrepends(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.Resize(20, 10)
	cellbuf := NewBuffer(20, 10)

	// Prepend multiple strings
	r.PrependString(cellbuf, "First line")
	r.PrependString(cellbuf, "Second line")

	// Prepend multiple lines
	line1 := make(Line, 10)
	line2 := make(Line, 10)
	for i, ch := range "Third line" {
		if i < len(line1) {
			line1[i] = Cell{Content: string(ch), Width: 1}
		}
	}
	for i, ch := range "Fourth lin" {
		if i < len(line2) {
			line2[i] = Cell{Content: string(ch), Width: 1}
		}
	}

	r.PrependString(cellbuf, strings.Join([]string{line1.Render(), line2.Render()}, "\n"))

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	expectedStrings := []string{"First line", "Second line", "Third line", "Fourth lin"}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %q", expected, output)
		}
	}
}

// Test error conditions and edge cases
func TestRendererEdgeCases(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Test with empty buffer
	emptyBuf := NewBuffer(0, 0)
	r.Render(emptyBuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer with empty buffer: %v", err)
	}

	// Test with nil cells
	cellbuf := NewBuffer(3, 3)
	cellbuf.SetCell(1, 1, nil) // Set nil cell

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer with nil cells: %v", err)
	}

	// Test with very large buffer
	largeBuf := NewBuffer(1000, 1000)
	cell := Cell{Content: "X", Width: 1}
	largeBuf.SetCell(999, 999, &cell)

	r.Render(largeBuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer with large buffer: %v", err)
	}
}

// Test terminal-specific optimizations
func TestRendererTerminalOptimizations(t *testing.T) {
	tests := []struct {
		name string
		term string
		test func(*testing.T, *TerminalRenderer)
	}{
		{
			name: "alacritty optimizations",
			term: "alacritty",
			test: func(t *testing.T, r *TerminalRenderer) {
				// Alacritty has specific capability limitations
				if r.caps.Contains(capCHT) {
					t.Error("expected alacritty to not support CHT")
				}
			},
		},
		{
			name: "screen optimizations",
			term: "screen",
			test: func(t *testing.T, r *TerminalRenderer) {
				// Screen terminal has specific limitations
				if r.caps.Contains(capREP) {
					t.Error("expected screen to not support REP")
				}
			},
		},
		{
			name: "tmux optimizations",
			term: "tmux",
			test: func(t *testing.T, r *TerminalRenderer) {
				// tmux should support most capabilities
				if !r.caps.Contains(capVPA) {
					t.Error("expected tmux to support VPA")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewTerminalRenderer(&buf, []string{"TERM=" + tt.term})
			tt.test(t, r)
		})
	}
}

// Test cursor movement optimizations
func TestRendererCursorMovementOptimizations(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	// Test tab optimization
	r.SetTabStops(8)
	cellbuf := NewBuffer(20, 1)

	// Place content at tab stops
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(8, 0, &cell)  // First tab stop
	cellbuf.SetCell(16, 0, &cell) // Second tab stop

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}
}

// Test backspace optimization
func TestRendererBackspaceOptimization(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.SetBackspace(true)
	cellbuf := NewBuffer(10, 1)

	// Place content that would benefit from backspace optimization
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(5, 0, &cell)

	// Move cursor to position that would use backspace
	r.MoveTo(8, 0)
	r.MoveTo(3, 0) // Should use backspace to move left

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}
}

// Test newline mapping
func TestRendererNewlineMapping(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.SetMapNewline(true)
	r.SetRelativeCursor(true)

	cellbuf := NewBuffer(10, 3)
	cell := Cell{Content: "X", Width: 1}
	cellbuf.SetCell(0, 0, &cell)
	cellbuf.SetCell(0, 1, &cell)
	cellbuf.SetCell(0, 2, &cell)

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	// Should contain newlines for multi-line content
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}
}

// Test underline styles
func TestRendererUnderlineStyles(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 1)

	// Test different underline styles
	styles := []Style{
		{Underline: UnderlineStyleSingle},
		{Underline: UnderlineStyleDouble},
		{Underline: UnderlineStyleCurly},
		{Underline: UnderlineStyleDotted},
		{Underline: UnderlineStyleDashed},
	}

	for i, style := range styles {
		if i < cellbuf.Width() {
			cell := Cell{Content: "U", Width: 1, Style: style}
			cellbuf.SetCell(i, 0, &cell)
		}
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "U") {
		t.Errorf("expected output to contain 'U', got: %q", output)
	}
}

// Test italic and other text attributes
func TestRendererTextAttributes(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 1)

	// Test various text attributes
	styles := []Style{
		{Attrs: AttrItalic},
		{Attrs: AttrFaint},
		{Attrs: AttrBlink},
		{Attrs: AttrReverse},
		{Attrs: AttrStrikethrough},
	}

	for i, style := range styles {
		if i < cellbuf.Width() {
			cell := Cell{Content: "A", Width: 1, Style: style}
			cellbuf.SetCell(i, 0, &cell)
		}
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A") {
		t.Errorf("expected output to contain 'A', got: %q", output)
	}
}

// Test color downsampling
func TestRendererColorDownsampling(t *testing.T) {
	tests := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{"TrueColor", colorprofile.TrueColor},
		{"ANSI256", colorprofile.ANSI256},
		{"ANSI", colorprofile.ANSI},
		{"Ascii", colorprofile.Ascii},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})
			r.SetColorProfile(tt.profile)

			cellbuf := NewBuffer(3, 1)

			// Test with high-precision color that needs downsampling
			cell := Cell{
				Content: "C",
				Width:   1,
				Style:   Style{Fg: color.RGBA{R: 123, G: 234, B: 45, A: 255}},
			}
			cellbuf.SetCell(0, 0, &cell)

			r.Render(cellbuf)
			if err := r.Flush(); err != nil {
				t.Fatalf("failed to flush renderer: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "C") {
				t.Errorf("expected output to contain 'C', got: %q", output)
			}
		})
	}
}

// Test phantom cursor handling
func TestRendererPhantomCursor(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})
	r.SetColorProfile(colorprofile.TrueColor)

	r.SetFullscreen(true) // Use fullscreen rendering optimizations
	r.SetRelativeCursor(false)
	cellbuf := NewBuffer(5, 3)

	// Fill the last column to trigger phantom cursor behavior
	cell := Cell{Content: "X", Width: 1}
	for y := 0; y < cellbuf.Height(); y++ {
		cellbuf.SetCell(4, y, &cell) // Last column
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	expected := "\x1b[1;5HX\r\n\x1b[5GX\r\n\x1b[5G\x1b[?7lX\x1b[?7h"
	if output != expected {
		t.Errorf("expected no output for phantom cursor case, got: %q", output)
	}
}

// Test line clearing optimizations
func TestRendererLineClearingOptimizations(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(10, 3)

	// Fill first line completely
	cell := Cell{Content: "X", Width: 1}
	for x := 0; x < 10; x++ {
		cellbuf.SetCell(x, 0, &cell)
	}

	// First render
	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	buf.Reset()

	// Clear the line by creating new buffer with empty line
	newBuf := NewBuffer(10, 3)
	// Only set one cell, leaving the rest empty
	newBuf.SetCell(0, 0, &cell)

	r.Render(newBuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "X") {
		t.Errorf("expected output to contain 'X', got: %q", output)
	}
}

// Test repeat character optimization
func TestRendererRepeatCharacterOptimization(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(20, 1)

	// Fill with repeated characters that should trigger REP optimization
	cell := Cell{Content: "A", Width: 1}
	for x := 0; x < 15; x++ {
		cellbuf.SetCell(x, 0, &cell)
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "A") {
		t.Errorf("expected output to contain 'A', got: %q", output)
	}
}

// Test erase character optimization
func TestRendererEraseCharacterOptimization(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	cellbuf := NewBuffer(20, 1)

	// Add some non-space content first
	cell := Cell{Content: "A", Width: 1}
	cellbuf.SetCell(0, 0, &cell)

	// Fill with spaces that should trigger ECH optimization
	spaceCell := Cell{Content: " ", Width: 1}
	for x := 5; x < 15; x++ {
		cellbuf.SetCell(x, 0, &spaceCell)
	}

	r.Render(cellbuf)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	// The output should use erase character optimization for consecutive spaces
	output := buf.String()
	// Just verify it doesn't crash and produces some output with our content
	if !strings.Contains(output, "A") {
		t.Errorf("expected output to contain 'A', got: %q", output)
	}
}

func TestRendererUpdates(t *testing.T) {
	cases := []struct {
		name     string
		frames   []string
		expected []string // expected ANSI escape sequence after each frame
	}{
		{
			name: "simple style change",
			frames: []string{
				"A",
				"\x1b[1mA",
			},
			expected: []string{
				"\rA\r\n\n",
				"\x1b[2A\x1b[1mA\x1b[m",
			},
		},
		{
			name:   "style and link change",
			frames: []string{"A", "\x1b[31m\x1b]8;;https://example.com\x1b\\A\x1b]8;;\x1b\\"}, // red + link
			expected: []string{
				"\rA\r\n\n",
				"\x1b[2A\x1b[31m\x1b]8;;https://example.com\aA\x1b[m\x1b]8;;\a",
			},
		},
		{
			// Covers comparing stored downsampled colors vs new true color styles
			// See commit 75d1e37ff1bb
			name: "the same true color style frames",
			frames: []string{
				" \x1b[38;2;255;128;0mABC\n DEF", // orange
				" \x1b[38;2;255;128;0mABC\n DEF", // orange
				" \x1b[38;2;255;128;0mABC\n DEF", // orange
			},
			expected: []string{
				"\r \x1b[38;5;208mABC\x1b[m\r\n\x1b[38;5;208m DEF\x1b[m\r\n",
				"",
				"",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color", "TTY_FORCE=1"})
			t.Logf("Profile: %v", r.profile)
			r.SetRelativeCursor(true) // Use absolute cursor movements since we're drawing fullscreen

			scr := NewScreenBuffer(5, 3)
			for i, frameStr := range tc.frames {
				NewStyledString(frameStr).Draw(scr, scr.Bounds())
				r.Render(scr.Buffer)
				if err := r.Flush(); err != nil {
					t.Fatalf("failed to flush renderer: %v", err)
				}

				output := buf.String()
				expected := tc.expected[i]
				if output != expected {
					t.Errorf("frame %d: expected output %q, got %q", i, expected, output)
				}

				buf.Reset()
			}
		})
	}
}

func TestRendererPrependOneLine(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})

	r.Resize(10, 5)
	cellbuf := NewScreenBuffer(10, 5)
	NewStyledString("This-is-a .").Draw(cellbuf, cellbuf.Bounds())
	r.Render(cellbuf.Buffer)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	NewStyledString("This-is-a .").Draw(cellbuf, cellbuf.Bounds())
	r.PrependString(cellbuf.Buffer, "Prepended-a-new-line")
	r.Render(cellbuf.Buffer)
	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	expected := "\x1b[HThis-is-a\r\n\n\n\n\n\n\x1b[H\x1b[2LPrepended-a-new-line\r\n"
	if output != expected {
		t.Errorf("expected output to be %q, got: %q", expected, output)
	}
}

func TestRendererEnterExitAltScreen(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{"TERM=xterm-256color"})
	cellbuf := NewBuffer(3, 3)

	// Simulate cursor change
	r.MoveTo(1, 1)

	// Save cursor to compare later
	cur := r.cur
	if cur.X != 1 || cur.Y != 1 {
		t.Errorf("expected cursor to be at (1,1), got (%d,%d)", cur.X, cur.Y)
	}

	// Enter alt screen
	r.EnterAltScreen()

	// Ensure we render before processing further
	r.Render(cellbuf)

	// Check fullscreen is enabled
	if !r.Fullscreen() {
		t.Errorf("expected fullscreen to be enabled in alt screen mode")
	}

	// Check cursor position reset
	if r.cur.X != 0 || r.cur.Y != 0 {
		t.Errorf("expected cursor to be reset to (0,0) in alt screen mode, got (%d,%d)", r.cur.X, r.cur.Y)
	}

	// Exit alt screen
	r.ExitAltScreen()

	// Check relative cursor are disabled
	if !r.flags.Contains(tRelativeCursor) {
		t.Errorf("expected relative cursor to be enabled after exiting alt screen mode")
	}

	// Check cursor position restored
	if r.cur.X != 1 || r.cur.Y != 1 {
		t.Errorf("expected cursor to be restored to (1,1) after exiting alt screen mode, got (%d,%d)", r.cur.X, r.cur.Y)
	}

	if err := r.Flush(); err != nil {
		t.Fatalf("failed to flush renderer: %v", err)
	}

	output := buf.String()
	expected := "\x1b[2;2H\x1b[?1049h\x1b[H\x1b[2J\x1b[?1049l"
	if output != expected {
		t.Errorf("expected output to be %q, got: %q", expected, output)
	}
}

// Helper type for testing logger
type testLogger struct {
	buf *bytes.Buffer
}

func (l *testLogger) Printf(format string, args ...interface{}) {
	l.buf.WriteString("LOG: ")
	l.buf.WriteString(format)
	l.buf.WriteByte('\n')
}
