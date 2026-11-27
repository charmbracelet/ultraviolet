package uv

import "strings"

// Capability represents a terminal capability that can be toggled.
// It is a bitmask value corresponding to a specific ANSI escape sequence
// optimization.
type Capability = capabilities

const (
	// CapVPA is the Vertical Position Absolute capability [ansi.VPA].
	CapVPA Capability = capVPA
	// CapHPA is the Horizontal Position Absolute capability [ansi.HPA].
	CapHPA Capability = capHPA
	// CapCHA is the Cursor Horizontal Absolute capability [ansi.CHA].
	CapCHA Capability = capCHA
	// CapCHT is the Cursor Horizontal Tab capability [ansi.CHT].
	CapCHT Capability = capCHT
	// CapCBT is the Cursor Backward Tab capability [ansi.CBT].
	CapCBT Capability = capCBT
	// CapREP is the Repeat Previous Character capability [ansi.REP].
	CapREP Capability = capREP
	// CapECH is the Erase Character capability [ansi.ECH].
	CapECH Capability = capECH
	// CapICH is the Insert Character capability [ansi.ICH].
	CapICH Capability = capICH
	// CapSD is the Scroll Down capability [ansi.SD].
	CapSD Capability = capSD
	// CapSU is the Scroll Up capability [ansi.SU].
	CapSU Capability = capSU
)

// DisableCaps clears the given capabilities from the renderer. Use this
// to disable specific terminal optimizations when a terminal emulator is
// known to mishandle certain escape sequences.
func (s *TerminalRenderer) DisableCaps(caps ...Capability) {
	for _, c := range caps {
		s.caps.Reset(c)
	}
}

// capNameMap maps UV_DISABLE_CAPS names to capability bits.
var capNameMap = map[string]capabilities{
	"VPA": capVPA,
	"HPA": capHPA,
	"CHA": capCHA,
	"CHT": capCHT,
	"CBT": capCBT,
	"REP": capREP,
	"ECH": capECH,
	"ICH": capICH,
	"SD":  capSD,
	"SU":  capSU,
}

// applyDisableCaps reads UV_DISABLE_CAPS from env and clears matching
// capability bits and flags. The pseudo-cap "SCROLL" clears tScrollOptim.
func applyDisableCaps(env Environ, caps *capabilities, flags *tFlag) {
	val := env.Getenv("UV_DISABLE_CAPS")
	if val == "" {
		return
	}
	for _, name := range strings.Split(val, ",") {
		name = strings.TrimSpace(name)
		if c, ok := capNameMap[name]; ok {
			caps.Reset(c)
		}
		if name == "SCROLL" {
			flags.Reset(tScrollOptim)
		}
	}
}
