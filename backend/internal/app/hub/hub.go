package hub

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/generator"
	"github.com/potibm/billedapparat/internal/app/repository"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	defaultShutdownTimeout    = 5 * time.Second
	defaultReadHeaderTimeout  = 3 * time.Second
	pathSlides                = "/slides"
	pathSlidesWithID          = "/slides/:id"
	pathFilterRules           = "/filter-rules"
	pathFilterRulesWithID     = "/filter-rules/:id"
	pathNews                  = "/news"
	pathNewsWithID            = "/news/:id"
	timeTablePath             = "/timetable"
	timeTablePathWithID       = "/timetable/:id"
	collectorSlidesPath       = "/slides"
	collectorNewsPath         = "/news"
	collectorTimeTablePath    = "/timetable"
	collectorPathWithIDSuffix = "/:source/:external_id"
)

var (
	meter                      = otel.Meter(config.OtelMeterName)
	receivedCollectorEvents, _ = meter.Int64Counter(
		"billedapparat_hub_collector_events_rcvd",
		metric.WithDescription("Received collector events"),
	)
)

type Config struct {
	Port               int
	StaticFiles        embed.FS
	SlideRepo          repository.SlideRepository
	NewsRepo           repository.NewsRepository
	TimetableEventRepo repository.TimetableEventRepository
	FilterRuleRepo     repository.FilterRuleRepository
	Cfg                config.Config
}

type Server struct {
	port               int
	staticFiles        embed.FS
	slideRepo          repository.SlideRepository
	filterRuleRepo     repository.FilterRuleRepository
	newsRepo           repository.NewsRepository
	timetableEventRepo repository.TimetableEventRepository
	cfg                config.Config
	streamer           *Streamer
	mediaProcessor     MediaProcessor
	mediaDownloader    *MediaDownloader
	logger             *slog.Logger
	sanitizer          *bluemonday.Policy
	generatorEngine    *generator.Engine
}

type MediaProcessor interface {
	ProcessSlideImage(c *gin.Context, formField string) (string, error)
}

