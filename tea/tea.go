package tea

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"sync"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

type Msg interface{}

type QuitMsg struct{}

func Quit() Msg {
	return QuitMsg{}
}

type (
	enterAltScreenMsg struct{}
	exitAltScreenMsg  struct{}
)

func EnterAltScreen() Msg {
	return enterAltScreenMsg{}
}

func ExitAltScreen() Msg {
	return exitAltScreenMsg{}
}

type Cmd func() Msg

// Cursor represents a cursor on the terminal screen.
type Cursor struct {
	// Position is a [Position] that determines the cursor's position on the
	// screen relative to the top left corner of the frame.
	uv.Position

	// Color is a [color.Color] that determines the cursor's color.
	Color color.Color

	// Shape is a [CursorShape] that determines the cursor's shape.
	Shape uv.CursorShape

	// Blink is a boolean that determines whether the cursor should blink.
	Blink bool
}

type View struct {
	Layer           uv.Drawable
	Cursor          *Cursor
	BackgroundColor color.Color
	ForegroundColor color.Color
	WindowTitle     string
}

func NewView(v any) View {
	var view View
	switch v := v.(type) {
	case string:
		view.Layer = uv.NewStyledString(v)
	case fmt.Stringer:
		view.Layer = uv.NewStyledString(v.String())
	case uv.Drawable:
		view.Layer = v
	default:
		view.Layer = uv.NewStyledString(fmt.Sprintf("%v", v))
	}
	return view
}

type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() View
}

type Program struct {
	in             io.Reader
	out            io.Writer
	env            []string
	model          Model
	t              *uv.Terminal
	evch           chan uv.Event
	evctx          context.Context
	evcancel       context.CancelFunc
	width, height  int
	fps            int
	ticker         *time.Ticker
	frame          uv.ScreenBuffer
	view, lastView View
	mu             sync.Mutex
	altscreen      bool

	options startupOptions
}

type startupOptions struct {
	altScreen bool
}

type ProgramOption func(*Program)

func WithInput(r io.Reader) ProgramOption {
	return func(p *Program) {
		p.in = r
	}
}

func WithOutput(w io.Writer) ProgramOption {
	return func(p *Program) {
		p.out = w
	}
}

func WithEnvironment(env []string) ProgramOption {
	return func(p *Program) {
		p.env = env
	}
}

func WithFPS(fps int) ProgramOption {
	return func(p *Program) {
		p.fps = fps
	}
}

func WithAltScreen() ProgramOption {
	return func(p *Program) {
		p.options.altScreen = true
	}
}

const DefaultFPS = 60

