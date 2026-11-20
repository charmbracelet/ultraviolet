package uv

import (
	"reflect"
	"strings"
	"testing"
)

func sameMap(a, b map[string]Key) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// TestSharedKeysTable covers #197: readers with the same legacy flags share one
// table instead of each building its own, and the shared table is the one
// buildKeysTable would have built.
func TestSharedKeysTable(t *testing.T) {
	const term = "xterm-256color"
	find := LegacyKeyEncoding(0).Find(true)

	a, b := sharedKeysTable(0, term, false), sharedKeysTable(0, "dumb", false)
	if !sameMap(a, b) {
		t.Error("readers with the same flags got different tables")
	}
	if !reflect.DeepEqual(a, buildKeysTable(0, term, false)) {
		t.Error("the shared table differs from a built one")
	}

	withFind := sharedKeysTable(find, term, false)
	if sameMap(a, withFind) {
		t.Error("readers with different flags share a table")
	}
	if !reflect.DeepEqual(withFind, buildKeysTable(find, term, false)) {
		t.Error("the shared table for other flags differs from a built one")
	}

	// With terminfo the table depends on term, which a remote client chooses,
	// so it is built each time rather than cached.
	if sameMap(sharedKeysTable(0, term, true), sharedKeysTable(0, term, true)) {
		t.Error("terminfo tables are cached")
	}
}

func TestNewTerminalReaderSharesItsTable(t *testing.T) {
	a := NewTerminalReader(strings.NewReader(""), "xterm-256color")
	b := NewTerminalReader(strings.NewReader(""), "xterm-256color")
	if !sameMap(a.table, b.table) {
		t.Error("two readers built their own tables")
	}
}

func BenchmarkNewTerminalReader(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = NewTerminalReader(strings.NewReader(""), "xterm-256color")
	}
}
