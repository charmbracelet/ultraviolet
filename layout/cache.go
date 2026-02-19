package layout

import (
	"fmt"

	uv "github.com/charmbracelet/ultraviolet"
	lru "github.com/hashicorp/golang-lru/v2"
)

// This is a somewhat arbitrary size for the layout cache based on adding
// the columns and rows (171+51 = 222) and doubling it for good measure and then adding a
// bit more to make it a round number. This gives enough entries to store a layout for every
// row and every column, twice over, which should be enough for most apps.
const globalCacheSize = 500

var globalCache = newCache(globalCacheSize)

type cacheKey struct {
	Area            uv.Rectangle
	Direction       Direction
	ConstraintsHash uint64
	Padding         Padding
	Spacing         int
	Flex            Flex
}

type cacheValue struct{ Segments, Spacers Splitted }

func newCache(size int) *lru.Cache[cacheKey, cacheValue] {
	cache, err := lru.New[cacheKey, cacheValue](size)
	if err != nil {
		// fails only when given negative size.
		panic(fmt.Sprintf("layout: failed to create lru cache of size %d: %v", size, err))
	}

	return cache
}
