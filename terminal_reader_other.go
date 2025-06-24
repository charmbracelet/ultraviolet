//go:build !windows
// +build !windows

package uv

// ReadEvents reads input events from the terminal. It takes a slice of [Event]
// as an argument that will be filled with the read events. It returns the
// number of events read and an error if any occurred during reading.
func (d *TerminalReader) ReadEvents(events []Event) (int, error) {
	return d.readEvents(events)
}

// parseWin32InputKeyEvent parses a Win32 input key events. This function is
// only available on Windows.
func (p *SequenceParser) parseWin32InputKeyEvent(*win32InputState, uint16, uint16, rune, bool, uint32, uint16, Logger) Event {
	return nil
}
