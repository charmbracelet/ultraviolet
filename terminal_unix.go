//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tv

import (
	"fmt"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func (t *Terminal) makeRaw() error {
	var err error

	// Check if input is a terminal
	if f, ok := t.in.(term.File); ok && term.IsTerminal(f.Fd()) {
		t.inTty = f
		t.inTtyState, err = term.MakeRaw(t.inTty.Fd())
		if err != nil {
			return fmt.Errorf("error entering raw mode: %w", err)
		}
	}

	if f, ok := t.out.(term.File); ok && term.IsTerminal(f.Fd()) {
		t.outTty = f
	}

	return nil
}

func (t *Terminal) optimizeMovements() {
	if f, ok := t.in.(term.File); ok {
		state, err := term.GetState(f.Fd())
		if err != nil {
			return
		}
		t.scr.UseHardTabs(state.Oflag&unix.TABDLY == unix.TAB0)
		t.scr.UseBackspaces(state.Lflag&unix.BSDLY == unix.BS0)
	}
}
