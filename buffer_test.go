package uv

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

func TestBufferUniseg(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty buffer",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "multiple lines",
			input:    "Hello, World!\nThis is a test.\nGoodbye!",
			expected: "Hello, World!\nThis is a test.\nGoodbye!",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := NewBuffer(width(c.input), height(c.input))
			for y, line := range strings.Split(c.input, "\n") {
				var x int
				seg := uniseg.NewGraphemes(line)
				for seg.Next() {
					cell := &Cell{
						Content: seg.Str(),
						Width:   seg.Width(),
					}
					buf.SetCell(x, y, cell)
					x += cell.Width
				}
			}

			if got := TrimSpace(buf.String()); got != c.expected {
				t.Errorf("expected %q, got %q", c.expected, got)
			}
		})
	}
}

// TestLineSet tests the Set method of Line
func TestLineSet(t *testing.T) {
	tests := []struct {
		name     string
		lineLen  int
		x        int
		cell     *Cell
		validate func(t *testing.T, l Line)
	}{
		{
			name:    "set simple cell",
			lineLen: 10,
			x:       5,
			cell:    &Cell{Content: "a", Width: 1},
			validate: func(t *testing.T, l Line) {
				if l[5].Content != "a" {
					t.Errorf("expected 'a' at position 5, got %q", l[5].Content)
				}
			},
		},
		{
			name:    "set out of bounds negative",
			lineLen: 10,
			x:       -1,
			cell:    &Cell{Content: "a", Width: 1},
			validate: func(t *testing.T, l Line) {
				// Should not panic, just return
			},
		},
		{
			name:    "set out of bounds positive",
			lineLen: 10,
			x:       10,
			cell:    &Cell{Content: "a", Width: 1},
			validate: func(t *testing.T, l Line) {
				// Should not panic, just return
			},
		},
		{
			name:    "overwrite wide cell",
			lineLen: 10,
			x:       2,
			cell:    &Cell{Content: "你", Width: 2},
			validate: func(t *testing.T, l Line) {
				// First set a wide cell
				l.Set(2, &Cell{Content: "你", Width: 2})
				// Then overwrite it
				l.Set(2, &Cell{Content: "a", Width: 1})
				if l[2].Content != "a" {
					t.Errorf("expected 'a' at position 2, got %q", l[2].Content)
				}
			},
		},
		{
			name:    "overwrite middle of wide cell",
			lineLen: 10,
			x:       3,
			cell:    &Cell{Content: "a", Width: 1},
			validate: func(t *testing.T, l Line) {
				// First set a wide cell at position 2
				l.Set(2, &Cell{Content: "你", Width: 2})
				// Then overwrite the middle
				l.Set(3, &Cell{Content: "a", Width: 1})
				if l[3].Content != "a" {
					t.Errorf("expected 'a' at position 3, got %q", l[3].Content)
				}
				// The wide cell should be cleared
				if l[2].Content != " " {
					t.Errorf("expected empty at position 2, got %q", l[2].Content)
				}
			},
		},
		{
			name:    "set wide cell at end",
			lineLen: 10,
			x:       9,
			cell:    &Cell{Content: "你", Width: 2},
			validate: func(t *testing.T, l Line) {
				l.Set(9, &Cell{Content: "你", Width: 2})
				if l[9].Content != " " {
					// We replace the wide cell with empty spaces because it
					// doesn't fit in the line.
					t.Errorf("expected ' ' at position 9, got %q", l[9].Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := make(Line, tt.lineLen)
			l.Set(tt.x, tt.cell)
			if tt.validate != nil {
				tt.validate(t, l)
			}
		})
	}
}

// TestLineString tests the String method of Line
func TestLineString(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Line
		expected string
	}{
		{
			name: "empty line",
			setup: func() Line {
				return NewLine(5)
			},
			expected: "",
		},
		{
			name: "simple text",
			setup: func() Line {
				l := make(Line, 5)
				l[0] = Cell{Content: "H", Width: 1}
				l[1] = Cell{Content: "e", Width: 1}
				l[2] = Cell{Content: "l", Width: 1}
				l[3] = Cell{Content: "l", Width: 1}
				l[4] = Cell{Content: "o", Width: 1}
				return l
			},
			expected: "Hello",
		},
		{
			name: "with wide characters",
			setup: func() Line {
				l := make(Line, 6)
				l[0] = Cell{Content: "你", Width: 2}
				l[1] = Cell{} // continuation
				l[2] = Cell{Content: "好", Width: 2}
				l[3] = Cell{} // continuation
				l[4] = Cell{Content: "!", Width: 1}
				l[5] = Cell{Content: " ", Width: 1}
				return l
			},
			expected: "你好!",
		},
		{
			name: "trailing spaces trimmed",
			setup: func() Line {
				l := make(Line, 10)
				l[0] = Cell{Content: "H", Width: 1}
				l[1] = Cell{Content: "i", Width: 1}
				// Rest are empty
				return l
			},
			expected: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			got := l.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBufferMethods tests various Buffer methods
func TestBufferMethods(t *testing.T) {
	t.Run("Width", func(t *testing.T) {
		// Test empty buffer
		b := &Buffer{}
		if w := b.Width(); w != 0 {
			t.Errorf("Width() = %d, want 0", w)
		}

		// Test buffer with lines
		b = NewBuffer(10, 5)
		if w := b.Width(); w != 10 {
			t.Errorf("Width() = %d, want 10", w)
		}
	})

	t.Run("CellAt", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.SetCell(2, 1, &Cell{Content: "X", Width: 1})

		// Test valid position
		if c := b.CellAt(2, 1); c == nil || c.Content != "X" {
			t.Errorf("CellAt(2, 1) = %v, want X", c)
		}

		// Test out of bounds
		if c := b.CellAt(-1, 0); c != nil {
			t.Errorf("CellAt(-1, 0) = %v, want nil", c)
		}
		if c := b.CellAt(0, -1); c != nil {
			t.Errorf("CellAt(0, -1) = %v, want nil", c)
		}
		if c := b.CellAt(10, 0); c != nil {
			t.Errorf("CellAt(10, 0) = %v, want nil", c)
		}
		if c := b.CellAt(0, 5); c != nil {
			t.Errorf("CellAt(0, 5) = %v, want nil", c)
		}
	})

	t.Run("SetCell", func(t *testing.T) {
		b := NewBuffer(10, 5)

		// Test setting cell
		b.SetCell(2, 1, &Cell{Content: "A", Width: 1})
		if c := b.CellAt(2, 1); c == nil || c.Content != "A" {
			t.Errorf("SetCell failed, got %v", c)
		}

		// Test out of bounds
		b.SetCell(-1, 0, &Cell{Content: "B", Width: 1})
		b.SetCell(0, -1, &Cell{Content: "C", Width: 1})
		b.SetCell(10, 0, &Cell{Content: "D", Width: 1})
		b.SetCell(0, 5, &Cell{Content: "E", Width: 1})
		// Should not panic

		// Test nil cell
		b.SetCell(3, 1, nil)
		// Should not panic
	})

	t.Run("Resize", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.SetCell(2, 1, &Cell{Content: "X", Width: 1})

		// Resize smaller
		b.Resize(5, 3)
		if b.Width() != 5 || b.Height() != 3 {
			t.Errorf("Resize(5, 3) failed, got %dx%d", b.Width(), b.Height())
		}

		// Resize larger
		b.Resize(15, 10)
		if b.Width() != 15 || b.Height() != 10 {
			t.Errorf("Resize(15, 10) failed, got %dx%d", b.Width(), b.Height())
		}

		// Resize to same size
		b.Resize(15, 10)
		if b.Width() != 15 || b.Height() != 10 {
			t.Errorf("Resize(15, 10) failed, got %dx%d", b.Width(), b.Height())
		}
	})

	t.Run("FillArea", func(t *testing.T) {
		b := NewBuffer(10, 5)
		cell := &Cell{Content: "X", Width: 1}

		// Fill a rectangle
		b.FillArea(cell, Rect(2, 1, 3, 2))

		// Check filled area
		for y := 1; y < 3; y++ {
			for x := 2; x < 5; x++ {
				if c := b.CellAt(x, y); c == nil || c.Content != "X" {
					t.Errorf("FillArea failed at (%d, %d)", x, y)
				}
			}
		}

		// Check unfilled area
		if c := b.CellAt(1, 1); c != nil && c.Content == "X" {
			t.Error("FillArea filled outside rectangle")
		}

		// Test with nil cell
		b.FillArea(nil, Rect(0, 0, 2, 2))
		// Should not panic
	})

	t.Run("TouchLine", func(t *testing.T) {
		b := NewBuffer(10, 5)

		// Touch valid line
		b.TouchLine(2, 1, 3)
		// Should mark line as dirty

		// Touch out of bounds
		b.TouchLine(0, -1, 1)
		b.TouchLine(0, 5, 1)
		// Should not panic
	})

	t.Run("Touch", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.Touch(2, 1)
		// Should mark cell as dirty
	})

	t.Run("Clear", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.SetCell(2, 1, &Cell{Content: "X", Width: 1})
		b.Clear()

		// Check all cells are cleared
		for y := 0; y < b.Height(); y++ {
			for x := 0; x < b.Width(); x++ {
				if c := b.CellAt(x, y); c != nil && c.Content != " " {
					t.Errorf("Clear failed at (%d, %d)", x, y)
				}
			}
		}
	})

	t.Run("Clone", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.SetCell(2, 1, &Cell{Content: "X", Width: 1})

		clone := b.Clone()

		// Check clone has same content
		if c := clone.CellAt(2, 1); c == nil || c.Content != "X" {
			t.Error("Clone failed to copy content")
		}

		// Modify clone shouldn't affect original
		clone.SetCell(2, 1, &Cell{Content: "Y", Width: 1})
		if c := b.CellAt(2, 1); c == nil || c.Content != "X" {
			t.Error("Clone is not independent")
		}
	})

	t.Run("CloneArea", func(t *testing.T) {
		b := NewBuffer(10, 5)
		b.SetCell(2, 1, &Cell{Content: "X", Width: 1})
		b.SetCell(3, 2, &Cell{Content: "Y", Width: 1})

		clone := b.CloneArea(Rect(2, 1, 2, 2))

		// Check clone has correct size
		if clone.Width() != 2 || clone.Height() != 2 {
			t.Errorf("CloneArea size = %dx%d, want 2x2", clone.Width(), clone.Height())
		}

		// Check content
		if c := clone.CellAt(0, 0); c == nil || c.Content != "X" {
			t.Error("CloneArea failed to copy X")
		}
		if c := clone.CellAt(1, 1); c == nil || c.Content != "Y" {
			t.Error("CloneArea failed to copy Y")
		}
	})

	t.Run("Draw", func(t *testing.T) {
		// Create source buffer
		src := NewBuffer(3, 3)
		src.SetCell(1, 1, &Cell{Content: "S", Width: 1})

		// Create destination screen buffer (which implements Screen interface)
		dst := NewScreenBuffer(10, 5)
		dst.SetCell(2, 2, &Cell{Content: "D", Width: 1})

		// Draw source onto destination at area (1,1) to (4,4)
		src.Draw(&dst, Rect(1, 1, 4, 4))

		// Check drawn content
		if c := dst.CellAt(2, 2); c == nil || c.Content != "S" {
			t.Error("Draw failed to copy content")
		}

		// Check original destination content is preserved outside draw area
		if c := dst.CellAt(0, 0); c != nil && c.Content != " " {
			t.Error("Draw modified content outside draw area")
		}
	})

	t.Run("Render", func(t *testing.T) {
		b := NewBuffer(5, 2)
		b.SetCell(0, 0, &Cell{Content: "H", Width: 1})
		b.SetCell(1, 0, &Cell{Content: "i", Width: 1})
		b.SetCell(0, 1, &Cell{Content: "!", Width: 1})

		got := b.Render()

		// The exact output depends on the rendering logic
		if !strings.Contains(got, "Hi") {
			t.Errorf("Render output doesn't contain expected text")
		}
	})
}

