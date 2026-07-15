package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"readmeai/core/models"
	"readmeai/core/services"
)

// RepositoryHandler handles GET /api/repository
type RepositoryHandler struct {
	svc *services.ReadmeService
}

// NewRepositoryHandler creates a new RepositoryHandler.
func NewRepositoryHandler(svc *services.ReadmeService) *RepositoryHandler {
	return &RepositoryHandler{svc: svc}
}

// Handle processes the GET /api/repository?url= request.
func (h *RepositoryHandler) Handle(c *gin.Context) {
	repoURL := c.Query("url")
	if repoURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "url query parameter is required",
		})
		return
	}

	resp, err := h.svc.GetRepoMeta(repoURL)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid URL") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "rate limit") {
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
