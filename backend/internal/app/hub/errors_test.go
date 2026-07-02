package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondWithProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generic problem response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)

		respondWithProblem(c, http.StatusBadRequest, "Bad Request", "Something went wrong")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/problem+json", w.Header().Get("Content-Type"))

		var body ProblemDetails

		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)

		assert.Equal(t, "about:blank", body.Type)
		assert.Equal(t, "Bad Request", body.Title)
		assert.Equal(t, http.StatusBadRequest, body.Status)
		assert.Equal(t, "Something went wrong", body.Detail)
		assert.Equal(t, "/api/test", body.Instance)
	})
}

func TestRespondWithInvalidIDFormatProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/slides/abc", http.NoBody)

	respondWithInvalidIDFormatProblem(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Bad Request", body.Title)
	assert.Contains(t, body.Detail, "Invalid ID format")
}

func TestRespondWithInternalServerProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/slides", http.NoBody)

	respondWithInternalServerProblem(c, "database connection failed")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Internal Server Error", body.Title)
	assert.Equal(t, "database connection failed", body.Detail)
}

func TestRespondWithNotFoundProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/slides/999", http.NoBody)

	respondWithNotFoundProblem(c, "Slide not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Not Found", body.Title)
	assert.Equal(t, "Slide not found", body.Detail)
}

func TestRespondWithFailedToParsePayloadProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/slides", http.NoBody)

	respondWithFailedToParsePayloadProblem(c, assert.AnError)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Bad Request", body.Title)
	assert.Contains(t, body.Detail, "Failed to parse payload")
}

func TestRespondWithBadRequestProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/slides", http.NoBody)

	respondWithBadRequestProblem(c, "Missing required field: title")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Bad Request", body.Title)
	assert.Equal(t, "Missing required field: title", body.Detail)
}

func TestRespondWithUnauthorizedProblem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/slides", http.NoBody)

	respondWithUnauthorizedProblem(c, "Invalid API key")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body ProblemDetails

	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "Unauthorized", body.Title)
	assert.Equal(t, "Invalid API key", body.Detail)
}
