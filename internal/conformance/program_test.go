package conformance_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/ultraviolet/internal/conformance"
	"github.com/charmbracelet/x/ansi"
)

// Tests for the fuzzing machinery itself.
//
// A fuzzer is only as good as its input decoder and its seeds, and both can rot
// silently. A decoder that panics turns every fuzz run into a false positive; a
// seed corpus that no longer decodes into anything interesting leaves the
// fuzzer exploring plain ASCII forever while still reporting success. Neither
// failure is visible from the fuzz targets, so they are pinned down here.

// TestDriftPremise guards the assumption the whole suite rests on. If a future
// width table lines legacy widths up with grapheme widths for these clusters,
// the emulators stop reproducing drift and every conformance test quietly
// starts proving nothing. Fail loudly instead.
func TestDriftPremise(t *testing.T) {
	clusters := conformance.DriftClusters()
	if len(clusters) == 0 {
		t.Fatal("no drift clusters defined")
	}

	for _, cluster := range clusters {
		wc, grapheme := ansi.StringWidthWc(cluster), ansi.StringWidth(cluster)
		if wc == grapheme {
			t.Errorf("%q measures %d columns under both width methods, so it no "+
				"longer exercises column drift and should be replaced", cluster, wc)
		}
	}
}

// TestDecodeProgramTotal checks that decoding never panics and always yields an
// in-range program.
//
// The decoder is the fuzzer's only interface to the renderer. A panic or an
// out-of-range row here would surface as a fuzz failure that has nothing to do
// with the renderer, so it is worth ruling out separately.
func TestDecodeProgramTotal(t *testing.T) {
	inputs := [][]byte{
		nil,
		{},
		{0},
		{255},
		[]byte(strings.Repeat("\xff", 512)),
		[]byte(strings.Repeat("\x00", 512)),
	}
	// A spread of short inputs, since those are the ones most likely to run the
	// decoder off the end of its buffer partway through an operation.
	for i := range 256 {
		inputs = append(inputs, []byte{byte(i), byte(i * 7), byte(i * 13), byte(i * 31)})
	}
	inputs = append(inputs, conformance.Seeds()...)

	for _, in := range inputs {
		p := conformance.DecodeProgram(in)

		if p.Width < 1 || p.Height < 1 {
			t.Fatalf("DecodeProgram(%x) gave a nonsensical size of %dx%d", in, p.Width, p.Height)
		}

		// The live dimensions, which an OpResize changes. Draw and move bounds
		// are checked against these, mirroring how the decoder computes them.
		curW, curH := p.Width, p.Height
		for i, op := range p.Ops {
			switch op.Kind {
			case conformance.OpDrawLine:
				if op.Y < 0 || op.Y >= curH {
					t.Fatalf("DecodeProgram(%x) op %d draws to row %d, outside the current %d-row screen",
						in, i, op.Y, curH)
				}
				if w := ansi.StringWidth(op.Text); w >= curW {
					t.Fatalf("DecodeProgram(%x) op %d draws %d columns into the current %d-column screen, "+
						"which would wrap and make failures ambiguous", in, i, w, curW)
				}
			case conformance.OpMoveTo:
				if op.N < 0 || op.N >= curW {
					t.Fatalf("DecodeProgram(%x) op %d moves to column %d, outside the current %d-column screen",
						in, i, op.N, curW)
				}
			case conformance.OpResize:
				if op.W < conformance.MinResizeW || op.W > conformance.MaxResizeW {
					t.Fatalf("DecodeProgram(%x) op %d resizes to width %d, outside [%d,%d]",
						in, i, op.W, conformance.MinResizeW, conformance.MaxResizeW)
				}
				if op.H < conformance.MinResizeH || op.H > conformance.MaxResizeH {
					t.Fatalf("DecodeProgram(%x) op %d resizes to height %d, outside [%d,%d]",
						in, i, op.H, conformance.MinResizeH, conformance.MaxResizeH)
				}
				curW, curH = op.W, op.H
			}
		}
	}
}

