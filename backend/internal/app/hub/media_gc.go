package hub

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (s *Server) StartMediaGarbageCollector(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runGarbageCollectionCycle()
			}
		}
	}()
}

func (s *Server) runGarbageCollectionCycle() {
	logger := s.logger.With("cycle", "media_gc")
	logger.Info("Starting Media Garbage Collection...")

	activeURLs, err := s.slideRepo.GetAllMediaURLs(context.Background())
	if err != nil {
		logger.Warn("Error fetching URLs from DB", "error", err)

		return
	}

	urlMap := make(map[string]bool)
	for _, url := range activeURLs {
		urlMap[url] = true
	}

	mediaDir := filepath.Join("data", "media")

	files, err := os.ReadDir(mediaDir)
	if err != nil {
		logger.Warn("Error reading media directory", "error", err)

		return
	}

	deletedCount := 0

	for _, file := range files {
		if s.deleteIfOrphaned(mediaDir, file, urlMap, logger) {
			deletedCount++
		}
	}

	logger.Info("Media Garbage Collection finished.", "deleted", deletedCount)
}

func (s *Server) deleteIfOrphaned(mediaDir string, file os.DirEntry, urlMap map[string]bool, logger *slog.Logger) bool {
	info, err := file.Info()
	if err != nil || info.IsDir() {
		return false
	}

	// ignore files modified in the last hour to avoid race conditions
	if time.Since(info.ModTime()) < 1*time.Hour {
		return false
	}

	publicURL := "/media/" + file.Name()
	if urlMap[publicURL] {
		return false
	}

	logger.Info("Deleting orphaned file", "url", publicURL)

	if err := os.Remove(filepath.Join(mediaDir, file.Name())); err != nil {
		logger.Warn("Failed to delete", "file", file.Name(), "error", err)

		return false
	}

	return true
}
