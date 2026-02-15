package layout

import (
	"reflect"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
)

// TODO: translate more tests from Ratatui.

func TestStrengthIsValid(t *testing.T) {
	t.Parallel()

	assert := func(ok bool) {
		t.Helper()

		if !ok {
			t.Error("invalid strength")
		}
	}

	// Ensures that the constants are defined in the correct order of priority.

	assert(spacerSizeEq > maxSizeLTE)
	assert(maxSizeLTE > maxSizeEq)
	assert(minSizeGTE == maxSizeLTE)
	assert(maxSizeLTE > lengthSizeEq)
	assert(lengthSizeEq > percentageSizeEq)
	assert(percentageSizeEq > ratioSizeEq)
	assert(ratioSizeEq > maxSizeEq)
	assert(minSizeGTE > fillGrow)
	assert(fillGrow > grow)
	assert(grow > spaceGrow)
	assert(spaceGrow > allSegmentGrow)
}

type LayoutSplitTestCase struct {
	Name        string
	Flex        Flex
	Width       int
	Constraints []Constraint
	Want        string
}

func (tc LayoutSplitTestCase) Test(t *testing.T) {
	t.Helper()

	t.Parallel()

	letters(t, tc.Flex, tc.Constraints, tc.Width, tc.Want)
}

func TestLength(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{
			Name:        "width 1 zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(2)},
			Want:        "a",
		},
		{
			Name:        "width 2 zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(3)},
			Want:        "aa",
		},
		{
			Name:        "width 1 zero zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(0), Len(0)},
			Want:        "b",
		},
		{
			Name:        "width 1 zero exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(0), Len(1)},
			Want:        "b",
		},
		{
			Name:        "width 1 zero overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(0), Len(2)},
			Want:        "b",
		},
		{
			Name:        "width 1 exact zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(1), Len(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(1), Len(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(1), Len(2)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(2), Len(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(2), Len(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Len(2), Len(2)},
			Want:        "a",
		},
		{
			Name:        "width 2 zero zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(0), Len(0)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(0), Len(1)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(0), Len(2)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(0), Len(3)},
			Want:        "bb",
		},
		{
			Name:        "width 2 underflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(1), Len(0)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(1), Len(1)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(1), Len(2)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(1), Len(3)},
			Want:        "ab",
		},
		{
			Name:        "width 2 exact zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(2), Len(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(2), Len(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(2), Len(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(2), Len(3)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(3), Len(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(3), Len(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(3), Len(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Len(3), Len(3)},
			Want:        "aa",
		},
		{
			Name:        "width 3 with stretch last",
			Flex:        FlexLegacy,
			Width:       3,
			Constraints: []Constraint{Len(2), Len(2)},
			Want:        "aab",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, tc.Test)
	}
}

func TestMax(t *testing.T) {
	testCases := []LayoutSplitTestCase{
		{
			Name:        "width 1 zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(2)},
			Want:        "a",
		},
		{
			Name:        "width 2 zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(3)},
			Want:        "aa",
		},
		{
			Name:        "width 1 zero zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(0), Max(0)},
			Want:        "b",
		},
		{
			Name:        "width 1 zero exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(0), Max(1)},
			Want:        "b",
		},
		{
			Name:        "width 1 zero overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(0), Max(2)},
			Want:        "b",
		},
		{
			Name:        "width 1 exact zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(1), Max(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(1), Max(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 exact overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(1), Max(2)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(2), Max(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(2), Max(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 overflow overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Max(2), Max(2)},
			Want:        "a",
		},
		{
			Name:        "width 2 zero zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(0), Max(0)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(0), Max(1)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(0), Max(2)},
			Want:        "bb",
		},
		{
			Name:        "width 2 zero overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(0), Max(3)},
			Want:        "bb",
		},
		{
			Name:        "width 2 underflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(1), Max(0)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(1), Max(1)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(1), Max(2)},
			Want:        "ab",
		},
		{
			Name:        "width 2 underflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(1), Max(3)},
			Want:        "ab",
		},
		{
			Name:        "width 2 exact zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(2), Max(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(2), Max(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(2), Max(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 exact overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(2), Max(3)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(3), Max(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(3), Max(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(3), Max(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 overflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Max(3), Max(3)},
			Want:        "aa",
		},
		{
			Name:        "width 3 with stretch last",
			Flex:        FlexLegacy,
			Width:       3,
			Constraints: []Constraint{Max(2), Max(2)},
			Want:        "aab",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, tc.Test)
	}
}

