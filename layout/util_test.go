package layout

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestSplitVertical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		area           uv.Rectangle
		constraint     Constraint
		expectedTop    uv.Rectangle
		expectedBottom uv.Rectangle
	}{
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 200}},
			Percent(50),
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 100}},
			uv.Rectangle{Min: uv.Position{X: 0, Y: 100}, Max: uv.Position{X: 100, Y: 200}},
		},
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 200}},
			Len(80),
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 80}},
			uv.Rectangle{Min: uv.Position{X: 0, Y: 80}, Max: uv.Position{X: 100, Y: 200}},
		},
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 200}},
			Percent(150), // Edge case: percent greater than 100
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 200}},
			uv.Rectangle{Min: uv.Position{X: 0, Y: 200}, Max: uv.Position{X: 100, Y: 200}},
		},
	}

	for _, test := range tests {
		top, bottom := SplitVertical(test.area, test.constraint)
		if top != test.expectedTop {
			t.Errorf("SplitVertical(%v, %v) top = %v; want %v", test.area, test.constraint, top, test.expectedTop)
		}
		if bottom != test.expectedBottom {
			t.Errorf("SplitVertical(%v, %v) bottom = %v; want %v", test.area, test.constraint, bottom, test.expectedBottom)
		}
	}
}

func TestSplitHorizontal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		area          uv.Rectangle
		constraint    Constraint
		expectedLeft  uv.Rectangle
		expectedRight uv.Rectangle
	}{
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
			Percent(50),
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 100, Y: 100}},
			uv.Rectangle{Min: uv.Position{X: 100, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
		},
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
			Len(80),
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 80, Y: 100}},
			uv.Rectangle{Min: uv.Position{X: 80, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
		},
		{
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
			Percent(150), // Edge case: percent greater than 100
			uv.Rectangle{Min: uv.Position{X: 0, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
			uv.Rectangle{Min: uv.Position{X: 200, Y: 0}, Max: uv.Position{X: 200, Y: 100}},
		},
	}

	for _, test := range tests {
		left, right := SplitHorizontal(test.area, test.constraint)
		if left != test.expectedLeft {
			t.Errorf("SplitHorizontal(%v, %v) left = %v; want %v", test.area, test.constraint, left, test.expectedLeft)
		}
		if right != test.expectedRight {
			t.Errorf("SplitHorizontal(%v, %v) right = %v; want %v", test.area, test.constraint, right, test.expectedRight)
		}
	}
}
