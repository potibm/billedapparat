package bluesky

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashtags_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		hashtags Hashtags
		expected Hashtags
	}{
		{
			name:     "empty list",
			hashtags: Hashtags{},
			expected: Hashtags{},
		},
		{
			name:     "no modifications needed",
			hashtags: Hashtags{"demoscene", "party"},
			expected: Hashtags{"demoscene", "party"},
		},
		{
			name:     "with leading hash symbols",
			hashtags: Hashtags{"#demoscene", "#party"},
			expected: Hashtags{"demoscene", "party"},
		},
		{
			name:     "with upper case characters",
			hashtags: Hashtags{"#DemoScene", "Party"},
			expected: Hashtags{"demoscene", "party"},
		},
		{
			name:     "with empty string and hash only",
			hashtags: Hashtags{"", "#", "valid"},
			expected: Hashtags{"valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.hashtags.Normalize()
			assert.Equal(t, tt.expected, got)
		})
	}
}