func TestMin(t *testing.T) {
	testCases := []LayoutSplitTestCase{
		{
			Name:        "width 1 min zero zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(0), Min(0)},
			Want:        "b",
		},
		{
			Name:        "width 1 min zero exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(0), Min(1)},
			Want:        "b",
		},
		{
			Name:        "width 1 min zero overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(0), Min(2)},
			Want:        "b",
		},
		{
			Name:        "width 1 min exact zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(1), Min(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 min exact exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(1), Min(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 min exact overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(1), Min(2)},
			Want:        "a",
		},
		{
			Name:        "width 1 min overflow zero",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(2), Min(0)},
			Want:        "a",
		},
		{
			Name:        "width 1 min overflow exact",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(2), Min(1)},
			Want:        "a",
		},
		{
			Name:        "width 1 min overflow overflow",
			Flex:        FlexLegacy,
			Width:       1,
			Constraints: []Constraint{Min(2), Min(2)},
			Want:        "a",
		},
		{
			Name:        "width 2 min zero zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(0), Min(0)},
			Want:        "bb",
		},
		{
			Name:        "width 2 min zero underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(0), Min(1)},
			Want:        "bb",
		},
		{
			Name:        "width 2 min zero exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(0), Min(2)},
			Want:        "bb",
		},
		{
			Name:        "width 2 min zero overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(0), Min(3)},
			Want:        "bb",
		},
		{
			Name:        "width 2 min underflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(1), Min(0)},
			Want:        "ab",
		},
		{
			Name:        "width 2 min underflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(1), Min(1)},
			Want:        "ab",
		},
		{
			Name:        "width 2 min underflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(1), Min(2)},
			Want:        "ab",
		},
		{
			Name:        "width 2 min underflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(1), Min(3)},
			Want:        "ab",
		},
		{
			Name:        "width 2 min exact zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(2), Min(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min exact underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(2), Min(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min exact exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(2), Min(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min exact overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(2), Min(3)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min overflow zero",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(3), Min(0)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min overflow underflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(3), Min(1)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min overflow exact",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(3), Min(2)},
			Want:        "aa",
		},
		{
			Name:        "width 2 min overflow overflow",
			Flex:        FlexLegacy,
			Width:       2,
			Constraints: []Constraint{Min(3), Min(3)},
			Want:        "aa",
		},
		{
			Name:        "width 3 min with stretch last",
			Flex:        FlexLegacy,
			Width:       3,
			Constraints: []Constraint{Min(2), Min(2)},
			Want:        "aab",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, tc.Test)
	}
}

