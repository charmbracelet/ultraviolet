package layout

// Position defines how a component should be positioned within its area. A
// position is a value between 0 and 100, where 0 represents the start of the
// area and
// 100 represents the end.
type Position = int

// Predefined positions for components. Some positions are special cases that
// represent the left, top, right, bottom.
const (
	Start  Position = 0
	Left   Position = 1
	Top    Position = 1
	Center Position = 50
	Right  Position = 99
	Bottom Position = 99
	End    Position = 100
)
