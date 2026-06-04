package bluesky

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (c *Collector) handleCreate(ctx context.Context, event *JetstreamEvent) {
	if event.Commit.Record == nil {
		return
	}

	if !c.hasRelevantHashtag(event.Commit.Record) {
		return
	}

	if c.upsertSlide(ctx, event, "create") {
		c.knownPosts.Add(event.Commit.Rkey)
	}
}

func (c *Collector) handleUpdate(ctx context.Context, event *JetstreamEvent) {
	if event.Commit.Record == nil {
		return
	}

	hasHashtag := c.hasRelevantHashtag(event.Commit.Record)
	isKnown := c.knownPosts.Contains(event.Commit.Rkey)

	if hasHashtag {
		if c.upsertSlide(ctx, event, "update") {
			c.knownPosts.Add(event.Commit.Rkey)
		}
	} else if isKnown && !hasHashtag {
		c.deleteSlide(ctx, event.Commit.Rkey, "hashtag_removed")
	}
}

func (c *Collector) handleDelete(ctx context.Context, event *JetstreamEvent) {
	if c.knownPosts.Contains(event.Commit.Rkey) {
		c.deleteSlide(ctx, event.Commit.Rkey, "post_deleted")
	}
}

func (c *Collector) deleteSlide(ctx context.Context, rkey, reason string) {
	c.logger.Info("Deleting slide from hub", "rkey", rkey, "reason", reason)

	if err := c.hubClient.DeleteSlide(ctx, collectorName, rkey); err != nil {
		c.logger.Error("Failed to delete slide from hub", "rkey", rkey, "error", err)
	}

	c.knownPosts.Remove(rkey)

	c.metrics.EventsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "delete"),
		attribute.String("reason", reason),
		attribute.String("collector", collectorName),
	))
}

func (c *Collector) upsertSlide(ctx context.Context, event *JetstreamEvent, opLabel string) bool {
	profile, err := c.getProfile(ctx, event.Did)
	if err != nil {
		c.logger.Error("Failed to fetch profile", "did", event.Did, "error", err)

		return false
	}

	c.logger.Info(
		"Post upserted",
		"rkey", event.Commit.Rkey,
		"text", event.Commit.Record.Text,
		"profile_handle", profile.Handle,
		"op", opLabel,
	)

	slideReq := mapEventToIngestSlide(event, event.Did, profile)
	if slideReq == nil {
		c.logger.Info("Unable to map event to ingest slide", "rkey", event.Commit.Rkey)

		return false
	}

	if err := c.hubClient.SendSlide(ctx, *slideReq); err != nil {
		c.logger.Error("Failed to send slide", "rkey", event.Commit.Rkey, "error", err)

		return false
	}

	c.metrics.EventsMatched.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", opLabel),
		attribute.String("collector", collectorName),
	))

	return true
}

func (c *Collector) hasRelevantHashtag(postRecord *PostRecord) bool {
	postTags := postRecord.Hashtags()

	if len(postTags) == 0 {
		return false
	}

	for _, searchTag := range c.searchTags {
		for _, postTag := range postTags {
			if searchTag == postTag {
				return true
			}
		}
	}

	return false
}
