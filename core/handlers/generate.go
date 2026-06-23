package handlers

import (
	"net/http"

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
		if contains(errMsg, "not found") || contains(errMsg, "invalid URL") {
			statusCode = http.StatusBadRequest
		} else if contains(errMsg, "rate limit") {
			statusCode = http.StatusTooManyRequests
		} else if contains(errMsg, "invalid OpenAI API key") || contains(errMsg, "OPENAI_API_KEY") {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error: errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
