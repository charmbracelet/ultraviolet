package main

import (
	"context"
	"log"
	"unicode"
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func main() {
	t := uv.DefaultTerminal()
	if err := t.Start(); err != nil {
		log.Fatalf("failed to start program: %v", err)
	}

	modes := []ansi.Mode{
		ansi.ButtonEventMouseMode,
		ansi.SgrExtMouseMode,
		ansi.FocusEventMode,
	}

	t.WriteString(ansi.SetMode(modes...))

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

	helpComp := uv.NewStyledString(help)
	helpArea := helpComp.Bounds()
	helpW, helpH := helpArea.Dx(), helpArea.Dy()

	var prevHelpBuf *uv.Buffer
	showingHelp := true
	displayHelp := func(show bool) {
		midX, midY := width/2, height/2
		x, y := midX-helpW/2, midY-helpH/2
		midArea := uv.Rect(x, y, helpW, helpH)
		if show {
			// Save the area under the help to restore it later.
			prevHelpBuf = screen.CloneArea(t, midArea)
			helpComp.Draw(t, midArea)
		} else if prevHelpBuf != nil {
			// Restore saved area under the help.
			prevHelpBuf.Draw(t, midArea)
		}
		t.Display()
	}

	clearScreen := func() {
		screen.Clear(t)
		t.Display()
	}

	// Display first frame.
	displayHelp(showingHelp)

	const defaultChar = "█"
	pen := uv.EmptyCell
	pen.Content = defaultChar
	draw := func(ev uv.MouseEvent) {
		m := ev.Mouse()
		cur := t.CellAt(m.X, m.Y)
		if cur == nil {
			// Position out of bounds.
			return
		}

		if cur.IsZero() && pen.Width == 1 {
			// Find the previous wide cell.
			var wide *uv.Cell
			var wideX, wideY int
			for i := 1; i < 5 && m.X-i >= 0; i++ {
				wide = t.CellAt(m.X-i, m.Y)
				if wide != nil && !wide.IsZero() && wide.Width > 1 {
					wideX, wideY = m.X-i, m.Y
					break
				}
			}

			if wide != nil {
				// Found a wide cell, make all cells blank.
				wc := *wide
				wc.Empty()
				t.SetCell(wideX, wideY, &wc)
			}
		}

		// Can we fit the cell?
		fit := true
		if w := pen.Width; w > 1 {
			if cur.IsZero() || cur.Width > 1 {
				fit = false
			} else {
				for i := 1; i < w; i++ {
					cur = t.CellAt(m.X+i, m.Y)
					if cur == nil || cur.IsZero() || cur.Width > 1 {
						// Position out of bounds or not empty.
						fit = false
						break
					}
				}
			}
		}
		if !fit {
			// Cell is too wide, ignore it.
			return
		}

		t.SetCell(m.X, m.Y, &pen)
		t.Display()
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case ev := <-t.Events():
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				if showingHelp {
					displayHelp(false)
				}
				width, height = ev.Width, ev.Height
				t.Resize(ev.Width, ev.Height)
				t.Erase()
				if showingHelp {
					displayHelp(showingHelp)
				}
			case uv.KeyPressEvent:
				if showingHelp {
					showingHelp = false
					displayHelp(showingHelp)
					break
				}
				switch {
				case ev.MatchString("ctrl+c"):
					cancel()
				case ev.MatchString("alt+esc"):
					pen.Style = uv.Style{}
					pen.Content = defaultChar
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
						pen.Style = pen.Style.Foreground(ansi.Black + ansi.BasicColor(r-'0'))
						break
					}
					pen.Content = text
					pen.Width = runewidth.RuneWidth(r)
				}
			case uv.MouseClickEvent:
				if showingHelp {
					break
				}
				draw(ev)
			case uv.MouseMotionEvent:
				if showingHelp || ev.Button == uv.MouseNone {
					break
				}
				draw(ev)
			}
		}
	}

	t.WriteString(ansi.ResetMode(modes...))

	// Shutdown the program.
	if err := t.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown program: %v", err)
	}
}
