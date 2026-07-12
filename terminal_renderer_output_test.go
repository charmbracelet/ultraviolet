package uv

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRendererOutput(t *testing.T) {
	cases := []struct {
		name      string
		input     []string
		wrap      []bool
		relative  bool
		altscreen bool
		expected  []string
	}{
		{
			name:     "scroll to bottom in inline mode",
			input:    []string{"ABC", "XXX"},
			expected: []string{"\rABC", "\rXXX"},
			relative: true,
		},
		{
			name: "scroll one line",
			input: []string{
				loremIpsum[0],
				loremIpsum[0][10:],
			},
			wrap: []bool{
				true,
				true,
			},
			expected: func() []string {
				if isWindows {
					return []string{
						"\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h",
						"\x1b[Hm dolor si\r\nt amet, co\r\nnsectetur\x1b[K\r\nadipiscing\r\n elit. Vi\x1b[?7lv\x1b[?7h",
					}
				} else {
					return []string{
						"\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h",
						"\r\n elit. Vi\x1b[?7lv\x1b[?7h",
					}
				}
			}(),
			altscreen: true,
		},
		{
			name: "scroll two lines",
			input: []string{
				loremIpsum[0],
				loremIpsum[0][20:],
			},
			wrap: []bool{
				true,
				true,
			},
			expected: func() []string {
				if isWindows {
					return []string{
						"\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h",
						"\x1b[Ht amet, co\r\nnsectetur\x1b[K\r\nadipiscing\r\n elit. Viv\r\namus at o\x1b[?7lr\x1b[?7h",
					}
				} else {
					return []string{
						"\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h",
						"\r\x1b[2S\x1bM elit. Viv\r\namus at o\x1b[?7lr\x1b[?7h",
					}
				}
			}(),
			altscreen: true,
		},
		{
			name: "insert line in the middle",
			input: []string{
				"ABC\nDEF\nGHI\n",
				"ABC\n\nDEF\nGHI",
			},
			wrap: []bool{
				true,
				true,
			},
			expected: func() []string {
				if isWindows {
					return []string{
						"\x1b[H\x1b[2JABC\r\nDEF\r\nGHI",
						"\r\x1bM\x1b[K\nDEF\r\nGHI",
					}
				} else {
					return []string{
						"\x1b[H\x1b[2JABC\r\nDEF\r\nGHI",
						"\r\x1bM\x1b[L",
					}
				}
			}(),
			altscreen: true,
		},
		{
			name: "erase until end of line",
			input: []string{
				"\nABCEFGHIJK",
				"\nABCE      ",
			},
			expected: []string{
				"\x1b[2;1HABCEFGHIJK",
				"\r\x1b[5G\x1b[K",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewTerminalRenderer(&buf, []string{
				"TERM=xterm-256color", // Enable 256 colors
				"COLORTERM=truecolor", // Enable true color support
			})

			s.SetScrollOptim(!isWindows) // Disable scroll optimization on Windows for consistent results
			s.SetFullscreen(c.altscreen)
			s.SetRelativeCursor(c.relative)
			if c.altscreen {
				s.SaveCursor()
				s.Erase()
			}

			scr := NewScreenBuffer(10, 5)
			for i := range c.input {
				buf.Reset()

				comp := NewStyledString(c.input[i])
				if i < len(c.wrap) {
					comp.Wrap = c.wrap[i]
				}
				comp.Draw(scr, scr.Bounds())
				s.Render(scr.RenderBuffer)
				if err := s.Flush(); err != nil {
					t.Fatalf("Flush failed: %v", err)
				}

				if buf.String() != c.expected[i] {
					t.Errorf("Expected output[%d]:\n%q\nGot:\n%q", i, c.expected[i], buf.String())
				}
			}
		})
	}
}

func TestRendererWideCellReanchor(t *testing.T) {
	render := func(grapheme bool) string {
		var buf bytes.Buffer
		s := NewTerminalRenderer(&buf, []string{
			"TERM=xterm-256color",
			"COLORTERM=truecolor",
		})
		s.SetFullscreen(true)
		s.SetGraphemeWidth(grapheme)
		s.SaveCursor()
		s.Erase()

		scr := NewScreenBuffer(10, 1)
		buf.Reset()
		NewStyledString("世界").Draw(scr, scr.Bounds())
		s.Render(scr.RenderBuffer)
		if err := s.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}
		return buf.String()
	}

	out := render(false)
	reanchors := strings.Count(out, "\x1b[5G")
	if reanchors != 1 {
		t.Errorf("non-grapheme line with wide cells: want 1 re-anchor, got %d in %q", reanchors, out)
	}
	if perCell := strings.Count(out, "\x1b[3G"); perCell != 0 {
		t.Errorf("adjacent wide cells emitted a per-cell CHA: %q", out)
	}

	gout := render(true)
	if n := strings.Count(gout, "\x1b[5G"); n != 0 {
		t.Errorf("grapheme-mode line should not re-anchor, got %d in %q", n, gout)
	}
}

