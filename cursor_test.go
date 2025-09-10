package uv

import "testing"

func TestCursorShape_Encode(t *testing.T) {
	tests := []struct {
		shape   CursorShape
		blink   bool
		wantEnc int
	}{
		{CursorBlock, true, 1},
		{CursorBlock, false, 2},
		{CursorUnderline, true, 3},
		{CursorUnderline, false, 4},
		{CursorBar, true, 5},
		{CursorBar, false, 6},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if gotEnc := tt.shape.Encode(tt.blink); gotEnc != tt.wantEnc {
				t.Errorf("CursorShape.Encode() = %v, want %v", gotEnc, tt.wantEnc)
			}
		})
	}
}
