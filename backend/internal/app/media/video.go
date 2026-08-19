package media

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/potibm/billedapparat/internal/app/config"
)

const maxVideoSize = 100 << 20 // 100 MB

const bytesPerMB = 1 << 20

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

	defer func() {
		out.Close()

		if err != nil {
			os.Remove(localFilePath)
		}
	}()

	limited := io.LimitReader(file, maxVideoSize+1)

	n, err := io.Copy(out, limited)
	if err != nil {
		return "", err
	}

	if n > maxVideoSize {
		err = fmt.Errorf("video file exceeds maximum size of %d MB", maxVideoSize/bytesPerMB)

		return "", err
	}

	publicURL := config.MediaURL + filename

	return publicURL, nil
}
