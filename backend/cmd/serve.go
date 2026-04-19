package cmd

import (
	"embed"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/potibm/billedapparat/internal/app/hub"
	"github.com/potibm/billedapparat/internal/app/initializer"
	store "github.com/potibm/billedapparat/internal/app/store/gorm"
)

//go:embed assets
var staticFiles embed.FS

const (
	defaultPort = 3100
)

var (
	port         int
	otelEndpoint string
)

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Runs the HTTP server for the Billedapparat application",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Context
			ctx := cmd.Context()

			// 2. Initialize Telemetry
			shutdownFn, err := initializer.InitTelemetry(ctx, otelEndpoint, Cfg.App.Version)
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

			/*

				// 5. Dependency Injection (Repositories & Middleware)
				sqliteRepository := sqliteRepo.NewRepository(db, Cfg.Format.Currency.FractionDigitsMax)
				sumupRepository := sumupRepo.NewRepository(initializer.GetSumupService())
				mailer := initializer.InitializeMailer(Cfg.Mailer)
				jwtMiddleware := initializer.InitializeJwtMiddleware(sqliteRepository, Cfg.Jwt, &Cfg.App.RedisURL)

				// 6. Services & Handler
				purchaseSvc := purchaseService.NewPurchaseService(
					sqliteRepository,
					sumupRepository,
					&mailer,
					Cfg.Format.Currency.FractionDigitsMax,
					Cfg.Format.Currency.Code,
				)

				websocketHandler := websocket.NewHandler(
					sqliteRepository,
					sumupRepository,
					purchaseSvc,
					jwtMiddleware,
					&Cfg.App.CorsAllowOrigins,
				)
				publisher := &websocket.WebsocketPublisher{}
				poller := monitor.NewPoller(sumupRepository, sqliteRepository, purchaseSvc, publisher)

				httpHandlerConfig := handlerHttp.HandlerConfig{
					Repo:            sqliteRepository,
					SumupRepository: sumupRepository,
					PurchaseService: purchaseSvc,
					Monitor:         poller,
					StatusPublisher: publisher,
					Mailer:          mailer,
					AppConfig:       Cfg,
				}
				httpHandler := handlerHttp.NewHandler(httpHandlerConfig)

				// 7. Initialize HTTP Server
				router, err := initializer.InitializeHTTPServer(
					*httpHandler,
					websocketHandler,
					*sqliteRepository,
					staticFiles,
					jwtMiddleware,
					Cfg,
					slog.Default(),
				)
				if err != nil {
					return fmt.Errorf("failed to initialize HTTP server: %w", err)
				}

				// 8. Start background tasks
				startPollerForPendingPurchases(poller, sqliteRepository)
				startCleanupForWebsocketConnections()

				// 9. Server hochfahren
				portStr := ":" + strconv.Itoa(port)
				slog.Info("HTTP server listening", slog.Int("port", port))

				if err := router.Run(portStr); err != nil {
					return fmt.Errorf("failed to start server: %w", err)
				}
			*/
			server, err := hub.NewServer(hub.Config{
				Port:        port,
				StaticFiles: staticFiles,
				SlideRepo:   dbStore.NewSlideRepository(),
				Cfg:         Cfg,
			})
			if err != nil {
				return fmt.Errorf("failed to initialize server: %w", err)
			}

			return server.Run(ctx)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Set the port number for the server to listen on")
	cmd.Flags().
		StringVar(&otelEndpoint, "otel-endpoint", "", "Set the OpenTelemetry endpoint (e.g., localhost:4317)")

	return cmd
}
