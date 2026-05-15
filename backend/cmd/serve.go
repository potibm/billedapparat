package cmd

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/hub"
	"github.com/potibm/billedapparat/internal/app/initializer"
	store "github.com/potibm/billedapparat/internal/app/store/gorm"
)

//go:embed assets
var staticFiles embed.FS

const (
	otelEndpointFlagName = "otel-endpoint"
	portFlagName         = "port"
)

var (
	port         int
	otelEndpoint string
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Runs the HTTP server for the Billedapparat application",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ensureAppInfrastructure()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Context
			ctx := cmd.Context()

			// 2. Initialize Telemetry
			shutdownFn, err := initializer.InitTelemetry(ctx, Cfg.App.OtelEndpoint, Cfg.App.Version)
			if err != nil {
				return fmt.Errorf("failed to initialize telemetry: %w", err)
			}

			if shutdownFn != nil {
				defer shutdownFn()
			}

			dbStore, err := store.NewSqliteStore(Cfg.App.DbFilename)
			if err != nil {
				return fmt.Errorf("database error: %w", err)
			}

			defer func() {
				if err := dbStore.Close(); err != nil {
					slog.Error("failed to close database", "error", err)
				}
			}()

			// 4. Initialize external services (Sentry)
			initializer.InitializeSentry(Cfg.Sentry)

			// 5. Start the server
			server, err := hub.NewServer(hub.Config{
				Port:               Cfg.App.Port,
				StaticFiles:        staticFiles,
				SlideRepo:          dbStore.NewSlideRepository(),
				FilterRuleRepo:     dbStore.NewFilterRuleRepository(),
				NewsRepo:           dbStore.NewNewsRepository(),
				TimetableEventRepo: dbStore.NewTimetableEventRepository(),
				Cfg:                Cfg,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize server: %w", err)
			}

			return server.Run(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, portFlagName, "p", config.DefaultPort, "Set the port number for the server to listen on")
	_ = viper.BindPFlag("app.port", cmd.Flags().Lookup(portFlagName))

	cmd.Flags().
		StringVar(&otelEndpoint, otelEndpointFlagName, "", "Set the OpenTelemetry endpoint (e.g., localhost:4317)")
	_ = viper.BindPFlag("app.otel_endpoint", cmd.Flags().Lookup(otelEndpointFlagName))

	return cmd
}
