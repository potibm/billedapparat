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

type eventRepository struct {
	db *gorm.DB
}

func (s *Store) NewEventRepository() repository.TimetableEventRepository {
	return NewEventRepository(s.db)
}

func NewEventRepository(db *gorm.DB) repository.TimetableEventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Save(ctx context.Context, event *domain.TimetableEvent) error {
	dbObj := fromDomainEvent(event)

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(dbObj).Error
	if err == nil {
		event.ID = dbObj.ID
	}

	return err
}

func (r *eventRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&dbTimetableEvent{}, id).Error
}

func (r *eventRepository) GetAll(ctx context.Context) (domain.Timetable, error) {
	var dbEventList []dbTimetableEvent

	err := r.db.WithContext(ctx).Find(&dbEventList).Error
	if err != nil {
		return nil, err
	}

	eventList := toDomainTimetable(dbEventList)

	return eventList, nil
}

func (r *eventRepository) GetByID(ctx context.Context, id int64) (*domain.TimetableEvent, error) {
	var dbObj dbTimetableEvent

	err := r.db.WithContext(ctx).First(&dbObj, id).Error
	if err != nil {
		return nil, err
	}

	return dbObj.toDomain(), nil
}

func (r *eventRepository) List(
	ctx context.Context,
	params repository.EventListParams,
	filters repository.EventListFilters,
) (domain.Timetable, int64, error) {
	var (
		dbTimetableEvents []dbTimetableEvent
		count             int64
	)

	query := r.db.WithContext(ctx).Model(&dbTimetableEvent{})

	query = r.applyFilters(query, filters)

	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	safeOrderClause := r.getOrderClause(params.Sort, params.Order)

	err = query.Order(safeOrderClause).
		Limit(params.Limit).
		Offset(params.Offset).
		Find(&dbTimetableEvents).Error
	if err != nil {
		return nil, 0, err
	}

	eventList := toDomainTimetable(dbTimetableEvents)

	return eventList, count, nil
}

func (r *eventRepository) applyFilters(query *gorm.DB, filters repository.EventListFilters) *gorm.DB {
	if filters.Query != nil {
		query = query.Where("title ILIKE ?", "%"+*filters.Query+"%")
	}

	if filters.IsHidden != nil {
		query = query.Where("is_hidden = ?", *filters.IsHidden)
	}

	return query
}

func (r *eventRepository) getOrderClause(sortField, order string) string {
	var sortCols []string

	switch sortField {
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
