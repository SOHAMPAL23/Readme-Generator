package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"readmeai/core/models"
	"readmeai/core/services"
)

// GenerateHandler handles POST /api/generate
type GenerateHandler struct {
	svc *services.ReadmeService
}

// NewGenerateHandler creates a new GenerateHandler.
func NewGenerateHandler(svc *services.ReadmeService) *GenerateHandler {
	return &GenerateHandler{svc: svc}
}

// Handle processes the POST /api/generate request.
func (h *GenerateHandler) Handle(c *gin.Context) {
	var req models.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if req.RepoURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "repo_url is required",
		})
		return
	}

	resp, err := h.svc.Generate(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()

		// Map specific errors to appropriate HTTP status codes
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "invalid URL") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(errMsg, "rate limit") {
			statusCode = http.StatusTooManyRequests
		} else if strings.Contains(errMsg, "invalid OpenAI API key") || strings.Contains(errMsg, "OPENAI_API_KEY") {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
