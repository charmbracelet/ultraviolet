//go:build !windows
// +build !windows

package uv

import (
	"context"
)

type inputRecord = struct{}

func (p *TerminalReader) processRecords([]inputRecord, chan<- Event) {
	// This is a no-op on non-Windows platforms.
}

// streamData sends data from the input stream to the event channel.
func (p *TerminalReader) streamData(ctx context.Context, readc chan []byte, _ chan []inputRecord) error {
	return p.sendBytes(ctx, readc)
}
