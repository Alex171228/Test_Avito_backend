package service

import (
	"testing"
)

func TestToMinutes(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"00:00", 0},
		{"09:00", 540},
		{"09:30", 570},
		{"12:00", 720},
		{"18:00", 1080},
		{"23:59", 1439},
	}
	for _, tt := range tests {
		got := ToMinutes(tt.input)
		if got != tt.want {
			t.Errorf("ToMinutes(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestTimePatternValidation(t *testing.T) {
	valid := []string{"00:00", "09:00", "9:00", "12:30", "23:59"}
	for _, v := range valid {
		if !timePattern.MatchString(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}

	invalid := []string{"24:00", "25:00", "12:60", "abc", "12:345", ""}
	for _, v := range invalid {
		if timePattern.MatchString(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}
