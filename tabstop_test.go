package tv

import (
	"testing"
)

func TestTabStops(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		interval int
		checks   []struct {
			col      int
			expected bool
		}
	}{
		{
			name:     "default interval of 8",
			width:    24,
			interval: DefaultTabInterval,
			checks: []struct {
				col      int
				expected bool
			}{
				{0, true},   // First tab stop
				{7, false},  // Not a tab stop
				{8, true},   // Second tab stop
				{15, false}, // Not a tab stop
				{16, true},  // Third tab stop
				{23, false}, // Not a tab stop
			},
		},
		{
			name:     "custom interval of 4",
			width:    16,
			interval: 4,
			checks: []struct {
				col      int
				expected bool
			}{
				{0, true},   // First tab stop
				{3, false},  // Not a tab stop
				{4, true},   // Second tab stop
				{7, false},  // Not a tab stop
				{8, true},   // Third tab stop
				{12, true},  // Fourth tab stop
				{15, false}, // Not a tab stop
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTabStops(tt.width, tt.interval)

			// Test initial tab stops
			for _, check := range tt.checks {
				if got := ts.IsStop(check.col); got != check.expected {
					t.Errorf("IsStop(%d) = %v, want %v", check.col, got, check.expected)
				}
			}

			// Test setting a custom tab stop
			customCol := tt.interval + 1
			ts.Set(customCol)
			if !ts.IsStop(customCol) {
				t.Errorf("After Set(%d), IsStop(%d) = false, want true", customCol, customCol)
			}

			// Test resetting a tab stop
			regularStop := tt.interval
			ts.Reset(regularStop)
			if ts.IsStop(regularStop) {
				t.Errorf("After Reset(%d), IsStop(%d) = true, want false", regularStop, regularStop)
			}
		})
	}
}

func TestTabStopsNavigation(t *testing.T) {
	ts := NewTabStops(24, DefaultTabInterval)

	tests := []struct {
		name     string
		col      int
		wantNext int
		wantPrev int
	}{
		{
			name:     "from column 0",
			col:      0,
			wantNext: 8,
			wantPrev: 0,
		},
		{
			name:     "from column 4",
			col:      4,
			wantNext: 8,
			wantPrev: 0,
		},
		{
			name:     "from column 8",
			col:      8,
			wantNext: 16,
			wantPrev: 0,
		},
		{
			name:     "from column 20",
			col:      20,
			wantNext: 23,
			wantPrev: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ts.Next(tt.col); got != tt.wantNext {
				t.Errorf("Next(%d) = %v, want %v", tt.col, got, tt.wantNext)
			}
			if got := ts.Prev(tt.col); got != tt.wantPrev {
				t.Errorf("Prev(%d) = %v, want %v", tt.col, got, tt.wantPrev)
			}
		})
	}
}

func TestTabStopsClear(t *testing.T) {
	ts := NewTabStops(24, DefaultTabInterval)

	// Verify initial state
	if !ts.IsStop(0) || !ts.IsStop(8) || !ts.IsStop(16) {
		t.Error("Initial tab stops not set correctly")
	}

	// Clear all tab stops
	ts.Clear()

	// Verify all stops are cleared
	for i := 0; i < 24; i++ {
		if ts.IsStop(i) {
			t.Errorf("Tab stop at column %d still set after Clear()", i)
		}
	}
}

func TestTabStopsResize(t *testing.T) {
	tests := []struct {
		name        string
		initialSize int
		newSize     int
		checks      []struct {
			col      int
			expected bool
		}
	}{
		{
			name:        "grow buffer",
			initialSize: 16,
			newSize:     24,
			checks: []struct {
				col      int
				expected bool
			}{
				{0, true},   // Original tab stop
				{8, true},   // Original tab stop
				{16, true},  // New tab stop
				{23, false}, // Not a tab stop
			},
		},
		{
			name:        "same size - no change",
			initialSize: 16,
			newSize:     16,
			checks: []struct {
				col      int
				expected bool
			}{
				{0, true},   // Original tab stop
				{8, true},   // Original tab stop
				{15, false}, // Not a tab stop
			},
		},
		{
			name:        "resize with custom interval",
			initialSize: 8,
			newSize:     16,
			checks: []struct {
				col      int
				expected bool
			}{
				{0, true},   // First tab stop
				{4, true},   // Second tab stop
				{8, true},   // Third tab stop
				{12, true},  // Fourth tab stop
				{15, false}, // Not a tab stop
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts *TabStops
			if tt.name == "resize with custom interval" {
				ts = NewTabStops(tt.initialSize, 4) // Custom interval of 4
			} else {
				ts = DefaultTabStops(tt.initialSize)
			}

			// Verify initial state
			if ts.width != tt.initialSize {
				t.Errorf("Initial width = %d, want %d", ts.width, tt.initialSize)
			}

			// Perform resize
			ts.Resize(tt.newSize)

			// Verify new size
			if ts.width != tt.newSize {
				t.Errorf("After resize, width = %d, want %d", ts.width, tt.newSize)
			}

			// Check tab stops after resize
			for _, check := range tt.checks {
				if got := ts.IsStop(check.col); got != check.expected {
					t.Errorf("After resize, IsStop(%d) = %v, want %v",
						check.col, got, check.expected)
				}
			}

			// Verify stops slice has correct length
			expectedStopsLen := (tt.newSize + (ts.interval - 1)) / ts.interval
			if len(ts.stops) != expectedStopsLen {
				t.Errorf("stops slice length = %d, want %d",
					len(ts.stops), expectedStopsLen)
			}
		})
	}
}

func TestTabStopsResizeEdgeCases(t *testing.T) {
	t.Run("resize to zero", func(t *testing.T) {
		ts := DefaultTabStops(8)
		ts.Resize(0)

		if ts.width != 0 {
			t.Errorf("width = %d, want 0", ts.width)
		}

		// Verify no tab stops are accessible
		if ts.IsStop(0) {
			t.Error("IsStop(0) should return false for zero width")
		}
	})

	t.Run("resize to very large width", func(t *testing.T) {
		ts := DefaultTabStops(8)
		largeWidth := 1000
		ts.Resize(largeWidth)

		// Check some tab stops at higher positions
		checks := []struct {
			col      int
			expected bool
		}{
			{992, true},  // Multiple of 8
			{999, false}, // Not a tab stop
		}

		for _, check := range checks {
			if got := ts.IsStop(check.col); got != check.expected {
				t.Errorf("IsStop(%d) = %v, want %v",
					check.col, got, check.expected)
			}
		}
	})

	t.Run("multiple resizes", func(t *testing.T) {
		ts := DefaultTabStops(8)

		// Perform multiple resizes
		sizes := []int{16, 8, 24, 4}
		for _, size := range sizes {
			ts.Resize(size)

			// Verify basic properties after each resize
			if ts.width != size {
				t.Errorf("width = %d, want %d", ts.width, size)
			}

			// Check first tab stop is always set
			if !ts.IsStop(0) {
				t.Errorf("After resize to %d, IsStop(0) = false, want true", size)
			}
		}
	})
}
