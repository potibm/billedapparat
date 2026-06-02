package hubclient

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func setupTestClient(mockTransport http.RoundTripper) *HubClient {
	logger := slog.New(slog.DiscardHandler)
	client := New("http://dummy-hub.local", "test-key", logger)
	client.HTTPClient.Transport = mockTransport

	return client
}

func TestHubClient_WaitForServer_SuccessImmediately(t *testing.T) {
	transport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	client := setupTestClient(transport)
	ctx := context.Background()

	err := client.WaitForServer(ctx, 3, 1*time.Millisecond)

	assert.NoError(t, err, "Es wurde kein Fehler erwartet, da der Server direkt antwortet")
}

func TestHubClient_WaitForServer_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	transport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			attempts++
			// the first two requests fail
			if attempts < 3 {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			}
			// the third request succeeds
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	client := setupTestClient(transport)
	ctx := context.Background()

	err := client.WaitForServer(ctx, 5, 1*time.Millisecond)

	assert.NoError(t, err, "The server should have become reachable within the retry limit")
	assert.Equal(t, 3, attempts, "There should have been exactly 3 attempts before success")
}

func TestHubClient_WaitForServer_ContextCancellation(t *testing.T) {
	transport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			// The server is permanently down (HTTP 500)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	client := setupTestClient(transport)

	// We create a context that cancels itself after 10 milliseconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Action: We tell it to try 100 times and wait 5 milliseconds between each attempt
	// (with exponential backoff, this will become too long quickly).
	// The context will cancel the loop early.
	err := client.WaitForServer(ctx, 100, 5*time.Millisecond)

	// Verification
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded, "Der Fehler sollte vom Context-Timeout stammen")
}

func TestHubClient_GetExternalIDs(t *testing.T) {
	transport := &MockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "/api/collectors/slides/bluesky", req.URL.Path)
			assert.Equal(t, "0", req.URL.Query().Get("_start"))
			assert.Equal(t, "10", req.URL.Query().Get("_end"))
			assert.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))

			respBody := `["ext-1", "ext-2"]`
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(respBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("X-Total-Count", "42")

			return resp, nil
		},
	}

	client := setupTestClient(transport)
	ctx := context.Background()

	ids, total, err := client.GetExternalIDs(ctx, "bluesky", 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ext-1", "ext-2"}, ids)
	assert.Equal(t, 42, total)
}
