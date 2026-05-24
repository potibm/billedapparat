package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) adminListNews(c *gin.Context) {
	start, _ := strconv.Atoi(c.DefaultQuery("_start", "0"))
	end, _ := strconv.Atoi(c.DefaultQuery("_end", "20"))
	sort := c.DefaultQuery("_sort", "id")
	order := c.DefaultQuery("_order", "DESC")

	params := repository.NewsListParams{
		ListParams: repository.ListParams{
			Offset: max(start, 0),
			Limit:  max(end-start, 0),
			Sort:   sort,
			Order:  order,
		},
	}

	filters := repository.NewsListFilters{
		Query:    nil,
		IsUrgent: nil,
		IsHidden: nil,
	}
	if c.Query("q") != "" {
		query := c.Query("q")
		filters.Query = &query
	}

	if c.Query("is_urgent") != "" {
		isUrgent, err := strconv.ParseBool(c.Query("is_urgent"))
		if err == nil {
			filters.IsUrgent = &isUrgent
		}
	}

	if c.Query("is_hidden") != "" {
		isHidden, err := strconv.ParseBool(c.Query("is_hidden"))
		if err == nil {
			filters.IsHidden = &isHidden
		}
	}

	slides, total, err := s.newsRepo.List(c.Request.Context(), params, filters)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list news: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))

	c.JSON(http.StatusOK, slides)
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
