package config

import "net/url"

const redacted = "***REDACTED***"

func (c Config) RedactConfigForDisplay() Config {
	result := c

	result.Sentry.DSN = redacted

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
