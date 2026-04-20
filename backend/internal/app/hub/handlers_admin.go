package hub

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/media"
	"github.com/potibm/billedapparat/internal/app/repository"
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

	s.streamer.Broadcast("CREATE", slide)

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

	s.streamer.Broadcast("UPDATE", slide)

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

	s.streamer.Broadcast("DELETE", id)

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Server) processSlideImage(c *gin.Context, fieldName string) (string, error) {
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

func (s *Server) parseSlidePayload(c *gin.Context) (*domain.Slide, error) {
	var slide domain.Slide

	contentType := c.GetHeader("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		priority, err := strconv.Atoi(c.PostForm("display_options.priority"))
		if err != nil {
			priority = 1
		}

		slide.Status = c.PostForm("status")
		slide.Content.Type = domain.SlideType(c.PostForm("content.type"))
		slide.Content.Text = c.PostForm("content.text")
		slide.Author.DisplayName = c.PostForm("author.displa_nName")
		slide.DisplayOptions.Priority = priority
		slide.DisplayOptions.AllowSocialOverlay = c.PostForm("display_options.allow_social_overlay") == "true"

		newPath, err := s.processSlideImage(c, "image_upload")
		if err == nil && newPath != "" {
			slide.MediaURLOriginal = newPath
		}

		return &slide, nil
	}

	if err := c.ShouldBindJSON(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}
