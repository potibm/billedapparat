package bluesky

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRKeyList(t *testing.T) {
	rList := NewRKeyList()
	assert.NotNil(t, rList)
	assert.Equal(t, 0, rList.Len())

	// Test Contains on empty list
	assert.False(t, rList.Contains("key1"))

	// Test Add
	rList.Add("key1")
	assert.Equal(t, 1, rList.Len())
	assert.True(t, rList.Contains("key1"))

	// Test Add another
	rList.Add("key2")
	assert.Equal(t, 2, rList.Len())
	assert.True(t, rList.Contains("key2"))

	// Test Remove
	rList.Remove("key1")
	assert.Equal(t, 1, rList.Len())
	assert.False(t, rList.Contains("key1"))
	assert.True(t, rList.Contains("key2"))

	// Test Remove nonexistent doesn't crash or error
	rList.Remove("nonexistent")
	assert.Equal(t, 1, rList.Len())
}
