package uv

import (
	"strings"
	"testing"
)

func TestReaderNonFile(t *testing.T) {
	pr, err := newPollReader(strings.NewReader(""))
	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !pr.Cancel() {
		t.Errorf("expected cancellation to be success")
	}
}
