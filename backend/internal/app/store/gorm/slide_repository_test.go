package gorm

import (
	"testing"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

func makeSlide(id int64, externalID string, subID *int, status domain.SlideStatus, title string) domain.Slide {
	return domain.Slide{
		ID:            id,
		Source:        "test-source",
		ExternalID:    externalID,
		ExternalSubID: subID,
		Status:        status,
		Content: domain.Content{
			Title: title,
			Body:  "body",
			Type:  domain.TypeSocialMedia,
		},
		DisplayOptions: domain.DisplayOptions{
			Priority: 1,
			IsUrgent: false,
		},
	}
}

func intPtr(v int) *int { return &v }

func TestDiffSlides_EmptyExisting(t *testing.T) {
	incoming := []domain.Slide{
		makeSlide(0, "ext1", nil, domain.StatusActive, "Slide 1"),
		makeSlide(0, "ext2", nil, domain.StatusActive, "Slide 2"),
	}

	toCreate, toUpdate, toDelete := diffSlides(nil, incoming)

	assert.Len(t, toCreate, 2)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 0)
}

func TestDiffSlides_EmptyIncoming(t *testing.T) {
	existing := []domain.Slide{
		makeSlide(1, "ext1", nil, domain.StatusActive, "Slide 1"),
		makeSlide(2, "ext2", nil, domain.StatusActive, "Slide 2"),
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, nil)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 2)
}

func TestDiffSlides_MatchingKeyWithChanges(t *testing.T) {
	existing := []domain.Slide{
		makeSlide(1, "ext1", nil, domain.StatusActive, "Slide 1"),
	}

	incoming := []domain.Slide{
		makeSlide(0, "ext1", nil, domain.StatusActive, "Slide 1 Changed"),
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Equal(t, int64(1), toUpdate[0].ID, "ID should be copied from existing")
	assert.Equal(t, "Slide 1 Changed", toUpdate[0].Content.Title)
	assert.Len(t, toDelete, 0)
}

func TestDiffSlides_MatchingKeyNoChanges(t *testing.T) {
	existing := []domain.Slide{
		makeSlide(1, "ext1", nil, domain.StatusActive, "Slide 1"),
	}

	incoming := []domain.Slide{
		makeSlide(0, "ext1", nil, domain.StatusActive, "Slide 1"),
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 0)
}

func TestDiffSlides_Mixed(t *testing.T) {
	sub1 := intPtr(1)

	existing := []domain.Slide{
		makeSlide(1, "ext1", nil, domain.StatusActive, "Keep Same"),
		makeSlide(2, "ext2", nil, domain.StatusActive, "Will Change"),
		makeSlide(3, "ext3", nil, domain.StatusActive, "Will Delete"),
	}

	incoming := []domain.Slide{
		makeSlide(0, "ext1", nil, domain.StatusActive, "Keep Same"),
		makeSlide(0, "ext2", nil, domain.StatusActive, "Changed Title"),
		makeSlide(0, "ext4", sub1, domain.StatusActive, "New Slide"),
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, incoming)

	assert.Len(t, toCreate, 1)
	assert.Equal(t, "New Slide", toCreate[0].Content.Title)
	assert.Equal(t, "ext4", toCreate[0].ExternalID)

	assert.Len(t, toUpdate, 1)
	assert.Equal(t, int64(2), toUpdate[0].ID)
	assert.Equal(t, "Changed Title", toUpdate[0].Content.Title)

	assert.Len(t, toDelete, 1)
	assert.Equal(t, int64(3), toDelete[0].ID)
}

func TestDiffSlides_ExternalSubIDMatching(t *testing.T) {
	sub1 := intPtr(1)
	sub2 := intPtr(2)

	existing := []domain.Slide{
		makeSlide(1, "extID", sub1, domain.StatusActive, "Slide with sub 1"),
		makeSlide(2, "extID", sub2, domain.StatusActive, "Slide with sub 2"),
	}

	// "sub 1 unchanged" must have same title as existing[0] to test no HasChanged
	incoming := []domain.Slide{
		makeSlide(0, "extID", sub1, domain.StatusActive, "Slide with sub 1"),
		makeSlide(0, "extID", sub2, domain.StatusActive, "Slide with sub 2 changed"),
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Equal(t, int64(2), toUpdate[0].ID)
	assert.Len(t, toDelete, 0)
}

func TestDiffSlides_StatusChanged(t *testing.T) {
	existing := []domain.Slide{
		{
			ID:         1,
			Source:     "test-source",
			ExternalID: "ext1",
			Status:     domain.StatusPending,
			Content: domain.Content{
				Title: "Slide 1",
				Body:  "body",
				Type:  domain.TypeSocialMedia,
			},
			DisplayOptions: domain.DisplayOptions{Priority: 1, IsUrgent: false},
		},
	}

	incoming := []domain.Slide{
		{
			Source:     "test-source",
			ExternalID: "ext1",
			Status:     domain.StatusActive,
			Content: domain.Content{
				Title: "Slide 1",
				Body:  "body",
				Type:  domain.TypeSocialMedia,
			},
			DisplayOptions: domain.DisplayOptions{Priority: 1, IsUrgent: false},
		},
	}

	toCreate, toUpdate, toDelete := diffSlides(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Equal(t, domain.StatusActive, toUpdate[0].Status)
	assert.Len(t, toDelete, 0)
}
