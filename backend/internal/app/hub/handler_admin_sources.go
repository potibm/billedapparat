package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (s *Server) adminListSources(c *gin.Context) {
	sources := []map[string]string{}
	titleCaser := cases.Title(language.Und)

	for name := range s.cfg.Collectors {
		sources = append(sources, map[string]string{
			"id":   name,
			"name": titleCaser.String(name),
		})
	}

	c.Header("X-Total-Count", strconv.Itoa(len(sources)))
	c.JSON(http.StatusOK, sources)
}
