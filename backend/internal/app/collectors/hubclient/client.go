package hubclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
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

func (c *HubClient) sendPostRequest(endpoint, entityName, externalID string, payload any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	u := fmt.Sprintf("%s%s", strings.TrimRight(c.BaseURL, "/"), endpoint)

	c.Logger.Debug(fmt.Sprintf("Sending %s to hub", entityName), "url", u, "external_id", externalID)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u, bytes.NewBuffer(jsonData))
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
		c.Logger.Info(fmt.Sprintf("Successfully created %s", entityName), "external_id", externalID)

		return nil
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("hub returned status %d", resp.StatusCode)
}

func (c *HubClient) sendDeleteRequest(entityType, source, externalID string) error {
	endpoint := fmt.Sprintf(
		"%s/api/collectors/%s/%s/%s",
		strings.TrimRight(c.BaseURL, "/"),
		entityType,
		url.PathEscape(source),
		url.PathEscape(externalID),
	)

	c.Logger.Debug(
		fmt.Sprintf("Deleting %s from hub", entityType),
		"url",
		endpoint,
		"source",
		source,
		"external_id",
		externalID,
	)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Logger.Info(fmt.Sprintf("Successfully deleted %s", entityType), "source", source, "external_id", externalID)

		return nil
	}

	return fmt.Errorf("error deleting %s from hub: %s", entityType, resp.Status)
}
