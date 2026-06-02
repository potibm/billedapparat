package hubclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/potibm/billedapparat/internal/app/contracts"
)

func (c *HubClient) SendSlide(ctx context.Context, payload contracts.IngestSlideRequest) error {
	return c.sendPostRequest(ctx, "slides", "slide", payload.ExternalID, payload)
}

func (c *HubClient) DeleteSlide(ctx context.Context, source, externalID string) error {
	return c.sendDeleteRequest(ctx, "slides", source, externalID)
}

func (c *HubClient) GetExternalIDs(
	ctx context.Context,
	source string,
	start, end int,
) (externalIDs []string, total int, err error) {
	u := fmt.Sprintf("%s/%s?_start=%d&_end=%d", c.getURL("slides"), url.PathEscape(source), start, end)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, 0, err
	}

	c.setHeaders(req, false)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf(errFmtNetwork, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf(errFmtHubStatus, resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(&externalIDs); err != nil {
		return nil, 0, fmt.Errorf(errFmtMarshal, err)
	}

	total = 0

	if totalStr := resp.Header.Get("X-Total-Count"); totalStr != "" {
		if t, err := strconv.Atoi(totalStr); err == nil {
			total = t
		}
	}

	return externalIDs, total, nil
}
