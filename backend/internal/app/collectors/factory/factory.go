package factory

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/bluesky"
	"github.com/potibm/billedapparat/internal/app/collectors/discord"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/collectors/lorem"
	"github.com/potibm/billedapparat/internal/app/collectors/mastodon"
	"github.com/potibm/billedapparat/internal/app/collectors/pouet"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_news"
	"github.com/potibm/billedapparat/internal/app/collectors/protokolapparat_timetable"
	"github.com/potibm/billedapparat/internal/app/collectors/twitch"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/spf13/viper"
)

func Build(source string, v *viper.Viper, client *hubclient.HubClient) (collectors.Collector, error) {
	validate := validator.New()

	switch source {
	case "mastodon":
		return buildMastodon(v, client, validate)
	case "bluesky":
		return buildBluesky(v, client, validate)
	case "pouet":
		return buildPouet(v, client, validate)
	case "discord":
		return buildDiscord(v, client, validate)
	case "twitch":
		return buildTwitch(v, client, validate)
	case "lorem":
		return buildLorem(v, client, validate)
	case "protokolapparat-news":
		return buildProtokolapparatNews(v, client, validate)
	case "protokolapparat-timetable":
		return buildProtokolapparatTimetable(v, client, validate)
	default:
		return nil, fmt.Errorf("unknown collector source: %s", source)
	}
}

func buildDiscord(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Discord", discord.NewCollector)
}

func buildTwitch(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Twitch", twitch.NewCollector)
}

func buildBluesky(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Bluesky", bluesky.NewCollector)
}

func buildMastodon(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Mastodon", mastodon.NewCollector)
}

func buildPouet(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Pouet", pouet.NewCollector)
}

func buildLorem(v *viper.Viper, c *hubclient.HubClient, validate *validator.Validate) (collectors.Collector, error) {
	return buildCollector(v, c, validate, "Lorem", lorem.NewCollector)
}

func buildProtokolapparatNews(
	v *viper.Viper,
	c *hubclient.HubClient,
	validate *validator.Validate,
) (collectors.Collector, error) {
	return buildRedisConsumer(
		v, c, validate,
		func(cfg protokolapparat_news.Config) config.RedisURL { return cfg.RedisURL },
		protokolapparat_news.NewCollector,
	)
}

func buildProtokolapparatTimetable(
	v *viper.Viper,
	c *hubclient.HubClient,
	validate *validator.Validate,
) (collectors.Collector, error) {
	return buildRedisConsumer(
		v, c, validate,
		func(cfg protokolapparat_timetable.Config) config.RedisURL { return cfg.RedisURL },
		protokolapparat_timetable.NewCollector,
	)
}
