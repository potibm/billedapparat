package hub

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockMediaProcessor struct{}

func (m *MockMediaProcessor) ProcessSlideImage(c *gin.Context, formField string) (string, error) {
	return "/media/fake-pfad-fuer-test.webp", nil
}

func TestParseSlidePayload_MultipartSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s := &Server{
		mediaProcessor: &MockMediaProcessor{},
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("status", "pending")
	_ = writer.WriteField("content.title", "My Test Slide")
	_ = writer.WriteField("display_options.priority", "5")

	part, err := writer.CreateFormFile("image_upload", "testbild.jpg")
	require.NoError(t, err)

	_, _ = part.Write([]byte("fake image content..."))

	writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/slides", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	slide, err := s.parseSlidePayload(c)

	require.NoError(t, err, "Parsing should not return an error")
	require.NotNil(t, slide, "Slide is not nil")

	assert.Equal(t, domain.StatusPending, slide.Status, "Status should be 'pending'")
	assert.Equal(t, "My Test Slide", slide.Content.Title)
	assert.Equal(t, 5, slide.DisplayOptions.Priority)

	require.NotNil(t, slide.Content.Media, "Attention! Media should not be nil when an image is uploaded")
	assert.Equal(t, "image/webp", slide.Content.Media.MimeType)
	assert.Equal(t, "/media/fake-pfad-fuer-test.webp", slide.Content.Media.LocalURL)
}
