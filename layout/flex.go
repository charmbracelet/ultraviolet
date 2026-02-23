package layout

import (
	"fmt"
)

// Flex controls how leftover space is distributed once every segment's
// constraint has been resolved. It is analogous to the CSS
// justify-content property and is used together with [Layout].
type Flex int

const (

	// FlexStart pushes segments to the leading edge of the area, leaving
	// any surplus space at the trailing edge.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌────16 px─────┐┌──────20 px───────┐┌──────20 px───────┐
	// 	│  Percent(20) ││    Length(20)    ││     Fixed(20)    │
	// 	└──────────────┘└──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌──────20 px───────┐┌──────20 px───────┐
	// 	│      Max(20)     ││      Max(20)     │
	// 	└──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌──────20 px───────┐
	// 	│      Max(20)     │
	// 	└──────────────────┘
	FlexStart Flex = iota

	// FlexLegacy fills the entire area by assigning surplus space to the
	// lowest-priority trailing segment. This reproduces the original
	// Ratatui/tui-rs layout behaviour.
	//
	// The examples below show how surplus space is allocated for different
	// constraint combinations. Recall the priority order (highest first):
	//
	// 	- [Min]
	// 	- [Max]
	// 	- [Len]
	// 	- [Percent]
	// 	- [Ratio]
	// 	- [Fill]
	//
	// With all-[Len] constraints the surplus goes to the final segment.
	//
	// 	<----------------------------------- 80 px ------------------------------------>
	// 	┌──────20 px───────┐┌──────20 px───────┐┌────────────────40 px─────────────────┐
	// 	│    Length(20)    ││    Length(20)    ││              Length(20)              │
	// 	└──────────────────┘└──────────────────┘└──────────────────────────────────────┘
	// 	                                        ^^^^^^^^^^^^^^^^ EXCESS ^^^^^^^^^^^^^^^^
	//
	// [Fill] has the lowest priority, so it always absorbs surplus.
	//
	// 	<----------------------------------- 80 px ------------------------------------>
	// 	┌──────20 px───────┐┌──────20 px───────┐┌──────20 px───────┐┌──────20 px───────┐
	// 	│      Fill(0)     ││      Max(20)     ││    Length(20)    ││     Length(20)   │
	// 	└──────────────────┘└──────────────────┘└──────────────────┘└──────────────────┘
	// 	^^^^^^ EXCESS ^^^^^^
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌──────────────────────────60 px───────────────────────────┐┌──────20 px───────┐
	// 	│                          Min(20)                         ││      Max(20)     │
	// 	└──────────────────────────────────────────────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌────────────────────────────────────80 px─────────────────────────────────────┐
	// 	│                                    Max(20)                                   │
	// 	└──────────────────────────────────────────────────────────────────────────────┘
	FlexLegacy

	// FlexEnd pushes segments to the trailing edge of the area, leaving
	// surplus space at the leading edge.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                        ┌────16 px─────┐┌──────20 px───────┐┌──────20 px───────┐
	// 	                        │  Percent(20) ││    Length(20)    ││     Length(20)   │
	// 	                        └──────────────┘└──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                                        ┌──────20 px───────┐┌──────20 px───────┐
	// 	                                        │      Max(20)     ││      Max(20)     │
	// 	                                        └──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                                                            ┌──────20 px───────┐
	// 	                                                            │      Max(20)     │
	// 	                                                            └──────────────────┘
	FlexEnd

	// FlexCenter places segments in the middle of the area, distributing
	// surplus space equally before the first and after the last segment.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	            ┌────16 px─────┐┌──────20 px───────┐┌──────20 px───────┐
	// 	            │  Percent(20) ││    Length(20)    ││     Length(20)   │
	// 	            └──────────────┘└──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                    ┌──────20 px───────┐┌──────20 px───────┐
	// 	                    │      Max(20)     ││      Max(20)     │
	// 	                    └──────────────────┘└──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                              ┌──────20 px───────┐
	// 	                              │      Max(20)     │
	// 	                              └──────────────────┘
	FlexCenter

	// FlexSpaceBetween distributes surplus space equally between adjacent
	// segments, with no space before the first or after the last.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌────16 px─────┐            ┌──────20 px───────┐            ┌──────20 px───────┐
	// 	│  Percent(20) │            │    Length(20)    │            │     Length(20)   │
	// 	└──────────────┘            └──────────────────┘            └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌──────20 px───────┐                                        ┌──────20 px───────┐
	// 	│      Max(20)     │                                        │      Max(20)     │
	// 	└──────────────────┘                                        └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	┌────────────────────────────────────80 px─────────────────────────────────────┐
	// 	│                                    Max(20)                                   │
	// 	└──────────────────────────────────────────────────────────────────────────────┘
	FlexSpaceBetween

	// FlexSpaceEvenly distributes surplus space so that every gap
	// (including before the first and after the last segment) is the same width.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	      ┌────16 px─────┐      ┌──────20 px───────┐      ┌──────20 px───────┐
	// 	      │  Percent(20) │      │    Length(20)    │      │     Length(20)   │
	// 	      └──────────────┘      └──────────────────┘      └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	             ┌──────20 px───────┐              ┌──────20 px───────┐
	// 	             │      Max(20)     │              │      Max(20)     │
	// 	             └──────────────────┘              └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                              ┌──────20 px───────┐
	// 	                              │      Max(20)     │
	// 	                              └──────────────────┘
	FlexSpaceEvenly

	// FlexSpaceAround places equal space on both sides of each segment.
	// Adjacent segments therefore have twice the gap of the outer edges.
	//
	// # Examples
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	    ┌────16 px─────┐       ┌──────20 px───────┐       ┌──────20 px───────┐
	// 	    │  Percent(20) │       │    Length(20)    │       │     Length(20)   │
	// 	    └──────────────┘       └──────────────────┘       └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	     ┌──────20 px───────┐                      ┌──────20 px───────┐
	// 	     │      Max(20)     │                      │      Max(20)     │
	// 	     └──────────────────┘                      └──────────────────┘
	//
	// 	<------------------------------------80 px------------------------------------->
	// 	                              ┌──────20 px───────┐
	// 	                              │      Max(20)     │
	// 	                              └──────────────────┘
	FlexSpaceAround
)

func (f Flex) String() string {
	switch f {
	case FlexCenter:
		return "Center"

	case FlexEnd:
		return "End"

	case FlexLegacy:
		return "Legacy"

	case FlexSpaceAround:
		return "Space Around"

	case FlexSpaceBetween:
		return "Space Between"

	case FlexSpaceEvenly:
		return "Space Evenly"

	case FlexStart:
		return "Start"

	default:
		return fmt.Sprintf("Flex(%d)", f)
	}
}
