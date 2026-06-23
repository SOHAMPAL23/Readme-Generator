package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	githubclient "readmeai/internal/github"
	"readmeai/internal/ai"
	"readmeai/internal/models"
	"readmeai/internal/utils"
)

// ReadmeService orchestrates the full pipeline: GitHub → Analyze → AI → Response.
type ReadmeService struct {
	ghClient  *githubclient.Client
	analyzer  *githubclient.Analyzer
	generator *ai.Generator
}

// NewReadmeService creates a new ReadmeService with all dependencies initialized.
func NewReadmeService() *ReadmeService {
	client := githubclient.NewClient()
	return &ReadmeService{
		ghClient:  client,
		analyzer:  githubclient.NewAnalyzer(client),
		generator: ai.NewGenerator(),
	}
}

// Generate runs the full pipeline and returns a GenerateResponse.
func (s *ReadmeService) Generate(req *models.GenerateRequest) (*models.GenerateResponse, error) {
	// 1. Validate and parse URL
	owner, repo, err := utils.ParseGitHubURL(req.RepoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	style := models.ReadmeStyle(utils.ValidateStyle(string(req.Style)))

	// 2. Fetch repo data in parallel
	analysis, err := s.fetchAndAnalyze(owner, repo)
	if err != nil {
		return nil, err
	}

	// 3. Generate README with AI
	readme, err := s.generator.Generate(analysis, style)
	if err != nil {
		return nil, fmt.Errorf("README generation failed: %w", err)
	}

	return &models.GenerateResponse{
		Repository:  analysis.RepoInfo,
		TechStack:   analysis.TechStack,
		HealthScore: analysis.HealthScore,
		Readme:      readme,
		Style:       style,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// GetRepoMeta returns repository metadata without calling AI.
func (s *ReadmeService) GetRepoMeta(repoURL string) (*models.RepoMetaResponse, error) {
	owner, repo, err := utils.ParseGitHubURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	analysis, err := s.fetchAndAnalyze(owner, repo)
	if err != nil {
		return nil, err
	}

	return &models.RepoMetaResponse{
		Repository:  analysis.RepoInfo,
		TechStack:   analysis.TechStack,
		HealthScore: analysis.HealthScore,
	}, nil
}

// fetchAndAnalyze runs parallel GitHub API calls and then analyzes the results.
func (s *ReadmeService) fetchAndAnalyze(owner, repo string) (*models.AnalysisResult, error) {
	var (
		rawRepo   map[string]interface{}
		rawLangs  map[string]interface{}
		topics    []string
		repoErr   error
		langErr   error
		topicsErr error
		wg        sync.WaitGroup
	)

	// Parallel fetch: repo info, languages, topics
	wg.Add(3)

	go func() {
		defer wg.Done()
		rawRepo, repoErr = s.ghClient.GetRepo(owner, repo)
	}()

	go func() {
		defer wg.Done()
		rawLangs, langErr = s.ghClient.GetLanguages(owner, repo)
	}()

	go func() {
		defer wg.Done()
		topics, topicsErr = s.ghClient.GetTopics(owner, repo)
	}()

	wg.Wait()

	if repoErr != nil {
		return nil, repoErr
	}
	if langErr != nil {
		langErr = nil // Non-fatal
	}
	if topicsErr != nil {
		topicsErr = nil // Non-fatal
	}

	// Parse repository info
	repoInfo := githubclient.ParseRepoInfo(rawRepo)
	if repoInfo == nil {
		return nil, fmt.Errorf("failed to parse repository information")
	}

	// Apply topics (they come from a separate endpoint)
	repoInfo.Topics = topics

	// Apply languages
	bytes, pct := githubclient.ParseLanguages(rawLangs)
	repoInfo.Languages = bytes
	repoInfo.LanguagesPercent = pct

	branch := repoInfo.DefaultBranch
	if branch == "" {
		branch = "main"
	}

	// Analyze file structure and dependencies
	keyFiles, fileTree, depFiles, techStack := s.analyzer.Analyze(owner, repo, branch)

	// Check for README existence
	repoInfo.HasReadme = hasReadme(keyFiles)

	// Calculate health score
	healthScore := githubclient.CalculateHealthScore(repoInfo, keyFiles, techStack)

	return &models.AnalysisResult{
		RepoInfo:     repoInfo,
		TechStack:    techStack,
		FileTree:     fileTree,
		Dependencies: depFiles,
		HealthScore:  healthScore,
		KeyFiles:     keyFiles,
	}, nil
}

func hasReadme(keyFiles []string) bool {
	for _, f := range keyFiles {
		lower := strings.ToLower(f)
		if lower == "readme.md" || lower == "readme" || lower == "readme.txt" || lower == "readme.rst" {
			return true
		}
	}
	return false
}

