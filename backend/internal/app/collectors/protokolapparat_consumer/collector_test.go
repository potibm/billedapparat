package redisconsumer

import (
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name       string
		errorCount int
		expected   time.Duration
	}{
		{
			name:       "Invalid Value",
			errorCount: -1,
			expected:   1 * time.Second,
		},
		{
			name:       "No Errors",
			errorCount: 0,
			expected:   1 * time.Second,
		},
		{
			name:       "1. Error (Base Delay)",
			errorCount: 1,
			expected:   1 * time.Second,
		},
		{
			name:       "2. Error (2s Delay)",
			errorCount: 2,
			expected:   2 * time.Second,
		},
		{
			name:       "3. Error",
			errorCount: 3,
			expected:   4 * time.Second,
		},
		{
			name:       "4. Error",
			errorCount: 4,
			expected:   8 * time.Second,
		},
		{
			name:       "Max Delay Reached",
			errorCount: 10, // 2^9 = 512 Seconds (> 5 Minuten)
			expected:   5 * time.Minute,
		},
		{
			name:       "Just before Max Shifts",
			errorCount: 29,
			expected:   5 * time.Minute,
		},
		{
			name:       "Extreme Error Count",
			errorCount: 1000,
			expected:   5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := calculateBackoff(tt.errorCount)

			if actual != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v; want %v", tt.errorCount, actual, tt.expected)
			}
		})
	}
}
