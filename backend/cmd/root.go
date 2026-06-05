package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
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
	logFormatFlagName              = "log-format"
	logLevelFlagName               = "log-level"
	databaseFileFlagName           = "db-file"
	skipConfigValidationAnnotation = "skip-config-validation"
)

var rootCmd = &cobra.Command{
	Use:           "billedapparat",
	Short:         "A party information system for the bigscreen at demoparties.",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Annotations: map[string]string{
		skipConfigValidationAnnotation: "true",
	},
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

		if len(Cfg.App.CorsAllowOrigins) == 1 && strings.Contains(Cfg.App.CorsAllowOrigins[0], ",") {
			rawOrigins := strings.Split(Cfg.App.CorsAllowOrigins[0], ",")

			var cleanOrigins []string

			for _, o := range rawOrigins {
				cleanOrigins = append(cleanOrigins, strings.TrimSpace(o))
			}

			Cfg.App.CorsAllowOrigins = cleanOrigins
		}

		if err := Cfg.Validate(); err != nil {
			if !skipConfigValidation(cmd) {
				return fmt.Errorf("invalid configuration: %w", err)
			}
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

	rootCmd.PersistentFlags().String(databaseFileFlagName, "billedapparat.db", "Filename of the SQLite database")
	_ = viper.BindPFlag("app.db_filename", rootCmd.PersistentFlags().Lookup(databaseFileFlagName))

	rootCmd.AddCommand(NewServeCmd())

	configCmd := NewConfigCmd()
	configCmd.AddCommand(
		NewConfigExportCmd(),
		NewConfigCreateCmd(),
	)
	rootCmd.AddCommand(configCmd)

	importCmd := NewImportCmd()
	importCmd.AddCommand(
		NewImportSlidesCmd(),
	)
	rootCmd.AddCommand(importCmd)

	databaseCmd := NewDatabaseCmd()
	databaseCmd.AddCommand(
		NewDatabaseResetCmd(),
	)
	rootCmd.AddCommand(databaseCmd)

	collectorCmd := NewCollectorCmd()
	rootCmd.AddCommand(collectorCmd)

	return rootCmd.ExecuteContext(ctx)
}

func loadConfig() error {
	_ = godotenv.Load()

	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	viper.SetConfigName("config.local")

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error merging local config: %w", err)
		}
	}

	config.InitViper()

	return nil
}

func setupLogger(format, level string) {
	initializer.InitLogger(format, level)
}

func confirm(question string) bool {
	fmt.Printf("WARNING: %s\n", question)
	fmt.Print("Are you sure? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func skipConfigValidation(cmd *cobra.Command) bool {
	cmdPath := cmd.CommandPath()

	if cmd.Annotations[skipConfigValidationAnnotation] == "true" {
		return true
	}

	if strings.HasPrefix(cmdPath, "billedapparat completion") {
		return true
	}

	if strings.HasPrefix(cmdPath, "billedapparat help") {
		return true
	}

	return false
}

func ensureAppInfrastructure() error {
	subDirs := []string{
		config.DatabaseDirname,
		config.MediaDirname,
		config.StylesDirname,
		config.ImportDirname,
		config.SeedCacheDirname,
	}

	for _, dir := range subDirs {
		if err := os.MkdirAll(dir, config.DataDirPerm); err != nil {
			return fmt.Errorf("error creating the directory %s: %w", dir, err)
		}
	}

	const fileMode = 0o600

	cssPath := filepath.Join(config.StylesDirname, "custom.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		initialCSS := []byte(
			"/* Custom style for the frontend */\n/* Override Tailwind classes or CSS variables here  */\n",
		)
		if err := os.WriteFile(cssPath, initialCSS, fileMode); err != nil {
			log.Printf("Warning: Could not create custom.css: %v", err)
		} else {
			slog.Debug("custom.css successfully created")
		}
	}

	return nil
}

func setTerminalTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J")
}
