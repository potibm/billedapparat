package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) adminListTimetable(c *gin.Context) {
	params := repository.TimetableEventListParams{
		ListParams: s.getListParams(c),
	}

	filters := newTimetableEventListFiltersFromContext(c)

	slides, total, err := s.timetableEventRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list timetable events: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	c.JSON(http.StatusOK, slides)
}

func newTimetableEventListFiltersFromContext(c *gin.Context) repository.TimetableEventListFilters {
	filters := repository.TimetableEventListFilters{
		Query:      nil,
		Source:     nil,
		IsHidden:   nil,
		ExternalID: nil,
	}
	if c.Query("q") != "" {
		query := c.Query("q")
		filters.Query = &query
	}

	if c.Query("source") != "" {
		source := c.Query("source")
		filters.Source = &source
	}

	if c.Query("hidden_yes") == "true" {
		isHidden := true
		filters.IsHidden = &isHidden
	} else if c.Query("hidden_no") == "true" {
		isHidden := false
		filters.IsHidden = &isHidden
	}

	if c.Query("external_id") != "" {
		externalID := c.Query("external_id")
		filters.ExternalID = &externalID
	}

	return filters
}

func (s *Server) adminGetTimetable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	slide, err := s.timetableEventRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		respondWithNotFoundProblem(c, "Timetable item with ID "+strconv.FormatInt(id, 10)+" not found")

		return
	}

	c.JSON(http.StatusOK, slide)
}
