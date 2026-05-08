package hashing

import "testing"

func TestIsInRange(t *testing.T) {
	tests := []struct {
		name          string
		start, end, k uint64
		expected      bool
	}{
		{"start<end in range", 100, 1000, 500, true},
		{"start<end outside range 1", 100, 1000, 0, false},
		{"start<end outside range 2", 100, 1000, 1000000, false},
		{"end<start in range 1", 1000, 100, 1000000, true},
		{"end<start in range 1", 1000, 100, 50, true},
		{"end<start outside range", 1000, 100, 500, false},
		{"end=start in range 1", 100, 100, 50, true},
		{"end=start in range 2", 100, 100, 150, true},
		{"end=start in range 3", 100, 100, 100, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInRange(tt.start, tt.end, tt.k)

			if result != tt.expected {
				t.Errorf("expected %v got %v\n", tt.expected, result)
			}
		})
	}
}
