package media

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/potibm/billedapparat/internal/app/config"
)

func ProcessAndSaveVideo(file io.Reader) (string, error) {
	filename := fmt.Sprintf("slide_%s.mp4", uuid.New().String())
	localFilePath := filepath.Join(config.MediaDirname, filename)

	if err := os.MkdirAll(filepath.Dir(localFilePath), config.DataDirPerm); err != nil {
		return "", err
	}

	out, err := os.Create(localFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	publicURL := config.MediaURL + filename

	return publicURL, nil
}
