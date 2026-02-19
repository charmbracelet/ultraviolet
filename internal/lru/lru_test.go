package lru

import (
	"strconv"
	"testing"
)

func TestCache(t *testing.T) {
	const size = 20

	cache := New[int, string](size)

	val := strconv.Itoa

	for i := range size {
		v := val(i)

		if cache.Add(i, v) {
			t.Fatalf("Cache.Add: value evicted before size limit at %d", i)
		}

		got, ok := cache.Get(i)
		if !ok {
			t.Fatalf("Cache.Get: value not found at key %d", i)
		}

		if v != got {
			t.Fatalf("Cache.Get: value at key %d not equal: want %q, got %q", i, v, got)
		}
	}

	if !cache.Add(size, val(size)) {
		t.Fatalf("Cache.Add: value not evicted after limit at %d", size)
	}

	if _, ok := cache.Get(0); ok {
		t.Fatalf("Cache.Get: value at key %d not evicted", 0)
	}

	for i := 1; i <= size; i++ {
		got, ok := cache.Get(i)
		if !ok {
			t.Fatalf("Cache.Get: value not found at key %d", i)
		}

		want := val(i)

		if want != got {
			t.Fatalf("Cache.Get: value at key %d not equal: want %q, got %q", i, want, got)
		}
	}
}
