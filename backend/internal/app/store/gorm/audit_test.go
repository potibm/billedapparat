package gorm

import (
	"context"
	"testing"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type TestLocation struct {
	ID         uint `gorm:"primaryKey"`
	Name       string
	CreatedBy  string
	ModifiedBy string
	DeletedBy  *string
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	store, err := NewSqliteInMemoryStore()
	require.NoError(t, err, "Should be able to create in-memory store")

	err = store.db.AutoMigrate(&TestLocation{})
	require.NoError(t, err)

	return store.db
}

func TestAuditCallbacks(t *testing.T) {
	db := setupTestDB(t)

	creatorID := "user-creator-123"
	updaterID := "user-updater-456"
	deleterID := "user-deleter-789"

	t.Run("BeforeCreate sets CreatedBy and ModifiedBy", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), domain.UserIDKey, creatorID)

		loc := TestLocation{Name: "Main Stage"}
		err := db.WithContext(ctx).Create(&loc).Error

		require.NoError(t, err)
		assert.Equal(t, creatorID, loc.CreatedBy, "CreatedBy should be set")
		assert.Equal(t, creatorID, loc.ModifiedBy, "ModifiedBy should be equal to CreatedBy initially")
	})

	t.Run("BeforeUpdate changes ONLY ModifiedBy", func(t *testing.T) {
		var loc TestLocation
		db.First(&loc, 1)

		ctx := context.WithValue(context.Background(), domain.UserIDKey, updaterID)

		loc.Name = "Main Stage (Updated)"
		err := db.WithContext(ctx).Save(&loc).Error

		require.NoError(t, err)

		var updatedLoc TestLocation
		db.First(&updatedLoc, 1)

		assert.Equal(t, creatorID, updatedLoc.CreatedBy, "CreatedBy should NOT change on update")
		assert.Equal(t, updaterID, updatedLoc.ModifiedBy, "ModifiedBy should be updated")
	})

	t.Run("BeforeDelete sets DeletedBy on soft delete", func(t *testing.T) {
		untouched := TestLocation{Name: "Untouched Location"}
		require.NoError(t, db.Create(&untouched).Error)

		var loc TestLocation
		db.First(&loc, 1)

		ctx := context.WithValue(context.Background(), domain.UserIDKey, deleterID)

		err := db.WithContext(ctx).Delete(&loc).Error
		require.NoError(t, err)

		var deletedLoc TestLocation

		err = db.Unscoped().First(&deletedLoc, 1).Error

		require.NoError(t, err)
		assert.True(t, deletedLoc.DeletedAt.Valid, "GORM should have set DeletedAt for soft-deleted record")
		assert.NotNil(t, deletedLoc.DeletedBy, "DeletedBy should be set")
		assert.Equal(t, deleterID, *deletedLoc.DeletedBy, "DeletedBy should be set to the deleter's user ID")

		var untouchedReloaded TestLocation
		db.First(&untouchedReloaded, untouched.ID)
		assert.Nil(t, untouchedReloaded.DeletedBy, "DeletedBy should remain empty for non-deleted records")
		assert.False(t, untouchedReloaded.DeletedAt.Valid, "DeletedAt should remain null for non-deleted records")
	})
}
