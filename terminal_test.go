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

func TestStopBeforeStart(t *testing.T) {
	term := DefaultTerminal()
	if err := term.Stop(); err != nil {
		t.Fatalf("Stop before Start should not panic or error: %v", err)
	}
}

func TestStopIdempotent(t *testing.T) {
	term := DefaultTerminal()
	for i := 0; i < 3; i++ {
		if err := term.Stop(); err != nil {
			t.Fatalf("Stop call %d should not panic or error: %v", i+1, err)
		}
	}
}
