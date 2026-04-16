package hub

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func (s *Server) StartMediaGarbageCollector() {
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for range ticker.C {
			slog.Info("[GC] Starting Media Garbage Collection...")

			activeURLs, err := s.slideRepo.GetAllMediaURLs(context.Background())
			if err != nil {
				slog.Warn("[GC] Error fetching URLs from DB", "error", err)

				continue
			}

			urlMap := make(map[string]bool)
			for _, url := range activeURLs {
				urlMap[url] = true
			}

			mediaDir := filepath.Join("data", "media")

			files, err := os.ReadDir(mediaDir)
			if err != nil {
				slog.Warn("[GC] Error reading media directory", "error", err)

				continue
			}

			deletedCount := 0

			for _, file := range files {
				info, err := file.Info()
				if err != nil || info.IsDir() {
					continue
				}

				// ignore files modified in the last hour to avoid race conditions
				if time.Since(info.ModTime()) < 1*time.Hour {
					continue
				}

				publicURL := "/media/" + file.Name()
				if !urlMap[publicURL] {
					slog.Info("[GC] Deleting orphaned file", "url", publicURL)

					err := os.Remove(filepath.Join(mediaDir, file.Name()))
					if err == nil {
						deletedCount++
					} else {
						slog.Warn("[GC] Failed to delete", "file", file.Name(), "error", err)
					}
				}
			}

			slog.Info("[GC] Media Garbage Collection finished.", "deleted", deletedCount)
		}
	}()
}
