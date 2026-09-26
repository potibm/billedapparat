package gorm

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These cover only what the probe can observe; see Store.Ping.

func TestStorePing(t *testing.T) {
	store, err := NewSqliteInMemoryStore()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = store.Close()
	})

	assert.NoError(t, store.Ping(context.Background()))
}

func TestStorePingDatabaseClosed(t *testing.T) {
	store, err := NewSqliteInMemoryStore()
	require.NoError(t, err)

	require.NoError(t, store.Close())

	err = store.Ping(context.Background())
	assert.Error(t, err)
}

func TestStorePingContextCancelled(t *testing.T) {
	store, err := NewSqliteInMemoryStore()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = store.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = store.Ping(ctx)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

// TestStoreConnectionPoolIsBounded covers the cap that keeps the unauthenticated
// /ready endpoint from opening an unbounded number of SQLite connections.
func TestStoreConnectionPoolIsBounded(t *testing.T) {
	store, err := NewSqliteInMemoryStore()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = store.Close()
	})

	sqlDB, err := store.db.DB()
	require.NoError(t, err)

	assert.Equal(t, maxOpenConns, sqlDB.Stats().MaxOpenConnections)
}

// TestStorePragmasAreApplied covers the pragmas newStore sets up front. They used
// to sit behind an "if err == nil" guard that could skip them silently.
func TestStorePragmasAreApplied(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "pragmas.db")

	store, err := newStore("file:" + dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = store.Close()
	})

	var journalMode string
	require.NoError(t, store.db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error)
	assert.Equal(t, "wal", strings.ToLower(journalMode))

	var foreignKeys int64
	require.NoError(t, store.db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error)
	assert.Equal(t, int64(1), foreignKeys)
}
