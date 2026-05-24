package gorm

import (
	"context"
	"fmt"
	"log/slog"
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
	dbObj := fromDomainSlide(slide)

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

	slides := toDomainSlideList(dbSlides)

	return slides, nil
}

func (r *slideRepository) AdminList(
	ctx context.Context,
	p repository.SlideListParams,
	filters repository.SlideListFilters,
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

	query = r.applyFilters(query, filters)

	// determine count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// sorting
	safeOrderClause := r.getOrderClause(p.Sort, p.Order)

	// perform query with pagination and sorting
	err := query.Order(safeOrderClause).
		Limit(p.Limit).
		Offset(p.Offset).
		Find(&dbSlides).Error
	if err != nil {
		return nil, 0, err
	}

	slides := toDomainSlideList(dbSlides)

	return slides, total, nil
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

	slides := toDomainSlideList(dbSlides)

	return slides, nil
}

func (r *slideRepository) Sync(
	ctx context.Context,
	source string,
	incoming []domain.Slide,
) (*repository.SlideSyncResult, error) {
	result := &repository.SlideSyncResult{}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbExisting []dbSlide
		if err := tx.Where("source = ?", source).Find(&dbExisting).Error; err != nil {
			return err
		}

		slog.Debug("Existing slides fetched from DB", "count", len(dbExisting), "source", source)

		existing := toDomainSlideList(dbExisting)

		toCreate, toUpdate, toDelete := diffSlides(existing, incoming)
		slog.Debug(
			"Sync diff calculated",
			"to_create",
			len(toCreate),
			"to_update",
			len(toUpdate),
			"to_delete",
			len(toDelete),
		)

		if err := r.insertNew(tx, toCreate, result); err != nil {
			return err
		}

		if err := r.updateExisting(tx, toUpdate, result); err != nil {
			return err
		}

		if err := r.deleteObsolete(tx, toDelete, result); err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func (r *slideRepository) getOrderClause(sortField, order string) string {
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

func (r *slideRepository) applyFilters(query *gorm.DB, filters repository.SlideListFilters) *gorm.DB {
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

	return query
}

func diffSlides(existing, incoming []domain.Slide) (toCreate, toUpdate, toDelete []domain.Slide) {
	existingMap := make(map[string]domain.Slide)
	for _, e := range existing {
		existingMap[e.SyncKey()] = e
	}

	incomingMap := make(map[string]bool)

	for _, inc := range incoming {
		key := inc.SyncKey()
		incomingMap[key] = true

		if old, exists := existingMap[key]; exists {
			if old.HasChanged(inc) {
				inc.ID = old.ID
				toUpdate = append(toUpdate, inc)
			}
		} else {
			toCreate = append(toCreate, inc)
		}
	}

	for _, ext := range existing {
		if !incomingMap[ext.SyncKey()] {
			toDelete = append(toDelete, ext)
		}
	}

	return toCreate, toUpdate, toDelete
}

func (r *slideRepository) insertNew(tx *gorm.DB, items []domain.Slide, res *repository.SlideSyncResult) error {
	if len(items) == 0 {
		return nil
	}

	dbItems := fromDomainSlideList(items)

	if err := tx.Create(&dbItems).Error; err != nil {
		return err
	}

	for i, dbItem := range dbItems {
		items[i].ID = dbItem.ID
	}

	res.Created = items

	return nil
}

func (r *slideRepository) updateExisting(tx *gorm.DB, items []domain.Slide, res *repository.SlideSyncResult) error {
	if len(items) == 0 {
		return nil
	}

	for _, item := range items {
		itemCopy := item
		dbObj := fromDomainSlide(&itemCopy)

		if err := tx.Save(dbObj).Error; err != nil {
			return err
		}

		item.ID = dbObj.ID
		res.Updated = append(res.Updated, item)
	}

	return nil
}

func (r *slideRepository) deleteObsolete(tx *gorm.DB, items []domain.Slide, res *repository.SlideSyncResult) error {
	if len(items) == 0 {
		return nil
	}

	dbItems := fromDomainSlideList(items)

	if err := tx.Delete(&dbItems).Error; err != nil {
		return err
	}

	res.Deleted = items

	return nil
}
