package github

import (
	"encoding/base64"
	"strings"
	"time"

	"readmeai/internal/models"
)

// ParseRepoInfo maps the raw GitHub API JSON to a RepoInfo model.
func ParseRepoInfo(raw map[string]interface{}) *models.RepoInfo {
	info := &models.RepoInfo{}

	if v, ok := raw["name"].(string); ok {
		info.Name = v
	}
	if v, ok := raw["full_name"].(string); ok {
		info.FullName = v
	}
	if owner, ok := raw["owner"].(map[string]interface{}); ok {
		if login, ok := owner["login"].(string); ok {
			info.Owner = login
		}
	}
	if v, ok := raw["description"].(string); ok {
		info.Description = v
	}
	if v, ok := raw["stargazers_count"].(float64); ok {
		info.Stars = int(v)
	}
	if v, ok := raw["forks_count"].(float64); ok {
		info.Forks = int(v)
	}
	if v, ok := raw["watchers_count"].(float64); ok {
		info.Watchers = int(v)
	}
	if v, ok := raw["open_issues_count"].(float64); ok {
		info.OpenIssues = int(v)
	}
	if v, ok := raw["language"].(string); ok {
		info.PrimaryLanguage = v
	}
	if v, ok := raw["size"].(float64); ok {
		info.Size = int(v)
	}
	if v, ok := raw["default_branch"].(string); ok {
		info.DefaultBranch = v
	}
	if v, ok := raw["private"].(bool); ok {
		info.IsPrivate = v
	}
	if v, ok := raw["homepage"].(string); ok {
		info.Homepage = v
	}
	if v, ok := raw["html_url"].(string); ok {
		info.HTMLURL = v
	}
	if v, ok := raw["clone_url"].(string); ok {
		info.CloneURL = v
	}
	if license, ok := raw["license"].(map[string]interface{}); ok {
		if name, ok := license["name"].(string); ok {
			info.License = name
		}
	}
	if v, ok := raw["created_at"].(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			info.CreatedAt = t
		}
	}
	if v, ok := raw["updated_at"].(string); ok {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			info.UpdatedAt = t
		}
	}

	return info
}

// ParseLanguages converts raw language bytes to percentages.
func ParseLanguages(raw map[string]interface{}) (map[string]int64, map[string]float64) {
	bytes := make(map[string]int64)
	pct := make(map[string]float64)

	if raw == nil {
		return bytes, pct
	}

	var total int64
	for _, v := range raw {
		if b, ok := v.(float64); ok {
			total += int64(b)
		}
	}

	for lang, v := range raw {
		if b, ok := v.(float64); ok {
			bytes[lang] = int64(b)
			if total > 0 {
				pct[lang] = float64(int64(b)*1000/total) / 10
			}
		}
	}
	return bytes, pct
}

// DecodeBase64Content decodes a base64-encoded file from the GitHub contents API.
func DecodeBase64Content(encoded string) string {
	// GitHub adds newlines every 60 chars; strip them
	encoded = strings.ReplaceAll(encoded, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(decoded)
}

// CalculateHealthScore evaluates the repo and returns a 0-100 score + suggestions.
func CalculateHealthScore(info *models.RepoInfo, keyFiles []string, techStack *models.TechStack) *models.HealthScore {
	score := 0
	breakdown := make(map[string]int)
	var suggestions []string

	fileSet := make(map[string]bool)
	for _, f := range keyFiles {
		fileSet[strings.ToLower(f)] = true
	}

	// Description: 15 pts
	if info.Description != "" {
		score += 15
		breakdown["description"] = 15
	} else {
		suggestions = append(suggestions, "Add a repository description to help visitors understand the project")
	}

	// README: 20 pts
	if info.HasReadme {
		score += 20
		breakdown["readme"] = 20
	} else {
		suggestions = append(suggestions, "Create a README.md — it's the first thing visitors see")
	}

	// License: 10 pts
	if info.License != "" && info.License != "Other" {
		score += 10
		breakdown["license"] = 10
	} else {
		suggestions = append(suggestions, "Add a LICENSE file to clarify how others can use your code")
	}

	// Topics: 5 pts
	if len(info.Topics) >= 3 {
		score += 5
		breakdown["topics"] = 5
	} else {
		suggestions = append(suggestions, "Add at least 3 repository topics to improve discoverability")
	}

	// CI/CD: 15 pts
	if fileSet[".github/workflows"] || fileSet["jenkinsfile"] || fileSet[".circleci"] || fileSet["travis.yml"] || fileSet[".travis.yml"] {
		score += 15
		breakdown["ci_cd"] = 15
	} else {
		suggestions = append(suggestions, "Set up GitHub Actions or another CI/CD pipeline")
	}

	// Tests: 15 pts
	if fileSet["test"] || fileSet["tests"] || fileSet["__tests__"] || fileSet["spec"] || fileSet["testing"] {
		score += 15
		breakdown["tests"] = 15
	} else {
		suggestions = append(suggestions, "Add tests to improve code reliability and reviewer confidence")
	}

	// Contributing guide: 5 pts
	if fileSet["contributing.md"] || fileSet["contributing"] {
		score += 5
		breakdown["contributing"] = 5
	} else {
		suggestions = append(suggestions, "Add a CONTRIBUTING.md to guide open-source contributors")
	}

	// Docker: 5 pts
	if fileSet["dockerfile"] || fileSet["docker-compose.yml"] {
		score += 5
		breakdown["docker"] = 5
	} else {
		suggestions = append(suggestions, "Add a Dockerfile for easy deployment and reproducibility")
	}

	// Stars bonus: up to 5 pts
	if info.Stars >= 100 {
		score += 5
		breakdown["stars"] = 5
	} else if info.Stars >= 10 {
		score += 3
		breakdown["stars"] = 3
	}

	// Has tech stack: 5 pts
	if len(techStack.Backend) > 0 || len(techStack.Frontend) > 0 {
		score += 5
		breakdown["tech_stack"] = 5
	}

	if score > 100 {
		score = 100
	}

	return &models.HealthScore{
		Score:       score,
		Suggestions: suggestions,
		Breakdown:   breakdown,
	}
}
