package hub

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) adminListSlides(c *gin.Context) {
	slideType := domain.SlideType(c.Query("type"))

	params := repository.SlideListParams{
		ListParams: s.getListParams(c),
		Type:       slideType,
	}

	filters := newSlideListFiltersFromContext(c)

	slides, total, err := s.slideRepo.AdminList(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list slides: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	c.JSON(http.StatusOK, slides)
}

func newSlideListFiltersFromContext(c *gin.Context) repository.SlideListFilters {
	filters := repository.SlideListFilters{}

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

	return filters
}

func (s *Server) adminGetSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	slide, err := s.slideRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondWithNotFoundProblem(c, "Slide with ID "+strconv.FormatInt(id, 10)+" not found")

		return
	}

	c.JSON(http.StatusOK, slide)
}

func (s *Server) adminCreateSlide(c *gin.Context) {
	slog.Debug("Admin Create Slide: Received request to create a new slide")

	slide, err := s.parseSlidePayload(c)
	if err != nil {
		slog.Debug("Admin Create Slide: Error parsing slide payload", "error", err)
		respondWithFailedToParsePayloadProblem(c, err)

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
		respondWithInternalServerProblem(c, "Failed to create slide: "+err.Error())

		return
	} else {
		slog.Info("Admin Create Slide: Successfully created slide", "id", slide.ID)
	}

	s.streamer.Broadcast(domain.EventCreate, slide)

	c.JSON(http.StatusCreated, slide)
}

func (s *Server) adminUpdateSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	slide, err := s.parseSlidePayload(c)
	if err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	_, _ = s.ensureSlideIsNotReadonly(c, id)

	slide.ID = id

	if err := s.slideRepo.Save(c.Request.Context(), slide); err != nil {
		respondWithInternalServerProblem(c, "Failed to update slide: "+err.Error())

		return
	}

	s.streamer.Broadcast(domain.EventUpdate, slide)

	c.JSON(http.StatusOK, slide)
}

func (s *Server) ensureSlideIsNotReadonly(c *gin.Context, id int64) (*domain.Slide, bool) {
	slide, err := s.slideRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondWithNotFoundProblem(c, "Slide with ID "+strconv.FormatInt(id, 10)+" not found")

		return nil, false
	}

	if slide.Content.Type.IsReadonly() {
		respondWithBadRequestProblem(c, "Cannot modify news or timetable slide")

		return nil, false
	}

	return slide, true
}

func (s *Server) adminDeleteSlide(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	_, _ = s.ensureSlideIsNotReadonly(c, id)

	if err := s.slideRepo.Delete(c.Request.Context(), id); err != nil {
		respondWithInternalServerProblem(c, "Failed to delete slide: "+err.Error())

		return
	}

	s.streamer.Broadcast(domain.EventDelete, id)

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (s *Server) parseSlidePayload(c *gin.Context) (*domain.Slide, error) {
	if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
		return s.parseMultipartSlide(c)
	}

	var slide domain.Slide
	if err := c.ShouldBindJSON(&slide); err != nil {
		return nil, err
	}

	return &slide, nil
}

func (s *Server) parseMultipartSlide(c *gin.Context) (*domain.Slide, error) {
	var slide domain.Slide

	slide.Author = &domain.Author{}

	priority := 1
	if p, err := strconv.Atoi(c.PostForm("display_options.priority")); err == nil {
		priority = p
	}

	slide.Status = domain.SlideStatus(c.PostForm("status"))
	slide.Content.Type = domain.SlideType(c.PostForm("content.type"))
	slide.Content.Title = c.PostForm("content.title")
	slide.Content.Body = c.PostForm("content.body")
	slide.Author.DisplayName = c.PostForm("author.display_name")
	slide.DisplayOptions.Priority = priority
	slide.DisplayOptions.AllowSocialOverlay = c.PostForm("display_options.allow_social_overlay") == "true"

	_, fileErr := c.FormFile("image_upload")
	if fileErr != nil {
		return &slide, nil
	}

	newPath, err := s.mediaProcessor.ProcessSlideImage(c, "image_upload")
	if err != nil {
		return nil, fmt.Errorf("failed to process image_upload: %w", err)
	}

	if newPath != "" {
		slide.Content.Media = &domain.Media{
			LocalURL: newPath,
			MimeType: "image/webp",
		}
	}

	return &slide, nil
}
