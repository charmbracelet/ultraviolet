//go:build !windows
// +build !windows

package uv

import "context"

type inputRecord = struct{}

// streamData sends data from the input stream to the event channel.
func (p *TerminalReader) streamData(ctx context.Context, readc chan []byte, _ chan []inputRecord) error {
	return p.sendBytes(ctx, readc)
}

// parseWin32InputKeyEvent parses a Win32 input key events. This function is
// only available on Windows.
func (p *EventDecoder) parseWin32InputKeyEvent(*win32InputState, uint16, uint16, rune, bool, uint32, uint16, Logger) Event {
	return nil
}