// TestBufferLineOperations tests line insertion and deletion
func TestBufferLineOperations(t *testing.T) {
	t.Run("InsertLine", func(t *testing.T) {
		b := NewBuffer(5, 3)
		b.SetCell(0, 0, &Cell{Content: "A", Width: 1})
		b.SetCell(0, 1, &Cell{Content: "B", Width: 1})
		b.SetCell(0, 2, &Cell{Content: "C", Width: 1})

		// Insert line at position 1
		b.InsertLine(1, 1, nil)

		// Check that B moved to line 2
		if c := b.CellAt(0, 2); c == nil || c.Content != "B" {
			t.Error("InsertLine failed to move line down")
		}

		// Check that new line is empty
		if c := b.CellAt(0, 1); c != nil && c.Content != " " {
			t.Error("InsertLine didn't create empty line")
		}
	})

	t.Run("InsertLineArea", func(t *testing.T) {
		b := NewBuffer(5, 5)
		b.SetCell(0, 1, &Cell{Content: "A", Width: 1})
		b.SetCell(0, 2, &Cell{Content: "B", Width: 1})

		// Insert line in area
		b.InsertLineArea(2, 1, nil, Rect(0, 1, 5, 4))

		// Check that B moved down
		if c := b.CellAt(0, 3); c == nil || c.Content != "B" {
			t.Error("InsertLineArea failed to move line down")
		}
	})

	t.Run("DeleteLine", func(t *testing.T) {
		b := NewBuffer(5, 3)
		b.SetCell(0, 0, &Cell{Content: "A", Width: 1})
		b.SetCell(0, 1, &Cell{Content: "B", Width: 1})
		b.SetCell(0, 2, &Cell{Content: "C", Width: 1})

		// Delete line at position 1
		b.DeleteLine(1, 1, nil)

		// Check that C moved to line 1
		if c := b.CellAt(0, 1); c == nil || c.Content != "C" {
			t.Error("DeleteLine failed to move line up")
		}

		// Check that last line is empty
		if c := b.CellAt(0, 2); c != nil && c.Content != " " {
			t.Error("DeleteLine didn't clear last line")
		}
	})

	t.Run("DeleteLineArea", func(t *testing.T) {
		b := NewBuffer(5, 5)
		b.SetCell(0, 1, &Cell{Content: "A", Width: 1})
		b.SetCell(0, 2, &Cell{Content: "B", Width: 1})
		b.SetCell(0, 3, &Cell{Content: "C", Width: 1})

		// Delete line in area
		b.DeleteLineArea(2, 1, nil, Rect(0, 1, 5, 4))

		// Check that C moved up
		if c := b.CellAt(0, 2); c == nil || c.Content != "C" {
			t.Error("DeleteLineArea failed to move line up")
		}
	})
}

