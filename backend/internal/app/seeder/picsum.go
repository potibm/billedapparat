package seeder

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/media"
)

func (s *Seeder) getFakeImageURL(id int64, blackAndWhite bool) (string, error) {
	cachedImagePath, err := s.getCachedSeedImage(id, blackAndWhite)
	if err != nil {
		slog.Error("Failed to get seed image", "error", err, "index", id)

		return "", err
	}

	file, err := os.Open(cachedImagePath)
	if err != nil {
		slog.Error("Failed to open cached image", "error", err)

		return "", err
	}
	defer file.Close()

	publicURL, err := media.ProcessAndSaveSlide(file)
	if err != nil {
		slog.Error("Failed to process seed image", "error", err)

		return "", err
	}

	return publicURL, nil
}

func (s *Seeder) getCachedSeedImage(index int64, blackAndWhite bool) (string, error) {
	bwStr := "color"
	if blackAndWhite {
		bwStr = "bw"
	}

	fileName := fmt.Sprintf("picsum_%d_%s.jpg", index, bwStr)
	filePath := filepath.Join(config.SeedCacheDirname, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Info("Downloading seed image to cache", "index", index)

		if err := s.downloadFromPicsum(index, blackAndWhite, filePath); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

func (s *Seeder) downloadFromPicsum(index int64, blackAndWhite bool, dstFilename string) error {
	picSumURL := fmt.Sprintf("https://picsum.photos/seed/billedapparat%d/1920/1080", index)
	if blackAndWhite {
		picSumURL += "?grayscale"
	}

	//nolint:gosec // G107: The url is not user-controlled, it's generated based on a fixed pattern and a seed value.
	resp, err := http.Get(picSumURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dstFilename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to copy image: %w", err)
	}

	return nil
}
