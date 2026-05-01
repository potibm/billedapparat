package gorm

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (r *slideRepository) AdminList(
	ctx context.Context,
	p repository.AdminListParams,
	filters repository.AdminListFilters,
) ([]domain.Slide, int64, error) {
	var (
		dbSlides []dbSlide
		total    int64
	)

	query := r.db.WithContext(ctx).Model(&dbSlide{})

	// filters
	if p.Type != "" {
		query = query.Where("type = ?", p.Type)
	}

	if filters.Query != nil {
		likeQuery := fmt.Sprintf("%%%s%%", *filters.Query)
		query = query.Where(
			"content_title LIKE ? OR content_body LIKE ? OR author_display_name LIKE ? OR author_username LIKE ?",
			likeQuery,
			likeQuery,
			likeQuery,
			likeQuery,
		)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}

	if filters.ID != nil {
		query = query.Where("id = ?", *filters.ID)
	}

	if filters.Source != nil {
		query = query.Where("source = ?", *filters.Source)
	}

	// determine count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// sorting
	safeOrderClause := getOrderClause(p.Sort, p.Order)

	// perform query with pagination and sorting
	err := query.Order(safeOrderClause).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&dbSlides).Error
	if err != nil {
		return nil, 0, err
	}

	slides := toDomainSlice(dbSlides)

	return slides, total, nil
}

func getOrderClause(sortField, order string) string {
	var sortCols []string

	switch sortField {
	case "content.title":
		sortCols = []string{"content_title"}
	case "display_options.priority":
		sortCols = []string{"priority", "id"}
	case "author.display_name":
		sortCols = []string{"author_display_name", "id"}
	case "source":
		sortCols = []string{"source", "id"}
	default:
		sortCols = []string{"id"}
	}

	orderDir := "ASC"
	if strings.ToUpper(order) == "DESC" {
		orderDir = "DESC"
	}

	var orderClauses []string
	for _, col := range sortCols {
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", col, orderDir))
	}

	safeOrderClause := strings.Join(orderClauses, ", ")

	return safeOrderClause
}

func (r *slideRepository) GetByID(ctx context.Context, id int64) (*domain.Slide, error) {
	var dbModel dbSlide
	if err := r.db.WithContext(ctx).First(&dbModel, id).Error; err != nil {
		return nil, err
	}

	return dbModel.toDomain(), nil
}

func (r *slideRepository) GetAllMediaURLs(ctx context.Context) ([]string, error) {
	var urls []string

	err := r.db.WithContext(ctx).
		Model(&domain.Slide{}).
		Where("media_url_original != ?", "").
		Pluck("media_url_original", &urls).Error

	return urls, err
}

func (r *slideRepository) SlideExists(source, externalID string, subID *int) (bool, error) {
	var count int64

	query := r.db.Model(&dbSlide{}).Where("source = ? AND external_id = ?", source, externalID)

	if subID != nil {
		query = query.Where("external_sub_id = ?", *subID)
	} else {
		query = query.Where("external_sub_id IS NULL")
	}

	err := query.Count(&count).Error

	return count > 0, err
}

func (r *slideRepository) FindLocalURLByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	var localURL string

	query := `
        SELECT content_media_url_local AS local_url 
        FROM slides 
        WHERE content_media_url_original = ? AND content_media_url_local != ''
        UNION ALL
        SELECT author_avatar_url_local AS local_url 
        FROM slides 
        WHERE author_avatar_url_original = ? AND author_avatar_url_local != ''
        LIMIT 1
    `

	err := r.db.WithContext(ctx).Raw(query, originalURL, originalURL).Scan(&localURL).Error

	if err != nil || localURL == "" {
		return "", false // Cache Miss
	}

	return localURL, true // Cache Hit!
}

func (r *slideRepository) MarkAsDeleted(ctx context.Context, source, externalID string) error {
	return r.db.WithContext(ctx).
		Model(&dbSlide{}).
		Where("source = ? AND external_id = ?", source, externalID).
		Update("status", domain.StatusDeleted).Error
}

func (r *slideRepository) FindExpiredSlidesByType(
	ctx context.Context,
	slideType string,
	cutoff time.Time,
) ([]domain.Slide, error) {
	var dbSlides []dbSlide

	err := r.db.WithContext(ctx).
		Where("type = ? AND status = ? AND created_at < ?", slideType, domain.StatusActive, cutoff).
		Find(&dbSlides).Error
	if err != nil {
		return nil, err
	}

	slides := toDomainSlice(dbSlides)

	return slides, nil
}
