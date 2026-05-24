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

type eventRepository struct {
	db *gorm.DB
}

func (s *Store) NewTimetableEventRepository() repository.TimetableEventRepository {
	return NewTimetableEventRepository(s.db)
}

func NewTimetableEventRepository(db *gorm.DB) repository.TimetableEventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Save(ctx context.Context, event *domain.TimetableEvent) error {
	dbObj := fromDomainTimetableEvent(event)

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source"}, {Name: "external_id"}},
		UpdateAll: true,
	}).Create(dbObj).Error
	if err == nil {
		event.ID = dbObj.ID
	}

	return err
}

func (r *eventRepository) Delete(ctx context.Context, source, externalID string) error {
	return r.db.WithContext(ctx).Delete(&dbTimetableEvent{}, "source = ? AND external_id = ?", source, externalID).Error
}

func (r *eventRepository) GetAll(ctx context.Context) (domain.Timetable, error) {
	var dbEventList []dbTimetableEvent

	err := r.db.WithContext(ctx).Find(&dbEventList).Error
	if err != nil {
		return nil, err
	}

	eventList := toDomainTimetableEventList(dbEventList)

	return eventList, nil
}

func (r *eventRepository) GetActive(ctx context.Context) (domain.Timetable, error) {
	var dbEventList []dbTimetableEvent

	err := r.db.WithContext(ctx).
		Where("is_hidden = ? AND end_time > ?", false, time.Now()).
		Order("start_time ASC, end_time ASC").
		Find(&dbEventList).Error
	if err != nil {
		return nil, err
	}

	eventList := toDomainTimetableEventList(dbEventList)

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
	params repository.TimetableEventListParams,
	filters repository.TimetableEventListFilters,
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

	eventList := toDomainTimetableEventList(dbTimetableEvents)

	return eventList, count, nil
}

//nolint:dupl // some code is similar to news repository but the domain models are different enough.
func (r *eventRepository) Sync(
	ctx context.Context,
	source string,
	incoming []domain.TimetableEvent,
) (*repository.TimetableEventSyncResult, error) {
	result := &repository.TimetableEventSyncResult{}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dbExisting []dbTimetableEvent
		if err := tx.Where("source = ?", source).Find(&dbExisting).Error; err != nil {
			return err
		}

		existing := toDomainTimetableEventList(dbExisting)

		toCreate, toUpdate, toDelete := diffTimetableEvents(existing, incoming)

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

func (r *eventRepository) insertNew(
	tx *gorm.DB,
	items []domain.TimetableEvent,
	res *repository.TimetableEventSyncResult,
) error {
	if len(items) == 0 {
		return nil // Early return
	}

	dbItems := make([]*dbTimetableEvent, 0, len(items))
	for _, item := range items {
		itemCopy := item
		dbItems = append(dbItems, fromDomainTimetableEvent(&itemCopy))
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

func (r *eventRepository) updateExisting(
	tx *gorm.DB,
	items []domain.TimetableEvent,
	res *repository.TimetableEventSyncResult,
) error {
	for _, item := range items {
		itemCopy := item
		dbObj := fromDomainTimetableEvent(&itemCopy)

		if err := tx.Save(dbObj).Error; err != nil {
			return err
		}

		item.ID = dbObj.ID
		res.Updated = append(res.Updated, item)
	}

	return nil
}

func (r *eventRepository) deleteObsolete(
	tx *gorm.DB,
	items []domain.TimetableEvent,
	res *repository.TimetableEventSyncResult,
) error {
	if len(items) == 0 {
		return nil // Early return
	}

	dbItems := make([]*dbTimetableEvent, 0, len(items))
	for _, item := range items {
		itemCopy := item
		dbItems = append(dbItems, fromDomainTimetableEvent(&itemCopy))
	}

	if err := tx.Delete(&dbItems).Error; err != nil {
		return err
	}

	res.Deleted = items

	return nil
}

func diffTimetableEvents(
	existing, incoming []domain.TimetableEvent,
) (toCreate, toUpdate, toDelete []domain.TimetableEvent) {
	existingMap := make(map[string]domain.TimetableEvent)
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

func (r *eventRepository) applyFilters(query *gorm.DB, filters repository.TimetableEventListFilters) *gorm.DB {
	if filters.Query != nil {
		likeQuery := fmt.Sprintf("%%%s%%", *filters.Query)
		query = query.Where("title LIKE ?", likeQuery)
	}

	if filters.IsHidden != nil {
		query = query.Where("is_hidden = ?", *filters.IsHidden)
	}

	if filters.Source != nil {
		query = query.Where("source = ?", *filters.Source)
	}

	if filters.ExternalID != nil {
		query = query.Where("external_id = ?", *filters.ExternalID)
	}

	return query
}

func (r *eventRepository) getOrderClause(sortField, order string) string {
	var sortCols []string

	switch sortField {
	case "id":
		sortCols = []string{"id"}
	case "source":
		sortCols = []string{"source", "external_id"}
	case "external_id":
		sortCols = []string{"external_id"}
	case "title":
		sortCols = []string{"title", "start_time", "end_time", "id"}
	case "start_time":
		sortCols = []string{"start_time", "end_time", "id"}
	case "end_time":
		sortCols = []string{"end_time", "id"}
	case "is_hidden":
		sortCols = []string{"is_hidden", "start_time", "end_time", "id"}
	default:
		sortCols = []string{"start_time", "end_time", "id"}
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