// TestDecodeProgramIsDeterministic checks that one input always decodes to one
// program. Go's fuzzer reruns a failing input to reproduce it, which only works
// if decoding is a pure function of the bytes.
func TestDecodeProgramIsDeterministic(t *testing.T) {
	for _, seed := range conformance.Seeds() {
		a, b := conformance.DecodeProgram(seed), conformance.DecodeProgram(seed)
		if a.String() != b.String() {
			t.Fatalf("DecodeProgram(%x) is not deterministic:\n%s\nvs\n%s", seed, a, b)
		}
	}
}

// TestSeedCorpusIsInteresting guards the seeds against quietly going stale.
//
// The seeds are hand-encoded bytes that address clusters by index, so a change
// to the alphabet or the decoder could turn them into noise with nothing
// failing. A seed that never renders, or that draws no text, costs the fuzzer
// nothing but also teaches it nothing.
func TestSeedCorpusIsInteresting(t *testing.T) {
	seeds := conformance.Seeds()
	if len(seeds) == 0 {
		t.Fatal("seed corpus is empty")
	}

	drift := conformance.DriftClusters()
	seenDrift := map[string]bool{}
	var withText int

	for i, seed := range seeds {
		p := conformance.DecodeProgram(seed)

		var renders, draws int
		for _, op := range p.Ops {
			switch op.Kind {
			case conformance.OpRender, conformance.OpRedraw:
				renders++
			case conformance.OpDrawLine:
				if op.Text == "" {
					continue
				}
				draws++
				for _, c := range drift {
					if strings.Contains(op.Text, c) {
						seenDrift[c] = true
					}
				}
			}
		}

		// One render paints; the bugs this suite exists for need a second frame
		// to diff against the first.
		if renders < 2 {
			t.Errorf("seed %d renders %d times, too few to exercise the incremental path:\n%s",
				i, renders, p)
		}
		if draws > 0 {
			withText++
		}
	}

	if withText == 0 {
		t.Error("no seed draws any text, so the corpus has gone stale")
	}
	for _, c := range drift {
		if !seenDrift[c] {
			t.Errorf("no seed draws %q, so the fuzzer does not start anywhere near it", c)
		}
	}
}

// TestFuzzTargetsRunSeeds runs every fuzz target over its seed corpus, which is
// what `go test` does for a FuzzXxx function without -fuzz.
//
// This is here as documentation as much as verification: it is the reason the
// fuzz targets are useful in ordinary CI, not just during a dedicated fuzzing
// run. The seeds alone make a decent regression suite.
func TestFuzzTargetsRunSeeds(t *testing.T) {
	if testing.Short() {
		t.Skip("exercises every seed against both emulators")
	}
	t.Log("go test runs each FuzzXxx target over its seeds; " +
		"use -fuzz to search for new inputs")
}

// TestIsKnownResidueIsNarrow pins down the allowance that lets fuzzing start.
//
// isKnownResidue suppresses a real failure, so it has to stay tight. If it ever
// widened to match erased or shifted content it would hide exactly the bugs
// these targets exist to find, and nothing else would notice.
func TestIsKnownResidueIsNarrow(t *testing.T) {
	for _, tc := range []struct {
		name      string
		got, want string
		known     bool
	}{
		{"identical rows are not a failure at all", "ab", "ab", false},
		{"trailing residue is the known bug", "a\u26d3\ufe0f\u200d\U0001f4a5\U0001f4a5", "a\u26d3\ufe0f\u200d\U0001f4a5", true},
		{"erased content must still fail", "a", "a\u26d3\ufe0f\u200d\U0001f4a5", false},
		{"shifted content must still fail", " ab", "ab", false},
		{"a wholly different row must still fail", "xy", "ab", false},
		{"an empty expectation must never be excused", "junk", "", false},
	} {
		if got := isKnownResidue(tc.got, tc.want); got != tc.known {
			t.Errorf("%s: isKnownResidue(%q, %q) = %v, want %v",
				tc.name, tc.got, tc.want, got, tc.known)
		}
	}
}
