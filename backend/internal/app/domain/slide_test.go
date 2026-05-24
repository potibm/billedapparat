package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlide_SyncKey(t *testing.T) {
	sub5 := 5

	tests := []struct {
		name string
		s    Slide
		want string
	}{
		{
			name: "ExternalSubID is nil",
			s:    Slide{ExternalID: "extID", ExternalSubID: nil},
			want: "extID|",
		},
		{
			name: "ExternalSubID is &5",
			s:    Slide{ExternalID: "extID", ExternalSubID: &sub5},
			want: "extID|5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.SyncKey()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlide_HasChanged(t *testing.T) {
	makeSlide := func(status SlideStatus, title string, priority int, urgent bool) Slide {
		return Slide{
			Status:         status,
			Content:        Content{Title: title, Body: "body", Type: TypeSocialMedia},
			DisplayOptions: DisplayOptions{Priority: priority, IsUrgent: urgent},
		}
	}

	tests := []struct {
		name       string
		old        Slide
		newSlide   Slide
		wantChange bool
	}{
		{
			name:       "Same slides",
			old:        makeSlide(StatusActive, "Hello", 1, false),
			newSlide:   makeSlide(StatusActive, "Hello", 1, false),
			wantChange: false,
		},
		{
			name:       "Different Status",
			old:        makeSlide(StatusActive, "Hello", 1, false),
			newSlide:   makeSlide(StatusPending, "Hello", 1, false),
			wantChange: true,
		},
		{
			name:       "Different Content.Title",
			old:        makeSlide(StatusActive, "Hello", 1, false),
			newSlide:   makeSlide(StatusActive, "World", 1, false),
			wantChange: true,
		},
		{
			name:       "Different DisplayOptions.Priority",
			old:        makeSlide(StatusActive, "Hello", 1, false),
			newSlide:   makeSlide(StatusActive, "Hello", 2, false),
			wantChange: true,
		},
		{
			name:       "Different DisplayOptions.IsUrgent",
			old:        makeSlide(StatusActive, "Hello", 1, false),
			newSlide:   makeSlide(StatusActive, "Hello", 1, true),
			wantChange: true,
		},
		{
			name: "Different ExternalID but same visual state",
			old: Slide{
				ExternalID:     "old-ext-123",
				Status:         StatusActive,
				Content:        Content{Title: "Hello", Body: "body", Type: TypeSocialMedia},
				DisplayOptions: DisplayOptions{Priority: 1, IsUrgent: false},
			},
			newSlide: Slide{
				ExternalID:     "new-ext-456",
				Status:         StatusActive,
				Content:        Content{Title: "Hello", Body: "body", Type: TypeSocialMedia},
				DisplayOptions: DisplayOptions{Priority: 1, IsUrgent: false},
			},
			wantChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.old.HasChanged(tt.newSlide)
			assert.Equal(t, tt.wantChange, got)
		})
	}
}

func TestIsValidSlideType(t *testing.T) {
	tests := []struct {
		name string
		t    SlideType
		want bool
	}{
		{name: "TypeSocialMedia", t: TypeSocialMedia, want: true},
		{name: "TypeSocialText", t: TypeSocialText, want: true},
		{name: "TypeSponsor", t: TypeSponsor, want: true},
		{name: "TypeNews", t: TypeNews, want: true},
		{name: "TypeTimetable", t: TypeTimetable, want: true},
		{name: "Unknown string", t: "unknown", want: false},
		{name: "Empty string", t: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidSlideType(tt.t)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlideType_IsReadonly(t *testing.T) {
	tests := []struct {
		name string
		t    SlideType
		want bool
	}{
		{name: "TypeNews", t: TypeNews, want: true},
		{name: "TypeTimetable", t: TypeTimetable, want: true},
		{name: "TypeSocialMedia", t: TypeSocialMedia, want: false},
		{name: "TypeSocialText", t: TypeSocialText, want: false},
		{name: "TypeSponsor", t: TypeSponsor, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.IsReadonly()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSlide_HasChanged_TimeFieldsIgnored(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)

	old := Slide{
		Status:          StatusActive,
		Content:         Content{Title: "Hello", Body: "body", Type: TypeSocialMedia},
		DisplayOptions:  DisplayOptions{Priority: 1, IsUrgent: false},
		CreatedAt:       now,
		OriginCreatedAt: now,
	}

	newSlide := Slide{
		Status:          StatusActive,
		Content:         Content{Title: "Hello", Body: "body", Type: TypeSocialMedia},
		DisplayOptions:  DisplayOptions{Priority: 1, IsUrgent: false},
		CreatedAt:       later,
		OriginCreatedAt: later,
	}

	assert.False(t, old.HasChanged(newSlide),
		"Time fields (CreatedAt, OriginCreatedAt) should not affect HasChanged")
}
