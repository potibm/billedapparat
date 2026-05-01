package hubclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/potibm/billedapparat/internal/app/contracts"
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

func (c *HubClient) SendSlide(payload contracts.IngestRequest) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/api/collectors/ingest", strings.TrimRight(c.BaseURL, "/"))
	slog.Debug("Sending slide to hub", "url", url, "external_id", payload.ExternalID)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		c.Logger.Info("Slide successfully created", "external_id", payload.ExternalID)

		return nil
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("hub returned status %d", resp.StatusCode)
}

func (c *HubClient) DeleteSlide(source, externalID string) error {
	url := fmt.Sprintf("%s/api/collectors/ingest/%s/%s", strings.TrimRight(c.BaseURL, "/"), source, externalID)

	req, err := http.NewRequest(http.MethodDelete, url, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error deleting slide from hub: %s", resp.Status)
	}

	return nil
}
