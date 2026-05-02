package hub

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/potibm/billedapparat/internal/app/domain"
)

type filterRulePayload struct {
	Source string            `json:"source" binding:"required"`
	Type   domain.FilterType `json:"type"   binding:"required,oneof=language username display_name"`
	Value  string            `json:"value"  binding:"required"`
}

func (s *Server) adminListFilterRules(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	rules, total, err := s.filterRuleRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to list filter rules: "+err.Error())

		return
	}

	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, rules)
}

func (s *Server) adminGetFilterRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	rule, err := s.filterRuleRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		respondWithInternalServerProblem(c, "Failed to fetch rule: "+err.Error())

		return
	}

	if rule == nil {
		respondWithNotFoundProblem(c, "Filter rule with ID "+strconv.FormatUint(id, 10)+" not found")

		return
	}

	c.JSON(http.StatusOK, rule)
}

func (s *Server) adminCreateFilterRule(c *gin.Context) {
	var payload filterRulePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	rule := &domain.FilterRule{
		Source: payload.Source,
		Type:   payload.Type,
		Value:  payload.Value,
	}

	if err := s.filterRuleRepo.Create(c.Request.Context(), rule); err != nil {
		respondWithInternalServerProblem(c, "Failed to create filter rule: "+err.Error())

		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (s *Server) adminUpdateFilterRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	var payload filterRulePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithFailedToParsePayloadProblem(c, err)

		return
	}

	rule := &domain.FilterRule{
		ID:     int64(id),
		Source: payload.Source,
		Type:   payload.Type,
		Value:  payload.Value,
	}

	if err := s.filterRuleRepo.Update(c.Request.Context(), rule); err != nil {
		respondWithInternalServerProblem(c, "Failed to update filter rule: "+err.Error())

		return
	}

	updatedRule, _ := s.filterRuleRepo.GetByID(c.Request.Context(), uint(id))

	c.JSON(http.StatusOK, updatedRule)
}

func (s *Server) adminDeleteFilterRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondWithInvalidIDFormatProblem(c)

		return
	}

	if err := s.filterRuleRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		respondWithInternalServerProblem(c, "Failed to delete filter rule: "+err.Error())

		return
	}

	c.Status(http.StatusNoContent) // 204
}
