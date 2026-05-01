package hub

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
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

	filters := repository.AdminListFilters{
		Query:    nil,
		Status:   nil,
		Priority: nil,
		Source:   nil,
		ID:       nil,
	}
	if c.Query("q") != "" {
		query := c.Query("q")
		filters.Query = &query
	}

	switch {
	case c.Query("status_active") == "true":
		status := "active"
		filters.Status = &status
	case c.Query("status_inactive") == "true":
		status := "inactive"
		filters.Status = &status
	case c.Query("status_pending") == "true":
		status := "pending"
		filters.Status = &status
	case c.Query("status_deleted") == "true":
		status := "deleted"
		filters.Status = &status
	}

	if priorityStr := c.Query("display_options.priority"); priorityStr != "" {
		priority64, err := strconv.ParseInt(priorityStr, 10, 32)

		if err == nil && priority64 >= 1 && priority64 <= 10 {
			priority32 := int32(priority64)
			filters.Priority = &priority32
		}
	}

	if c.Query("id") != "" {
		id, err := strconv.ParseInt(c.Query("id"), 10, 64)
		if err == nil {
			filters.ID = &id
		}
	}

	if c.Query("source") != "" {
		source := c.Query("source")
		filters.Source = &source
	}

	slides, total, err := s.slideRepo.AdminList(c.Request.Context(), params, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

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
	slog.Debug("Admin Create Slide: Received request to create a new slide")

	slide, err := s.parseSlidePayload(c)
	if err != nil {
		slog.Debug("Admin Create Slide: Error parsing slide payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})

		return
	} else {
		slog.Info(
			"Admin Create Slide: Successfully parsed slide payload",
			"title",
			slide.Content.Title,
			"type",
			slide.Content.Type,
		)
	}

	if err := s.slideRepo.Save(c.Request.Context(), slide); err != nil {
		slog.Error("Admin Create Slide: Failed to create slide", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create slide"})

		return
	} else {
		slog.Info("Admin Create Slide: Successfully created slide", "id", slide.ID)
	}

	s.streamer.Broadcast(EventCreate, slide)

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

	s.streamer.Broadcast(EventUpdate, slide)

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

	s.streamer.Broadcast(EventDelete, id)

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Server) parseSlidePayload(c *gin.Context) (*domain.Slide, error) {
	var slide domain.Slide

	contentType := c.GetHeader("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		priority, err := strconv.Atoi(c.PostForm("display_options.priority"))
		if err != nil {
			priority = 1
		}

		slide.Status = domain.SlideStatus(c.PostForm("status"))
		slide.Content.Type = domain.SlideType(c.PostForm("content.type"))
		slide.Content.Title = c.PostForm("content.title")
		slide.Content.Body = c.PostForm("content.body")
		slide.Author.DisplayName = c.PostForm("author.display_name")
		slide.DisplayOptions.Priority = priority
		slide.DisplayOptions.AllowSocialOverlay = c.PostForm("display_options.allow_social_overlay") == "true"

		newPath, err := s.mediaProcessor.ProcessSlideImage(c, "image_upload")
		if err == nil && newPath != "" {
			slide.Content.Media = &domain.Media{
				LocalURL: newPath,
				MimeType: "image/webp",
			}
		}

		return &slide, nil
	}

	if err := c.ShouldBindJSON(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}
