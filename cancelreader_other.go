//go:build !windows
// +build !windows

package uv

import (
	"io"

	"github.com/muesli/cancelreader"
)

// NewCancelreader creates a new [cancelreader.CancelReader] that provides a
// cancelable reader interface that can be used to cancel reads.
func NewCancelreader(r io.Reader) (cancelreader.CancelReader, error) {
	return cancelreader.NewReader(r) //nolint:wrapcheck
}
