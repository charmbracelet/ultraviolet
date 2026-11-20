//go:build windows
// +build windows

package uv

import (
	"os"
	"testing"
	"time"
)

// openConsoleInput opens CONIN$ the same way a caller does when stdio is
// redirected and it still wants to talk to the console. The handle is a console
// handle, but it is not os.Stdin.
func openConsoleInput(t *testing.T) *os.File {
	t.Helper()

	f, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("no console attached to this process: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	return f
}

// A CONIN$ handle is a console handle, so it must get the console reader, not the
// fallback reader. The fallback cannot cancel an in-flight read.
func TestNewCancelReaderConsoleHandleIsNotFallback(t *testing.T) {
	r, err := NewCancelReader(openConsoleInput(t))
	if err != nil {
		t.Fatalf("NewCancelReader: %v", err)
	}
	defer r.Close() //nolint:errcheck

	if _, ok := r.(*conInputReader); !ok {
		t.Errorf("got %T for a console handle, want *conInputReader", r)
	}
}

// Cancel must interrupt a read that is already blocked on the console.
//
// The console input buffer is empty and nobody is typing, so the read below never
// completes on its own. If Cancel cannot interrupt it, the read blocks forever:
// callers that rely on Cancel to enforce a timeout (lipgloss's terminal query does)
// hang instead of timing out.
func TestCancelUnblocksBlockedConsoleRead(t *testing.T) {
	r, err := NewCancelReader(openConsoleInput(t))
	if err != nil {
		t.Fatalf("NewCancelReader: %v", err)
	}
	defer r.Close() //nolint:errcheck

	read := make(chan error, 1)
	go func() {
		buf := make([]byte, 16)
		_, err := r.Read(buf)
		read <- err
	}()

	// Let the read block on the empty console input buffer.
	time.Sleep(100 * time.Millisecond)

	if !r.Cancel() {
		t.Error("Cancel returned false: the blocked read was not interrupted")
	}

	select {
	case <-read:
	case <-time.After(5 * time.Second):
		t.Fatal("Read still blocked 5s after Cancel")
	}
}
