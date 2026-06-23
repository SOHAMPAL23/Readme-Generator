package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var githubURLRegex = regexp.MustCompile(`^https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+?)(?:\.git)?(?:/.*)?$`)

// ParseGitHubURL validates a GitHub URL and returns the owner and repo name.
// Returns an error if the URL is not a valid GitHub repository URL.
func ParseGitHubURL(rawURL string) (owner, repo string, err error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return "", "", fmt.Errorf("repository URL cannot be empty")
	}

	matches := githubURLRegex.FindStringSubmatch(rawURL)
	if matches == nil {
		return "", "", fmt.Errorf("invalid GitHub URL: must be in format https://github.com/owner/repository")
	}

	owner = matches[1]
	repo = matches[2]

	// Strip trailing slashes/paths from repo name
	if idx := strings.IndexAny(repo, "/#?"); idx != -1 {
		repo = repo[:idx]
	}
	repo = strings.TrimSuffix(repo, ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not extract owner or repository name from URL")
	}

	return owner, repo, nil
}

// ValidateStyle ensures the style is one of the accepted values, defaulting to "developer".
func ValidateStyle(style string) string {
	valid := map[string]bool{
		"developer":  true,
		"startup":    true,
		"portfolio":  true,
		"opensource": true,
	}

	style = strings.ToLower(strings.TrimSpace(style))
	if !valid[style] {
		return "developer"
	}
	return style
}
