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
	return r.db.WithContext(ctx).Delete(&dbNews{}, "source = ? AND external_id = ?", source, externalID).Error
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

func (r *newsRepository) Sync(
	ctx context.Context,
	source string,
	incoming []domain.News,
) (*repository.NewsSyncResult, error) {
	result := &repository.NewsSyncResult{}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []domain.News
		if err := tx.Where("source = ?", source).Find(&existing).Error; err != nil {
			return err
		}

		existingMap := make(map[string]domain.News)
		for _, e := range existing {
			existingMap[e.ExternalID] = e
		}

		incomingMap := make(map[string]bool)

		for _, inc := range incoming {
			incomingMap[inc.ExternalID] = true

			if old, exists := existingMap[inc.ExternalID]; exists {
				inc.ID = old.ID
				if err := tx.Save(&inc).Error; err != nil {
					return err
				}

				result.Updated = append(result.Updated, inc)
			} else {
				if err := tx.Create(&inc).Error; err != nil {
					return err
				}

				result.Created = append(result.Created, inc)
			}
		}

		for _, ext := range existing {
			if !incomingMap[ext.ExternalID] {
				if err := tx.Delete(&ext).Error; err != nil {
					return err
				}

				result.Deleted = append(result.Deleted, ext)
			}
		}

		return nil
	})

	return result, err
}

func (r *newsRepository) applyFilters(query *gorm.DB, filters repository.NewsListFilters) *gorm.DB {
	if filters.Query != nil {
		query = query.Where("title ILIKE ?", "%"+*filters.Query+"%")
	}

	if filters.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *filters.IsUrgent)
	}

	if filters.IsHidden != nil {
		query = query.Where("is_hidden = ?", *filters.IsHidden)
	}

	return query
}

func (r *newsRepository) getOrderClause(_, order string) string {
	sortCols := []string{"id"}

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
