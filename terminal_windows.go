//go:build windows
// +build windows

package tv

import (
	"fmt"

	"github.com/charmbracelet/x/term"
	"golang.org/x/sys/windows"
)

func (t *Terminal) makeRaw() (err error) {
	// Save stdin state and enable VT input.
	// We also need to enable VT input here.
	if f, ok := t.in.(term.File); ok && term.IsTerminal(f.Fd()) {
		t.inTty = f
		t.inTtyState, err = term.MakeRaw(t.inTty.Fd())
		if err != nil {
			return fmt.Errorf("error making terminal raw: %w", err)
		}

		// Enable VT input
		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(t.inTty.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(t.inTty.Fd()), mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT); err != nil {
			return fmt.Errorf("error setting console mode: %w", err)
		}
	}

	// Save output screen buffer state and enable VT processing.
	if f, ok := t.out.(term.File); ok && term.IsTerminal(f.Fd()) {
		t.outTty = f
		t.outTtyState, err = term.GetState(f.Fd())
		if err != nil {
			return fmt.Errorf("error getting terminal state: %w", err)
		}

		var mode uint32
		if err := windows.GetConsoleMode(windows.Handle(t.outTty.Fd()), &mode); err != nil {
			return fmt.Errorf("error getting console mode: %w", err)
		}

		if err := windows.SetConsoleMode(windows.Handle(t.outTty.Fd()),
			mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING|
				windows.DISABLE_NEWLINE_AUTO_RETURN); err != nil {
			return fmt.Errorf("error setting console mode: %w", err)
		}
	}

	return //nolint:nakedret
}

func (t *Terminal) optimizeMovements() {
	// TODO: check if we can optimize cursor movements on Windows.
}
