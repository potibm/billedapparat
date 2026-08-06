package gorm

import (
	"context"
	"fmt"
	"strings"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type newsRepository struct {
	db *gorm.DB
}

func (s *Store) NewNewsRepository() repository.NewsRepository {
	return NewNewsRepository(s.db)
}

func NewNewsRepository(db *gorm.DB) repository.NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Save(ctx context.Context, news *domain.News) error {
	dbObj := fromDomainNews(news)

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source"}, {Name: "external_id"}},
		UpdateAll: true,
	}).Create(dbObj).Error
	if err == nil {
		news.ID = dbObj.ID
	}

	return err
}

func (r *newsRepository) Delete(ctx context.Context, source, externalID string) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Delete(&dbNews{}, "source = ? AND external_id = ?", source, externalID).
		Error
}

func (r *newsRepository) GetAll(ctx context.Context) ([]domain.News, error) {
	var dbNewsList []dbNews

	err := r.db.WithContext(ctx).Find(&dbNewsList).Error
	if err != nil {
		return nil, err
	}

	newsList := toDomainNewsList(dbNewsList)

	return newsList, nil
}

func (r *newsRepository) GetByID(ctx context.Context, id int64) (*domain.News, error) {
	var dbObj dbNews

	err := r.db.WithContext(ctx).First(&dbObj, id).Error
	if err != nil {
		return nil, err
	}

	return dbObj.toDomain(), nil
}

func (r *newsRepository) List(
	ctx context.Context,
	params repository.NewsListParams,
	filters repository.NewsListFilters,
) ([]domain.News, int64, error) {
	var (
		dbNewsList []dbNews
		count      int64
	)

	query := r.db.WithContext(ctx).Model(&dbNews{})

	query = r.applyFilters(query, filters)

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	safeOrderClause := r.getOrderClause(params.Sort, params.Order)

	err = query.Order(safeOrderClause).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&dbNewsList).Error
	if err != nil {
		return nil, 0, err
	}

	newsList := toDomainNewsList(dbNewsList)

	return newsList, count, nil
}

//nolint:dupl // some code is similar to timetable repository but the domain models are different enough.
func (r *newsRepository) Sync(
	ctx context.Context,
	source string,
	incoming []domain.News,
) (*repository.NewsSyncResult, error) {
	result := &repository.NewsSyncResult{}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbExisting []dbNews
		if err := tx.Where("source = ?", source).Find(&dbExisting).Error; err != nil {
			return err
		}

		existing := toDomainNewsList(dbExisting)

		toCreate, toUpdate, toDelete := diffNews(existing, incoming)

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

func (r *newsRepository) insertNew(tx *gorm.DB, items []domain.News, res *repository.NewsSyncResult) error {
	if len(items) == 0 {
		return nil // Early return
	}

	dbItems := make([]*dbNews, 0, len(items))
	for _, item := range items {
		itemCopy := item
		dbItems = append(dbItems, fromDomainNews(&itemCopy))
	}

	if err := tx.Create(&dbItems).Error; err != nil {
		return err
	}

	for i, dbItem := range dbItems {
		items[i].ID = dbItem.ID
	}

	res.Created = items

	return nil
}

func (r *newsRepository) updateExisting(tx *gorm.DB, items []domain.News, res *repository.NewsSyncResult) error {
	for _, item := range items {
		itemCopy := item
		dbObj := fromDomainNews(&itemCopy)

		if err := tx.Save(dbObj).Error; err != nil {
			return err
		}

		item.ID = dbObj.ID
		res.Updated = append(res.Updated, item)
	}

	return nil
}

func (r *newsRepository) deleteObsolete(tx *gorm.DB, items []domain.News, res *repository.NewsSyncResult) error {
	if len(items) == 0 {
		return nil // Early return
	}

	dbItems := make([]*dbNews, 0, len(items))
	for _, item := range items {
		itemCopy := item
		dbItems = append(dbItems, fromDomainNews(&itemCopy))
	}

	if err := tx.Unscoped().Delete(&dbItems).Error; err != nil {
		return err
	}

	res.Deleted = items

	return nil
}

func diffNews(existing, incoming []domain.News) (toCreate, toUpdate, toDelete []domain.News) {
	existingMap := make(map[string]domain.News)
	for _, e := range existing {
		existingMap[e.ExternalID] = e
	}

	incomingMap := make(map[string]bool)

	for _, inc := range incoming {
		incomingMap[inc.ExternalID] = true

		if old, exists := existingMap[inc.ExternalID]; exists {
			inc.ID = old.ID // use the existing ID for updates
			toUpdate = append(toUpdate, inc)
		} else {
			toCreate = append(toCreate, inc)
		}
	}

	for _, ext := range existing {
		if !incomingMap[ext.ExternalID] {
			toDelete = append(toDelete, ext)
		}
	}

	return toCreate, toUpdate, toDelete
}

func (r *newsRepository) applyFilters(query *gorm.DB, filters repository.NewsListFilters) *gorm.DB {
	if filters.Query != nil {
		likeQuery := fmt.Sprintf("%%%s%%", *filters.Query)
		query = query.Where("title LIKE ?", likeQuery)
	}

	if filters.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *filters.IsUrgent)
	}

	if filters.IsHidden != nil {
		query = query.Where("is_hidden = ?", *filters.IsHidden)
	}

	if filters.ExternalID != nil {
		query = query.Where("external_id = ?", *filters.ExternalID)
	}

	return query
}

func (r *newsRepository) getOrderClause(sortField, order string) string {
	var sortCols []string

	switch sortField {
	case "id":
		sortCols = []string{"id"}
	case "source":
		sortCols = []string{"source", "external_id"}
	case "external_id":
		sortCols = []string{"external_id"}
	case "title":
		sortCols = []string{"title", "id"}
	case "is_urgent":
		sortCols = []string{"is_urgent", "id"}
	case "is_hidden":
		sortCols = []string{"is_hidden", "id"}
	default:
		sortCols = []string{"is_urgent", "id"}
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
