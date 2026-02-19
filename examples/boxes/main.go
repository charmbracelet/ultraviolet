package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"slices"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

const rootID = "root"

// AppWindow represents a window in the application.
type AppWindow struct {
	id  string
	win *uv.Window
	ctx *screen.Context
	z   int
	st  uv.Style
}

// Bounds returns the bounds of the window.
func (aw *AppWindow) Bounds() uv.Rectangle {
	return aw.win.Bounds()
}

// Draw draws the window onto the given target window at the specified position.
func (aw *AppWindow) Draw(scr uv.Screen, rect uv.Rectangle) {
	aw.win.Draw(scr, rect)
}

// Resize resizes the window to the given width and height.
func (aw *AppWindow) Resize(width, height int) {
	aw.win.Resize(width, height)
}

// Context returns a new drawing context for the window.
func (aw *AppWindow) Context() *screen.Context {
	return aw.ctx
}

type App struct {
	scr         *uv.Window
	root        *AppWindow
	wins        map[string]*AppWindow
	zwins       []*AppWindow
	active      string
	mtx         sync.RWMutex
	quit        bool
	lastClicked string
	mouseDown   bool
}

// EventHandler represents an event handler function. It receives the focused
// window and the event as parameters. It returns true if the event was
// handled, false otherwise.
type EventHandler func(win *uv.Window, ev uv.Event) bool

// NewApp creates a new [App] instance.
func NewApp(width, height int) *App {
	a := new(App)
	a.active = rootID
	a.scr = uv.NewScreen(width, height)
	a.scr.SetWidthMethod(ansi.GraphemeWidth)
	root := &AppWindow{
		id:  rootID,
		win: a.scr.NewWindow(0, 0, width, height),
		ctx: screen.NewContext(a.scr),
		z:   0,
	}
	a.root = root
	a.wins = map[string]*AppWindow{
		root.id: root,
	}
	a.zwins = []*AppWindow{root}
	return a
}

// BringToFront brings the window with the given id to the front.
func (a *App) BringToFront(id string) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	awin, ok := a.wins[id]
	if !ok {
		return
	}

	// Remove the window from its current position.
	for i, zw := range a.zwins {
		if zw.id == id {
			a.zwins = append(a.zwins[:i], a.zwins[i+1:]...)
			break
		}
	}

	// Append it to the end of the slice.
	a.zwins = append(a.zwins, awin)
}

// CreateWindow creates a new window with the given id, position and size.
func (a *App) CreateWindow(id string, x, y, width, height int) *AppWindow {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	var style uv.Style
	style.Bg = ansi.IndexedColor(rand.Intn(256))

	win := a.root.win.NewWindow(x, y, width, height)
	win.Fill(&uv.Cell{
		Content: " ",
		Width:   1,
		Style:   style,
	})

	awin := &AppWindow{
		id:  id,
		win: win,
		ctx: screen.NewContext(win),
		st:  style,
	}
	a.wins[id] = awin
	a.zwins = append(a.zwins, awin)
	return awin
}

// DestroyWindow destroys the window with the given id.
func (a *App) DestroyWindow(id string) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	_, ok := a.wins[id]
	if !ok {
		return
	}
	log.Printf("destroying window %q", id)
	delete(a.wins, id)
	for i, zw := range a.zwins {
		if zw.id == id {
			a.zwins = append(a.zwins[:i], a.zwins[i+1:]...)
			break
		}
	}
	if a.active == id {
		a.active = rootID
	}
}

// SetActiveID sets the currently active window to the given id.
func (a *App) SetActiveID(id string) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.active = id
}

// ActiveID returns the currently active window. If no active window is set, it
// returns an empty string.
func (a *App) ActiveID() string {
	return a.active
}

// Window returns the window associated with the given id. If no window is
// found, it returns nil.
func (a *App) Window(id string) *uv.Window {
	a.mtx.RLock()
	defer a.mtx.RUnlock()
	if win, ok := a.wins[id]; ok {
		return win.win
	}
	return nil
}

// ParentID returns the parent window ID of the given window ID. If no parent is
// found, it returns an empty string.
func (a *App) ParentID(id string) string {
	a.mtx.RLock()
	defer a.mtx.RUnlock()
	win, ok := a.wins[id]
	if !ok {
		return ""
	}
	parent := win.win.Parent()
	if parent == nil {
		return ""
	}
	for _, aw := range a.wins {
		if aw.win == parent {
			return aw.id
		}
	}
	return ""
}

// Draw draws the applications windows to the root window.
func (a *App) Draw(scr uv.Screen, area uv.Rectangle) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	slices.SortStableFunc(a.zwins, func(aw1, aw2 *AppWindow) int {
		return aw1.z - aw2.z
	})

	screen.Clear(a.root.win)
	for _, zw := range a.zwins {
		if zw.id == rootID {
			continue
		}
		cell0 := zw.win.CellAt(0, 0)
		var bg color.Color
		if cell0 != nil {
			bg = cell0.Style.Bg
		}
		log.Printf("drawing window %q at z-index %d bg color %v", zw.id, zw.z, bg)
		zw.Draw(a.root.win, zw.Bounds())
	}

	a.root.Draw(scr, area)
}

