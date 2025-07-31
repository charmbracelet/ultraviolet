//go:build !windows
// +build !windows

package uv

// Scan advances the scanner to the next event and returns whether it was
// successful. If the scanner is at the end of the input, it returns false.
func (d *InputScanner) Scan() bool {
	return d.scan()
}

// parseWin32InputKeyEvent parses a Win32 input key events. This function is
// only available on Windows.
func (p *EventDecoder) parseWin32InputKeyEvent(*win32InputState, uint16, uint16, rune, bool, uint32, uint16, Logger) Event {
	return nil
}

func (d *InputScanner) processEvents(expired bool) bool {
	return d.processEventsDefault(expired)
}

func (d *InputScanner) run() {
	d.runDefault()
}
