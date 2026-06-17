package gorm

import (
	"github.com/potibm/billedapparat/internal/app/domain"
)

type dbFilterRule struct {
	GormModel
	AuditModel

	Source string            `gorm:"type:varchar(50);index"`
	Type   domain.FilterType `gorm:"type:varchar(50);index"`
	Value  string            `gorm:"type:varchar(255);index"`
}

func (dbFilterRule) TableName() string {
	return "filter_rules"
}

func fromDomainFilterRule(s *domain.FilterRule) *dbFilterRule {
	return &dbFilterRule{
		GormModel: GormModel{ID: s.ID},

		Source: s.Source,
		Type:   s.Type,
		Value:  s.Value,
	}
}

func (s *dbFilterRule) toDomain() *domain.FilterRule {
	fr := &domain.FilterRule{
		ID:     s.ID,
		Source: s.Source,
		Type:   s.Type,
		Value:  s.Value,
	}

	return fr
}

func toDomainFilterRuleSlice(dbFilterRules []dbFilterRule) []domain.FilterRule {
	filterRules := make([]domain.FilterRule, len(dbFilterRules))
	for i, s := range dbFilterRules {
		filterRules[i] = *s.toDomain()
	}

	return filterRules
}
