package hub

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
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

type mockCollectorSlideRepo struct {
	slides []domain.Slide
	total  int64
	err    error
}

func (m *mockCollectorSlideRepo) AdminList(
	ctx context.Context,
	params repository.SlideListParams,
	filters repository.SlideListFilters,
) ([]domain.Slide, int64, error) {
	return m.slides, m.total, m.err
}

func (m *mockCollectorSlideRepo) Save(ctx context.Context, slide *domain.Slide) error {
	return nil
}

func (m *mockCollectorSlideRepo) GetActive(ctx context.Context) ([]domain.Slide, error) {
	return nil, nil
}

func (m *mockCollectorSlideRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockCollectorSlideRepo) GetByID(ctx context.Context, id int64) (*domain.Slide, error) {
	return nil, nil
}

func (m *mockCollectorSlideRepo) MarkAsDeleted(ctx context.Context, source, externalID string) error {
	return nil
}

func (m *mockCollectorSlideRepo) SlideExists(source, externalID string, subID *int) (bool, error) {
	return false, nil
}

func (m *mockCollectorSlideRepo) GetAllMediaURLs(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockCollectorSlideRepo) FindLocalURLByOriginalURL(
	ctx context.Context,
	originalURL string,
) (string, bool) {
	return "", false
}

func (m *mockCollectorSlideRepo) FindExpiredSlidesByType(
	ctx context.Context,
	slideType string,
	cutoff time.Time,
) ([]domain.Slide, error) {
	return nil, nil
}

func (m *mockCollectorSlideRepo) Sync(
	ctx context.Context,
	source string,
	newSlides []domain.Slide,
) (*repository.SlideSyncResult, error) {
	return nil, nil
}

func TestCollectorListExternalIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		repo := &mockCollectorSlideRepo{
			slides: []domain.Slide{
				{
					Source:     "bluesky",
					ExternalID: "ext-1",
				},
				{
					Source:     "bluesky",
					ExternalID: "ext-2",
				},
			},
			total: 2,
		}

		s := &Server{
			slideRepo: repo,
			logger:    newTestLoggerForHandler(),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = gin.Params{{Key: "source", Value: "bluesky"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/collectors/slides/bluesky?_start=0&_end=20", http.NoBody)

		c.Set(collectorSourceKey, "bluesky")
		c.Set(collectorTypeKey, "slide")

		s.collectorListExternalIDs(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-Total-Count"))

		var result []string

		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, []string{"ext-1", "ext-2"}, result)
	})

	t.Run("source mismatch unauthorized", func(t *testing.T) {
		s := &Server{
			logger: newTestLoggerForHandler(),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Params = gin.Params{{Key: "source", Value: "bluesky"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/api/collectors/slides/bluesky", http.NoBody)

		c.Set(collectorSourceKey, "different")
		c.Set(collectorTypeKey, "slide")

		s.collectorListExternalIDs(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
