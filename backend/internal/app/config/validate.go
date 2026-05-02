package config

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	validDbFilename = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	validLocale     = regexp.MustCompile(`^[a-zA-Z]{2}-[A-Z]{2}$`)
)

func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := c.App.Validate(); err != nil {
		return err
	}

	if err := c.Format.Validate(); err != nil {
		return err
	}

	if err := c.API.Validate(c.App.Environment); err != nil {
		return err
	}

	for i := range c.Playlists {
		if err := c.Playlists[i].Validate(); err != nil {
			return fmt.Errorf("playlist[%d] '%s' is invalid: %w", i, c.Playlists[i].Name, err)
		}
	}

	return nil
}

func (f *APIConfig) Validate(environment string) error {
	if f.AdminAPIKey == DefaultAPIAdminKey && environment == "production" {
		return fmt.Errorf(
			"admin_api_key is set to the default value in production environment, which is not allowed for security reasons",
		)
	}

	return nil
}

func (f *AppConfig) Validate() error {
	if !validDbFilename.MatchString(f.DbFilename) {
		return fmt.Errorf("db_filename '%s' contains invalid characters", f.DbFilename)
	}

	return nil
}

func (f *FormatConfig) Validate() error {
	if err := f.Date.Validate(); err != nil {
		return err
	}

	return nil
}

func (f *DateFormatConfig) Validate() error {
	if !validLocale.MatchString(f.Locale) {
		return fmt.Errorf("date.locale '%s' is not a valid locale", f.Locale)
	}

	return nil
}

func (p *PlaylistConfig) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("playlist name cannot be empty")
	}

	for i := range p.Steps {
		p.Steps[i].SetDefaults()

		if p.Steps[i].Type == "" {
			return fmt.Errorf("step[%d] has no slide type defined", i)
		}
	}

	return nil
}