func TestPercentageFlexStart(t *testing.T) {
	testCases := []LayoutSplitTestCase{
		{
			Name:        "Flex Start with Percentage 0, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(0)},
			Want:        "          ",
		},
		{
			Name:        "Flex Start with Percentage 0, 25",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(25)},
			Want:        "bbb       ",
		},
		{
			Name:        "Flex Start with Percentage 0, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(50)},
			Want:        "bbbbb     ",
		},
		{
			Name:        "Flex Start with Percentage 0, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(100)},
			Want:        "bbbbbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 0, 200",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(200)},
			Want:        "bbbbbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 10, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(0)},
			Want:        "a         ",
		},
		{
			Name:        "Flex Start with Percentage 10, 25",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(25)},
			Want:        "abbb      ",
		},
		{
			Name:        "Flex Start with Percentage 10, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(50)},
			Want:        "abbbbb    ",
		},
		{
			Name:        "Flex Start with Percentage 10, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(100)},
			Want:        "abbbbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 10, 200",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(200)},
			Want:        "abbbbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 25, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(0)},
			Want:        "aaa       ",
		},
		{
			Name:        "Flex Start with Percentage 25, 25",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(25)},
			Want:        "aaabb     ",
		},
		{
			Name:        "Flex Start with Percentage 25, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(50)},
			Want:        "aaabbbbb  ",
		},
		{
			Name:        "Flex Start with Percentage 25, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(100)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 25, 200",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(200)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 33, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(0)},
			Want:        "aaa       ",
		},
		{
			Name:        "Flex Start with Percentage 33, 25",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(25)},
			Want:        "aaabbb    ",
		},
		{
			Name:        "Flex Start with Percentage 33, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(50)},
			Want:        "aaabbbbb  ",
		},
		{
			Name:        "Flex Start with Percentage 33, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(100)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 33, 200",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(200)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex Start with Percentage 50, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(0)},
			Want:        "aaaaa     ",
		},
		{
			Name:        "Flex Start with Percentage 50, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(50)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex Start with Percentage 50, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(100)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex Start with Percentage 100, 0",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(0)},
			Want:        "aaaaaaaaaa",
		},
		{
			Name:        "Flex Start with Percentage 100, 50",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(50)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex Start with Percentage 100, 100",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(100)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex Start with Percentage 100, 200",
			Flex:        FlexStart,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(200)},
			Want:        "aaaaabbbbb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, tc.Test)
	}
}

func TestPercentageFlexSpaceBetween(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{
			Name:        "Flex SpaceBetween with Percentage 0, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(0)},
			Want:        "          ",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 0, 25",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(25)},
			Want:        "        bb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 0, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(50)},
			Want:        "     bbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 0, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(100)},
			Want:        "bbbbbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 0, 200",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(0), Percentage(200)},
			Want:        "bbbbbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 10, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(0)},
			Want:        "a         ",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 10, 25",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(25)},
			Want:        "a       bb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 10, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(50)},
			Want:        "a    bbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 10, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(100)},
			Want:        "abbbbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 10, 200",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(10), Percentage(200)},
			Want:        "abbbbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 25, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(0)},
			Want:        "aaa       ",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 25, 25",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(25)},
			Want:        "aaa     bb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 25, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(50)},
			Want:        "aaa  bbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 25, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(100)},
			Want:        "aaabbbbbbb",
		},
		{
			Name: "Flex SpaceBetween with Percentage 25, 200", Flex: FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(25), Percentage(200)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 33, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(0)},
			Want:        "aaa       ",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 33, 25",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(25)},
			Want:        "aaa     bb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 33, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(50)},
			Want:        "aaa  bbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 33, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(100)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 33, 200",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(33), Percentage(200)},
			Want:        "aaabbbbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 50, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(0)},
			Want:        "aaaaa     ",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 50, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(50)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 50, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(50), Percentage(100)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 100, 0",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(0)},
			Want:        "aaaaaaaaaa",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 100, 50",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(50)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 100, 100",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(100)},
			Want:        "aaaaabbbbb",
		},
		{
			Name:        "Flex SpaceBetween with Percentage 100, 200",
			Flex:        FlexSpaceBetween,
			Width:       10,
			Constraints: []Constraint{Percentage(100), Percentage(200)},
			Want:        "aaaaabbbbb",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, tc.Test)
	}
}

