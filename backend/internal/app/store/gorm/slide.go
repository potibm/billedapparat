package gorm

import (
	"context"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dbSlide struct {
	gorm.Model
	Source     string `gorm:"uniqueIndex:idx_ext"`
	ExternalID string `gorm:"uniqueIndex:idx_ext"`
}

type slideRepository struct {
	db *gorm.DB
}

func NewSlideRepository(db *gorm.DB) repository.SlideRepository {
	return &slideRepository{db: db}
}

func (r *slideRepository) Save(ctx context.Context, slide *domain.Slide) error {
	dbObj := fromDomain(slide)

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(dbObj).Error

	if err == nil {
		slide.ID = int64(dbObj.ID)
	}
	return err
}

func (s *dbSlide) toDomain() *domain.Slide {
	return &domain.Slide{
		ID:     int64(s.ID),
		Source: s.Source,
	}
}

func fromDomain(s *domain.Slide) *dbSlide {
	return &dbSlide{
		Source:     s.Source,
		ExternalID: s.ExternalID,
	}
}

func (r *slideRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&dbSlide{}, id).Error
}

func (r *slideRepository) GetActive(ctx context.Context) ([]domain.Slide, error) {
	var dbSlides []dbSlide
	err := r.db.WithContext(ctx).Find(&dbSlides).Error
	if err != nil {
		return nil, err
	}

	slides := make([]domain.Slide, len(dbSlides))
	for i, dbSlide := range dbSlides {
		slides[i] = *dbSlide.toDomain()
	}

	return slides, nil
}
