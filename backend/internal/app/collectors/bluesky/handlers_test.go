package bluesky

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector_HasRelevantHashtag(t *testing.T) {
	c := &Collector{
		searchTags: Hashtags{"demoscene", "party"},
	}

	tests := []struct {
		name     string
		record   *PostRecord
		expected bool
	}{
		{
			name: "no tags in record",
			record: &PostRecord{
				Facets: nil,
			},
			expected: false,
		},
		{
			name: "no matching tags",
			record: &PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "other"},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "matching tag exact case",
			record: &PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "demoscene"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "matching tag different case",
			record: &PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "Party"},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "multiple tags, one matches",
			record: &PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "other"},
							{Type: "app.bsky.richtext.facet#tag", Tag: "demoscene"},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.hasRelevantHashtag(tt.record)
			assert.Equal(t, tt.expected, got)
		})
	}
}
