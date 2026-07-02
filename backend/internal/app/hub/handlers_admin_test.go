package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
	"github.com/stretchr/testify/assert"
)

func TestGetListParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := &Server{}

	tests := []struct {
		name       string
		query      string
		expected   repository.ListParams
		headerOpts map[string]string
	}{
		{
			name:  "default pagination",
			query: "",
			expected: repository.ListParams{
				Offset: 0,
				Limit:  20,
				Sort:   "id",
				Order:  "DESC",
			},
		},
		{
			name:  "custom start and end",
			query: "_start=10&_end=30",
			expected: repository.ListParams{
				Offset: 10,
				Limit:  20,
				Sort:   "id",
				Order:  "DESC",
			},
		},
		{
			name:  "custom sort and order",
			query: "_sort=created_at&_order=ASC",
			expected: repository.ListParams{
				Offset: 0,
				Limit:  20,
				Sort:   "created_at",
				Order:  "ASC",
			},
		},
		{
			name:  "negative start clamped to zero",
			query: "_start=-5&_end=10",
			expected: repository.ListParams{
				Offset: 0,
				Limit:  15,
				Sort:   "id",
				Order:  "DESC",
			},
		},
		{
			name:  "end smaller than start gives zero limit",
			query: "_start=50&_end=20",
			expected: repository.ListParams{
				Offset: 50,
				Limit:  0,
				Sort:   "id",
				Order:  "DESC",
			},
		},
		{
			name:  "all custom params",
			query: "_start=5&_end=15&_sort=title&_order=ASC",
			expected: repository.ListParams{
				Offset: 5,
				Limit:  10,
				Sort:   "title",
				Order:  "ASC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/slides?"+tt.query, http.NoBody)

			got := s.getListParams(c)
			assert.Equal(t, tt.expected, got)
		})
	}
}
