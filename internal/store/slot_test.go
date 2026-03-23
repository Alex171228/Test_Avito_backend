package store

import (
	"testing"
)

func TestParseTimeStr(t *testing.T) {
	tests := []struct {
		input      string
		wantH      int
		wantM      int
	}{
		{"09:00", 9, 0},
		{"18:30", 18, 30},
		{"00:00", 0, 0},
		{"23:59", 23, 59},
		{"9:00", 9, 0},
		{"12:15", 12, 15},
	}
	for _, tt := range tests {
		h, m := ParseTimeStr(tt.input)
		if h != tt.wantH || m != tt.wantM {
			t.Errorf("ParseTimeStr(%q) = (%d, %d), want (%d, %d)",
				tt.input, h, m, tt.wantH, tt.wantM)
		}
	}
}
