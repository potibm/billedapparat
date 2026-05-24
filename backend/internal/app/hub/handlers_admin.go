package hub

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/repository"
)

func (s *Server) getListParams(c *gin.Context) repository.ListParams {
	start, _ := strconv.Atoi(c.DefaultQuery("_start", "0"))
	end, _ := strconv.Atoi(c.DefaultQuery("_end", "20"))
	sort := c.DefaultQuery("_sort", "id")
	order := c.DefaultQuery("_order", "DESC")

	return repository.ListParams{
		Offset: max(start, 0),
		Limit:  max(end-start, 0),
		Sort:   sort,
		Order:  order,
	}
}
