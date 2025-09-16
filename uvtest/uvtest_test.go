package uvtest

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

type helloApp struct {
	*TestTerminal
}

func (h *helloApp) Draw(s uv.Screen, a uv.Rectangle) {
	uv.NewStyledString("Hello, World!").Draw(s, a)
}

func (h *helloApp) HandleEvent(ev uv.Event) {
	switch ev := ev.(type) {
	case uv.KeyPressEvent:
		if ev.MatchStrings("q", "ctrl+c") {
			_ = h.Close()
		}
	}
}

func TestHelloApp(t *testing.T) {
	tt := NewTestTerminal(t)
	ha := &helloApp{tt}
	go tt.Run(ha)
	go tt.SendText("helloq")
	tt.Wait()
}

func TestTerminalSimpleApp(t *testing.T) {
	t.Skip("flaky test, needs investigation")

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

	go func() {
		vt.e.SendText("Helloq")
	}()

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

			<-vt.sigCh
			golden.RequireEqual(tname(t, fname), vt.e.Render())
		}
	}

	if err := vt.Shutdown(t.Context()); err != nil {
		t.Fatalf("failed to shutdown vt terminal: %v", err)
	}

	golden.RequireEqual(tname(t, fname), vt.e.Render())
}

type vtTerm struct {
	*uv.Terminal
	e     *vt.Emulator
	p     xpty.Pty
	mu    *sync.RWMutex
	sigCh chan struct{}
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

	sigCh := make(chan struct{})
	term := uv.NewTerminal(in, out, env)
	e := vt.NewEmulator(w, h)
	mu := &sync.RWMutex{}

	go io.Copy(SignalWriter(t.Context(), e, sigCh), pty)
	go io.Copy(pty, e)

	return &vtTerm{
		Terminal: term,
		e:        e,
		p:        pty,
		mu:       mu,
		sigCh:    sigCh,
	}, nil
}

type safeWriter struct {
	io.Writer
	mu sync.Locker
}

func (s *safeWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Writer.Write(p)
}

func SafeWriter(w io.Writer) io.Writer {
	return &safeWriter{
		Writer: w,
		mu:     new(sync.Mutex),
	}
}

func SafeWriterLock(w io.Writer, mu sync.Locker) io.Writer {
	return &safeWriter{
		Writer: w,
		mu:     mu,
	}
}

type safeReader struct {
	io.Reader
	mu *sync.RWMutex
}

func (s *safeReader) Read(p []byte) (n int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Reader.Read(p)
}

func SafeReader(r io.Reader) io.Reader {
	return &safeReader{
		Reader: r,
		mu:     new(sync.RWMutex),
	}
}

func SafeReaderLock(r io.Reader, mu *sync.RWMutex) io.Reader {
	return &safeReader{
		Reader: r,
		mu:     mu,
	}
}
