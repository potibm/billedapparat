package factory

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/potibm/billedapparat/internal/app/collectors"
	"github.com/potibm/billedapparat/internal/app/collectors/hubclient"
	"github.com/spf13/viper"
)

func buildCollector[T any, C collectors.Collector](
	v *viper.Viper,
	c *hubclient.HubClient,
	validate *validator.Validate,
	name string,
	constructor func(T, *hubclient.HubClient) C,
) (C, error) {
	var (
		cfg  T
		zero C
	)
	if err := v.Unmarshal(&cfg); err != nil {
		return zero, fmt.Errorf("error parsing config for %s Collector: %w", name, err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return zero, fmt.Errorf("error parsing config for %s Collector: %w", name, err)
	}

	if err := validate.Struct(&cfg); err != nil {
		return zero, fmt.Errorf("invalid config for %s Collector: %w", name, err)
	}

	return constructor(cfg, c), nil
}