// TestBufferCellOperations tests cell insertion and deletion
func TestBufferCellOperations(t *testing.T) {
	t.Run("InsertCell", func(t *testing.T) {
		b := NewBuffer(5, 2)
		l := b.Line(0)
		l[0] = Cell{Content: "A", Width: 1}
		l[1] = Cell{Content: "B", Width: 1}
		l[2] = Cell{Content: "C", Width: 1}

		// Insert cell at position 1
		b.InsertCell(1, 0, 1, nil)

		// Check that B moved right
		if c := b.CellAt(2, 0); c == nil || c.Content != "B" {
			t.Error("InsertCell failed to move cell right")
		}
	})

	t.Run("InsertCellArea", func(t *testing.T) {
		b := NewBuffer(5, 3)
		l := b.Line(1)
		l[1] = Cell{Content: "A", Width: 1}
		l[2] = Cell{Content: "B", Width: 1}

		// Insert cell in area
		b.InsertCellArea(1, 1, 1, nil, Rect(1, 1, 4, 2))

		// Check that A moved right
		if c := b.CellAt(2, 1); c == nil || c.Content != "A" {
			t.Error("InsertCellArea failed to move cell right")
		}
	})

	t.Run("DeleteCell", func(t *testing.T) {
		b := NewBuffer(5, 2)
		l := b.Line(0)
		l[0] = Cell{Content: "A", Width: 1}
		l[1] = Cell{Content: "B", Width: 1}
		l[2] = Cell{Content: "C", Width: 1}

		// Delete cell at position 1
		b.DeleteCell(1, 0, 1, nil)

		// Check that C moved left
		if c := b.CellAt(1, 0); c == nil || c.Content != "C" {
			t.Error("DeleteCell failed to move cell left")
		}
	})

	t.Run("DeleteCellArea", func(t *testing.T) {
		b := NewBuffer(5, 3)
		l := b.Line(1)
		l[1] = Cell{Content: "A", Width: 1}
		l[2] = Cell{Content: "B", Width: 1}
		l[3] = Cell{Content: "C", Width: 1}

		// Delete cell in area
		b.DeleteCellArea(2, 1, 1, nil, Rect(1, 1, 4, 2))

		// Check that C moved left
		if c := b.CellAt(2, 1); c == nil || c.Content != "C" {
			t.Error("DeleteCellArea failed to move cell left")
		}
	})
}

