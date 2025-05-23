package tv

import (
	"testing"
)

func TestFixedSize(t *testing.T) {
	size := FixedSize(100)
	if size.Calculate(200) != 100 {
		t.Errorf("Expected 100, got %d", size.Calculate(200))
	}
	if size.IsFlexible() {
		t.Error("FixedSize should not be flexible")
	}
}

func TestRatioSize(t *testing.T) {
	tests := []struct {
		ratio     RatioSize
		available int
		expected  int
	}{
		{RatioSize(0.5), 100, 50},
		{RatioSize(0.25), 200, 50},
		{RatioSize(-0.1), 100, 0}, // Clamped to 0
		{RatioSize(1.5), 100, 100}, // Clamped to 1
	}

	for _, test := range tests {
		result := test.ratio.Calculate(test.available)
		if result != test.expected {
			t.Errorf("RatioSize(%f).Calculate(%d) = %d, expected %d", 
				test.ratio, test.available, result, test.expected)
		}
	}

	if RatioSize(0.5).IsFlexible() {
		t.Error("RatioSize should not be flexible")
	}
}

func TestFlexSize(t *testing.T) {
	flex := FlexSize{Grow: 1, Shrink: 1, Basis: FixedSize(50)}
	if flex.Calculate(100) != 50 {
		t.Errorf("Expected 50, got %d", flex.Calculate(100))
	}
	if !flex.IsFlexible() {
		t.Error("FlexSize should be flexible")
	}
}

func TestMinMaxSize(t *testing.T) {
	constraint := MinMaxSize{
		Constraint: FixedSize(100),
		Min:        50,
		Max:        150,
	}

	if constraint.Calculate(200) != 100 {
		t.Errorf("Expected 100, got %d", constraint.Calculate(200))
	}

	// Test with constraint that would be too small
	smallConstraint := MinMaxSize{
		Constraint: FixedSize(25),
		Min:        50,
		Max:        150,
	}
	if smallConstraint.Calculate(200) != 50 {
		t.Errorf("Expected 50 (min), got %d", smallConstraint.Calculate(200))
	}

	// Test with constraint that would be too large
	largeConstraint := MinMaxSize{
		Constraint: FixedSize(200),
		Min:        50,
		Max:        150,
	}
	if largeConstraint.Calculate(300) != 150 {
		t.Errorf("Expected 150 (max), got %d", largeConstraint.Calculate(300))
	}
}

func TestSpacing(t *testing.T) {
	uniform := Uniform(10)
	expected := Spacing{Top: 10, Right: 10, Bottom: 10, Left: 10}
	if uniform != expected {
		t.Errorf("Uniform(10) = %+v, expected %+v", uniform, expected)
	}

	hv := HorizontalVertical(20, 15)
	expected = Spacing{Top: 15, Right: 20, Bottom: 15, Left: 20}
	if hv != expected {
		t.Errorf("HorizontalVertical(20, 15) = %+v, expected %+v", hv, expected)
	}
}

