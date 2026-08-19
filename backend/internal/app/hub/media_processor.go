package hub

import (
	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/media"
)

type LocalDiskMediaProcessor struct{}

func (lmp *LocalDiskMediaProcessor) ProcessSlideImage(c *gin.Context, fieldName string) (string, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	return media.ProcessAndSaveSlide(file)
}

func (lmp *LocalDiskMediaProcessor) ProcessSlideVideo(c *gin.Context, fieldName string) (string, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	return media.ProcessAndSaveVideo(file)
}
