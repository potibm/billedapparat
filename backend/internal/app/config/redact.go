package config

import "net/url"

const redacted = "***REDACTED***"

func (c Config) RedactConfigForDisplay() Config {
	result := c

	result.Sentry.DSN = redacted
	result.API.AdminAPIKey = redacted

	for key, collector := range result.Collectors {
		collector.APIKey = redacted
		result.Collectors[key] = collector
	}

	return result
}

func redactURLPassword(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsedURL.User != nil {
		parsedURL.User = url.UserPassword(parsedURL.User.Username(), redacted)

		return parsedURL.String()
	}

	return rawURL
}
