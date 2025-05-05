//go:build !windows
// +build !windows

package tv

import (
	"io"

	"github.com/muesli/cancelreader"
)

func newCancelreader(r io.Reader, _ bool) (cancelreader.CancelReader, error) {
	return cancelreader.NewReader(r) //nolint:wrapcheck
}