type Rect = uv.Rectangle

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints []Constraint
		direction   Direction
		split       Rect
		want        Splitted
	}{
		{
			name: "50% 50% min(0) stretches into last",
			constraints: []Constraint{
				Percentage(50),
				Percentage(50),
				Min(0),
			},
			direction: DirectionVertical,
			split:     uv.Rect(0, 0, 1, 1),
			want: []Rect{
				uv.Rect(0, 0, 1, 1),
				uv.Rect(0, 1, 1, 0),
				uv.Rect(0, 1, 1, 0),
			},
		},
		{
			name: "max(1) 99% min(0) stretches into last",
			constraints: []Constraint{
				Max(1),
				Percentage(99),
				Min(0),
			},
			direction: DirectionVertical,
			split:     uv.Rect(0, 0, 1, 1),
			want: []Rect{
				uv.Rect(0, 0, 1, 0),
				uv.Rect(0, 0, 1, 1),
				uv.Rect(0, 1, 1, 0),
			},
		},
		{
			name: "min(1) length(0) min(1)",
			constraints: []Constraint{
				Min(1),
				Len(0),
				Min(1),
			},
			direction: DirectionHorizontal,
			split:     uv.Rect(0, 0, 1, 1),
			want: []Rect{
				uv.Rect(0, 0, 1, 1),
				uv.Rect(1, 0, 0, 1),
				uv.Rect(1, 0, 0, 1),
			},
		},
		{
			name: "stretches the 2nd last length instead of the last min based on ranking",
			constraints: []Constraint{
				Len(3),
				Min(4),
				Len(1),
				Min(4),
			},
			direction: DirectionHorizontal,
			split:     uv.Rect(0, 0, 7, 1),
			want: []Rect{
				uv.Rect(0, 0, 0, 1),
				uv.Rect(0, 0, 4, 1),
				uv.Rect(4, 0, 0, 1),
				uv.Rect(4, 0, 3, 1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			layout := Layout{
				Constraints: tc.constraints,
				Direction:   tc.direction,
			}.Split(tc.split)

			if !reflect.DeepEqual(tc.want, layout) {
				t.Fatalf("not equal: want %#+v, got %#+v", tc.want, layout)
			}
		})
	}
}

