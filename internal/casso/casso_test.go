package casso

import "testing"

func assertEqual(t *testing.T, expected, actual float64) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSymbol(t *testing.T) {
	v := next(external)
	if v.isZero() {
		t.Fatal("external symbol should not be zero")
	}
	if v.kind() != external {
		t.Fatalf("expected external, got %v", v.kind())
	}

	v = next(slack)
	if v.isZero() {
		t.Fatal("slack symbol should not be zero")
	}
	if v.kind() != slack {
		t.Fatalf("expected slack, got %v", v.kind())
	}

	v = next(errorSym)
	if v.isZero() {
		t.Fatal("error symbol should not be zero")
	}
	if v.kind() != errorSym {
		t.Fatalf("expected errorSym, got %v", v.kind())
	}

	v = next(dummy)
	if v.isZero() {
		t.Fatal("dummy symbol should not be zero")
	}
	if v.kind() != dummy {
		t.Fatalf("expected dummy, got %v", v.kind())
	}
}

func TestConstraint(t *testing.T) {
	l := New()
	m := New()
	r := New()

	a := NewConstraint(EQ, 0, r.T(1), l.T(1), m.T(-2))
	b := NewConstraint(GTE, -100, r.T(1), l.T(-1))
	c := NewConstraint(GTE, 0, l.T(1))

	s := NewSolver()

	_, err := s.Add(1e9, a)
	assertNoError(t, err)

	_, err = s.Add(1e9, b)
	assertNoError(t, err)

	_, err = s.Add(1e9, c)
	assertNoError(t, err)

	assertEqual(t, 0, s.Val(l))
	assertEqual(t, 50, s.Val(m))
	assertEqual(t, 100, s.Val(r))
}

func TestConstraintRequiringArtificialVariable(t *testing.T) {
	s := NewSolver()

	p1 := New()
	p2 := New()
	p3 := New()
	container := New()

	_, err := s.Add(1e9, NewConstraint(EQ, -100, container.T(1)))
	assertNoError(t, err)

	_, err = s.Add(1e6, NewConstraint(GTE, -30, p1.T(1)))
	assertNoError(t, err)

	_, err = s.Add(1e3, NewConstraint(EQ, 0, p1.T(1), p3.T(-1)))
	assertNoError(t, err)

	_, err = s.Add(1e9, NewConstraint(EQ, 0, p2.T(1), p1.T(-2)))
	assertNoError(t, err)

	_, err = s.Add(1e9, NewConstraint(EQ, 0, container.T(1), p1.T(-1), p2.T(-1), p3.T(-1)))
	assertNoError(t, err)

	assertEqual(t, 30, s.Val(p1))
	assertEqual(t, 60, s.Val(p2))
	assertEqual(t, 10, s.Val(p3))
	assertEqual(t, 100, s.Val(container))
}
