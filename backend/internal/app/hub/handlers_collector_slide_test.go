package hub

import (
	"context"
	"log/slog"
	"testing"

	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

// mockFilterRuleRepo implements repository.FilterRuleRepository for testing.
type mockFilterRuleRepo struct {
	rules []domain.FilterRule
	err   error
}

func (m *mockFilterRuleRepo) List(
	ctx context.Context,
	limit, offset int,
) ([]domain.FilterRule, int64, error) {
	return m.rules, int64(len(m.rules)), m.err
}

func (m *mockFilterRuleRepo) GetByID(ctx context.Context, id uint) (*domain.FilterRule, error) {
	return nil, nil
}

func (m *mockFilterRuleRepo) Create(ctx context.Context, rule *domain.FilterRule) error {
	return nil
}

func (m *mockFilterRuleRepo) Update(ctx context.Context, rule *domain.FilterRule) error {
	return nil
}

func (m *mockFilterRuleRepo) Delete(ctx context.Context, id uint) error {
	return nil
}

func newTestLoggerForHandler() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func authorPtr(username, displayName string) *contracts.IngestSlideRequestAuthor {
	return &contracts.IngestSlideRequestAuthor{
		ExternalID:  "auth-ext-1",
		Username:    username,
		DisplayName: displayName,
	}
}

func TestEvaluateModerationRules_NoRules(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{rules: nil},
		logger:         newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("testuser", "Test User"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusPending, status)
}

func TestEvaluateModerationRules_MatchingUsername(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeUsername, Value: "SpamBot"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("spambot", "Friendly Name"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusFiltered, status)
}

func TestEvaluateModerationRules_MatchingDisplayName(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeDisplayName, Value: "Crypto"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("normaluser", "Get your CRYPTO now!"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusFiltered, status)
}

func TestEvaluateModerationRules_MatchingLanguage(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeLanguage, Value: "ru"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source:   "mastodon",
		Author:   authorPtr("user", "User"),
		Language: "ru",
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusFiltered, status)
}

func TestEvaluateModerationRules_SourceMismatch(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "instagram", Type: domain.FilterTypeUsername, Value: "troll"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("troll", "Troll"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusPending, status, "rule source 'instagram' should not match 'mastodon'")
}

func TestEvaluateModerationRules_NilAuthor(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeUsername, Value: "someuser"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: nil,
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusPending, status, "nil author should not match any rule")
}

func TestEvaluateModerationRules_RepoError(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			err: assert.AnError,
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("baduser", "Bad User"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusPending, status, "repo error should fall back to pending")
}

func TestEvaluateModerationRules_WildcardSource(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeUsername, Value: "blocked"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source: "mastodon",
		Author: authorPtr("blocked", "Blocked User"),
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusFiltered, status)
}

func TestEvaluateModerationRules_MultipleRulesFirstMatch(t *testing.T) {
	s := &Server{
		filterRuleRepo: &mockFilterRuleRepo{
			rules: []domain.FilterRule{
				{Source: "*", Type: domain.FilterTypeDisplayName, Value: "Spam"},
				{Source: "*", Type: domain.FilterTypeLanguage, Value: "en"},
			},
		},
		logger: newTestLoggerForHandler(),
	}

	req := contracts.IngestSlideRequest{
		Source:   "mastodon",
		Author:   authorPtr("user", "Spam bot!"),
		Language: "en",
	}

	status := s.evaluateModerationRules(context.Background(), req)

	assert.Equal(t, domain.StatusFiltered, status, "first matching rule should filter")
}
