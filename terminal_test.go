package uv

import (
	"testing"
)

func TestSupportsBackspace(t *testing.T) {
	_ = supportsBackspace(0)
}

func TestSupportsHardTabs(t *testing.T) {
	_ = supportsHardTabs(0)
}

func TestOpenTTY(t *testing.T) {
	_, _, _ = OpenTTY()
}
