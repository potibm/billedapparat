package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisURL string

func (ru *RedisURL) URLObject() *url.URL {
	if ru == nil {
		return nil
	}

	parsedURL, err := url.Parse(string(*ru))
	if err != nil {
		return nil
	}

	return parsedURL
}

func (ru *RedisURL) IsValid() bool {
	if ru == nil {
		return true
	}

	_, err := url.ParseRequestURI(string(*ru))

	return err == nil
}

func (ru RedisURL) RedisOptions() *redis.Options {
	u := ru.URLObject()

	if u == nil {
		return nil
	}

	path := u.Path
	if path == "" {
		path = "/0"
	}

	db, err := strconv.Atoi(path[1:])
	if err != nil {
		db = 0
	}

	password, _ := u.User.Password()

	return &redis.Options{
		Addr:     u.Host,
		Password: password,
		DB:       db,
	}
}

func (r RedisURL) Redacted() RedisURL {
	return RedisURL(redactURLPassword(string(r)))
}

func (ru *RedisURL) Validate() error {
	rString := string(*ru)

	if !ru.IsValid() {
		return fmt.Errorf("redis_url '%s' is not a valid URL", rString)
	}

	redisURL := ru.URLObject()
	if redisURL.Scheme != "redis" && redisURL.Scheme != "rediss" {
		return fmt.Errorf(
			"redis_url '%s' has invalid scheme '%s' (expected 'redis' or 'rediss')",
			rString,
			redisURL.Scheme,
		)
	}

	host, _, err := net.SplitHostPort(redisURL.Host)
	if err != nil || host == "" {
		return fmt.Errorf("redis_url '%s' has missing host", rString)
	}

	return nil
}
