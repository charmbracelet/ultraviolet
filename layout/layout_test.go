package layout

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
)

func TestPriorityIsValid(t *testing.T) {
	t.Parallel()

	assert := func(ok bool) {
		t.Helper()

		if !ok {
			t.Error("invalid priority")
		}
	}

	// Ensures that the constants are defined in the correct order of priority.

	assert(spacerSizeEq > maxSizeLTE)
	assert(maxSizeLTE > maxSizeEq)
	assert(minSizeGTE == maxSizeLTE)
	assert(maxSizeLTE > lengthSizeEq)
	assert(lengthSizeEq > percentSizeEq)
	assert(percentSizeEq > ratioSizeEq)
	assert(ratioSizeEq > maxSizeEq)
	assert(minSizeGTE > fillGrow)
	assert(fillGrow > grow)
	assert(grow > spaceGrow)
	assert(spaceGrow > allSegmentGrow)
}

type LayoutSplitTestCase struct {
	Flex        Flex
	Width       int
	Constraints []Constraint
	Want        string
}

func (tc LayoutSplitTestCase) Name() string {
	return fmt.Sprintf("Flex(%s) Width(%d) Constraints(%s)", tc.Flex, tc.Width, tc.Constraints)
}

func (tc LayoutSplitTestCase) Test(t *testing.T) {
	t.Helper()

	t.Parallel()

	letters(t, tc.Flex, tc.Constraints, tc.Width, tc.Want)
}

