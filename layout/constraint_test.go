package layout

import (
	"math"
	"testing"
)

func TestConstraint_Apply(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		constraint Constraint
		size       int
		want       int
	}{
		{Percent(0), 100, 0},
		{Percent(50), 100, 50},
		{Percent(100), 100, 100},
		{Percent(200), 100, 100},
		{Percent(math.MaxInt), 100, 100},

		{Ratio{0, 0}, 100, 0},
		{Ratio{1, 0}, 100, 100},
		{Ratio{0, 1}, 100, 0},
		{Ratio{1, 2}, 100, 50},
		{Ratio{2, 2}, 100, 100},
		{Ratio{3, 2}, 100, 100},
		{Ratio{math.MaxInt, 2}, 100, 100},

		{Len(0), 100, 0},
		{Len(50), 100, 50},
		{Len(100), 100, 100},
		{Len(200), 100, 100},
		{Len(math.MaxInt), 100, 100},

		{Max(0), 100, 0},
		{Max(50), 100, 50},
		{Max(100), 100, 100},
		{Max(200), 100, 100},
		{Max(math.MaxInt), 100, 100},
		{Min(0), 100, 100},
		{Min(50), 100, 100},
		{Min(100), 100, 100},
		{Min(200), 100, 200},
		{Min(math.MaxInt), 100, math.MaxInt},
	} {
		got := tt.constraint.Apply(tt.size)

		if tt.want != got {
			t.Errorf("%s.Apply(%v): want %v, got %v", tt.constraint, tt.size, tt.want, got)
		}
	}
}
