package uv

import (
	"bytes"
	"testing"
)

func TestDisableCaps(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
	})

	// xterm-256color should have CBT enabled
	if !r.caps.Contains(capCBT) {
		t.Fatal("expected CBT to be enabled for xterm-256color")
	}

	r.DisableCaps(CapCBT)

	if r.caps.Contains(capCBT) {
		t.Fatal("expected CBT to be disabled after DisableCaps")
	}

	// Other caps should still be enabled
	if !r.caps.Contains(capVPA) {
		t.Fatal("expected VPA to still be enabled")
	}
}

func TestDisableCapsMultiple(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
	})

	r.DisableCaps(CapCBT, CapREP, CapECH)

	if r.caps.Contains(capCBT) {
		t.Fatal("expected CBT to be disabled")
	}
	if r.caps.Contains(capREP) {
		t.Fatal("expected REP to be disabled")
	}
	if r.caps.Contains(capECH) {
		t.Fatal("expected ECH to be disabled")
	}

	// Unaffected caps
	if !r.caps.Contains(capVPA) {
		t.Fatal("expected VPA to still be enabled")
	}
	if !r.caps.Contains(capCHA) {
		t.Fatal("expected CHA to still be enabled")
	}
}

func TestExportedCapabilityConstants(t *testing.T) {
	// Verify exported constants match internal values.
	tests := []struct {
		name     string
		exported Capability
		internal capabilities
	}{
		{"VPA", CapVPA, capVPA},
		{"HPA", CapHPA, capHPA},
		{"CHA", CapCHA, capCHA},
		{"CHT", CapCHT, capCHT},
		{"CBT", CapCBT, capCBT},
		{"REP", CapREP, capREP},
		{"ECH", CapECH, capECH},
		{"ICH", CapICH, capICH},
		{"SD", CapSD, capSD},
		{"SU", CapSU, capSU},
	}
	for _, tt := range tests {
		if tt.exported != tt.internal {
			t.Errorf("Cap%s: exported %d != internal %d", tt.name, tt.exported, tt.internal)
		}
	}
}

func TestApplyDisableCapsEnv(t *testing.T) {
	var buf bytes.Buffer
	r := NewTerminalRenderer(&buf, []string{
		"TERM=xterm-256color",
		"UV_DISABLE_CAPS=CBT,REP",
	})

	if r.caps.Contains(capCBT) {
		t.Fatal("expected CBT to be disabled via UV_DISABLE_CAPS")
	}
	if r.caps.Contains(capREP) {
		t.Fatal("expected REP to be disabled via UV_DISABLE_CAPS")
	}

	// Unaffected caps
	if !r.caps.Contains(capVPA) {
		t.Fatal("expected VPA to still be enabled")
	}
}
