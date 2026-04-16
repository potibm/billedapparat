package media

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
)

const (
	slideMaxWidth  = 1920
	slideMaxHeight = 1080
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
	const mode = 0o755
	if err := os.MkdirAll(filepath.Dir(destPath), mode); err != nil {
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

func ResizeAndSaveSlide(srcReader io.Reader, destPath string) error {
	return ResizeAndSave(srcReader, destPath, slideMaxWidth, slideMaxHeight)
}
