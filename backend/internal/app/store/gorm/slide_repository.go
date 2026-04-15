package gorm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type slideRepository struct {
	db *gorm.DB
}

func (s *Store) NewSlideRepository() repository.SlideRepository {
	return NewSlideRepository(s.db)
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
		slide.ID = dbObj.ID
	}

	return err
}

func (r *slideRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&dbSlide{}, id).Error
}

func (r *slideRepository) GetActive(ctx context.Context) ([]domain.Slide, error) {
	var dbSlides []dbSlide

	err := r.db.WithContext(ctx).Find(&dbSlides).Error
	if err != nil {
		return nil, err
	}

	slides := toDomainSlice(dbSlides)

	return slides, nil
}

func (r *slideRepository) AdminList(ctx context.Context, p repository.AdminListParams) ([]domain.Slide, int64, error) {
	slog.Info("AdminLists called")

	var (
		dbSlides []dbSlide
		total    int64
	)

	query := r.db.WithContext(ctx).Model(&dbSlide{})

	if p.Type != "" {
		query = query.Where("type = ?", p.Type)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := fmt.Sprintf("%s %s", p.Sort, p.Order)

	err := query.Order(orderClause).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&dbSlides).Error
	if err != nil {
		return nil, 0, err
	}

	slides := toDomainSlice(dbSlides)

	return slides, total, nil
}

func (r *slideRepository) GetByID(ctx context.Context, id int64) (*domain.Slide, error) {
	var dbModel dbSlide
	if err := r.db.WithContext(ctx).First(&dbModel, id).Error; err != nil {
		return nil, err
	}

	return dbModel.toDomain(), nil
}
