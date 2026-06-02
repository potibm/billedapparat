package bluesky

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostRecord_FirstLanguage(t *testing.T) {
	tests := []struct {
		name     string
		record   PostRecord
		expected string
	}{
		{
			name:     "nil or empty languages",
			record:   PostRecord{Langs: nil},
			expected: "",
		},
		{
			name:     "empty language slice",
			record:   PostRecord{Langs: []string{}},
			expected: "",
		},
		{
			name:     "single language",
			record:   PostRecord{Langs: []string{"en"}},
			expected: "en",
		},
		{
			name:     "multiple languages returns the first one",
			record:   PostRecord{Langs: []string{"de", "en", "fr"}},
			expected: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.FirstLanguage()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestJetstreamEvent_ExtractImageURLs(t *testing.T) {
	tests := []struct {
		name     string
		event    JetstreamEvent
		expected []string
	}{
		{
			name:     "nil commit",
			event:    JetstreamEvent{Did: "did:plc:123", Commit: nil},
			expected: nil,
		},
		{
			name: "nil record",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: nil,
				},
			},
			expected: nil,
		},
		{
			name: "nil embed",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: nil,
					},
				},
			},
			expected: nil,
		},
		{
			name: "embed is not images type",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type: "app.bsky.embed.external",
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "embed is images with empty images slice",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type:   "app.bsky.embed.images",
							Images: []EmbedImage{},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "embed has images with valid and empty link CIDs",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type: "app.bsky.embed.images",
							Images: []EmbedImage{
								{
									Image: MediaBlob{
										Ref: ImageRef{Link: "cid1"},
									},
								},
								{
									Image: MediaBlob{
										Ref: ImageRef{Link: ""},
									},
								},
								{
									Image: MediaBlob{
										Ref: ImageRef{Link: "cid2"},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{
				"https://cdn.bsky.app/img/feed_fullsize/plain/did:plc:123/cid1@jpeg",
				"https://cdn.bsky.app/img/feed_fullsize/plain/did:plc:123/cid2@jpeg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.ExtractImageURLs()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestJetstreamEvent_ExtractVideoURLs(t *testing.T) {
	tests := []struct {
		name             string
		event            JetstreamEvent
		wantStreamURL    string
		wantThumbnailURL string
		wantHasVideo     bool
	}{
		{
			name:             "nil commit",
			event:            JetstreamEvent{Did: "did:plc:123", Commit: nil},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "nil record",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: nil,
				},
			},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "nil embed",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: nil,
					},
				},
			},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "embed is not video type",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type: "app.bsky.embed.images",
						},
					},
				},
			},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "embed is video type but video field is nil",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type:  "app.bsky.embed.video",
							Video: nil,
						},
					},
				},
			},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "embed is video but Link is empty",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type: "app.bsky.embed.video",
							Video: &MediaBlob{
								Ref: ImageRef{Link: ""},
							},
						},
					},
				},
			},
			wantStreamURL:    "",
			wantThumbnailURL: "",
			wantHasVideo:     false,
		},
		{
			name: "valid video embed",
			event: JetstreamEvent{
				Did: "did:plc:123",
				Commit: &JetstreamCommit{
					Record: &PostRecord{
						Embed: &PostEmbed{
							Type: "app.bsky.embed.video",
							Video: &MediaBlob{
								Ref: ImageRef{Link: "videocid"},
							},
						},
					},
				},
			},
			wantStreamURL:    "https://video.bsky.app/watch/did:plc:123/videocid/playlist.m3u8",
			wantThumbnailURL: "https://video.bsky.app/watch/did:plc:123/videocid/thumbnail.jpg",
			wantHasVideo:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStream, gotThumb, gotHas := tt.event.ExtractVideoURLs()
			assert.Equal(t, tt.wantStreamURL, gotStream)
			assert.Equal(t, tt.wantThumbnailURL, gotThumb)
			assert.Equal(t, tt.wantHasVideo, gotHas)
		})
	}
}

func TestPostRecord_Hashtags(t *testing.T) {
	tests := []struct {
		name     string
		record   PostRecord
		expected []string
	}{
		{
			name:     "no facets",
			record:   PostRecord{Facets: nil},
			expected: nil,
		},
		{
			name: "empty features inside facets",
			record: PostRecord{
				Facets: []Facet{
					{Features: []FacetFeature{}},
				},
			},
			expected: nil,
		},
		{
			name: "features with non-tag types",
			record: PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#mention", Tag: "mentiontag"},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "features with tag types and tags are lowercase normalized",
			record: PostRecord{
				Facets: []Facet{
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "DemoScene"},
							{Type: "app.bsky.richtext.facet#mention", Tag: "ignored"},
						},
					},
					{
						Features: []FacetFeature{
							{Type: "app.bsky.richtext.facet#tag", Tag: "Party"},
							{Type: "app.bsky.richtext.facet#tag", Tag: ""},
						},
					},
				},
			},
			expected: []string{"demoscene", "party"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.record.Hashtags()
			assert.Equal(t, tt.expected, got)
		})
	}
}
