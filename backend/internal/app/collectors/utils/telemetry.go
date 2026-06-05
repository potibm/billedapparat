package utils

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const Namespace = "billedapparat_collector_"

type CollectorCounters struct {
	EventsReceived metric.Int64Counter
	EventsMatched  metric.Int64Counter
	EventsDropped  metric.Int64Counter
	Reconnects     metric.Int64Counter
}

func NewCollectorCounters(meter metric.Meter) CollectorCounters {
	received, _ := meter.Int64Counter(Namespace+"events_received_total",
		metric.WithDescription("Total number of events received"))
	matched, _ := meter.Int64Counter(Namespace+"events_matched_total",
		metric.WithDescription("Total number of events matched/retained"))
	dropped, _ := meter.Int64Counter(Namespace+"events_dropped_total",
		metric.WithDescription("Total number of events dropped (e.g. buffer full)"))
	reconnects, _ := meter.Int64Counter(Namespace+"reconnects_total",
		metric.WithDescription("Total number of reconnects"))

	return CollectorCounters{
		EventsReceived: received,
		EventsMatched:  matched,
		EventsDropped:  dropped,
		Reconnects:     reconnects,
	}
}

func RegisterQueueDepthGauge(meter metric.Meter, collectorName string, getLen func() int) {
	_, _ = meter.Int64ObservableGauge(
		Namespace+"worker_queue_depth",
		metric.WithDescription("Number of unprocessed events in the channel"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(getLen()), metric.WithAttributes(attribute.String("collector", collectorName)))

			return nil
		}),
	)
}

func RegisterCacheSizeGauge(meter metric.Meter, collectorName, cacheName string, getLen func() int) {
	_, _ = meter.Int64ObservableGauge(
		Namespace+"cache_size",
		metric.WithDescription("Number of items in cache"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(int64(getLen()), metric.WithAttributes(
				attribute.String("collector", collectorName),
				attribute.String("cache_type", cacheName),
			))

			return nil
		}),
	)
}

type CacheStatsProvider interface {
	Stats() CacheStats
}

func RegisterCacheMetrics(meter metric.Meter, collectorName, cacheName string, cache CacheStatsProvider) {
	attrs := metric.WithAttributes(
		attribute.String("collector", collectorName),
		attribute.String("cache_type", cacheName),
	)

	_, _ = meter.Int64ObservableGauge(Namespace+"cache_size",
		metric.WithDescription("Number of items in cache"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(cache.Stats().Size), attrs)

			return nil
		}))

	_, _ = meter.Int64ObservableGauge(Namespace+"cache_evictions",
		metric.WithDescription("Number of cache evictions"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(cache.Stats().Evictions), attrs)

			return nil
		}))

	_, _ = meter.Int64ObservableGauge(Namespace+"cache_hits",
		metric.WithDescription("Number of cache hits"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(cache.Stats().Hits), attrs)

			return nil
		}))

	_, _ = meter.Int64ObservableGauge(Namespace+"cache_misses",
		metric.WithDescription("Number of cache misses"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(cache.Stats().Misses), attrs)

			return nil
		}))
}
