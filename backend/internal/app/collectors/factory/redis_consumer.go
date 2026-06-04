package factory

import (
	"context"
	"fmt"

	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func buildRedisConsumer[T any, C collectors.Collector](
	v *viper.Viper,
	c *hubclient.HubClient,
	getRedisURL func(T) config.RedisURL,
	newCollector func(T, *hubclient.HubClient, *redis.Client) C,
) (collectors.Collector, error) {
	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	rdb, err := initializeRedisClient(getRedisURL(cfg))
	if err != nil {
		return nil, fmt.Errorf("error initializing Redis client: %w", err)
	}

	return newCollector(cfg, c, rdb), nil
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
