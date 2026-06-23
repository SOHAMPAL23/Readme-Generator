package models

// ReadmeStyle defines the tone and structure of the generated README.
type ReadmeStyle string

const (
	StyleDeveloper  ReadmeStyle = "developer"
	StyleStartup    ReadmeStyle = "startup"
	StylePortfolio  ReadmeStyle = "portfolio"
	StyleOpenSource ReadmeStyle = "opensource"
)

// GenerateRequest is the incoming API request body.
type GenerateRequest struct {
	RepoURL string      `json:"repo_url" binding:"required"`
	Style   ReadmeStyle `json:"style"`
}

// GenerateResponse is the full API response returned after generation.
type GenerateResponse struct {
	Repository  *RepoInfo    `json:"repository"`
	TechStack   *TechStack   `json:"tech_stack"`
	HealthScore *HealthScore `json:"health_score"`
	Readme      string       `json:"readme"`
	Style       ReadmeStyle  `json:"style"`
	GeneratedAt string       `json:"generated_at"`
}

// RepoMetaResponse is returned by GET /api/repository (no AI call).
type RepoMetaResponse struct {
	Repository  *RepoInfo    `json:"repository"`
	TechStack   *TechStack   `json:"tech_stack"`
	HealthScore *HealthScore `json:"health_score"`
}

// ErrorResponse is a standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