func NewServer(cfg Config) (*Server, error) {
	logger := slog.Default()
	streamer := NewStreamer(logger.With("component", "Streamer"))

	engine := generator.NewEngine(
		cfg.SlideRepo,
		streamer,
		logger.With("component", "GeneratorEngine"),
		generator.NewNewsGenerator(cfg.NewsRepo, logger.With("component", "NewsGenerator")),
		generator.NewTimetableGenerator(cfg.TimetableEventRepo, logger.With("component", "TimetableGenerator")),
	)

	mediaDownloader := NewMediaDownloader(cfg.SlideRepo, streamer, logger.With("component", "MediaDownloader"))
	mediaProcessor := &LocalDiskMediaProcessor{}

	sanitizer := bluemonday.StrictPolicy()

	return &Server{
		port:               cfg.Port,
		staticFiles:        cfg.StaticFiles,
		slideRepo:          cfg.SlideRepo,
		filterRuleRepo:     cfg.FilterRuleRepo,
		newsRepo:           cfg.NewsRepo,
		timetableEventRepo: cfg.TimetableEventRepo,
		cfg:                cfg.Cfg,
		streamer:           streamer,
		mediaDownloader:    mediaDownloader,
		mediaProcessor:     mediaProcessor,
		logger:             logger.With("component", "HubServer"),
		sanitizer:          sanitizer,
		generatorEngine:    engine,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	router, err := s.setupRouter()
	if err != nil {
		return fmt.Errorf("setup router: %w", err)
	}

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(s.port),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		Handler:           router,
	}

	s.StartMediaGarbageCollector(ctx)
	s.StartCollectorTextGarbageCollector(ctx)

	serverErr := make(chan error, 1)

	// Start generator engine in Goroutine
	go func() {
		if err := s.generatorEngine.Run(ctx); err != nil {
			s.logger.Error("Generator engine stopped with error", "error", err)
		}
	}()

	// Start server in Goroutine
	go func() {
		s.logger.Info("Starting server...", "port", s.port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server failed to start: %w", err)

	case <-ctx.Done():
		s.logger.Info("Shutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		s.logger.Info("Server stopped cleanly")

		return nil
	}
}

func (s *Server) setupRouter() (*gin.Engine, error) {
	gin.SetMode(s.cfg.App.GinMode)

	r := gin.New()

	r.Use(
		// middleware.ErrorHandlingMiddleware(),
		gin.Recovery(),
		sentrygin.New(sentrygin.Options{Repanic: false}),
		sloggin.New(s.logger),
		otelgin.Middleware(config.OtelBackendServiceName),
	)
	s.registerCorsMiddleware(r)

	r.Static("/media", "./data/media")
	r.Static("/style", "./data/style")

	folder, err := static.EmbedFolder(s.staticFiles, "assets")
	if err != nil {
		return nil, fmt.Errorf("create embedded folder: %w", err)
	}

	r.Use(static.Serve("/", folder))

	api := r.Group("/api")
	api.GET("/config", s.handleGetPublicConfig)
	api.GET("/stream", s.streamSlides)

	admin := r.Group("/api/admin")
	admin.GET(pathSlides, s.adminListSlides)
	admin.POST(pathSlides, s.adminCreateSlide)
	admin.GET(pathSlidesWithID, s.adminGetSlide)
	admin.PUT(pathSlidesWithID, s.adminUpdateSlide)
	admin.DELETE(pathSlidesWithID, s.adminDeleteSlide)

	admin.GET(pathFilterRules, s.adminListFilterRules)
	admin.POST(pathFilterRules, s.adminCreateFilterRule)
	admin.GET(pathFilterRulesWithID, s.adminGetFilterRule)
	admin.PUT(pathFilterRulesWithID, s.adminUpdateFilterRule)
	admin.DELETE(pathFilterRulesWithID, s.adminDeleteFilterRule)

	admin.GET(pathNews, s.adminListNews)
	admin.GET(pathNewsWithID, s.adminGetNews)

	admin.GET(timeTablePath, s.adminListTimetable)
	admin.GET(timeTablePathWithID, s.adminGetTimetable)

	admin.GET("/sources", s.adminListSources)

	internal := r.Group("/api/internal")
	internal.Use(APIKeyAuthMiddleware(s.cfg.API.AdminAPIKey))
	internal.POST("/import", s.internalImportDirectory)

	collectors := r.Group("/api/collectors")
	collectors.Use(CollectorAuthMiddleware(s.cfg.Collectors))
	collectors.POST(collectorSlidesPath, s.collectorIngestSlide)
	collectors.GET(collectorSlidesPath+"/:source", s.collectorListExternalIDs)
	collectors.DELETE(collectorSlidesPath+collectorPathWithIDSuffix, s.collectorDeleteSlide)
	collectors.POST(collectorNewsPath, s.collectorUpsertNews)
	collectors.PUT(collectorNewsPath, s.collectorSyncNews)
	collectors.DELETE(collectorNewsPath+collectorPathWithIDSuffix, s.collectorDeleteNews)
	collectors.POST(collectorTimeTablePath, s.collectorUpsertTimetable)
	collectors.PUT(collectorTimeTablePath, s.collectorSyncTimetable)
	collectors.DELETE(collectorTimeTablePath+collectorPathWithIDSuffix, s.collectorDeleteTimetable)

	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.RequestURI, "/api") && !strings.Contains(c.Request.RequestURI, ".") {
			file, _ := s.staticFiles.ReadFile("assets/index.html")
			c.Data(
				http.StatusOK,
				"text/html; charset=utf-8",
				file,
			)
		}
	})

	return r, nil
}

func (s *Server) registerCorsMiddleware(r *gin.Engine) {
	if len(s.cfg.App.CorsAllowOrigins) > 0 {
		s.logger.Info("CORS middleware enabled", "origins", s.cfg.App.CorsAllowOrigins)
		r.Use(s.createCorsMiddleware())
	} else {
		s.logger.Info("CORS middleware disabled (no origins configured)")
	}
}

func (s *Server) createCorsMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = s.cfg.App.CorsAllowOrigins
	corsConfig.AllowAllOrigins = false
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization", "Credentials", "Content-Type", "X-API-Key", "Accept")
	corsConfig.AddExposeHeaders("X-Total-Count", "Content-Disposition")

	return cors.New(corsConfig)
}
