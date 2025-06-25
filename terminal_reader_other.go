//go:build !windows
// +build !windows

package uv

// ReadEvents reads input events from the terminal. It returns a slice of
// events. The events are parsed from the input buffer and translated into
// input events that can be used by applications to handle user input.
func (d *TerminalReader) ReadEvents(evs []Event) (int, error) {
	return d.readEvents(evs)
}

// parseWin32InputKeyEvent parses a Win32 input key events. This function is
// only available on Windows.
func (p *SequenceParser) parseWin32InputKeyEvent(*win32InputState, uint16, uint16, rune, bool, uint32, uint16, Logger) Event {
	return nil
}
