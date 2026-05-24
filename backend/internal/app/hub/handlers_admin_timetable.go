package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) adminListTimetable(c *gin.Context) {
	start, _ := strconv.Atoi(c.DefaultQuery("_start", "0"))
	end, _ := strconv.Atoi(c.DefaultQuery("_end", "20"))
	sort := c.DefaultQuery("_sort", "id")
	order := c.DefaultQuery("_order", "DESC")

	params := repository.TimetableEventListParams{
		ListParams: repository.ListParams{
			Offset: max(start, 0),
			Limit:  max(end-start, 0),
			Sort:   sort,
			Order:  order,
		},
	}

	filters := repository.TimetableEventListFilters{
		Query:    nil,
		Source:   nil,
		IsHidden: nil,
	}
	if c.Query("q") != "" {
		query := c.Query("q")
		filters.Query = &query
	}

	if c.Query("source") != "" {
		source := c.Query("source")
		filters.Source = &source
	}

	if c.Query("is_hidden") != "" {
		isHidden, err := strconv.ParseBool(c.Query("is_hidden"))
		if err == nil {
			filters.IsHidden = &isHidden
		}
	}

	slides, total, err := s.timetableEventRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list timetable events: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	c.JSON(http.StatusOK, slides)
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