// TestScreenBuffer tests ScreenBuffer specific functionality
func TestScreenBuffer(t *testing.T) {
	t.Run("NewScreenBuffer", func(t *testing.T) {
		sb := NewScreenBuffer(10, 5)
		if sb.Width() != 10 || sb.Height() != 5 {
			t.Errorf("NewScreenBuffer size = %dx%d, want 10x5", sb.Width(), sb.Height())
		}
	})

	t.Run("WidthMethod", func(t *testing.T) {
		sb := NewScreenBuffer(10, 5)

		// Get the default width method
		method := sb.WidthMethod()
		if method == nil {
			t.Error("WidthMethod returned nil")
		}

		// The method should be set to ansi.WcWidth by default
		// We can test it works by checking a simple character
		if width := method.StringWidth("a"); width != 1 {
			t.Errorf("Default width method returned %d for 'a', want 1", width)
		}
	})
}

// TestLineRenderLine tests the renderLine function
func TestLineRenderLine(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Line
		validate func(t *testing.T, output string)
	}{
		{
			name: "render with styles",
			setup: func() Line {
				l := make(Line, 5)
				l[0] = Cell{Content: "H", Width: 1, Style: Style{Fg: ansi.Red}}
				l[1] = Cell{Content: "i", Width: 1}
				return l
			},
			validate: func(t *testing.T, output string) {
				// Should contain style sequences
				if !strings.Contains(output, "H") || !strings.Contains(output, "i") {
					t.Error("renderLine failed to include content")
				}
			},
		},
		{
			name: "render with hyperlink",
			setup: func() Line {
				l := make(Line, 5)
				l[0] = Cell{Content: "L", Width: 1, Link: Link{URL: "http://example.com"}}
				l[1] = Cell{Content: "i", Width: 1, Link: Link{URL: "http://example.com"}}
				l[2] = Cell{Content: "n", Width: 1, Link: Link{URL: "http://example.com"}}
				l[3] = Cell{Content: "k", Width: 1, Link: Link{URL: "http://example.com"}}
				return l
			},
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Link") {
					t.Error("renderLine failed to include link content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.setup()
			var sb strings.Builder
			renderLine(&sb, l)
			output := sb.String()
			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func BenchmarkBufferSetCell(b *testing.B) {
	buf := NewBuffer(80, 24)
	cell := &Cell{Content: "A", Width: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 80
		y := (i / 80) % 24
		buf.SetCell(x, y, cell)
	}
}

func BenchmarkBufferResize(b *testing.B) {
	buf := NewBuffer(80, 24)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		width := 80 + (i % 20)
		height := 24 + (i % 10)
		buf.Resize(width, height)
		bufWidth, bufHeight := buf.Width(), buf.Height()
		if bufWidth != width || bufHeight != height {
			b.Errorf("Resize failed: got %dx%d, want %dx%d", bufWidth, bufHeight, width, height)
		}
	}
}

func width(s string) int {
	width := 0
	for _, line := range strings.Split(s, "\n") {
		width = max(width, ansi.StringWidth(line))
	}
	return width
}

func height(s string) int {
	return strings.Count(s, "\n") + 1
}
