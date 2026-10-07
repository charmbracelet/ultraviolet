package uv

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/x/term"
)

// pipeConsole is a Console over an io.Pipe: a terminal that is not there, so a
// test can feed the input loop bytes exactly as a terminal would. its input is
// not a File, so newPollReader takes the fallback reader on every platform.
type pipeConsole struct {
	in   *io.PipeReader
	feed *io.PipeWriter

	// ended is closed the first time a read of the input fails. the input
	// loop returns on that error without touching the terminal again, so from
	// then on Stop cannot race it.
	ended     chan struct{}
	endedOnce sync.Once
}

var _ Console = (*pipeConsole)(nil)

func newPipeConsole() *pipeConsole {
	r, w := io.Pipe()
	return &pipeConsole{in: r, feed: w, ended: make(chan struct{})}
}

// send writes s as the terminal would. it returns once the input loop has
// read it, which is how a pipe works.
func (c *pipeConsole) send(s string) { _, _ = io.WriteString(c.feed, s) }

func (c *pipeConsole) Read(p []byte) (int, error) {
	n, err := c.in.Read(p)
	if err != nil {
		c.endedOnce.Do(func() { close(c.ended) })
	}
	return n, err
}

func (c *pipeConsole) Write(p []byte) (int, error) { return io.Discard.Write(p) }
func (c *pipeConsole) Close() error                { return c.in.Close() }
func (c *pipeConsole) Environ() []string           { return []string{"TERM=xterm"} }

func (c *pipeConsole) Getenv(k string) string {
	v, _ := c.LookupEnv(k)
	return v
}

func (c *pipeConsole) LookupEnv(k string) (string, bool) {
	if k == "TERM" {
		return "xterm", true
	}
	return "", false
}

func (c *pipeConsole) Reader() io.Reader             { return c }
func (c *pipeConsole) Writer() io.Writer             { return io.Discard }
func (c *pipeConsole) MakeRaw() (*term.State, error) { return nil, nil }
func (c *pipeConsole) Restore() error                { return nil }
func (c *pipeConsole) GetSize() (int, int, error)    { return 80, 24, nil }
func (c *pipeConsole) GetWinsize() (*Winsize, error) { return &Winsize{Col: 80, Row: 24}, nil }

// stopAfterInputEnds ends the input and waits until the input loop has seen
// that before calling Stop, so a test that is not about Stop cannot trip over
// a Stop that races the input loop. events are drained meanwhile, since an
// event loop blocked on delivery would keep the input loop from reading.
func stopAfterInputEnds(t *testing.T, tm *Terminal, c *pipeConsole) {
	t.Helper()
	_ = c.feed.Close()
	deadline := time.After(5 * time.Second)
	for waiting := true; waiting; {
		select {
		case <-c.ended:
			waiting = false
		case <-tm.Events():
		case <-deadline:
			t.Fatal("the input loop never saw the end of its input")
		}
	}
	if err := tm.Stop(); err != nil {
		t.Errorf("Stop: %v", err)
	}
}
