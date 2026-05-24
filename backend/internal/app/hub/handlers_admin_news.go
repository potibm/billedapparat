package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) adminListNews(c *gin.Context) {
	params := repository.NewsListParams{
		ListParams: s.getListParams(c),
	}

	filters := newNewsListFiltersFromContext(c)

	slides, total, err := s.newsRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list news: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	c.JSON(http.StatusOK, slides)
}

func newNewsListFiltersFromContext(c *gin.Context) repository.NewsListFilters {
	filters := repository.NewsListFilters{}

	if q := c.Query("q"); q != "" {
		filters.Query = &q
	}

	if c.Query("hidden_yes") == "true" {
		isHidden := true
		filters.IsHidden = &isHidden
	} else if c.Query("hidden_no") == "true" {
		isHidden := false
		filters.IsHidden = &isHidden
	}

	if c.Query("urgent_yes") == "true" {
		isUrgent := true
		filters.IsUrgent = &isUrgent
	} else if c.Query("urgent_no") == "true" {
		isUrgent := false
		filters.IsUrgent = &isUrgent
	}

	if c.Query("external_id") != "" {
		externalID := c.Query("external_id")
		filters.ExternalID = &externalID
	}

	return filters
}

func (s *Server) adminGetNews(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	slide, err := s.newsRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondWithNotFoundProblem(c, "News item with ID "+strconv.FormatInt(id, 10)+" not found")

		return
	}

	c.JSON(http.StatusOK, slide)
}
