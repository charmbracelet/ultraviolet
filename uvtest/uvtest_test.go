package uvtest

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

func TestTerminalSimpleApp(t *testing.T) {
	vt, err := newVtTerm(t)
	if err != nil {
		t.Fatalf("failed to create vt terminal: %v", err)
	}
	defer func() { vt.Close() }()

	if err := vt.Start(); err != nil {
		t.Fatalf("failed to start vt terminal: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)

	defer func() {
		cancel()
	}()

	evch := make(chan uv.Event)
	go vt.StreamEvents(ctx, evch)

	go vt.e.SendText("Helloq")

	var fname string
	var counter int
	var text string

OUT:
	for {
		select {
		case <-ctx.Done():
			break OUT
		case ev := <-evch:
			t.Logf("event: %T %#v", ev, ev)
			counter++
			fname = fmt.Sprintf("frame-%d", counter)
			switch ev := ev.(type) {
			case uv.KeyPressEvent:
				switch {
				case ev.MatchStrings("q", "ctrl+c"):
					cancel()
					break OUT
				default:
					if len(ev.Text) > 0 {
						text += ev.Text
					}
				}
			}

			uv.NewStyledString(text).Draw(vt, vt.Bounds())
			if err := vt.Display(); err != nil {
				t.Errorf("failed to display: %v", err)
			}

			golden.RequireEqual(TName(t, fname), vt.e.Render())
		}
	}

	if err := vt.Shutdown(t.Context()); err != nil {
		t.Fatalf("failed to shutdown vt terminal: %v", err)
	}

	golden.RequireEqual(TName(t, fname), vt.e.Render())
}

type vtTerm struct {
	*uv.Terminal
	e *vt.Emulator
	p xpty.Pty
}

func (v *vtTerm) Close() error {
	_ = v.e.Close()
	defer func() {
		_ = v.p.Close()
	}()
	return v.Terminal.Close()
}

func newVtTerm(t testing.TB, env ...string) (*vtTerm, error) {
	const w, h = 80, 24
	pty, err := xpty.NewPty(w, h)
	if err != nil {
		t.Fatalf("failed to create pty: %v", err)
	}

	if len(env) == 0 {
		env = os.Environ()
	}

	var (
		in  io.Reader = pty
		out io.Writer = pty
	)
	switch p := pty.(type) {
	case *xpty.UnixPty:
		in = p.Slave()
		out = p.Slave()
	case *xpty.ConPty:
		in = p.InPipe()
		out = p.OutPipe()
	}

	term := uv.NewTerminal(in, out, env)
	e := vt.NewEmulator(w, h)

	go io.Copy(e, pty)
	go io.Copy(pty, e)

	return &vtTerm{
		Terminal: term,
		e:        e,
		p:        pty,
	}, nil
}

// tbName is a helper to change the name of a [testing.TB] instance.
type tbName struct {
	*testing.T
	name string
}

// Name implements [testing.TB].
func (t *tbName) Name() string {
	return path.Join(t.T.Name(), t.name)
}

func TName(t *testing.T, name string) *tbName {
	return &tbName{T: t, name: name}
}
