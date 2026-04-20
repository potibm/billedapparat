package gorm

import (
	"time"

	"gorm.io/gorm"
)

type GormModel struct {
	ID        int64          `json:"id"        gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`
} // @name models.gormModel
