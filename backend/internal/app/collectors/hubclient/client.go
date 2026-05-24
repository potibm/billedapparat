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

const (
	errFmtNetwork   = "network error: %w"
	errFmtMarshal   = "json marshal error: %w"
	errFmtHubStatus = "hub returned status %d"
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

func (c *HubClient) sendPostRequest(ctx context.Context, entityType, entityName, externalID string, payload any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf(errFmtMarshal, err)
	}

	u := c.getURL(entityType)

	c.Logger.Debug(fmt.Sprintf("Sending %s to hub", entityName), "url", u, "external_id", externalID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf(errFmtNetwork, err)
	}

	c.setHeaders(req, true)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(errFmtNetwork, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		c.Logger.Info(fmt.Sprintf("Successfully created %s", entityName), "external_id", externalID)

		return nil
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf(errFmtHubStatus, resp.StatusCode)
}

func (c *HubClient) sendDeleteRequest(ctx context.Context, entityType, source, externalID string) error {
	endpoint := fmt.Sprintf(
		"%s/%s/%s",
		c.getURL(entityType),
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

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	c.setHeaders(req, false)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(errFmtNetwork, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Logger.Info(fmt.Sprintf("Successfully deleted %s", entityType), "source", source, "external_id", externalID)

		return nil
	}

	return fmt.Errorf(errFmtHubStatus, resp.StatusCode)
}

func (c *HubClient) sendPutRequest(ctx context.Context, entityType, entityName string, payload any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf(errFmtMarshal, err)
	}

	u := c.getURL(entityType)

	c.Logger.Debug(fmt.Sprintf("Syncing %s to hub", entityName), "url", u)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf(errFmtNetwork, err)
	}

	c.setHeaders(req, true)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(errFmtNetwork, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.Logger.Info(fmt.Sprintf("Successfully synced %s", entityName))

		return nil
	}

	return fmt.Errorf(errFmtHubStatus, resp.StatusCode)
}

func (c *HubClient) getURL(entityType string) string {
	return fmt.Sprintf("%s/api/collectors/%s", strings.TrimRight(c.BaseURL, "/"), entityType)
}

func (c *HubClient) setHeaders(req *http.Request, isJSONPayload bool) {
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	if isJSONPayload {
		req.Header.Set("Content-Type", "application/json")
	}
}