func NewProgram(model Model, opts ...ProgramOption) *Program {
	p := &Program{
		model: model,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.in == nil {
		p.in = os.Stdin
	}
	if p.out == nil {
		p.out = os.Stdout
	}
	if p.env == nil {
		p.env = os.Environ()
	}
	if p.fps <= 0 {
		p.fps = DefaultFPS
	}

	p.t = uv.NewTerminal(p.in, p.out, p.env)
	p.evch = make(chan uv.Event)
	p.evctx, p.evcancel = context.WithCancel(context.Background())
	p.ticker = time.NewTicker(time.Second / time.Duration(p.fps))

	return p
}

func (p *Program) Run() (Model, error) {
	cmds := make(chan Cmd)

	if err := p.t.Start(); err != nil {
		return p.model, err
	}

	w, h, err := p.t.GetSize()
	if err != nil {
		return p.model, err
	}

	p.width, p.height = w, h
	p.frame = uv.NewScreenBuffer(p.width, p.height)

	if cmd := p.model.Init(); cmd != nil {
		go func() {
			select {
			case <-p.evctx.Done():
				return
			case cmds <- cmd:
			}
		}()
	}

	if p.options.altScreen {
		p.altscreen = true
		p.t.EnterAltScreen()
	}

	// Command processing loop
	go func() {
		for {
			select {
			case <-p.evctx.Done():
				return
			case cmd := <-cmds:
				if cmd == nil {
					continue
				}
				select {
				case <-p.evctx.Done():
					return
				case p.evch <- cmd():
				}
			}
		}
	}()

	// Input event handling loop
	go func() {
		for {
			select {
			case <-p.evctx.Done():
				return
			case ev := <-p.evch:
				if ev == nil {
					continue
				}

				switch ev := ev.(type) {
				case uv.WindowSizeEvent:
					p.mu.Lock()
					p.width, p.height = ev.Width, ev.Height
					if p.altscreen {
						p.t.Erase()
						_ = p.t.Resize(p.width, p.height)
					}
					p.mu.Unlock()
				case QuitMsg:
					p.evcancel()
				case enterAltScreenMsg:
					p.mu.Lock()
					p.t.Erase()
					p.t.EnterAltScreen()
					_ = p.t.Resize(p.width, p.height)
					p.altscreen = true
					p.mu.Unlock()
				case exitAltScreenMsg:
					p.mu.Lock()
					p.t.Erase()
					p.t.ExitAltScreen()
					_ = p.t.Resize(p.width, p.frame.Height())
					p.altscreen = false
					p.mu.Unlock()
				}

				var cmd Cmd
				p.model, cmd = p.model.Update(ev)
				p.draw(p.model)
				if cmd != nil {
					select {
					case <-p.evctx.Done():
						return
					case cmds <- cmd:
					}
				}

				// The render loop will handle displaying the changes.
			}
		}
	}()

	// Render loop
	go func() {
		for {
			select {
			case <-p.evctx.Done():
				return
			case <-p.ticker.C:
				if err := p.flush(); err != nil {
					p.evcancel()
					return
				}
			}
		}
	}()

	if err := p.t.StreamEvents(p.evctx, p.evch); err != nil {
		return p.model, err
	}

	// Last flush
	if err := p.flush(); err != nil {
		return p.model, err
	}

	err = p.t.Shutdown(context.Background())

	return p.model, err
}

func (p *Program) draw(m Model) {
	frameArea := uv.Rect(0, 0, p.width, p.height)
	view := m.View()
	switch l := view.Layer.(type) {
	case nil:
		frameArea.Max.Y = 0
	case *uv.StyledString:
		frameArea.Max.Y = l.Height()
	case interface{ Bounds() uv.Rectangle }:
		frameArea.Max.Y = l.Bounds().Dy()
	case interface{ Height() int }:
		frameArea.Max.Y = l.Height()
	}

	p.frame.Resize(p.width, frameArea.Dy())

	if view.Layer == nil {
		return
	}

	p.frame.Clear()
	view.Layer.Draw(p.frame, p.frame.Bounds())

	// If the frame height is greater than the screen height, we drop the
	// lines from the top of the buffer.
	if frameHeight := frameArea.Dy(); frameHeight > p.height {
		p.frame.Lines = p.frame.Lines[frameHeight-p.height:]
	}

	p.mu.Lock()
	p.view = view
	p.mu.Unlock()
}

func (p *Program) flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.altscreen {
		_ = p.t.Resize(p.width, p.frame.Height())
	}

	p.t.Clear()
	p.frame.Draw(p.t, p.t.Bounds())

	if p.view.WindowTitle != p.lastView.WindowTitle {
		_, _ = p.t.WriteString(ansi.SetWindowTitle(p.view.WindowTitle))
	}

	var curColor, lastCurColor color.Color
	if p.view.Cursor != nil {
		curColor = p.view.Cursor.Color
	}
	if p.lastView.Cursor != nil {
		lastCurColor = p.lastView.Cursor.Color
	}

	for _, c := range []struct {
		cur, last color.Color
		reset     string
		setter    func(string) string
	}{
		{p.view.ForegroundColor, p.lastView.ForegroundColor, ansi.ResetForegroundColor, ansi.SetForegroundColor},
		{p.view.BackgroundColor, p.lastView.BackgroundColor, ansi.ResetBackgroundColor, ansi.SetBackgroundColor},
		{curColor, lastCurColor, ansi.ResetForegroundColor, ansi.SetForegroundColor},
	} {
		if c.cur != c.last {
			if c.cur == nil {
				_, _ = p.t.WriteString(c.reset)
			} else {
				col, ok := colorful.MakeColor(c.cur)
				if ok {
					_, _ = p.t.WriteString(c.setter(col.Hex()))
				}
			}
		}
	}

	if p.view.Cursor != nil {
		if p.lastView.Cursor == nil ||
			p.view.Cursor.Shape != p.lastView.Cursor.Shape ||
			p.view.Cursor.Blink != p.lastView.Cursor.Blink {
			curStyle := encodeCursorStyle(p.view.Cursor.Shape, p.view.Cursor.Blink)
			_, _ = p.t.WriteString(ansi.SetCursorStyle(curStyle))
		}
	}

	// Queue the changes
	p.t.Render()

	// Move the cursor if needed
	if p.view.Cursor != nil {
		p.t.ShowCursor()
		if p.lastView.Cursor == nil ||
			p.view.Cursor.X != p.lastView.Cursor.X ||
			p.view.Cursor.Y != p.lastView.Cursor.Y {
			p.t.MoveTo(p.view.Cursor.X, p.view.Cursor.Y)
		}
	} else {
		p.t.HideCursor()
	}

	if err := p.t.Flush(); err != nil {
		return err
	}

	p.lastView = p.view

	return nil
}

// encodeCursorStyle returns the integer value for the given cursor style and
// blink state.
func encodeCursorStyle(style uv.CursorShape, blink bool) int {
	// We're using the ANSI escape sequence values for cursor styles.
	// We need to map both [style] and [steady] to the correct value.
	style = (style * 2) + 1 //nolint:mnd
	if !blink {
		style++
	}
	return int(style)
}
