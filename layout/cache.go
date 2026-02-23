package layout

import (
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/internal/lru"
)

// globalCacheSize is chosen to comfortably hold one entry per row and column
// of a typical terminal, with headroom to spare. A 171-column x 51-row
// display yields 222 unique keys; doubling and rounding up gives 500.
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
