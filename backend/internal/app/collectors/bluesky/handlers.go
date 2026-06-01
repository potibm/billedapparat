package bluesky

import (
	"context"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (c *Collector) handleCreate(ctx context.Context, event *JetstreamEvent) {
	if event.Commit.Record == nil {
		return
	}

	if !c.hasRelevantHashtag(event.Commit.Record.Text) {
		return
	}

	profile, err := c.getProfile(context.Background(), event.Did)
	if err != nil {
		slog.Error("Failed to fetch profile", "did", event.Did, "error", err)

		return
	}

	// send event to hub
	slog.Info(
		"Post upserted",
		"rkey",
		event.Commit.Rkey,
		"text",
		event.Commit.Record.Text,
		"profile_handle",
		profile.Handle,
	)

	c.postsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "create"),
	))

	c.knownPosts.Add(event.Commit.Rkey)
}

func (c *Collector) handleUpdate(ctx context.Context, event *JetstreamEvent) {
	if event.Commit.Record == nil {
		return
	}

	hasHashtag := c.hasRelevantHashtag(event.Commit.Record.Text)
	isKnown := c.knownPosts.Contains(event.Commit.Rkey)

	if hasHashtag {
		profile, err := c.getProfile(context.Background(), event.Did)
		if err != nil {
			slog.Error("Failed to fetch profile", "did", event.Did, "error", err)

			return
		}

		// send event to hub
		slog.Info(
			"Post upserted",
			"rkey",
			event.Commit.Rkey,
			"text",
			event.Commit.Record.Text,
			"profile_handle",
			profile.Handle,
		)

		c.knownPosts.Add(event.Commit.Rkey)

		c.postsMatched.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "update"),
		))
	} else if isKnown && !hasHashtag {
		// send event to hub
		slog.Info("Post deleted (hashtag removed)", "rkey", event.Commit.Rkey)

		c.knownPosts.Remove(event.Commit.Rkey)

		c.postsMatched.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "delete"),
		))
	}
}

func (c *Collector) handleDelete(ctx context.Context, event *JetstreamEvent) {
	if c.knownPosts.Contains(event.Commit.Rkey) {
		// send event to hub
		slog.Info("Post deleted", "rkey", event.Commit.Rkey)

		c.postsMatched.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "delete"),
		))

		c.knownPosts.Remove(event.Commit.Rkey)
	}
}

func (c *Collector) hasRelevantHashtag(text string) bool {
	textLower := strings.ToLower(text)
	for _, hashtag := range c.cfg.Hashtags.Lower() {
		if strings.Contains(textLower, hashtag) {
			return true
		}
	}

	return false
}
