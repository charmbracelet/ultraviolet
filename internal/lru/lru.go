package lru

import (
	"container/list"
	"fmt"
)

type LRU[K comparable, V any] struct {
	size int

	items map[K]*list.Element
	evict *list.List
}

type entry[K comparable, T any] struct {
	Key   K
	Value T
}

func New[K comparable, V any](size int) *LRU[K, V] {
	if size < 0 {
		panic(fmt.Sprintf("lru: negative size given: %d", size))
	}

	return &LRU[K, V]{
		size:  size,
		items: make(map[K]*list.Element),
		evict: list.New(),
	}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	e, ok := c.items[key]
	if !ok {
		var zero V

		return zero, false
	}

	c.evict.MoveToFront(e)

	return e.Value.(entry[K, V]).Value, true
}

func (c *LRU[K, V]) Add(key K, value V) (evicted bool) {
	if e, ok := c.items[key]; ok {
		c.evict.MoveToFront(e)
		e.Value = value

		return false
	}

	e := c.evict.PushFront(key)

	e.Value = entry[K, V]{Key: key, Value: value}

	c.items[key] = e

	if c.evict.Len() <= c.size {
		return false
	}

	if oldest := c.evict.Back(); oldest != nil {
		c.evict.Remove(oldest)
		delete(c.items, oldest.Value.(entry[K, V]).Key)
	}

	return true
}
