//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || aix || zos
// +build darwin dragonfly freebsd linux netbsd openbsd solaris aix zos

package tv

import (
	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

func (t *Terminal) makeRaw() error {
	var err error

	if t.inTty == nil && t.outTty == nil {
		return ErrNotTerminal
	}

	// Check if we have a terminal.
	for _, f := range []term.File{t.inTty, t.outTty} {
		if f == nil {
			continue
		}
		t.inTtyState, err = term.MakeRaw(f.Fd())
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (t *Terminal) getSize() (w, h int, err error) {
	var f term.File
	if t.inTty != nil {
		f = t.inTty
	}
	if f == nil && t.outTty != nil {
		f = t.outTty
	}
	w, h, err = term.GetSize(f.Fd())
	if err != nil {
		return 0, 0, err
	}
	return w, h, nil
}

func (t *Terminal) optimizeMovements() {
	var f term.File
	if t.inTty != nil {
		f = t.inTty
	}
	if f == nil && t.outTty != nil {
		f = t.outTty
	}
	state, err := term.GetState(f.Fd())
	if err != nil {
		return
	}
	t.useTabs = state.Oflag&unix.TABDLY == unix.TAB0
	t.useBspace = state.Lflag&unix.BSDLY == unix.BS0
}

func (*Terminal) enableWindowsMouse() error  { return nil }
func (*Terminal) disableWindowsMouse() error { return nil }
