package layout

import (
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/lru"
)

// This is a somewhat arbitrary size for the layout cache based on adding
// the columns and rows (171+51 = 222) and doubling it for good measure and then adding a
// bit more to make it a round number. This gives enough entries to store a layout for every
// row and every column, twice over, which should be enough for most apps.
const globalCacheSize = 500

var (
	globalCache   = lru.New[cacheKey, cacheValue](globalCacheSize)
	globalCacheMu sync.Mutex
)

type cacheKey struct {
	Area            uv.Rectangle
	Direction       Direction
	ConstraintsHash uint64
	Padding         Padding
	Spacing         int
	Flex            Flex
}

type cacheValue struct{ Segments, Spacers Splitted }
