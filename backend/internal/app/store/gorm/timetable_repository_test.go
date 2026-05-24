package gorm

import (
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

func makeEvent(id int64, externalID, title string) domain.TimetableEvent {
	return domain.TimetableEvent{
		ID:         id,
		ExternalID: externalID,
		Source:     "test-source",
		Title:      title,
		StartTime:  time.Date(2026, time.May, 25, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2026, time.May, 25, 10, 0, 0, 0, time.UTC),
	}
}

func TestDiffTimetableEvents_EmptyExisting(t *testing.T) {
	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-1", "Event 1"),
		makeEvent(0, "evt-2", "Event 2"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(nil, incoming)

	assert.Len(t, toCreate, 2)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 0)
}

func TestDiffTimetableEvents_EmptyIncoming(t *testing.T) {
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-1", "Event 1"),
		makeEvent(2, "evt-2", "Event 2"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, nil)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 2)
}

func TestDiffTimetableEvents_MatchingExternalID_AlwaysUpdated(t *testing.T) {
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-1", "Event 1"),
	}

	// Same ExternalID, different title — no HasChanged check, always update
	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-1", "Event 1 Same"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Equal(t, int64(1), toUpdate[0].ID, "ID should be copied from existing")
	assert.Len(t, toDelete, 0)
}

func TestDiffTimetableEvents_MatchingExternalID_IdenticalStillUpdated(t *testing.T) {
	// Unlike diffSlides, diffTimetableEvents does NOT check HasChanged.
	// All matching events are always marked for update.
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-1", "Event 1"),
	}

	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-1", "Event 1"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Equal(t, int64(1), toUpdate[0].ID)
	assert.Len(t, toDelete, 0)
}

func TestDiffTimetableEvents_NewExternalID(t *testing.T) {
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-1", "Event 1"),
	}

	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-2", "New Event"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

	assert.Len(t, toCreate, 1)
	assert.Equal(t, "evt-2", toCreate[0].ExternalID)
	assert.Len(t, toUpdate, 0)
	assert.Len(t, toDelete, 1)
	assert.Equal(t, "evt-1", toDelete[0].ExternalID)
}

func TestDiffTimetableEvents_RemovedExternalID(t *testing.T) {
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-1", "Event 1"),
		makeEvent(2, "evt-2", "Event 2"),
	}

	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-1", "Event 1"),
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

	assert.Len(t, toCreate, 0)
	assert.Len(t, toUpdate, 1)
	assert.Len(t, toDelete, 1)
	assert.Equal(t, "evt-2", toDelete[0].ExternalID)
}

func TestDiffTimetableEvents_Mixed(t *testing.T) {
	existing := []domain.TimetableEvent{
		makeEvent(1, "evt-keep", "Keep Event"),
		makeEvent(2, "evt-update", "Update Event"),
		makeEvent(3, "evt-delete", "Delete Event"),
	}

	incoming := []domain.TimetableEvent{
		makeEvent(0, "evt-keep", "Keep Event"),      // should update (always)
		makeEvent(0, "evt-update", "Updated Event"), // should update
		makeEvent(0, "evt-new", "New Event"),        // should create
	}

	toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

	assert.Len(t, toCreate, 1)
	assert.Equal(t, "evt-new", toCreate[0].ExternalID)

	assert.Len(t, toUpdate, 2)

	assert.Len(t, toDelete, 1)
	assert.Equal(t, "evt-delete", toDelete[0].ExternalID)
}