func TestLength(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{FlexLegacy, 1, []Constraint{Len(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(2)}, "a"},
		{FlexLegacy, 2, []Constraint{Len(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(3)}, "aa"},
		{FlexLegacy, 1, []Constraint{Len(0), Len(0)}, "b"},
		{FlexLegacy, 1, []Constraint{Len(0), Len(1)}, "b"},
		{FlexLegacy, 1, []Constraint{Len(0), Len(2)}, "b"},
		{FlexLegacy, 1, []Constraint{Len(1), Len(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(1), Len(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(1), Len(2)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(2), Len(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(2), Len(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Len(2), Len(2)}, "a"},
		{FlexLegacy, 2, []Constraint{Len(0), Len(0)}, "bb"},
		{FlexLegacy, 2, []Constraint{Len(0), Len(1)}, "bb"},
		{FlexLegacy, 2, []Constraint{Len(0), Len(2)}, "bb"},
		{FlexLegacy, 2, []Constraint{Len(0), Len(3)}, "bb"},
		{FlexLegacy, 2, []Constraint{Len(1), Len(0)}, "ab"},
		{FlexLegacy, 2, []Constraint{Len(1), Len(1)}, "ab"},
		{FlexLegacy, 2, []Constraint{Len(1), Len(2)}, "ab"},
		{FlexLegacy, 2, []Constraint{Len(1), Len(3)}, "ab"},
		{FlexLegacy, 2, []Constraint{Len(2), Len(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(2), Len(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(2), Len(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(2), Len(3)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(3), Len(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(3), Len(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(3), Len(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Len(3), Len(3)}, "aa"},
		{FlexLegacy, 3, []Constraint{Len(2), Len(2)}, "aab"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{FlexLegacy, 1, []Constraint{Max(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(2)}, "a"},
		{FlexLegacy, 2, []Constraint{Max(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(3)}, "aa"},
		{FlexLegacy, 1, []Constraint{Max(0), Max(0)}, "b"},
		{FlexLegacy, 1, []Constraint{Max(0), Max(1)}, "b"},
		{FlexLegacy, 1, []Constraint{Max(0), Max(2)}, "b"},
		{FlexLegacy, 1, []Constraint{Max(1), Max(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(1), Max(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(1), Max(2)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(2), Max(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(2), Max(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Max(2), Max(2)}, "a"},
		{FlexLegacy, 2, []Constraint{Max(0), Max(0)}, "bb"},
		{FlexLegacy, 2, []Constraint{Max(0), Max(1)}, "bb"},
		{FlexLegacy, 2, []Constraint{Max(0), Max(2)}, "bb"},
		{FlexLegacy, 2, []Constraint{Max(0), Max(3)}, "bb"},
		{FlexLegacy, 2, []Constraint{Max(1), Max(0)}, "ab"},
		{FlexLegacy, 2, []Constraint{Max(1), Max(1)}, "ab"},
		{FlexLegacy, 2, []Constraint{Max(1), Max(2)}, "ab"},
		{FlexLegacy, 2, []Constraint{Max(1), Max(3)}, "ab"},
		{FlexLegacy, 2, []Constraint{Max(2), Max(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(2), Max(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(2), Max(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(2), Max(3)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(3), Max(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(3), Max(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(3), Max(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Max(3), Max(3)}, "aa"},
		{FlexLegacy, 3, []Constraint{Max(2), Max(2)}, "aab"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{FlexLegacy, 1, []Constraint{Min(0), Min(0)}, "b"},
		{FlexLegacy, 1, []Constraint{Min(0), Min(1)}, "b"},
		{FlexLegacy, 1, []Constraint{Min(0), Min(2)}, "b"},
		{FlexLegacy, 1, []Constraint{Min(1), Min(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Min(1), Min(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Min(1), Min(2)}, "a"},
		{FlexLegacy, 1, []Constraint{Min(2), Min(0)}, "a"},
		{FlexLegacy, 1, []Constraint{Min(2), Min(1)}, "a"},
		{FlexLegacy, 1, []Constraint{Min(2), Min(2)}, "a"},
		{FlexLegacy, 2, []Constraint{Min(0), Min(0)}, "bb"},
		{FlexLegacy, 2, []Constraint{Min(0), Min(1)}, "bb"},
		{FlexLegacy, 2, []Constraint{Min(0), Min(2)}, "bb"},
		{FlexLegacy, 2, []Constraint{Min(0), Min(3)}, "bb"},
		{FlexLegacy, 2, []Constraint{Min(1), Min(0)}, "ab"},
		{FlexLegacy, 2, []Constraint{Min(1), Min(1)}, "ab"},
		{FlexLegacy, 2, []Constraint{Min(1), Min(2)}, "ab"},
		{FlexLegacy, 2, []Constraint{Min(1), Min(3)}, "ab"},
		{FlexLegacy, 2, []Constraint{Min(2), Min(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(2), Min(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(2), Min(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(2), Min(3)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(3), Min(0)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(3), Min(1)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(3), Min(2)}, "aa"},
		{FlexLegacy, 2, []Constraint{Min(3), Min(3)}, "aa"},
		{FlexLegacy, 3, []Constraint{Min(2), Min(2)}, "aab"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestPercent(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{FlexStart, 10, []Constraint{Percent(0), Percent(0)}, "          "},
		{FlexStart, 10, []Constraint{Percent(0), Percent(25)}, "bbb       "},
		{FlexStart, 10, []Constraint{Percent(0), Percent(50)}, "bbbbb     "},
		{FlexStart, 10, []Constraint{Percent(0), Percent(100)}, "bbbbbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(0), Percent(200)}, "bbbbbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(10), Percent(0)}, "a         "},
		{FlexStart, 10, []Constraint{Percent(10), Percent(25)}, "abbb      "},
		{FlexStart, 10, []Constraint{Percent(10), Percent(50)}, "abbbbb    "},
		{FlexStart, 10, []Constraint{Percent(10), Percent(100)}, "abbbbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(10), Percent(200)}, "abbbbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(25), Percent(0)}, "aaa       "},
		{FlexStart, 10, []Constraint{Percent(25), Percent(25)}, "aaabb     "},
		{FlexStart, 10, []Constraint{Percent(25), Percent(50)}, "aaabbbbb  "},
		{FlexStart, 10, []Constraint{Percent(25), Percent(100)}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(25), Percent(200)}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(33), Percent(0)}, "aaa       "},
		{FlexStart, 10, []Constraint{Percent(33), Percent(25)}, "aaabbb    "},
		{FlexStart, 10, []Constraint{Percent(33), Percent(50)}, "aaabbbbb  "},
		{FlexStart, 10, []Constraint{Percent(33), Percent(100)}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(33), Percent(200)}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Percent(50), Percent(0)}, "aaaaa     "},
		{FlexStart, 10, []Constraint{Percent(50), Percent(50)}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Percent(50), Percent(100)}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Percent(100), Percent(0)}, "aaaaaaaaaa"},
		{FlexStart, 10, []Constraint{Percent(100), Percent(50)}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Percent(100), Percent(100)}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Percent(100), Percent(200)}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(0), Percent(0)}, "          "},
		{FlexSpaceBetween, 10, []Constraint{Percent(0), Percent(25)}, "        bb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(0), Percent(50)}, "     bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(0), Percent(100)}, "bbbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(0), Percent(200)}, "bbbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(10), Percent(0)}, "a         "},
		{FlexSpaceBetween, 10, []Constraint{Percent(10), Percent(25)}, "a       bb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(10), Percent(50)}, "a    bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(10), Percent(100)}, "abbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(10), Percent(200)}, "abbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(25), Percent(0)}, "aaa       "},
		{FlexSpaceBetween, 10, []Constraint{Percent(25), Percent(25)}, "aaa     bb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(25), Percent(50)}, "aaa  bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(25), Percent(100)}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Percent(25), Percent(200)}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(33), Percent(0)}, "aaa       "},
		{FlexSpaceBetween, 10, []Constraint{Percent(33), Percent(25)}, "aaa     bb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(33), Percent(50)}, "aaa  bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(33), Percent(100)}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(33), Percent(200)}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(50), Percent(0)}, "aaaaa     "},
		{FlexSpaceBetween, 10, []Constraint{Percent(50), Percent(50)}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(50), Percent(100)}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(100), Percent(0)}, "aaaaaaaaaa"},
		{FlexSpaceBetween, 10, []Constraint{Percent(100), Percent(50)}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(100), Percent(100)}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Percent(100), Percent(200)}, "aaaaabbbbb"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
	}
}

func TestRatio(t *testing.T) {
	t.Parallel()

	testCases := []LayoutSplitTestCase{
		{FlexLegacy, 1, []Constraint{Ratio{0, 1}}, "a"},
		{FlexLegacy, 1, []Constraint{Ratio{0, 1}}, "a"},
		{FlexLegacy, 2, []Constraint{Ratio{0, 1}}, "aa"},
		{FlexLegacy, 10, []Constraint{Ratio{0, 1}, Ratio{0, 1}}, "bbbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{0, 1}, Ratio{1, 4}}, "bbbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{0, 1}, Ratio{1, 2}}, "bbbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{0, 1}, Ratio{1, 1}}, "bbbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{0, 1}, Ratio{2, 1}}, "bbbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 10}, Ratio{0, 1}}, "abbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 10}, Ratio{1, 4}}, "abbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 10}, Ratio{1, 2}}, "abbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 10}, Ratio{1, 1}}, "abbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 10}, Ratio{2, 1}}, "abbbbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 4}, Ratio{0, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 4}, Ratio{1, 4}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 4}, Ratio{1, 2}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 4}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 4}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 3}, Ratio{0, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 3}, Ratio{1, 4}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 3}, Ratio{1, 2}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 3}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 3}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 2}, Ratio{0, 1}}, "aaaaabbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 2}, Ratio{1, 2}}, "aaaaabbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 2}, Ratio{1, 1}}, "aaaaabbbbb"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 1}, Ratio{0, 1}}, "aaaaaaaaaa"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 1}, Ratio{1, 2}}, "aaaaaaaaaa"},
		{FlexLegacy, 10, []Constraint{Ratio{1, 1}, Ratio{1, 1}}, "aaaaaaaaaa"},
		{FlexStart, 10, []Constraint{Ratio{0, 1}, Ratio{0, 1}}, "          "},
		{FlexStart, 10, []Constraint{Ratio{0, 1}, Ratio{1, 4}}, "bbb       "},
		{FlexStart, 10, []Constraint{Ratio{0, 1}, Ratio{1, 2}}, "bbbbb     "},
		{FlexStart, 10, []Constraint{Ratio{0, 1}, Ratio{1, 1}}, "bbbbbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{0, 1}, Ratio{2, 1}}, "bbbbbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 10}, Ratio{0, 1}}, "a         "},
		{FlexStart, 10, []Constraint{Ratio{1, 10}, Ratio{1, 4}}, "abbb      "},
		{FlexStart, 10, []Constraint{Ratio{1, 10}, Ratio{1, 2}}, "abbbbb    "},
		{FlexStart, 10, []Constraint{Ratio{1, 10}, Ratio{1, 1}}, "abbbbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 10}, Ratio{2, 1}}, "abbbbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 4}, Ratio{0, 1}}, "aaa       "},
		{FlexStart, 10, []Constraint{Ratio{1, 4}, Ratio{1, 4}}, "aaabb     "},
		{FlexStart, 10, []Constraint{Ratio{1, 4}, Ratio{1, 2}}, "aaabbbbb  "},
		{FlexStart, 10, []Constraint{Ratio{1, 4}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 4}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 3}, Ratio{0, 1}}, "aaa       "},
		{FlexStart, 10, []Constraint{Ratio{1, 3}, Ratio{1, 4}}, "aaabbb    "},
		{FlexStart, 10, []Constraint{Ratio{1, 3}, Ratio{1, 2}}, "aaabbbbb  "},
		{FlexStart, 10, []Constraint{Ratio{1, 3}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 3}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 2}, Ratio{0, 1}}, "aaaaa     "},
		{FlexStart, 10, []Constraint{Ratio{1, 2}, Ratio{1, 2}}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 2}, Ratio{1, 1}}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 1}, Ratio{0, 1}}, "aaaaaaaaaa"},
		{FlexStart, 10, []Constraint{Ratio{1, 1}, Ratio{1, 2}}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 1}, Ratio{1, 1}}, "aaaaabbbbb"},
		{FlexStart, 10, []Constraint{Ratio{1, 1}, Ratio{2, 1}}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{0, 1}, Ratio{0, 1}}, "          "},
		{FlexSpaceBetween, 10, []Constraint{Ratio{0, 1}, Ratio{1, 4}}, "        bb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{0, 1}, Ratio{1, 2}}, "     bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{0, 1}, Ratio{1, 1}}, "bbbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{0, 1}, Ratio{2, 1}}, "bbbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 10}, Ratio{0, 1}}, "a         "},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 10}, Ratio{1, 4}}, "a       bb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 10}, Ratio{1, 2}}, "a    bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 10}, Ratio{1, 1}}, "abbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 10}, Ratio{2, 1}}, "abbbbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 4}, Ratio{0, 1}}, "aaa       "},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 4}, Ratio{1, 4}}, "aaa     bb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 4}, Ratio{1, 2}}, "aaa  bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 4}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 4}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 3}, Ratio{0, 1}}, "aaa       "},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 3}, Ratio{1, 4}}, "aaa     bb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 3}, Ratio{1, 2}}, "aaa  bbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 3}, Ratio{1, 1}}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 3}, Ratio{2, 1}}, "aaabbbbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 2}, Ratio{0, 1}}, "aaaaa     "},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 2}, Ratio{1, 2}}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 2}, Ratio{1, 1}}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 1}, Ratio{0, 1}}, "aaaaaaaaaa"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 1}, Ratio{1, 2}}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 1}, Ratio{1, 1}}, "aaaaabbbbb"},
		{FlexSpaceBetween, 10, []Constraint{Ratio{1, 1}, Ratio{2, 1}}, "aaaaabbbbb"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name(), tc.Test)
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
				Percent(50),
				Percent(50),
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
				Percent(99),
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
				Percent(50),
			},
			want: [][]int{{0, 100}},
			flex: FlexLegacy,
		},
		{
			name: "percent start",
			constraints: []Constraint{
				Percent(50),
			},
			want: [][]int{{0, 50}},
			flex: FlexStart,
		},
		{
			name: "percent end",
			constraints: []Constraint{
				Percent(50),
			},
			want: [][]int{{50, 100}},
			flex: FlexEnd,
		},
		{
			name: "percent center",
			constraints: []Constraint{
				Percent(50),
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
				Percent(25), Percent(25),
			},
			want: [][]int{{0, 25}, {25, 100}},
			flex: FlexLegacy,
		},
		{
			name: "percentage start",
			constraints: []Constraint{
				Percent(25), Percent(25),
			},
			want: [][]int{{0, 25}, {25, 50}},
			flex: FlexStart,
		},
		{
			name: "percentage center",
			constraints: []Constraint{
				Percent(25), Percent(25),
			},
			want: [][]int{{25, 50}, {50, 75}},
			flex: FlexCenter,
		},
		{
			name: "percentage end",
			constraints: []Constraint{
				Percent(25), Percent(25),
			},
			want: [][]int{{50, 75}, {75, 100}},
			flex: FlexEnd,
		},
		{
			name: "percentage space between",
			constraints: []Constraint{
				Percent(25), Percent(25),
			},
			want: [][]int{{0, 25}, {75, 100}},
			flex: FlexSpaceBetween,
		},
		{
			name: "percentage space evenly",
			constraints: []Constraint{
				Percent(25), Percent(25),
			},
			want: [][]int{{17, 42}, {58, 83}},
			flex: FlexSpaceEvenly,
		},
		{
			name: "percentage space around",
			constraints: []Constraint{
				Percent(25), Percent(25),
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

func TestFlexSpacing(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		want        [][]int
		constraints []Constraint
		flex        Flex
		spacing     int
	}{
		{
			name:        "length zero spacing",
			want:        [][]int{{0, 20}, {20, 20}, {40, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexStart,
			spacing:     0,
		},
		{
			name:        "length overlap 2",
			want:        [][]int{{0, 20}, {19, 20}, {38, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexStart,
			spacing:     -1,
		},
		{
			name:        "length overlap 3",
			want:        [][]int{{21, 20}, {40, 20}, {59, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexCenter,
			spacing:     -1,
		},
		{
			name:        "length overlap 4",
			want:        [][]int{{42, 20}, {61, 20}, {80, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexEnd,
			spacing:     -1,
		},
		{
			name:        "length overlap 5",
			want:        [][]int{{0, 20}, {19, 20}, {38, 62}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexLegacy,
			spacing:     -1,
		},
		{
			name:        "length overlap 6",
			want:        [][]int{{0, 20}, {40, 20}, {80, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceBetween,
			spacing:     -1,
		},
		{
			name:        "length overlap 7",
			want:        [][]int{{10, 20}, {40, 20}, {70, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceEvenly,
			spacing:     -1,
		},
		{
			name:        "length overlap 8",
			want:        [][]int{{7, 20}, {40, 20}, {73, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceAround,
			spacing:     -1,
		},

		{
			name:        "length spacing 1",
			want:        [][]int{{0, 20}, {22, 20}, {44, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexStart,
			spacing:     2,
		},
		{
			name:        "length spacing 2",
			want:        [][]int{{18, 20}, {40, 20}, {62, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexCenter,
			spacing:     2,
		},
		{
			name:        "length spacing 3",
			want:        [][]int{{36, 20}, {58, 20}, {80, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexEnd,
			spacing:     2,
		},
		{
			name:        "length spacing 4",
			want:        [][]int{{0, 20}, {22, 20}, {44, 56}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexLegacy,
			spacing:     2,
		},
		{
			name:        "length spacing 5",
			want:        [][]int{{0, 20}, {40, 20}, {80, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceBetween,
			spacing:     2,
		},
		{
			name:        "length spacing 6",
			want:        [][]int{{10, 20}, {40, 20}, {70, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceEvenly,
			spacing:     2,
		},
		{
			name:        "length spacing 7",
			want:        [][]int{{7, 20}, {40, 20}, {73, 20}},
			constraints: []Constraint{Len(20), Len(20), Len(20)},
			flex:        FlexSpaceAround,
			spacing:     2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rect := uv.Rect(0, 0, 100, 1)

			splitted := Horizontal(tc.constraints...).
				WithFlex(tc.flex).
				WithSpacing(tc.spacing).
				Split(rect)

			got := make([][]int, 0, len(splitted))

			for _, r := range splitted {
				got = append(got, []int{r.Min.X, r.Dx()})
			}

			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("not equal: want %#+v, got %#+v", tc.want, got)
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
