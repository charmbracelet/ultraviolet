package casso

import (
	"image"
	"testing"
)

func TestQuadrilateral(t *testing.T) {
	t.Parallel()

	type Point struct {
		X, Y Variable
	}

	type PointValue struct {
		X, Y float64
	}

	newPoint := func() Point { return Point{NewVariable(), NewVariable()} }

	valueOf, updateValues := newValues()

	points := []Point{newPoint(), newPoint(), newPoint(), newPoint()}
	pointStarts := []PointValue{{10, 10}, {10, 200}, {200, 200}, {200, 10}}
	midpoints := []Point{newPoint(), newPoint(), newPoint(), newPoint()}

	solver := NewSolver()

	weight := Strength(1.0)
	multiplier := 2.0

	for i := range 4 {
		err := solver.AddConstraints(
			Equal(Weak*weight).VariableLHS(points[i].X).ConstantRHS(pointStarts[i].X),
			Equal(Weak*weight).VariableLHS(points[i].Y).ConstantRHS(pointStarts[i].Y),
		)
		if err != nil {
			t.Fatalf("failed to add constraints: %v", err)
		}

		weight *= Strength(multiplier)
	}

	for _, p := range []image.Point{{0, 1}, {1, 2}, {2, 3}, {3, 0}} {
		start, end := p.X, p.Y

		err := solver.AddConstraints(
			Equal(Required).
				VariableLHS(midpoints[start].X).
				ExpressionRHS((points[start].X.Add(points[end].X)).DivConstant(2)),

			Equal(Required).
				VariableLHS(midpoints[start].Y).
				ExpressionRHS((points[start].Y.Add(points[end].Y)).DivConstant(2)),
		)
		if err != nil {
			t.Fatalf("failed to add constraints: %v", err)
		}
	}

	err := solver.AddConstraints(
		LessThanEqual(Strong).ExpressionLHS(points[0].X.AddConstant(20)).VariableRHS(points[2].X),
		LessThanEqual(Strong).ExpressionLHS(points[0].X.AddConstant(20)).VariableRHS(points[3].X),
		LessThanEqual(Strong).ExpressionLHS(points[1].X.AddConstant(20)).VariableRHS(points[2].X),
		LessThanEqual(Strong).ExpressionLHS(points[1].X.AddConstant(20)).VariableRHS(points[3].X),
		LessThanEqual(Strong).ExpressionLHS(points[0].Y.AddConstant(20)).VariableRHS(points[1].Y),
		LessThanEqual(Strong).ExpressionLHS(points[0].Y.AddConstant(20)).VariableRHS(points[2].Y),
		LessThanEqual(Strong).ExpressionLHS(points[3].Y.AddConstant(20)).VariableRHS(points[1].Y),
		LessThanEqual(Strong).ExpressionLHS(points[3].Y.AddConstant(20)).VariableRHS(points[2].Y),
	)
	if err != nil {
		t.Fatalf("failed to add constraints: %v", err)
	}

	for _, p := range points {
		err = solver.AddConstraints(
			GreaterThanEqual(Required).VariableLHS(p.X).ConstantRHS(0),
			GreaterThanEqual(Required).VariableLHS(p.Y).ConstantRHS(0),
			LessThanEqual(Required).VariableLHS(p.X).ConstantRHS(500),
			LessThanEqual(Required).VariableLHS(p.Y).ConstantRHS(500),
		)
		if err != nil {
			t.Fatalf("failed to add constraints: %v", err)
		}
	}

	updateValues(solver.FetchChanges())

	want := [4]PointValue{
		{10, 105}, {105, 200}, {200, 105}, {105, 10},
	}

	got := [4]PointValue{
		{valueOf(midpoints[0].X), valueOf(midpoints[0].Y)},
		{valueOf(midpoints[1].X), valueOf(midpoints[1].Y)},
		{valueOf(midpoints[2].X), valueOf(midpoints[2].Y)},
		{valueOf(midpoints[3].X), valueOf(midpoints[3].Y)},
	}

	if want != got {
		t.Fatalf("not equal: want %v, got %v", want, got)
	}
}

type Values map[Variable]float64

type ValueChange struct {
	Variable Variable
	Value    float64
}

func newValues() (
	func(Variable) float64,
	func([]PublicChange),
) {
	values := make(Values)

	valueOf := func(v Variable) float64 {
		return values[v]
	}

	updateValues := func(changes []PublicChange) {
		for _, change := range changes {
			values[change.Variable] = change.Constant
		}
	}

	return valueOf, updateValues
}
