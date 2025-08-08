//go:build !windows
// +build !windows

package uv

import (
	"bytes"
	"context"
)

type inputRecord = struct{}

func (p *TerminalReader) processRecords(records []inputRecord, buf *bytes.Buffer, eventc chan<- Event) error {
	// This is a no-op on non-Windows platforms.
	return nil
}

// streamData sends data from the input stream to the event channel.
func (p *TerminalReader) streamData(ctx context.Context, readc chan []byte, _ chan []inputRecord) error {
	return p.sendBytes(ctx, readc)
}