func TestHorizontalLayout(t *testing.T) {
	layout := NewHorizontalLayout(
		Item(FixedSize(100), FixedSize(50)),
		Item(FixedSize(150), FixedSize(75)),
	)

	area := Rect(0, 0, 300, 100)
	result := layout.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// First item should be at (0,0) with size 100x50
	expected1 := Rect(0, 0, 100, 50)
	if result.Areas[0] != expected1 {
		t.Errorf("First area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item should be at (100,0) with size 150x75
	expected2 := Rect(100, 0, 150, 75)
	if result.Areas[1] != expected2 {
		t.Errorf("Second area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestVerticalLayout(t *testing.T) {
	layout := NewVerticalLayout(
		Item(FixedSize(100), FixedSize(50)),
		Item(FixedSize(150), FixedSize(75)),
	)

	area := Rect(0, 0, 200, 200)
	result := layout.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// First item should be at (0,0) with size 100x50
	expected1 := Rect(0, 0, 100, 50)
	if result.Areas[0] != expected1 {
		t.Errorf("First area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item should be at (0,50) with size 150x75
	expected2 := Rect(0, 50, 150, 75)
	if result.Areas[1] != expected2 {
		t.Errorf("Second area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestLayoutWithGap(t *testing.T) {
	layout := &Layout{
		Direction: Horizontal,
		Items: []LayoutItem{
			Item(FixedSize(100), FixedSize(50)),
			Item(FixedSize(100), FixedSize(50)),
		},
		Gap: 20,
	}

	area := Rect(0, 0, 300, 100)
	result := layout.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// First item should be at (0,0)
	expected1 := Rect(0, 0, 100, 50)
	if result.Areas[0] != expected1 {
		t.Errorf("First area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item should be at (120,0) due to 20px gap
	expected2 := Rect(120, 0, 100, 50)
	if result.Areas[1] != expected2 {
		t.Errorf("Second area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestLayoutWithMargin(t *testing.T) {
	item := Item(FixedSize(100), FixedSize(50))
	item.Margin = Uniform(10)

	layout := NewHorizontalLayout(item)
	area := Rect(0, 0, 200, 100)
	result := layout.Calculate(area)

	if len(result.Areas) != 1 {
		t.Fatalf("Expected 1 area, got %d", len(result.Areas))
	}

	// Item should be at (10,10) with size 80x30 (reduced by margin)
	expected := Rect(10, 10, 80, 30)
	if result.Areas[0] != expected {
		t.Errorf("Area = %+v, expected %+v", result.Areas[0], expected)
	}
}

func TestFlexLayout(t *testing.T) {
	layout := &Layout{
		Direction: Horizontal,
		Items: []LayoutItem{
			Item(FixedSize(100), FixedSize(50)),
			Item(FlexSize{Grow: 1}, FixedSize(50)),
			Item(FixedSize(100), FixedSize(50)),
		},
	}

	area := Rect(0, 0, 400, 100)
	result := layout.Calculate(area)

	if len(result.Areas) != 3 {
		t.Fatalf("Expected 3 areas, got %d", len(result.Areas))
	}

	// First item: fixed 100px
	expected1 := Rect(0, 0, 100, 50)
	if result.Areas[0] != expected1 {
		t.Errorf("First area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item: should take remaining space (200px)
	expected2 := Rect(100, 0, 200, 50)
	if result.Areas[1] != expected2 {
		t.Errorf("Second area = %+v, expected %+v", result.Areas[1], expected2)
	}

	// Third item: fixed 100px
	expected3 := Rect(300, 0, 100, 50)
	if result.Areas[2] != expected3 {
		t.Errorf("Third area = %+v, expected %+v", result.Areas[2], expected3)
	}
}

func TestRatioLayout(t *testing.T) {
	layout := NewHorizontalLayout(
		Item(RatioSize(0.3), FixedSize(50)),
		Item(RatioSize(0.7), FixedSize(50)),
	)

	area := Rect(0, 0, 200, 100)
	result := layout.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// First item: 30% of 200 = 60px
	expected1 := Rect(0, 0, 60, 50)
	if result.Areas[0] != expected1 {
		t.Errorf("First area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item: 70% of 200 = 140px
	expected2 := Rect(60, 0, 140, 50)
	if result.Areas[1] != expected2 {
		t.Errorf("Second area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestGrid(t *testing.T) {
	grid := NewGrid(2,
		GridItemAt(1, 1, Item(nil, nil)),
		GridItemAt(2, 1, Item(nil, nil)),
		GridItemAt(1, 2, Item(nil, nil)),
		GridItemAt(2, 2, Item(nil, nil)),
	)

	area := Rect(0, 0, 200, 100)
	result := grid.Calculate(area)

	if len(result.Areas) != 4 {
		t.Fatalf("Expected 4 areas, got %d", len(result.Areas))
	}

	// Each cell should be 100x50
	expected := []Rectangle{
		Rect(0, 0, 100, 50),   // (1,1)
		Rect(100, 0, 100, 50), // (2,1)
		Rect(0, 50, 100, 50),  // (1,2)
		Rect(100, 50, 100, 50), // (2,2)
	}

	for i, expectedArea := range expected {
		if result.Areas[i] != expectedArea {
			t.Errorf("Grid area %d = %+v, expected %+v", i, result.Areas[i], expectedArea)
		}
	}
}

func TestGridWithGap(t *testing.T) {
	grid := &Grid{
		Columns:   2,
		ColumnGap: 10,
		RowGap:    20,
		Items: []GridItem{
			GridItemAt(1, 1, Item(nil, nil)),
			GridItemAt(2, 1, Item(nil, nil)),
		},
	}

	area := Rect(0, 0, 210, 100) // 200 + 10 gap
	result := grid.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// Each cell should be 100x100 with 10px gap between columns
	expected1 := Rect(0, 0, 100, 100)
	expected2 := Rect(110, 0, 100, 100) // 100 + 10 gap

	if result.Areas[0] != expected1 {
		t.Errorf("First grid area = %+v, expected %+v", result.Areas[0], expected1)
	}
	if result.Areas[1] != expected2 {
		t.Errorf("Second grid area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestGridSpan(t *testing.T) {
	grid := &Grid{
		Columns: 3,
		Items: []GridItem{
			GridItemSpan(1, 1, 2, 1, Item(nil, nil)), // Spans 2 columns
			GridItemAt(3, 1, Item(nil, nil)),
		},
	}

	area := Rect(0, 0, 300, 100)
	result := grid.Calculate(area)

	if len(result.Areas) != 2 {
		t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
	}

	// First item spans 2 columns (200px wide)
	expected1 := Rect(0, 0, 200, 100)
	if result.Areas[0] != expected1 {
		t.Errorf("Spanning grid area = %+v, expected %+v", result.Areas[0], expected1)
	}

	// Second item is in third column
	expected2 := Rect(200, 0, 100, 100)
	if result.Areas[1] != expected2 {
		t.Errorf("Third column area = %+v, expected %+v", result.Areas[1], expected2)
	}
}

func TestMainAxisAlignment(t *testing.T) {
	tests := []struct {
		alignment MainAxisAlignment
		expected  []Rectangle
	}{
		{
			MainAxisStart,
			[]Rectangle{
				Rect(0, 0, 50, 100),
				Rect(50, 0, 50, 100),
			},
		},
		{
			MainAxisCenter,
			[]Rectangle{
				Rect(100, 0, 50, 100), // Centered with 100px offset
				Rect(150, 0, 50, 100),
			},
		},
		{
			MainAxisEnd,
			[]Rectangle{
				Rect(200, 0, 50, 100), // Right-aligned
				Rect(250, 0, 50, 100),
			},
		},
	}

	for _, test := range tests {
		layout := &Layout{
			Direction:         Horizontal,
			MainAxisAlignment: test.alignment,
			Items: []LayoutItem{
				Item(FixedSize(50), FixedSize(100)),
				Item(FixedSize(50), FixedSize(100)),
			},
		}

		area := Rect(0, 0, 300, 100) // Extra space for alignment
		result := layout.Calculate(area)

		if len(result.Areas) != 2 {
			t.Fatalf("Expected 2 areas, got %d", len(result.Areas))
		}

		for i, expected := range test.expected {
			if result.Areas[i] != expected {
				t.Errorf("Alignment %d, area %d = %+v, expected %+v", 
					test.alignment, i, result.Areas[i], expected)
			}
		}
	}
}

func TestCrossAxisAlignment(t *testing.T) {
	layout := &Layout{
		Direction:          Horizontal,
		CrossAxisAlignment: CrossAxisCenter,
		Items: []LayoutItem{
			Item(FixedSize(100), FixedSize(50)), // Smaller height
		},
	}

	area := Rect(0, 0, 200, 100) // Taller container
	result := layout.Calculate(area)

	if len(result.Areas) != 1 {
		t.Fatalf("Expected 1 area, got %d", len(result.Areas))
	}

	// Item should be vertically centered
	expected := Rect(0, 25, 100, 50) // Y offset = (100-50)/2 = 25
	if result.Areas[0] != expected {
		t.Errorf("Centered area = %+v, expected %+v", result.Areas[0], expected)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that the new SplitVertical works the same as the old one
	area := Rect(0, 0, 200, 100)
	left, right := SplitVertical(area, 0.3)

	expectedLeft := Rect(0, 0, 60, 100)  // 30% of 200
	expectedRight := Rect(60, 0, 140, 100) // 70% of 200

	if left != expectedLeft {
		t.Errorf("SplitVertical left = %+v, expected %+v", left, expectedLeft)
	}
	if right != expectedRight {
		t.Errorf("SplitVertical right = %+v, expected %+v", right, expectedRight)
	}

	// Test SplitHorizontal
	top, bottom := SplitHorizontal(area, 0.4)

	expectedTop := Rect(0, 0, 200, 40)    // 40% of 100
	expectedBottom := Rect(0, 40, 200, 60) // 60% of 100

	if top != expectedTop {
		t.Errorf("SplitHorizontal top = %+v, expected %+v", top, expectedTop)
	}
	if bottom != expectedBottom {
		t.Errorf("SplitHorizontal bottom = %+v, expected %+v", bottom, expectedBottom)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test Item helper
	item := Item(FixedSize(100), FixedSize(50))
	if item.Width.Calculate(200) != 100 {
		t.Error("Item helper width not set correctly")
	}
	if item.Height.Calculate(200) != 50 {
		t.Error("Item helper height not set correctly")
	}

	// Test FlexItem helper
	flexItem := FlexItem(2, 1, FixedSize(50))
	if !flexItem.Width.IsFlexible() {
		t.Error("FlexItem should create flexible width")
	}
	if !flexItem.Height.IsFlexible() {
		t.Error("FlexItem should create flexible height")
	}

	// Test GridItemAt helper
	gridItem := GridItemAt(2, 3, item)
	if gridItem.Column != 2 || gridItem.Row != 3 {
		t.Error("GridItemAt position not set correctly")
	}
	if gridItem.ColumnSpan != 1 || gridItem.RowSpan != 1 {
		t.Error("GridItemAt spans not set correctly")
	}

	// Test GridItemSpan helper
	spanItem := GridItemSpan(1, 1, 2, 3, item)
	if spanItem.ColumnSpan != 2 || spanItem.RowSpan != 3 {
		t.Error("GridItemSpan spans not set correctly")
	}
}