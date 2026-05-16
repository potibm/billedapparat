package hubclient

import (
	"log/slog"
	"net/http"
	"time"
)

type HubClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Logger     *slog.Logger
}

func New(baseURL, apiKey string, logger *slog.Logger) *HubClient {
	const defaultTimeout = 10 * time.Second

	return &HubClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Logger:  logger.With("component", "hubclient"),
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}
