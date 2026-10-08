package uv

import (
	"bytes"
	"errors"
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
	ps := &ProgramStatus{State: ProgramStateWorking, App: "uv", Progress: Percent(40)}
	if err := EncodeProgramStatus(&buf, ps); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "\x1b]7501;state=working:progress=40:app=uv\x07"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	buf.Reset()
	err := EncodeProgramStatus(&buf, &ProgramStatus{State: "bogus"})
	if !errors.Is(err, ansi.ErrProgramStatusState) {
		t.Errorf("got %v, want %v", err, ansi.ErrProgramStatusState)
	}
	if buf.Len() != 0 {
		t.Errorf("invalid status wrote %q", buf.String())
	}
}

func newProgramStatusScreen() (*TerminalScreen, *bytes.Buffer) {
	var out bytes.Buffer
	s := NewTerminalScreen(io.Discard, Environ{"TERM=xterm-256color"})
	s.w = &out
	return s, &out
}

func TestTerminalScreenProgramStatus(t *testing.T) {
	s, out := newProgramStatusScreen()
	ps := &ProgramStatus{State: ProgramStateDone, App: "uv"}
	want := ansi.SetProgramStatus(ps.toANSI())

	if err := s.SetProgramStatus(ps); err != nil {
		t.Fatal(err)
	}
	if got := s.ProgramStatus(); got == nil || *got != *ps {
		t.Fatalf("ProgramStatus() = %+v, want %+v", got, ps)
	}
	ps.Message = "mutated"
	if s.ProgramStatus().Message != "" {
		t.Error("the screen should keep a copy, not the caller's pointer")
	}
	_ = s.Flush()
	if !strings.Contains(out.String(), want) {
		t.Fatalf("SetProgramStatus: output %q missing %q", out.String(), want)
	}

	out.Reset()
	s.Reset()
	_ = s.Flush()
	if strings.Contains(out.String(), "7501") {
		t.Errorf("Reset should leave a done status, got %q", out.String())
	}

	out.Reset()
	s.Restore()
	_ = s.Flush()
	if !strings.Contains(out.String(), want) {
		t.Errorf("Restore: output %q missing %q", out.String(), want)
	}

	if err := s.SetProgramStatus(nil); err != nil || s.ProgramStatus() != nil {
		t.Errorf("nil status: err=%v, status=%+v", err, s.ProgramStatus())
	}
}

func TestTerminalScreenProgramStatusInvalid(t *testing.T) {
	s, out := newProgramStatusScreen()
	good := ProgramStatus{State: ProgramStateWorking, App: "uv"}
	if err := s.SetProgramStatus(&good); err != nil {
		t.Fatal(err)
	}
	_ = s.Flush()
	out.Reset()

	if err := s.SetProgramStatus(&ProgramStatus{State: "bogus"}); err == nil {
		t.Fatal("expected an error for an invalid state")
	}
	if got := s.ProgramStatus(); got == nil || *got != good {
		t.Errorf("an invalid status replaced the current one: %+v", got)
	}
	_ = s.Flush()
	if strings.Contains(out.String(), "7501") {
		t.Errorf("an invalid status was written: %q", out.String())
	}
}

func TestTerminalScreenResetClearsStaleProgramStatus(t *testing.T) {
	for _, state := range []ProgramState{ProgramStateIdle, ProgramStateWorking, ProgramStateBlocked} {
		t.Run(string(state), func(t *testing.T) {
			s, out := newProgramStatusScreen()
			ps := &ProgramStatus{State: state, App: "uv"}
			if err := s.SetProgramStatus(ps); err != nil {
				t.Fatal(err)
			}
			_ = s.Flush()

			out.Reset()
			s.Reset()
			_ = s.Flush()
			if !strings.Contains(out.String(), ansi.ClearProgramStatus) {
				t.Errorf("Reset should clear a %s status, got %q", state, out.String())
			}

			out.Reset()
			s.Restore()
			_ = s.Flush()
			if !strings.Contains(out.String(), ansi.SetProgramStatus(ps.toANSI())) {
				t.Errorf("Restore should re-send the status, got %q", out.String())
			}
		})
	}
}
