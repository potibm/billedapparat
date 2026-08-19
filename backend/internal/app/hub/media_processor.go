package hub

import (
	"mime/multipart"

	"github.com/potibm/billedapparat/internal/app/media"
)

type LocalDiskMediaProcessor struct{}

func (lmp *LocalDiskMediaProcessor) ProcessSlideImage(file multipart.File) (string, error) {
	defer file.Close()

	return media.ProcessAndSaveSlide(file)
}

func (lmp *LocalDiskMediaProcessor) ProcessSlideVideo(file multipart.File) (string, error) {
	defer file.Close()

	return media.ProcessAndSaveVideo(file)
}
