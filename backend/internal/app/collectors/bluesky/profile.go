package bluesky

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/potibm/billedapparat/internal/app/collectors/utils"
)

type ProfileList struct {
	mu sync.RWMutex
	m  map[string]*ProfileResponse
}

func NewProfileList() *ProfileList {
	return &ProfileList{
		m: make(map[string]*ProfileResponse),
	}
}

func (p *ProfileList) Add(did string, profile *ProfileResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.m[did] = profile
}

func (p *ProfileList) Get(did string) (*ProfileResponse, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	profile, exists := p.m[did]

	return profile, exists
}

func (p *ProfileList) Remove(did string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.m, did)
}

func (p *ProfileList) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.m)
}

func (c *Collector) getProfile(ctx context.Context, did string) (*ProfileResponse, error) {
	profile, exists := c.profiles.Get(did)
	if exists {
		return profile, nil
	}

	profile, err := fetchProfile(ctx, did)
	if err != nil {
		return nil, err
	}

	c.profiles.Add(did, profile)

	return profile, nil
}

func fetchProfile(ctx context.Context, did string) (*ProfileResponse, error) {
	apiURL := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=%s", url.QueryEscape(did))

	client := &http.Client{Timeout: profileRequestTimeout}

	// Ein Einzeiler!
	return utils.FetchJSON[ProfileResponse](ctx, client, apiURL, utils.RequestOptions{})
}
