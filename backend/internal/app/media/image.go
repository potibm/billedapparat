package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/HugoSmits86/nativewebp"
	"github.com/google/uuid"
	"github.com/potibm/billedapparat/internal/app/config"
	"golang.org/x/image/draw"
)

type ImageType string

const (
	slideMaxWidth   = 1920
	slideMaxHeight  = 1080
	avatarMaxWidth  = 256
	avatarMaxHeight = 256

	TypeAvatar ImageType = "avatar"
	TypeSlide  ImageType = "slide"
)

func ResizeAndSave(srcReader io.Reader, destPath string, maxWidth, maxHeight int) error {
	// 1. Decode image
	src, _, err := image.Decode(srcReader)
	if err != nil {
		return err
	}

	// 2. Calculate new dimensions
	bounds := src.Bounds()
	dstWidth := maxWidth

	dstHeight := (dstWidth * bounds.Dy()) / bounds.Dx()
	if dstHeight > maxHeight {
		dstHeight = maxHeight
		dstWidth = (dstHeight * bounds.Dx()) / bounds.Dy()
	}

	// 3. Resize
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	// 4. Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), config.DataDirPerm); err != nil {
		return err
	}

	// 5. Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 6. Save
	return nativewebp.Encode(out, dst, nil)
}

func ProcessAndSaveSlide(file io.Reader) (string, error) {
	filename := fmt.Sprintf("slide_%s.webp", uuid.New().String())
	localFilePath := filepath.Join(config.MediaDirname, filename)

	if err := ResizeAndSave(file, localFilePath, slideMaxWidth, slideMaxHeight); err != nil {
		return "", err
	}

	publicURL := config.MediaURL + filename

	return publicURL, nil
}

func ProcessAndSaveAvatar(file io.Reader) (string, error) {
	filename := fmt.Sprintf("avatar_%s.webp", uuid.New().String())
	localFilePath := filepath.Join(config.MediaDirname, filename)

	if err := ResizeAndSave(file, localFilePath, avatarMaxWidth, avatarMaxHeight); err != nil {
		return "", err
	}

	return config.MediaURL + filename, nil
}

func ProcessAndSave(file io.Reader, imageType ImageType) (string, error) {
	switch imageType {
	case TypeSlide:
		return ProcessAndSaveSlide(file)
	case TypeAvatar:
		return ProcessAndSaveAvatar(file)
	default:
		return "", fmt.Errorf("unsupported image type: %s", imageType)
	}
}
