package main

import (
	"context"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/widget/styledstring"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	t := tv.DefaultTerminal()
	t.SetTitle("Draw Example")
	p := tv.NewProgram(t)
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	// Set terminal info raw mode.
	if err := t.MakeRaw(); err != nil {
		log.Fatalf("failed to set raw mode: %v", err)
	}

	defer t.Restore() //nolint:errcheck

	// Use altscreen buffer.
	t.EnterAltScreen()      //nolint:errcheck
	defer t.ExitAltScreen() //nolint:errcheck

	// Enable mouse events.
	t.EnableMouse()        //nolint:errcheck
	defer t.DisableMouse() //nolint:errcheck

	// Listen for input and mouse events.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const help = `Welcome to Draw Example!

Use the mouse to draw on the screen.
Press ctrl+c to exit.
Press esc to clear the screen.
Press alt+esc to reset the pen character, color, and the screen.
Press 0-9 to set the foreground color.
Press any other key to set the pen character.
Press ctrl+h for this help message.

Press any key to continue...`

	method := ansi.WcWidth
	helpWidget := styledstring.New(method, help)
	helpW, helpH := helpWidget.Buffer.Width(), helpWidget.Buffer.Height()

	var prevHelpBuf *tv.Buffer
	showingHelp := true
	displayHelp := func(show bool) {
		p.Display(func(f *tv.Frame) error {
			midX, midY := f.Area.Max.X/2, f.Area.Max.Y/2
			x, y := midX-helpW/2, midY-helpH/2
			midArea := tv.Rect(x, y, helpW, helpH)
			if show {
				// Save the area under the help to restore it later.
				prevHelpBuf = f.Buffer.CloneArea(midArea)
				return f.RenderWidget(helpWidget, midArea)
			} else if prevHelpBuf != nil {
				// Restore saved area under the help.
				for y := 0; y < prevHelpBuf.Height(); y++ {
					for x := 0; x < prevHelpBuf.Width(); x++ {
						c := prevHelpBuf.CellAt(x, y)
						f.Buffer.SetCell(x+midArea.Min.X, y+midArea.Min.Y, c)
					}
				}
			}
			return nil
		})
	}

	clearScreen := func() {
		p.Display(func(f *tv.Frame) error {
			f.Buffer.Clear()
			return nil
		})
	}

	// Display first frame.
	displayHelp(showingHelp)

	var st tv.Style
	const defaultChar = '█'
	char := defaultChar
	draw := func(ev tv.MouseEvent) {
		m := ev.Mouse()
		p.Display(func(f *tv.Frame) error {
			f.Buffer.SetCell(m.X, m.Y, &tv.Cell{
				Rune:  char,
				Width: method.StringWidth(string(char)),
				Style: st,
			})
			return nil
		})
	}

	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			p.Resize(ev.Width, ev.Height)
		case tv.KeyPressEvent:
			if showingHelp {
				showingHelp = false
				displayHelp(showingHelp)
				break
			}
			switch {
			case ev.MatchStrings("ctrl+c"):
				cancel()
			case ev.MatchString("alt+esc"):
				st = tv.Style{}
				char = defaultChar
				fallthrough
			case ev.MatchString("esc"):
				clearScreen()
			case ev.MatchString("ctrl+h"):
				showingHelp = true
				displayHelp(showingHelp)
			default:
				text := ev.Text
				if len(text) == 0 {
					break
				}
				r, rw := utf8.DecodeRuneInString(text)
				if rw == 1 && unicode.IsDigit(r) {
					st.Foreground(ansi.Black + ansi.BasicColor(r-'0'))
					break
				}
				char = r
			}
		case tv.MouseClickEvent:
			if showingHelp {
				break
			}
			draw(ev)
		case tv.MouseMotionEvent:
			if showingHelp {
				break
			}
			draw(ev)
		}
	}

	// Shutdown the program.
	if err := p.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown program: %v", err)
	}
}
