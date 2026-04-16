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
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/repository"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	defaultShutdownTimeout   = 5 * time.Second
	defaultReadHeaderTimeout = 3 * time.Second
)

type Config struct {
	Port        int
	StaticFiles embed.FS
	SlideRepo   repository.SlideRepository
	Cfg         config.AppConfig
}

type Server struct {
	port        int
	staticFiles embed.FS
	slideRepo   repository.SlideRepository
	cfg         config.AppConfig
}

func NewServer(cfg Config) (*Server, error) {
	return &Server{
		port:        cfg.Port,
		staticFiles: cfg.StaticFiles,
		slideRepo:   cfg.SlideRepo,
		cfg:         cfg.Cfg,
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

	s.StartMediaGarbageCollector()

	// Start server in Goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", "error", err)
		}
	}()

	slog.Info("Hub server is up", "port", s.port)

	<-ctx.Done()
	slog.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

func (s *Server) setupRouter() (*gin.Engine, error) {
	gin.SetMode(s.cfg.GinMode)

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

	admin := r.Group("/api/admin")
	{
		admin.GET("/slides", s.adminListSlides)
		admin.POST("/slides", s.adminCreateSlide)
		admin.GET("/slides/:id", s.adminGetSlide)
		admin.PUT("/slides/:id", s.adminUpdateSlide)
		admin.DELETE("/slides/:id", s.adminDeleteSlide)
	}

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
	if len(s.cfg.CorsAllowOrigins) > 0 {
		slog.Info("CORS middleware enabled", "origins", s.cfg.CorsAllowOrigins)
		r.Use(s.createCorsMiddleware())
	} else {
		slog.Info("CORS middleware disabled (no origins configured)")
	}
}

func (s *Server) createCorsMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = s.cfg.CorsAllowOrigins
	corsConfig.AllowAllOrigins = false
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization", "Credentials")
	corsConfig.AddExposeHeaders("X-Total-Count", "Content-Disposition")

	return cors.New(corsConfig)
}
