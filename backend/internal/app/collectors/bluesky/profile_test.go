package bluesky

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfileList(t *testing.T) {
	pList := NewProfileList()
	assert.NotNil(t, pList)
	assert.Equal(t, 0, pList.Len())

	// Test Get on empty list
	profile, exists := pList.Get("nonexistent")
	assert.Nil(t, profile)
	assert.False(t, exists)

	// Test Add
	p1 := &ProfileResponse{
		Handle:      "handle1.bsky.social",
		DisplayName: "Display One",
		Avatar:      "https://example.com/avatar1.jpg",
	}
	pList.Add("did:plc:1", p1)
	assert.Equal(t, 1, pList.Len())

	// Test Get existing
	profile, exists = pList.Get("did:plc:1")
	assert.True(t, exists)
	assert.Equal(t, p1, profile)

	// Test Add another
	p2 := &ProfileResponse{
		Handle:      "handle2.bsky.social",
		DisplayName: "Display Two",
		Avatar:      "https://example.com/avatar2.jpg",
	}
	pList.Add("did:plc:2", p2)
	assert.Equal(t, 2, pList.Len())

	// Test Remove
	pList.Remove("did:plc:1")
	assert.Equal(t, 1, pList.Len())

	profile, exists = pList.Get("did:plc:1")
	assert.False(t, exists)
	assert.Nil(t, profile)

	// Test Remove nonexistent doesn't crash or error
	pList.Remove("nonexistent")
	assert.Equal(t, 1, pList.Len())
}
