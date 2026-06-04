package bluesky

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/potibm/billedapparat/internal/app/collectors/utils"
)

func (c *Collector) getProfile(ctx context.Context, did string) (*ProfileResponse, error) {
	profile, exists := c.profiles.Get(did)
	if exists {
		return profile, nil
	}

	fetchedProfile, err := fetchProfile(ctx, did)
	if err != nil {
		return nil, err
	}

	c.profiles.Set(did, fetchedProfile)

	return fetchedProfile, nil
}

func fetchProfile(ctx context.Context, did string) (*ProfileResponse, error) {
	apiURL := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=%s", url.QueryEscape(did))

	client := &http.Client{Timeout: profileRequestTimeout}

	return utils.FetchJSON[ProfileResponse](ctx, client, apiURL, utils.RequestOptions{})
}
