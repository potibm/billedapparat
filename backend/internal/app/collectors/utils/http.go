package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestOptions struct {
	Headers map[string]string
}

func DoGet(ctx context.Context, client *http.Client, url string, opts RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error during request: %w", err)
	}

	return resp, nil
}

func FetchJSON[T any](ctx context.Context, client *http.Client, url string, opts RequestOptions) (*T, error) {
	resp, err := DoGet(ctx, client, url, opts)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: unexpected status %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return &result, nil
}