func TestFlexConstraint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints []Constraint
		want        [][]int
		flex        Flex
	}{
		{
			name: "length legacy",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "length start",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "length end",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "length end",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexCenter,
		},
		{
			name: "ratio legacy",
			constraints: []Constraint{
				Ratio{1, 2},
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "ratio start",
			constraints: []Constraint{
				Ratio{1, 2},
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "ratio end",
			constraints: []Constraint{
				Ratio{1, 2},
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "ratio center",
			constraints: []Constraint{
				Ratio{1, 2},
			},
			want: [][]int{{25, 75}},
			flex: FlexCenter,
		},
		{
			name: "percent legacy",
			constraints: []Constraint{
				Percentage(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "percent start",
			constraints: []Constraint{
				Percentage(50),
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "percent end",
			constraints: []Constraint{
				Percentage(50),
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "percent center",
			constraints: []Constraint{
				Percentage(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexCenter,
		},
		{
			name: "min legacy",
			constraints: []Constraint{
				Min(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "min start",
			constraints: []Constraint{
				Min(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexStart,
		},
		{
			name: "min end",
			constraints: []Constraint{
				Min(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexEnd,
		},
		{
			name: "min center",
			constraints: []Constraint{
				Min(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexCenter,
		},
		{
			name: "min legacy",
			constraints: []Constraint{
				Min(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "max legacy",
			constraints: []Constraint{
				Max(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "max start",
			constraints: []Constraint{
				Max(50),
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "max end",
			constraints: []Constraint{
				Max(50),
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "max center",
			constraints: []Constraint{
				Max(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexCenter,
		},
		{
			name: "space between becomes stretch",
			constraints: []Constraint{
				Min(1),
			},
			want: [][]int{{0, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "space between becomes stretch",
			constraints: []Constraint{
				Max(20),
			},
			want: [][]int{{0, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "space between becomes stretch",
			constraints: []Constraint{
				Len(20),
			},
			want: [][]int{{0, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "len legacy 2",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{0, 25}, {25, 100}},
			flex: FlexLegacy,
		},
		{
			name: "len start 2",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{0, 25}, {25, 50}},
			flex: FlexStart,
		},
		{
			name: "len center 2",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{25, 50}, {50, 75}},
			flex: FlexCenter,
		},
		{
			name: "len end 2",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{50, 75}, {75, 100}},
			flex: FlexEnd,
		},
		{
			name: "len space between",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{0, 25}, {75, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "len space evenly",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{17, 42}, {58, 83}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "len space around",
			constraints: []Constraint{
				Len(25), Len(25),
			},
			want: [][]int{{13, 38}, {63, 88}},
			flex: FlexSpaceAround,
		},
		{
			name: "percentage around",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{0, 25}, {25, 100}},
			flex: FlexLegacy,
		},
		{
			name: "percentage start",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{0, 25}, {25, 50}},
			flex: FlexStart,
		},
		{
			name: "percentage center",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{25, 50}, {50, 75}},
			flex: FlexCenter,
		},
		{
			name: "percentage end",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{50, 75}, {75, 100}},
			flex: FlexEnd,
		},
		{
			name: "percentage space between",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{0, 25}, {75, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "percentage space evenly",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{17, 42}, {58, 83}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "percentage space around",
			constraints: []Constraint{
				Percentage(25), Percentage(25),
			},
			want: [][]int{{13, 38}, {63, 88}},
			flex: FlexSpaceAround,
		},
		{
			name: "min legacy 2",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 25}, {25, 100}},
			flex: FlexLegacy,
		},
		{
			name: "min start 2",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexStart,
		},
		{
			name: "min center 2",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexCenter,
		},
		{
			name: "min end 2",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexEnd,
		},
		{
			name: "min space between",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "min space evenly",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "min space around",
			constraints: []Constraint{
				Min(25), Min(25),
			},
			want: [][]int{{0, 50}, {50, 100}},
			flex: FlexSpaceAround,
		},
		{
			name: "max legacy 2",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{0, 25}, {25, 100}},
			flex: FlexLegacy,
		},
		{
			name: "max start 2",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{0, 25}, {25, 50}},
			flex: FlexStart,
		},
		{
			name: "max center 2",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{25, 50}, {50, 75}},
			flex: FlexCenter,
		},
		{
			name: "max end 2",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{50, 75}, {75, 100}},
			flex: FlexEnd,
		},
		{
			name: "max space between",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{0, 25}, {75, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "max space evenly",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{17, 42}, {58, 83}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "max space around",
			constraints: []Constraint{
				Max(25), Max(25),
			},
			want: [][]int{{13, 38}, {63, 88}},
			flex: FlexSpaceAround,
		},
		{
			name: "length spaced around",
			constraints: []Constraint{
				Len(25), Len(25), Len(25),
			},
			want: [][]int{{0, 25}, {38, 63}, {75, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "one segment legacy",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "one segment start",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "one segment end",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "one segment center",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexCenter,
		},
		{
			name: "one segment space between",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "one segment space evenly",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "one segment space around",
			constraints: []Constraint{
				Len(50),
			},
			want: [][]int{{25, 75}},
			flex: FlexSpaceAround,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rect := uv.Rect(0, 0, 100, 1)

			rects := Horizontal(tc.constraints...).WithFlex(tc.flex).Split(rect)

			ranges := make([][]int, 0, len(rects))

			for _, r := range rects {
				ranges = append(ranges, []int{r.Min.X, r.Max.X})
			}

			if !reflect.DeepEqual(tc.want, ranges) {
				t.Fatalf("not equal: want %#+v, got %#+v", tc.want, ranges)
			}
		})
	}
}

func letters(t *testing.T, flex Flex, constraints []Constraint, width int, expected string) {
	t.Helper()

	area := uv.Rect(0, 0, width, 1)

	layout := Layout{
		Direction:   DirectionHorizontal,
		Constraints: constraints,
		Flex:        flex,
		Spacing:     SpacingSpace(0),
	}.Split(area)

	got := uv.NewScreenBuffer(area.Dx(), area.Dy())

	latin := []rune("abcdefghijklmnopqrstuvwxyz")

	for i := 0; i < min(len(constraints), len(layout)); i++ {
		c := latin[i]
		area := layout[i]

		s := strings.Repeat(string(c), area.Dx())

		buffer := uv.NewScreenBuffer(area.Dx(), area.Dy())

		screen.NewContext(buffer).WriteString(s)

		buffer.Draw(got, area)
	}

	want := newBufferString(expected)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("not equal: want %#+v, got %#+v", want, got)
	}
}

func newBufferString(s string) uv.ScreenBuffer {
	var width, height int

	for line := range strings.Lines(s) {
		width = max(width, len(line))
		height++
	}

	buf := uv.NewScreenBuffer(width, height)

	screen.NewContext(buf).WriteString(s)

	return buf
}
