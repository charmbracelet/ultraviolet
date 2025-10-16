package uv

import "testing"

func TestRatio(t *testing.T) {
	tests := []struct {
		numerator, denominator int
		expected               Percent
	}{
		{1, 2, 50},
		{1, 4, 25},
		{3, 4, 75},
		{0, 1, 0},
		{1, 0, 0}, // Edge case: denominator is zero
		{5, 5, 100},
		{2, 3, 66}, // Integer division
	}

	for _, test := range tests {
		result := Ratio(test.numerator, test.denominator)
		if result != test.expected {
			t.Errorf("Ratio(%d, %d) = %d; want %d", test.numerator, test.denominator, result, test.expected)
		}
	}
}

func TestPercentApply(t *testing.T) {
	tests := []struct {
		percent  Percent
		size     int
		expected int
	}{
		{50, 200, 100},
		{25, 400, 100},
		{75, 800, 600},
		{0, 100, 0},
		{100, 100, 100},
		{-10, 100, 0},   // Edge case: negative percent
		{150, 100, 100}, // Edge case: percent greater than 100
	}

	for _, test := range tests {
		result := test.percent.Apply(test.size)
		if result != test.expected {
			t.Errorf("Percent(%d).Apply(%d) = %d; want %d", test.percent, test.size, result, test.expected)
		}
	}
}

func TestFixedApply(t *testing.T) {
	tests := []struct {
		fixed    Fixed
		size     int
		expected int
	}{
		{50, 200, 50},
		{150, 200, 150},
		{250, 200, 200}, // Edge case: fixed size greater than available size
		{0, 100, 0},
		{-10, 100, 0}, // Edge case: negative fixed size
	}

	for _, test := range tests {
		result := test.fixed.Apply(test.size)
		if result != test.expected {
			t.Errorf("Fixed(%d).Apply(%d) = %d; want %d", test.fixed, test.size, result, test.expected)
		}
	}
}

func TestSplitVertical(t *testing.T) {
	tests := []struct {
		area           Rectangle
		constraint     Constraint
		expectedTop    Rectangle
		expectedBottom Rectangle
	}{
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 200}},
			Percent(50),
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 100}},
			Rectangle{Min: Position{X: 0, Y: 100}, Max: Position{X: 100, Y: 200}},
		},
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 200}},
			Fixed(80),
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 80}},
			Rectangle{Min: Position{X: 0, Y: 80}, Max: Position{X: 100, Y: 200}},
		},
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 200}},
			Percent(150), // Edge case: percent greater than 100
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 200}},
			Rectangle{Min: Position{X: 0, Y: 200}, Max: Position{X: 100, Y: 200}},
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
	tests := []struct {
		area          Rectangle
		constraint    Constraint
		expectedLeft  Rectangle
		expectedRight Rectangle
	}{
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 200, Y: 100}},
			Percent(50),
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 100, Y: 100}},
			Rectangle{Min: Position{X: 100, Y: 0}, Max: Position{X: 200, Y: 100}},
		},
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 200, Y: 100}},
			Fixed(80),
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 80, Y: 100}},
			Rectangle{Min: Position{X: 80, Y: 0}, Max: Position{X: 200, Y: 100}},
		},
		{
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 200, Y: 100}},
			Percent(150), // Edge case: percent greater than 100
			Rectangle{Min: Position{X: 0, Y: 0}, Max: Position{X: 200, Y: 100}},
			Rectangle{Min: Position{X: 200, Y: 0}, Max: Position{X: 200, Y: 100}},
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
