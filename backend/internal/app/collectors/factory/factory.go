package factory

import (
	"fmt"

	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/bluesky"
	"github.com/potibm/billedapparat/internal/app/collectors/discord"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/potibm/billedapparat/internal/app/collectors/pouet"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_news"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_timetable"
	"github.com/potibm/billedapparat/internal/app/collectors/twitch"
	"github.com/potibm/billedapparat/internal/app/config"
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
	case "twitch":
		return buildTwitch(v, client)
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

func buildTwitch(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	var cfg twitch.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config for Twitch Collector: %w", err)
	}

	return twitch.NewCollector(cfg, c), nil
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
	return buildRedisConsumer(
		v, c,
		func(cfg protokolapparat_news.Config) config.RedisURL { return cfg.RedisURL },
		protokolapparat_news.NewCollector,
	)
}

func buildProtokolapparatTimetable(v *viper.Viper, c *hubclient.HubClient) (collectors.Collector, error) {
	return buildRedisConsumer(
		v, c,
		func(cfg protokolapparat_timetable.Config) config.RedisURL { return cfg.RedisURL },
		protokolapparat_timetable.NewCollector,
	)
}
