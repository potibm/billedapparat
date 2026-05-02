package gorm

import (
	"testing"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

func TestFilterRuleMapping(t *testing.T) {
	// 1. Test toDomain()
	t.Run("dbFilterRule to domain.FilterRule", func(t *testing.T) {
		dbRule := dbFilterRule{
			GormModel: GormModel{ID: 42},
			Source:    "mastodon",
			Type:      domain.FilterTypeUsername,
			Value:     "spambot",
		}

		domainRule := dbRule.toDomain()

		assert.Equal(t, int64(42), domainRule.ID)
		assert.Equal(t, "mastodon", domainRule.Source)
		assert.Equal(t, domain.FilterTypeUsername, domainRule.Type)
		assert.Equal(t, "spambot", domainRule.Value)
	})

	// 2. Test fromDomainFilterRule()
	t.Run("domain.FilterRule to dbFilterRule", func(t *testing.T) {
		domainRule := &domain.FilterRule{
			ID:     99,
			Source: "instagram",
			Type:   domain.FilterTypeDisplayName,
			Value:  "free money",
		}

		dbRule := fromDomainFilterRule(domainRule)

		assert.Equal(t, int64(99), dbRule.ID)
		assert.Equal(t, "instagram", dbRule.Source)
		assert.Equal(t, domain.FilterTypeDisplayName, dbRule.Type)
		assert.Equal(t, "free money", dbRule.Value)
	})

	// 3. Test toDomainFilterRuleSlice()
	t.Run("Slice mapping", func(t *testing.T) {
		dbRules := []dbFilterRule{
			{GormModel: GormModel{ID: 1}, Source: "src1", Type: domain.FilterTypeLanguage, Value: "ru"},
			{GormModel: GormModel{ID: 2}, Source: "src2", Type: domain.FilterTypeUsername, Value: "bot"},
		}

		domainRules := toDomainFilterRuleSlice(dbRules)

		assert.Len(t, domainRules, 2)
		assert.Equal(t, int64(1), domainRules[0].ID)
		assert.Equal(t, "src1", domainRules[0].Source)
		assert.Equal(t, int64(2), domainRules[1].ID)
		assert.Equal(t, "src2", domainRules[1].Source)
	})
}
