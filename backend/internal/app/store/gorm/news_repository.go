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
		UpdateAll: true,
	}).Create(dbObj).Error
	if err == nil {
		news.ID = dbObj.ID
	}

	return err
}

func (r *newsRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&dbNews{}, id).Error
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
