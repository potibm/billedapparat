package hub

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/media"
)

type ImportRequest struct {
	Directory string `json:"directory"`
}

func (s *Server) internalImportDirectory(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})

		return
	}

	baseDropzone, err := filepath.Abs(config.ImportDirname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server config error"})

		return
	}

	requestedPath := filepath.Join(baseDropzone, req.Directory)

	if !strings.HasPrefix(requestedPath, baseDropzone) {
		slog.Warn("Path traversal attempt detected!", "path", req.Directory, "ip", c.ClientIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})

		return
	}

	files, err := os.ReadDir(requestedPath)
	if err != nil {
		slog.Warn("Directory read failed", "path", requestedPath, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid directory path",
			"details": err.Error(),
		})

		return
	}

	var importedSlides []domain.Slide

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			continue
		}

		fullPath := filepath.Join(requestedPath, file.Name())

		slide, err := s.processImportedFile(c.Request.Context(), fullPath)
		if err != nil {
			slog.Error("Failed to process import file", "file", file.Name(), "error", err)

			continue
		}

		s.streamer.Broadcast("CREATE", slide)

		importedSlides = append(importedSlides, *slide)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Import finished",
		"count":   len(importedSlides),
	})
}

func (s *Server) processImportedFile(ctx context.Context, filePath string) (*domain.Slide, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	publicURL, err := media.ProcessAndSaveSlide(file)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	newSlide := &domain.Slide{
		Status: "active",
		Content: domain.Content{
			Type: "sponsor",
			Title: baseName,
		},
		MediaURLOriginal: publicURL,
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           false,
		},
	}

	if err := s.slideRepo.Save(ctx, newSlide); err != nil {
		return nil, err
	}

	return newSlide, nil
}
