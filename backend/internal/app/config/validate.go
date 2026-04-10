package config

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	validDbFilename = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	validLocale     = regexp.MustCompile(`^[a-zA-Z]{2}-[A-Z]{2}$`)
	validCurrency   = regexp.MustCompile(`^[A-Z]{3}$`)
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
