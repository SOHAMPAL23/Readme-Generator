package models

import "time"

// RepoInfo holds all metadata fetched from the GitHub API.
type RepoInfo struct {
	Name            string            `json:"name"`
	FullName        string            `json:"full_name"`
	Owner           string            `json:"owner"`
	Description     string            `json:"description"`
	Stars           int               `json:"stars"`
	Forks           int               `json:"forks"`
	Watchers        int               `json:"watchers"`
	OpenIssues      int               `json:"open_issues"`
	Topics          []string          `json:"topics"`
	PrimaryLanguage string            `json:"primary_language"`
	Languages       map[string]int64  `json:"languages"`
	LanguagesPercent map[string]float64 `json:"languages_percent"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	License         string            `json:"license"`
	HasReadme       bool              `json:"has_readme"`
	Size            int               `json:"size"`
	DefaultBranch   string            `json:"default_branch"`
	IsPrivate       bool              `json:"is_private"`
	Homepage        string            `json:"homepage"`
	HTMLURL         string            `json:"html_url"`
	CloneURL        string            `json:"clone_url"`
}

// TechStack holds detected technologies from repo analysis.
type TechStack struct {
	Frontend    []string `json:"frontend"`
	Backend     []string `json:"backend"`
	Databases   []string `json:"databases"`
	DevOps      []string `json:"devops"`
	Cloud       []string `json:"cloud"`
	Testing     []string `json:"testing"`
	Languages   []string `json:"languages"`
	Frameworks  []string `json:"frameworks"`
	PackageManager string `json:"package_manager"`
}

// FileNode represents a file or directory in the repo structure.
type FileNode struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "file" or "dir"
	Path     string      `json:"path"`
	Children []*FileNode `json:"children,omitempty"`
}

// HealthScore holds the repository health evaluation.
type HealthScore struct {
	Score       int      `json:"score"`
	Suggestions []string `json:"suggestions"`
	Breakdown   map[string]int `json:"breakdown"`
}

// DependencyFile represents a parsed dependency manifest.
type DependencyFile struct {
	FileName     string   `json:"file_name"`
	Dependencies []string `json:"dependencies"`
}

// AnalysisResult is the combined output from the GitHub analyzer.
type AnalysisResult struct {
	RepoInfo        *RepoInfo        `json:"repo_info"`
	TechStack       *TechStack       `json:"tech_stack"`
	FileTree        []*FileNode      `json:"file_tree"`
	Dependencies    []DependencyFile `json:"dependencies"`
	HealthScore     *HealthScore     `json:"health_score"`
	KeyFiles        []string         `json:"key_files"`
}
