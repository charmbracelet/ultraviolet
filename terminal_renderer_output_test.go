package uv

import (
	"bytes"
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
			name:     "Single Line",
			input:    []string{"ABC", "XXX"},
			expected: []string{"\x1b[?25l\rABC\r\n\n\n\n\x1b[?25h", "\x1b[?25l\x1b[4AXXX\x1b[?25h"},
			relative: true,
		},
		{
			name: "Scroll one line",
			input: []string{
				loremIpsum[0],
				loremIpsum[0][10:],
			},
			wrap: []bool{
				true,
				true,
			},
			expected: []string{
				"\x1b[?25l\x1b[?1049h\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h\x1b[?25h",
				"\x1b[?25l\r\n elit. Vi\x1b[?7lv\x1b[?7h\x1b[?25h",
			},
			altscreen: true,
		},
		{
			name: "Scroll two lines",
			input: []string{
				loremIpsum[0],
				loremIpsum[0][20:],
			},
			wrap: []bool{
				true,
				true,
			},
			expected: []string{
				"\x1b[?25l\x1b[?1049h\x1b[H\x1b[2JLorem ipsu\r\nm dolor si\r\nt amet, co\r\nnsectetur\r\nadipiscin\x1b[?7lg\x1b[?7h\x1b[?25h",
				"\x1b[?25l\r\x1b[2S\x1bM elit. Viv\r\namus at o\x1b[?7lr\x1b[?7h\x1b[?25h",
			},
			altscreen: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewTerminalRenderer(&buf, []string{
				"TERM=xterm-256color", // Enable 256 colors
				"COLORTERM=truecolor", // Enable true color support
			})

			s.SetRelativeCursor(c.relative)
			if c.altscreen {
				s.EnterAltScreen()
			}

			scr := NewScreenBuffer(10, 5)
			for i := range c.input {
				buf.Reset()

				comp := NewStyledString(c.input[i])
				if i < len(c.wrap) {
					comp.Wrap = c.wrap[i]
				}
				comp.Draw(scr, scr.Bounds())
				s.Render(scr.Buffer)
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

var loremIpsum = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus at ornare risus, quis lacinia magna. Suspendisse egestas purus risus, id rutrum diam porta non. Duis luctus tempus dictum. Maecenas luctus metus vitae nulla consectetur egestas. Curabitur faucibus nunc vel eros semper scelerisque. Proin dictum aliquam lacus dignissim fringilla. Praesent ut quam id dui aliquam vehicula in vitae orci. Fusce imperdiet aliquam quam. Nullam euismod magna tincidunt nisl ullamcorper, dignissim rutrum arcu rutrum. Nulla ac fringilla velit. Duis non pellentesque erat.",
	"In egestas ex et sem vulputate, congue bibendum diam ultrices. Nam auctor dictum enim, in rutrum nulla vestibulum sit amet. Vestibulum vel velit ac sem pellentesque accumsan. Vivamus pharetra mi non arcu tristique gravida. Interdum et malesuada fames ac ante ipsum primis in faucibus. Sed molestie lectus nunc, sit amet rhoncus orci laoreet vel. Nulla eget mattis massa. Nunc porta eros sollicitudin lorem dapibus luctus. Vestibulum ut turpis ut nibh tincidunt feugiat. Integer eget augue nunc. Morbi vitae ultrices neque. Nulla et convallis libero. Cras nec faucibus odio. Maecenas lacinia sed odio sit amet ultrices.",
	"Nunc at molestie massa. Phasellus commodo dui odio, quis pulvinar orci eleifend a. In et erat nec nisl auctor facilisis at at orci. Curabitur ut ligula in ipsum consequat consectetur. Suspendisse pulvinar arcu metus, et faucibus risus interdum pharetra. Vestibulum vulputate, arcu at malesuada varius, nisl turpis molestie risus, ut lobortis dolor neque vitae diam. Donec lectus libero, iaculis non diam sit amet, sagittis mattis lectus. Vestibulum a magna molestie neque molestie faucibus sagittis et ante. Etiam porta tincidunt nisi sit amet blandit. Vivamus et tellus diam. Vivamus id dolor placerat, tristique magna non, congue est. Nulla a condimentum nulla. Fusce maximus semper nunc, at bibendum mi. Nam malesuada vitae mi molestie tincidunt. Pellentesque sed vestibulum lectus, eu ultrices ligula. Phasellus id nibh tristique, ultricies diam vel, cursus odio.",
	"Integer sed mi viverra, convallis urna congue, efficitur libero. Duis non eros commodo, ultricies quam hendrerit, molestie velit. Nunc non eros vitae lectus hendrerit gravida. Nunc lacinia neque sapien, et accumsan orci elementum vel. Praesent vel interdum nisl. Duis eget diam turpis. Nunc gravida, lacus dictum congue pharetra, dui est laoreet massa, ac convallis elit est sed dui. Morbi luctus convallis dui id tristique.",
	"Praesent vitae laoreet risus. Sed ac facilisis justo. Morbi fringilla in est vel volutpat. Aliquam erat tortor, posuere ac libero sit amet, vehicula blandit sapien. Nullam feugiat purus eget sapien bibendum, id posuere risus finibus. Aliquam erat volutpat. Pellentesque ac purus accumsan, accumsan mi vel, viverra lectus. Ut sed porta erat, vitae mollis nibh. Nunc dignissim quis tellus sed blandit. Mauris id velit in odio commodo aliquet.",
}
