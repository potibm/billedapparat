package hubclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendNews(payload contracts.IngestNewsRequest) error {
	return c.sendPostRequest("/api/collectors/news", "news", payload.ExternalID, payload)
}

func (c *HubClient) DeleteNews(source, externalID string) error {
	return c.sendDeleteRequest("news", source, externalID)
}

func (c *HubClient) SendNewsSync(payload contracts.IngestNewsSyncRequest) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	u := fmt.Sprintf("%s/api/collectors/news", strings.TrimRight(c.BaseURL, "/"))
	c.Logger.Debug("Syncing news to hub", "url", u)

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