// HandleEvent implements an [EventHandler] for the application.
func (a *App) HandleEvent(id string, ev uv.Event) bool {
	win := a.Window(id)
	if win == nil {
		return false
	}

	switch ev := ev.(type) {
	case uv.KeyPressEvent:
		switch {
		case ev.MatchString("ctrl+c", "esc"):
			a.quit = true
			return true
		default:
			activeWin, ok := a.wins[a.ActiveID()]
			if !ok {
				log.Printf("no active window found for id %q", a.ActiveID())
				break
			}

			ctx := activeWin.Context()
			ctx.SetStyle(activeWin.st)
			ctx.SetForeground(color.Black)

			switch {
			case ev.MatchString("backspace"):
				x, y := ctx.Position()
				x--
				if x < 0 {
					x = activeWin.Bounds().Dx() - 1
					y--
					if y < 0 {
						y = 0
					}
				}
				ctx.SetPosition(x, y)
				ctx.Printf(" ")
				ctx.SetPosition(x, y)
				return true
			case ev.MatchString("enter"):
				ctx.Printf("\n")
				return true
			case len(ev.Text) > 0:
				ctx.Print(ev.Text)
				return true
			}
		}
	case uv.MouseMotionEvent:
		if a.mouseDown && a.lastClicked == id {
			// Move the window.
			bounds := win.Bounds()
			newX := ev.X - bounds.Dx()/2
			newY := ev.Y - bounds.Dy()/2
			log.Printf("moving window %q from %v to (%d, %d)", id, bounds.Min, newX, newY)
			win.MoveTo(newX, newY)
			return true
		}
	case uv.MouseReleaseEvent:
		a.mouseDown = false
		a.lastClicked = ""
		return true
	case uv.MouseClickEvent:
		a.mouseDown = true

		switch ev.Button {
		case uv.MouseLeft:
			log.Printf("mouse left click for %q at (%d, %d)", id, ev.X, ev.Y)

			for i := len(a.zwins) - 1; i >= 0; i-- {
				zw := a.zwins[i]
				pos := uv.Pos(ev.X, ev.Y)
				bounds := zw.win.Bounds()
				if zw.id != rootID && pos.In(bounds) {
					log.Printf("clicked window %s at %v (bounds: %v)", zw.id, pos, bounds)
					a.SetActiveID(zw.id)
					a.BringToFront(zw.id)
					a.lastClicked = zw.id
					return true
				}
			}

			log.Printf("no window clicked on at (%d, %d)", ev.X, ev.Y)

			if id == rootID {
				log.Printf("clicked root window at (%d, %d)", ev.X, ev.Y)

				// Create a new window when we click anywhere in the root window.
				rootSize := a.root.Bounds().Size()
				width := rand.Intn(20)
				height := rand.Intn(10)
				if width == 0 || height == 0 {
					// Try again
					return a.HandleEvent(id, ev)
				}

				x := ev.X - width/2
				y := ev.Y - height/2
				if x < 0 {
					x = 0
				}
				if y < 0 {
					y = 0
				}
				if x+width > rootSize.X {
					x = rootSize.X - width
				}
				if y+height > rootSize.Y {
					y = rootSize.Y - height
				}

				winID := fmt.Sprintf("win-%d", len(a.wins))
				a.CreateWindow(winID, x, y, width, height)
				a.SetActiveID(winID)

				return true
			}

		case uv.MouseRight:
			// Destroy the clicked window.
			for i := len(a.zwins) - 1; i >= 0; i-- {
				zw := a.zwins[i]
				pos := uv.Pos(ev.X, ev.Y)
				if pos.In(zw.win.Bounds()) && zw.id != rootID {
					log.Printf("right-clicked window %q at (%d, %d), destroying", zw.id, ev.X, ev.Y)
					a.DestroyWindow(zw.id)
				}
			}

			return true
		}
	}

	return false
}

// Resize resizes the screen to the given width and height.
func (a *App) Resize(width, height int) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.scr.Resize(width, height)
	return nil
}

// Run starts the application event loop on the given terminal.
func (a *App) Run(input uv.File, output uv.File, environ []string) error {
	term := uv.NewTerminal(uv.NewConsole(input, output, environ), nil)
	// We're using the alternate screen buffer, we need to ensure we're using
	// the fullscreen and absolute cursor movement flags.
	scr := term.Screen()
	scr.EnterAltScreen()
	scr.HideCursor()

	if err := term.Start(); err != nil {
		return fmt.Errorf("failed to start terminal: %w", err)
	}

	defer term.Stop()

	scr.SetMouseMode(uv.MouseModeDrag)

	for !a.quit {
		select {
		case ev := <-term.Events():
			log.Printf("event: %#v", ev)

			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				// We need to update our terminal size and root window size.
				scr.Resize(ev.Width, ev.Height)
				a.root.Resize(ev.Width, ev.Height)
			}

			focusedID := a.ActiveID()
			if len(focusedID) == 0 {
				// Ignore events if no window is focused.
				continue
			}

			for !a.HandleEvent(focusedID, ev) {
				if parentID := a.ParentID(focusedID); parentID != "" {
					log.Printf("event not handled by %q, passing to parent %q", focusedID, parentID)
					focusedID = parentID
				} else {
					break
				}
			}

			if err := scr.Display(a); err != nil {
				return fmt.Errorf("failed to display terminal: %w", err)
			}
		}
	}

	return nil
}

func init() {
	f, err := os.OpenFile("uv.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	stdin, stdout, environ := os.Stdin, os.Stdout, os.Environ()
	physicalWidth, physicalHeight, err := term.GetSize(stdout.Fd())
	if err != nil {
		log.Fatalf("failed to get terminal size: %v", err)
	}

	app := NewApp(physicalWidth, physicalHeight)
	if err := app.Run(stdin, stdout, environ); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
