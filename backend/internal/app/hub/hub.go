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

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/repository"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	defaultShutdownTimeout   = 5 * time.Second
	defaultReadHeaderTimeout = 3 * time.Second
	pathSlides               = "/slides"
	pathSlidesWithID         = "/slides/:id"
	pathFilterRules          = "/filter-rules"
	pathFilterRulesWithID    = "/filter-rules/:id"
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
	markdownConverter  *converter.Converter
}

type MediaProcessor interface {
	ProcessSlideImage(c *gin.Context, formField string) (string, error)
}

func NewServer(cfg Config) (*Server, error) {
	logger := slog.Default()
	streamer := NewStreamer(logger.With("component", "Streamer"))

	markdownConverter := converter.NewConverter()

	return &Server{
		port:               cfg.Port,
		staticFiles:        cfg.StaticFiles,
		slideRepo:          cfg.SlideRepo,
		filterRuleRepo:     cfg.FilterRuleRepo,
		newsRepo:           cfg.NewsRepo,
		timetableEventRepo: cfg.TimetableEventRepo,
		cfg:                cfg.Cfg,
		streamer:           streamer,
		mediaDownloader:    NewMediaDownloader(cfg.SlideRepo, streamer, logger.With("component", "MediaDownloader")),
		logger:             logger.With("component", "HubServer"),
		markdownConverter:  markdownConverter,
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

	// Start server in Goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "error", err)
		}
	}()

	slog.Info("Server is up", "port", s.port)

	<-ctx.Done()
	slog.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

func (s *Server) setupRouter() (*gin.Engine, error) {
	gin.SetMode(s.cfg.App.GinMode)

	r := gin.New()

	r.Use(
		// middleware.ErrorHandlingMiddleware(),
		gin.Recovery(),
		sentrygin.New(sentrygin.Options{Repanic: false}),
		sloggin.New(slog.Default()),
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

	admin.GET("/sources", s.adminListSources)

	internal := r.Group("/api/internal")
	internal.Use(APIKeyAuthMiddleware(s.cfg.API.AdminAPIKey))
	internal.POST("/import", s.internalImportDirectory)

	collectors := r.Group("/api/collectors")
	collectors.Use(CollectorAuthMiddleware(s.cfg.Collectors))
	collectors.POST("/slides", s.collectorIngestSlide)
	collectors.DELETE("/slides/:source/:external_id", s.collectorDeleteSlide)
	collectors.POST("/news", s.collectorUpsertNews)
	collectors.PUT("/news", s.collectorSyncNews)
	collectors.DELETE("/news/:source/:external_id", s.collectorDeleteNews)

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
		slog.Info("CORS middleware enabled", "origins", s.cfg.App.CorsAllowOrigins)
		r.Use(s.createCorsMiddleware())
	} else {
		slog.Info("CORS middleware disabled (no origins configured)")
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
