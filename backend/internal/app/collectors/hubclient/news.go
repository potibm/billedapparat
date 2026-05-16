package hubclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendNews(payload contracts.IngestNewsRequest) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	u := fmt.Sprintf("%s/api/collectors/news", strings.TrimRight(c.BaseURL, "/"))
	slog.Debug("Sending news to hub", "url", u, "external_id", payload.ExternalID)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(jsonData))
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
		c.Logger.Info("News successfully created", "external_id", payload.ExternalID)

		return nil
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("hub returned status %d", resp.StatusCode)
}

func (c *HubClient) DeleteNews(source, externalID string) error {
	endpoint := fmt.Sprintf(
		"%s/api/collectors/news/%s/%s",
		strings.TrimRight(c.BaseURL, "/"),
		url.PathEscape(source),
		url.PathEscape(externalID),
	)

	req, err := http.NewRequest(http.MethodDelete, endpoint, http.NoBody)
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
		return fmt.Errorf("error deleting news from hub: %s", resp.Status)
	}

	return nil
}

func (c *HubClient) SendNewsSync(payload contracts.IngestNewsSyncRequest) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	u := fmt.Sprintf("%s/api/collectors/news", strings.TrimRight(c.BaseURL, "/"))
	slog.Debug("Syncing news to hub", "url", u)

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode == http.StatusOK {
		c.Logger.Info("News successfully synced")

		return nil
	}

	return fmt.Errorf("hub returned status %d", resp.StatusCode)
}
