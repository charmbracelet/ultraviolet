package uv

import (
	"reflect"
	"strings"
	"sync"
	"testing"
)

// streamAll returns every event the reader produces until its input ends.
func streamAll(t *testing.T, r *TerminalReader) []Event {
	t.Helper()
	eventc := make(chan Event)
	go func() {
		defer close(eventc)
		if err := r.StreamEvents(t.Context(), eventc); err != nil {
			t.Errorf("error streaming events: %v", err)
		}
	}()
	var events []Event
	for ev := range eventc {
		events = append(events, ev)
	}
	return events
}

// TestTerminalReaderLegacySetAfterConstruction is the regression test for
// #196. NewTerminalReader copied Legacy into its lookup table and its event
// scanner before the caller could set it, so a reader told to report ESC[1~ as
// Find still reported Home, from the table and from the decoder alike.
func TestTerminalReaderLegacySetAfterConstruction(t *testing.T) {
	find := KeyPressEvent{Code: KeyFind}
	tests := []struct {
		name string
		in   string
		want []Event
	}{
		{"lookup table", "\x1b[1~", []Event{find}},
		// The trailing byte keeps the table from matching, so the decoder runs.
		{"decoder", "\x1b[1~a", []Event{find, KeyPressEvent{Code: 'a', Text: "a"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTerminalReader(strings.NewReader(tt.in), "xterm-256color")
			r.Legacy = LegacyKeyEncoding(0).Find(true)

			if got := streamAll(t, r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

// TestTerminalReaderUseTerminfoSetAfterConstruction is the regression test for
// #196: UseTerminfo set after NewTerminalReader never reached the lookup table.
func TestTerminalReaderUseTerminfoSetAfterConstruction(t *testing.T) {
	const term = "xterm-256color"
	want := buildKeysTable(0, term, true)
	if reflect.DeepEqual(want, buildKeysTable(0, term, false)) {
		t.Skipf("no terminfo entry for %s on this machine", term)
	}

	r := NewTerminalReader(strings.NewReader(""), term)
	r.UseTerminfo = true
	streamAll(t, r)

	if !reflect.DeepEqual(r.eventScanner.table, want) {
		t.Error("the lookup table does not include the terminfo key sequences")
	}
}

type recordingLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *recordingLogger) Printf(format string, _ ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, format)
}

// TestTerminalReaderSetLoggerReachesScanner is the regression test for #196:
// the scanner kept the nil logger it was given in NewTerminalReader, so only
// the reader's own lines reached a logger set with SetLogger.
func TestTerminalReaderSetLoggerReachesScanner(t *testing.T) {
	r := NewTerminalReader(strings.NewReader("a"), "xterm-256color")
	var l recordingLogger
	r.SetLogger(&l)
	streamAll(t, r)

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, line := range l.lines {
		if strings.HasPrefix(line, "processing buf") {
			return
		}
	}
	t.Errorf("no scanner line among %q", l.lines)
}