// TestRendererDeleteBeforeWideCell verifies that deleting a narrow cell that
// precedes a wide cell does not split the wide cell.
//
// When the tail of a line is shorter than before, transformLine takes the DCH
// (delete character) branch. The last reprinted cell may be a wide character
// whose zero-width continuation occupies the next column. Without guarding for
// that continuation, the renderer moves the cursor back onto it (ESC[D) and
// issues DCH there, deleting the right half of the wide cell and corrupting the
// glyph. In grapheme-width mode (mode 2027) there is no reanchor fallback, so
// the corruption is permanent.
//
// The wide cell may sit anywhere on the line, so the cases interleave wide and
// narrow cells: leading, medial (between two wide cells), and within multiple
// wide runs, not just a trailing CJK run. Each case removes one narrow cell
// that immediately precedes a wide cell, forcing the DCH branch to reprint the
// wide cell as its last cell.
func TestRendererDeleteBeforeWideCell(t *testing.T) {
	cases := []struct {
		name   string
		before string
		after  string
	}{
		{"trailing wide run", "engli世界", "engl世界"},
		{"delete leading narrow before wide", "a世b界c", "世b界c"},
		{"delete medial narrow between wides", "a世b界c", "a世界c"},
		{"wide neighbors on both sides", "世a界b世", "世界b世"},
		{"multiple wide runs, leading delete", "x世y界z国", "世y界z国"},
		{"multiple wide runs, medial delete", "x世y界z国", "x世界z国"},
	}

	render := func(grapheme bool, before, after string) string {
		var buf bytes.Buffer
		s := NewTerminalRenderer(&buf, []string{
			"TERM=xterm-256color",
			"COLORTERM=truecolor",
		})
		s.SetFullscreen(true)
		s.SetGraphemeWidth(grapheme)
		s.SaveCursor()
		s.Erase()

		scr := NewScreenBuffer(12, 1)
		for _, in := range []string{before, after} {
			buf.Reset()
			NewStyledString(in).Draw(scr, scr.Bounds())
			s.Render(scr.RenderBuffer)
			if err := s.Flush(); err != nil {
				t.Fatalf("Flush failed: %v", err)
			}
		}
		// Only the last frame (the delete) is asserted on.
		return buf.String()
	}

	for _, grapheme := range []bool{false, true} {
		for _, c := range cases {
			t.Run(fmt.Sprintf("%s/grapheme=%v", c.name, grapheme), func(t *testing.T) {
				out := render(grapheme, c.before, c.after)
				// A cursor-back (ESC[D) immediately before DCH (ESC[P) means the
				// delete landed on the wide continuation cell, splitting the glyph.
				if strings.Contains(out, "\x1b[D\x1b[P") {
					t.Errorf("DCH split the wide cell (spurious ESC[D before ESC[P): %q", out)
				}
			})
		}
	}
}

var loremIpsum = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus at ornare risus, quis lacinia magna. Suspendisse egestas purus risus, id rutrum diam porta non. Duis luctus tempus dictum. Maecenas luctus metus vitae nulla consectetur egestas. Curabitur faucibus nunc vel eros semper scelerisque. Proin dictum aliquam lacus dignissim fringilla. Praesent ut quam id dui aliquam vehicula in vitae orci. Fusce imperdiet aliquam quam. Nullam euismod magna tincidunt nisl ullamcorper, dignissim rutrum arcu rutrum. Nulla ac fringilla velit. Duis non pellentesque erat.",
	"In egestas ex et sem vulputate, congue bibendum diam ultrices. Nam auctor dictum enim, in rutrum nulla vestibulum sit amet. Vestibulum vel velit ac sem pellentesque accumsan. Vivamus pharetra mi non arcu tristique gravida. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed molestie lectus nunc, sit amet rhoncus orci laoreet vel. Nulla eget mattis massa. Nunc porta eros sollicitudin lorem dapibus luctus. Vestibulum ut turpis ut nibh tincidunt feugiat. Integer eget augue nunc. Morbi vitae ultrices neque. Nulla et convallis libero. Cras nec faucibus odio. Maecenas lacinia sed odio sit amet ultrices.",
	"Nunc at molestie massa. Phasellus commodo dui odio, quis pulvinar orci eleifend a. In et erat nec nisl auctor facilisis at at orci. Curabitur ut ligula in ipsum consequat consectetur. Suspendisse pulvinar arcu metus, et faucibus risus interdum pharetra. Vestibulum vulputate, arcu at malesuada varius, nisl turpis molestie risus, ut lobortis dolor neque vitae diam. Donec lectus libero, iaculis non diam sit amet, sagittis mattis lectus. Vestibulum a magna molestie neque molestie faucibus sagittis et ante. Etiam porta tincidunt nisi sit amet blandit. Vivamus et tellus diam. Vivamus id dolor placerat, tristique magna non, congue est. Nulla a condimentum nulla. Fusce maximus semper nunc, at bibendum mi. Nam malesuada vitae mi molestie tincidunt. Pellentesque sed vestibulum lectus, eu ultrices ligula. Phasellus id nibh tristique, ultricies diam vel, cursus odio.",
	"Integer sed mi viverra, convallis urna congue, efficitur libero. Duis non eros commodo, ultricies quam hendrerit, molestie velit. Nunc non eros vitae lectus hendrerit gravida. Nunc lacinia neque sapien, et accumsan orci elementum vel. Praesent vel interdum nisl. Duis eget diam turpis. Nunc gravida, lacus dictum congue pharetra, dui est laoreet massa, ac convallis elit est sed dui. Morbi luctus convallis dui id tristique.",
	"Praesent vitae laoreet risus. Sed ac facilisis justo. Morbi fringilla in est vel volutpat. Aliquam erat tortor, posuere ac libero sit amet, vehicula blandit sapien. Nullam feugiat purus eget sapien bibendum, id posuere risus finibus. Aliquam erat volutpat. Pellentesque ac purus accumsan, accumsan mi vel, viverra lectus. Ut sed porta erat, vitae mollis nibh. Nunc dignissim quis tellus sed blandit. Mauris id velit in odio commodo aliquet.",
}
