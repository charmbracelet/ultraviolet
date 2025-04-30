package tv

import (
	"fmt"

	"github.com/charmbracelet/x/term"
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
