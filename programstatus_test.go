package uv

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestEncodeProgramStatus(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeProgramStatus(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != ansi.ClearProgramStatus {
		t.Errorf("nil: got %q, want %q", got, ansi.ClearProgramStatus)
	}

	buf.Reset()
	ps := &ProgramStatus{State: ProgramStateWorking, App: "uv"}
	if err := EncodeProgramStatus(&buf, ps); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "\x1b]7501;state=working:app=uv\x07"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	buf.Reset()
	if err := EncodeProgramStatus(&buf, &ProgramStatus{State: "bogus"}); err == nil {
		t.Error("expected an error for an invalid state")
	}
	if buf.Len() != 0 {
		t.Errorf("invalid status wrote %q", buf.String())
	}
}

func TestTerminalScreenProgramStatus(t *testing.T) {
	var out bytes.Buffer
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})
	s.w = &out
	ps := &ProgramStatus{State: ProgramStateDone, App: "uv"}
	want := ansi.SetProgramStatus(*ps)

	s.SetProgramStatus(ps)
	if s.ProgramStatus() != ps {
		t.Fatal("ProgramStatus() did not return the set status")
	}
	_ = s.Flush()
	if !strings.Contains(out.String(), want) {
		t.Fatalf("SetProgramStatus: output %q missing %q", out.String(), want)
	}

	out.Reset()
	s.Reset()
	_ = s.Flush()
	if strings.Contains(out.String(), "7501") {
		t.Errorf("Reset should leave the program status, got %q", out.String())
	}

	out.Reset()
	s.Restore()
	_ = s.Flush()
	if !strings.Contains(out.String(), want) {
		t.Errorf("Restore: output %q missing %q", out.String(), want)
	}
}
