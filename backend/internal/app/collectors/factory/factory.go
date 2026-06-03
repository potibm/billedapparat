package factory

import (
	"context"
	"fmt"

	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/bluesky"
	"github.com/potibm/billedapparat/internal/app/collectors/discord"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/potibm/billedapparat/internal/app/collectors/pouet"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_news"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_timetable"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func Build(source string, v *viper.Viper, client *hubclient.HubClient) (collectors.Collector, error) {
	switch source {
	case "mastodon":
		return buildMastodon(v, client)
	case "bluesky":
		return buildBluesky(v, client)
	case "pouet":
		return buildPouet(v, client)
	case "discord":
		return buildDiscord(v, client)
	case "protokolapparat-news":
		return buildProtokolapparatNews(v, client)
	case "protokolapparat-timetable":
		return buildProtokolapparatTimetable(v, client)
	default:
		return nil, fmt.Errorf("unknown collector source: %s", source)
	}
}

func buildDiscord(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg discord.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Discord Collector: %w", err)
	}

	return discord.NewCollector(cfg, c), nil
}

func buildBluesky(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg bluesky.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Bluesky Collector: %w", err)
	}

	return bluesky.NewCollector(cfg, c), nil
}

func buildMastodon(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg mastodon.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Mastodon Collector: %w", err)
	}

	return mastodon.NewCollector(cfg, c), nil
}

func buildPouet(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg pouet.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Pouet Collector: %w", err)
	}

	return pouet.NewCollector(cfg, c), nil
}

func buildProtokolapparatNews(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg protokolapparat_news.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Protokolapparat News Collector: %w", err)
	}

	rdb, err := initializeRedisClient(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("error initializing Redis client: %w", err)
	}

	return protokolapparat_news.NewCollector(cfg, c, rdb), nil
}

func buildProtokolapparatTimetable(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg protokolapparat_timetable.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Protokolapparat Timetable Collector: %w", err)
	}

	rdb, err := initializeRedisClient(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("error initializing Redis client: %w", err)
	}

	return protokolapparat_timetable.NewCollector(cfg, c, rdb), nil
}

func initializeRedisClient(redisURL config.RedisURL) (*redis.Client, error) {
	options, err := redis.ParseURL(string(redisURL))
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	rdb := redis.NewClient(options)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to Redis: %w", err)
	}

	return rdb, nil
}
