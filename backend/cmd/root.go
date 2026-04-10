package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/initializer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "dev"

var Cfg config.Config

const (
	logFormatFlagName    = "log-format"
	logLevelFlagName     = "log-level"
	databaseFileFlagName = "db-file"
)

var rootCmd = &cobra.Command{
	Use:           "billedapparat",
	Short:         "A party information system for the bigscreen at demoparties.",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		if Version != "" {
			viper.Set("app.version", Version)
		}

		err := viper.Unmarshal(&Cfg, viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToSliceHookFunc(","),
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToTimeHookFunc(time.RFC3339),
			),
		))
		if err != nil {
			return fmt.Errorf("error parsing the config: %w", err)
		}

		Cfg.App.CorsAllowOrigins = strings.Split(viper.GetString("app.cors_allow_origins"), ",")

		if err := Cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		if !cmd.Flags().Changed(logFormatFlagName) {
			if cmd.Name() != "serve" {
				Cfg.App.LogFormat = "text"
			}
		}

		setupLogger(Cfg.App.LogFormat, Cfg.App.LogLevel)

		return nil
	},
}

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rootCmd.PersistentFlags().String(logLevelFlagName, "info", "Log Level (debug, info, warn, error)")
	_ = viper.BindPFlag("app.log_level", rootCmd.PersistentFlags().Lookup(logLevelFlagName))

	rootCmd.PersistentFlags().String(logFormatFlagName, "json", "Log Format (json, text)")
	_ = viper.BindPFlag("app.log_format", rootCmd.PersistentFlags().Lookup(logFormatFlagName))

	rootCmd.PersistentFlags().String(databaseFileFlagName, "billedapparat.db", "Dateiname der SQLite Datenbank")
	_ = viper.BindPFlag("app.db_filename", rootCmd.PersistentFlags().Lookup(databaseFileFlagName))

	rootCmd.AddCommand(NewServeCmd())

	rootCmd.AddCommand(NewConfigCmd())

	return rootCmd.ExecuteContext(ctx)
}

func loadConfig() error {
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	config.InitViper()

	return nil
}

func setupLogger(format, level string) {
	initializer.InitLogger(format, level)
}
