package conformance

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// corpusProgram decodes a saved fuzz entry the way the fuzz target would.
func corpusProgram(t *testing.T, path string) (Program, bool) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "[]byte(") {
			continue
		}
		quoted := strings.TrimSuffix(strings.TrimPrefix(line, "[]byte("), ")")
		data, err := strconv.Unquote(quoted)
		if err != nil {
			continue
		}
		return DecodeProgram([]byte(data)), true
	}
	return Program{}, false
}

// TestCorpusDecodesStably pins what every saved entry decodes to.
//
// A saved entry is only a regression test for as long as it keeps performing
// the operations it was minimised for. Choosing an operation by a count that
// grows re-pointed every byte the moment a kind was added, so entries kept
// passing while testing something else entirely. This is the alarm for that:
// if it fires, the decoder's mapping moved, and the entries have to be
// re-minted deliberately rather than left to drift.
func TestCorpusDecodesStably(t *testing.T) {
	golden := map[string]string{}
	raw, err := os.ReadFile("testdata/corpus_decode.txt")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, ops, found := strings.Cut(line, "\t")
		if !found {
			t.Fatalf("malformed golden line %q", line)
		}
		golden[name] = ops
	}

	seeds, err := filepath.Glob("testdata/fuzz/*/*")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, path := range seeds {
		p, ok := corpusProgram(t, path)
		if !ok {
			continue
		}
		name := filepath.Join(filepath.Base(filepath.Dir(path)), filepath.Base(path))
		var kinds []string
		for _, op := range p.Ops {
			kinds = append(kinds, op.Kind.String())
		}
		got := strings.Join(kinds, ",")
		seen[name] = true
		want, recorded := golden[name]
		if !recorded {
			t.Errorf("%s is not in testdata/corpus_decode.txt; add it with:\n\t%s\t%s", name, name, got)
			continue
		}
		if got != want {
			t.Errorf("%s decodes to a different program than it was minimised for\n got: %s\nwant: %s", name, got, want)
		}
	}
	for name := range golden {
		if !seen[name] {
			t.Errorf("%s is recorded in testdata/corpus_decode.txt but no longer exists", name)
		}
	}
}

// TestOpTableCoversEveryKind keeps the table honest: a kind the table never
// names is a kind the fuzzer can never generate, which is a silent hole in
// coverage rather than a failure anyone would notice.
func TestOpTableCoversEveryKind(t *testing.T) {
	seen := map[OpKind]int{}
	for _, kind := range opTable {
		seen[kind]++
	}
	for kind := OpKind(0); kind < opKindCount; kind++ {
		if seen[kind] == 0 {
			t.Errorf("%s is never reachable: no slot in opTable names it", kind)
		}
	}
	if len(seen) != int(opKindCount) {
		t.Errorf("opTable names %d kinds, but %d exist", len(seen), opKindCount)
	}
}

// TestSeedsDecodeToTheProgramsTheyDescribe checks that the hand-written seeds
// still mean what their comments say.
//
// A seed is a literal byte slice, so nothing connects it to the program it is
// meant to express. Two silent divergences have already happened: the op bytes
// were the enum's numeric values, which stopped selecting those ops when the
// decoder moved to a slot table, and the resize heights were encoded against the
// fullscreen low bound for the inline modes too, so every inline resize in them
// asked for a different size than intended. Both left the seeds passing while
// covering less than they claimed.
func TestSeedsDecodeToTheProgramsTheyDescribe(t *testing.T) {
	for i, seed := range Seeds() {
		p := DecodeProgram(seed)
		if len(p.Ops) == 0 {
			t.Errorf("seed %d decodes to no operations at all", i)
			continue
		}

		// Every seed exists to reach a render; one that cannot is inert.
		renders := 0
		for _, op := range p.Ops {
			if op.Kind == OpRender || op.Kind == OpRedraw {
				renders++
			}
		}
		if renders == 0 {
			t.Errorf("seed %d never renders, so nothing it sets up is ever tested: %v", i, kindsOf(p))
		}

		// A seed that resizes has to land inside the bounds the decoder
		// promises, or it is exercising a clamp rather than the case it names.
		for _, op := range p.Ops {
			if op.Kind != OpResize {
				continue
			}
			minH := MinResizeH
			if p.Inline {
				minH = 0
			}
			if op.W < MinResizeW || op.W > MaxResizeW {
				t.Errorf("seed %d resizes to width %d, outside [%d,%d]", i, op.W, MinResizeW, MaxResizeW)
			}
			if op.H < minH || op.H > MaxResizeH {
				t.Errorf("seed %d (inline=%v) resizes to height %d, outside [%d,%d]",
					i, p.Inline, op.H, minH, MaxResizeH)
			}
		}
	}
}

func kindsOf(p Program) []string {
	var out []string
	for _, op := range p.Ops {
		out = append(out, op.Kind.String())
	}
	return out
}
