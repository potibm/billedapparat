package hub

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
)

type AppConfigPublic struct {
	Version            string                   `json:"version"`
	Environment        string                   `json:"environment"`
	EnvironmentMessage string                   `json:"environment_message"`
	Sentry             config.SentryConfig      `json:"sentry"`
	Format             config.FormatConfig      `json:"format"`
	Playlists          []config.PlaylistConfig  `json:"playlists"`
	AdminURLs          config.ExternalAdminURLs `json:"admin_urls"`
	Auth               *config.AuthConfig       `json:"auth,omitempty"`
}

type SentryPublic struct {
	DSN         string `json:"dsn"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

func (s *Server) handleGetPublicConfig(c *gin.Context) {
	pub := mapToPublicConfig(&s.cfg)

	c.JSON(http.StatusOK, pub)
}

func mapToPublicConfig(cfg *config.Config) AppConfigPublic {
	return AppConfigPublic{
		Version:            cfg.App.Version,
		Environment:        cfg.App.Environment,
		EnvironmentMessage: cfg.App.EnvironmentMessage,
		Format:             cfg.Format,
		Playlists:          cfg.Playlists,
		Sentry:             cfg.Sentry,
		AdminURLs:          cfg.AdminURLs,
		Auth:               cfg.Auth,
	}
}
