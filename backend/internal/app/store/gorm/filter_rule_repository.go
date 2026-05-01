package gorm

import (
	"context"
	"errors"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"gorm.io/gorm"
)

type filterRuleRepository struct {
	db *gorm.DB
}

func (s *Store) NewFilterRuleRepository() repository.FilterRuleRepository {
	return NewFilterRuleRepository(s.db)
}

func NewFilterRuleRepository(db *gorm.DB) repository.FilterRuleRepository {
	return &filterRuleRepository{db: db}
}

func (r *filterRuleRepository) List(ctx context.Context, limit, offset int) ([]domain.FilterRule, int64, error) {
	var (
		dbRules []dbFilterRule
		total   int64
	)

	query := r.db.WithContext(ctx).Model(&dbFilterRule{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbRules).Error
	if err != nil {
		return nil, 0, err
	}

	rules := toDomainFilterRuleSlice(dbRules)

	return rules, total, nil
}

func (r *filterRuleRepository) GetByID(ctx context.Context, id uint) (*domain.FilterRule, error) {
	var dbRule dbFilterRule

	err := r.db.WithContext(ctx).First(&dbRule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return dbRule.toDomain(), nil
}

func (r *filterRuleRepository) Create(ctx context.Context, rule *domain.FilterRule) error {
	dbRule := fromDomainFilterRule(rule)

	if err := r.db.WithContext(ctx).Create(dbRule).Error; err != nil {
		return err
	}

	rule.ID = dbRule.ID

	return nil
}

func (r *filterRuleRepository) Update(ctx context.Context, rule *domain.FilterRule) error {
	dbRule := fromDomainFilterRule(rule)

	return r.db.WithContext(ctx).
		Model(&dbFilterRule{}).
		Where("id = ?", rule.ID).
		Omit("CreatedAt").
		Updates(dbRule).Error
}

func (r *filterRuleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&dbFilterRule{}, id).Error
}
