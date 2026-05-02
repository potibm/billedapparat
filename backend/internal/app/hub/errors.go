package hub

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProblemDetails struct {
	Type     string `json:"type"`               // URI zur Fehler-Dokumentation (oft "about:blank")
	Title    string `json:"title"`              // Kurze, menschenlesbare Zusammenfassung
	Status   int    `json:"status"`             // Der HTTP Status Code
	Detail   string `json:"detail,omitempty"`   // Spezifische Infos zu diesem Fehler
	Instance string `json:"instance,omitempty"` // URI des konkreten Requests
}

func respondWithProblem(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")

	c.AbortWithStatusJSON(status, ProblemDetails{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.RequestURI,
	})
}

func respondWithInvalidIDFormatProblem(c *gin.Context) {
	respondWithProblem(c, http.StatusBadRequest, "Bad Request", "Invalid ID format")
}

func respondWithInternalServerProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusInternalServerError, "Internal Server Error", detail)
}

func respondWithNotFoundProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusNotFound, "Not Found", detail)
}

func respondWithFailedToParsePayloadProblem(c *gin.Context, err error) {
	respondWithBadRequestProblem(c, "Failed to parse payload: "+err.Error())
}

func respondWithBadRequestProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusBadRequest, "Bad Request", detail)
}

func respondWithUnauthorizedProblem(c *gin.Context, detail string) {
	respondWithProblem(c, http.StatusUnauthorized, "Unauthorized", detail)
}
