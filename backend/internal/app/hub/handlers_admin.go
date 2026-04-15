package hub

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
	"golang.org/x/image/draw"
)

func (s *Server) adminListSlides(c *gin.Context) {
	start, _ := strconv.Atoi(c.DefaultQuery("_start", "0"))
	end, _ := strconv.Atoi(c.DefaultQuery("_end", "20"))
	sort := c.DefaultQuery("_sort", "id")
	order := c.DefaultQuery("_order", "DESC")
	slideType := domain.SlideType(c.Query("type"))

	params := repository.AdminListParams{
		Offset: start,
		Limit:  end - start,
		Sort:   sort,
		Order:  order,
		Type:   slideType,
	}

	slides, total, err := s.slideRepo.AdminList(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	//	c.Header("Access-Control-Expose-Headers", "X-Total-Count")

	c.JSON(http.StatusOK, slides)
}

func (s *Server) adminGetSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})

		return
	}

	slide, err := s.slideRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "slide not found"})

		return
	}

	c.JSON(http.StatusOK, slide)
}

func (s *Server) adminCreateSlide(c *gin.Context) {
	slide, err := s.parseSlidePayload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})

		return
	}

	if err := s.slideRepo.Save(c.Request.Context(), slide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create slide"})

		return
	}

	c.JSON(http.StatusCreated, slide)
}

func (s *Server) adminUpdateSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide ID format"})

		return
	}

	slide, err := s.parseSlidePayload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})

		return
	}

	slide.ID = id

	if err := s.slideRepo.Save(c.Request.Context(), slide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update slide"})

		return
	}

	c.JSON(http.StatusOK, slide)
}

func (s *Server) adminDeleteSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})

		return
	}

	if err := s.slideRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete slide"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Hilfsfunktion 1: Kümmert sich NUR um das Bild-Resizing und Speichern.
func (s *Server) processSlideImage(c *gin.Context, fieldName string) (string, error) {
	fileHeader, err := c.FormFile(fieldName)
	if err != nil {
		return "", err // Keine Datei hochgeladen (ist okay!)
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	src, format, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Resize auf max 800px Breite
	bounds := src.Bounds()
	dstWidth := 800
	dstHeight := (dstWidth * bounds.Dy()) / bounds.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
	localFilePath := filepath.Join("data", "media", filename)

	os.MkdirAll(filepath.Join("data", "media"), 0o755)

	out, _ := os.Create(localFilePath)
	defer out.Close()

	if format == "png" {
		png.Encode(out, dst)
	} else {
		jpeg.Encode(out, dst, &jpeg.Options{Quality: 90})
	}

	publicURL := "/media/" + filename

	return publicURL, nil
}

// Hilfsfunktion 2: Findet heraus, ob JSON oder FormData kam, und füllt das Struct.
func (s *Server) parseSlidePayload(c *gin.Context) (*domain.Slide, error) {
	var slide domain.Slide

	contentType := c.GetHeader("Content-Type")

	// Fall A: Jemand lädt ein Bild hoch (React-Admin schickt FormData)
	if strings.Contains(contentType, "multipart/form-data") {
		slide.Status = c.PostForm("status")
		slide.Content.Type = domain.SlideType(c.PostForm("content.type"))
		slide.Content.Text = c.PostForm("content.text")
		slide.Author.DisplayName = c.PostForm("author.displayName")

		// Versuchen, das Bild zu verarbeiten.
		// "image_upload" muss der Name des ImageInputs in React-Admin sein!
		newPath, err := s.processSlideImage(c, "image_upload")
		if err == nil && newPath != "" {
			slide.MediaURLOriginal = newPath
		}

		return &slide, nil
	}

	// Fall B: Ein ganz normales Formular ohne Datei (React-Admin schickt JSON)
	if err := c.ShouldBindJSON(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}
