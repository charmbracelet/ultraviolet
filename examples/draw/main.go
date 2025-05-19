package main

import (
	"context"
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/tv"
	"github.com/charmbracelet/tv/component/styledstring"
	"github.com/charmbracelet/x/ansi"
)

func main() {
	t := tv.DefaultTerminal()
	t.SetTitle("Draw Example")
	if err := t.Start(); err != nil {
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

	t.EnableMode(ansi.FocusEventMode)

	width, height, err := t.GetSize()
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

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
	helpComp := styledstring.New(method, help)
	helpW, helpH := helpComp.Buffer.Width(), helpComp.Buffer.Height()

	var prevHelpBuf *tv.Buffer
	showingHelp := true
	f := &tv.Frame{
		Buffer:   tv.NewBuffer(width, height),
		Viewport: tv.FullViewport{},
		Area:     tv.Rect(0, 0, width, height),
	}
	displayHelp := func(show bool) {
		midX, midY := f.Area.Max.X/2, f.Area.Max.Y/2
		x, y := midX-helpW/2, midY-helpH/2
		midArea := tv.Rect(x, y, helpW, helpH)
		if show {
			// Save the area under the help to restore it later.
			prevHelpBuf = f.Buffer.CloneArea(midArea)
			f.RenderComponent(helpComp, midArea)
		} else if prevHelpBuf != nil {
			// Restore saved area under the help.
			for y := 0; y < prevHelpBuf.Height(); y++ {
				for x := 0; x < prevHelpBuf.Width(); x++ {
					c := prevHelpBuf.CellAt(x, y)
					f.Buffer.SetCell(x+midArea.Min.X, y+midArea.Min.Y, c)
				}
			}
		}
		t.Display(f)
	}

	clearScreen := func() {
		f.Buffer.Clear()
		t.Display(f)
	}

	// Display first frame.
	displayHelp(showingHelp)

	var st tv.Style
	const defaultChar = '█'
	char := defaultChar
	draw := func(ev tv.MouseEvent) {
		m := ev.Mouse()
		f.Buffer.SetCell(m.X, m.Y, &tv.Cell{
			Rune:  char,
			Width: method.StringWidth(string(char)),
			Style: st,
		})
		t.Display(f)
	}

	for ev := range t.Events(ctx) {
		switch ev := ev.(type) {
		case tv.WindowSizeEvent:
			if showingHelp {
				displayHelp(false)
			}
			width, height = ev.Width, ev.Height
			f.Area = tv.Rect(0, 0, ev.Width, ev.Height)
			t.Resize(ev.Width, ev.Height)
			t.ClearScreen()
			if showingHelp {
				displayHelp(showingHelp)
			}
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
	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown program: %v", err)
	}
}
