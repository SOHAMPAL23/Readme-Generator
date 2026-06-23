package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"readmeai/internal/models"
)

// BuildPrompt constructs a structured prompt for the OpenAI API based on style.
func BuildPrompt(analysis *models.AnalysisResult, style models.ReadmeStyle) (string, string) {
	systemPrompt := buildSystemPrompt(style)
	userPrompt := buildUserPrompt(analysis, style)
	return systemPrompt, userPrompt
}

func buildSystemPrompt(style models.ReadmeStyle) string {
	base := `You are an expert technical writer and software engineer specializing in writing professional GitHub README files.
Your task is to generate a high-quality, complete README.md file based on the provided repository analysis data.

CRITICAL RULES:
- Only use technologies and facts provided in the analysis — do NOT invent features
- Use proper Markdown formatting with headers, code blocks, tables, and badges
- Include working badge URLs from shields.io where applicable
- In shields.io badge URLs, any percentage symbol (%) must be written as %25 (e.g. HTML-58.1%25-orange)
- Write in clear, professional English
- Do not include placeholder sections unless the actual data supports them
- For any image or screenshot placeholders, use placehold.co (e.g., https://placehold.co/600x400) and NEVER use via.placeholder.com
- Output ONLY the Markdown content — no preamble, no explanation`

	var styleGuide string
	switch style {
	case models.StyleDeveloper:
		styleGuide = `
STYLE: Developer-focused technical README
- Lead with technical depth and implementation details
- Include detailed installation steps, environment variables, and API documentation
- Use code blocks extensively for commands, code samples, and config
- Add a detailed architecture section
- Include contributing guidelines and development setup`

	case models.StyleStartup:
		styleGuide = `
STYLE: Startup / product-focused README
- Lead with the problem being solved and the value proposition
- Focus on WHY this project exists and WHAT problem it solves
- Make it compelling for investors, early adopters, and potential hires
- Keep technical details high-level but include a quick-start section
- Include a roadmap section highlighting future vision`

	case models.StylePortfolio:
		styleGuide = `
STYLE: Portfolio / recruiter-friendly README
- Write as if you're showcasing this project to a potential employer
- Highlight the developer's skills and technical decisions
- Include a "What I Learned" or "Technical Challenges" section
- Make it scannable — recruiters spend 30 seconds on a README
- Include screenshots placeholder section prominently
- Emphasize the tech stack clearly at the top`

	case models.StyleOpenSource:
		styleGuide = `
STYLE: Open Source community README
- Focus on welcoming contributors
- Include detailed CONTRIBUTING section with code of conduct
- Add issue templates guidance
- Include community links and communication channels
- Add a section for "Good first issues" or contribution guidelines
- Include acknowledgements and credits section`
	}

	return base + styleGuide
}

func buildUserPrompt(analysis *models.AnalysisResult, style models.ReadmeStyle) string {
	// Serialize tech stack
	techJSON, _ := json.MarshalIndent(analysis.TechStack, "", "  ")
	healthJSON, _ := json.MarshalIndent(analysis.HealthScore, "", "  ")

	// Build file tree summary
	fileList := buildFileList(analysis.FileTree)

	// Build dependency summary
	depSummary := buildDepSummary(analysis.Dependencies)

	var sb strings.Builder
	sb.WriteString("# Repository Analysis Data\n\n")

	// Core info
	sb.WriteString(fmt.Sprintf("## Repository: %s\n", analysis.RepoInfo.FullName))
	sb.WriteString(fmt.Sprintf("- **Owner**: %s\n", analysis.RepoInfo.Owner))
	sb.WriteString(fmt.Sprintf("- **Description**: %s\n", orDefault(analysis.RepoInfo.Description, "No description provided")))
	sb.WriteString(fmt.Sprintf("- **Primary Language**: %s\n", orDefault(analysis.RepoInfo.PrimaryLanguage, "Unknown")))
	sb.WriteString(fmt.Sprintf("- **Stars**: %d | **Forks**: %d | **Watchers**: %d\n",
		analysis.RepoInfo.Stars, analysis.RepoInfo.Forks, analysis.RepoInfo.Watchers))
	sb.WriteString(fmt.Sprintf("- **Open Issues**: %d\n", analysis.RepoInfo.OpenIssues))
	sb.WriteString(fmt.Sprintf("- **License**: %s\n", orDefault(analysis.RepoInfo.License, "No license detected")))
	sb.WriteString(fmt.Sprintf("- **Homepage**: %s\n", orDefault(analysis.RepoInfo.Homepage, "None")))
	sb.WriteString(fmt.Sprintf("- **Repository URL**: %s\n", analysis.RepoInfo.HTMLURL))
	sb.WriteString(fmt.Sprintf("- **Default Branch**: %s\n", analysis.RepoInfo.DefaultBranch))
	sb.WriteString(fmt.Sprintf("- **Size**: %d KB\n", analysis.RepoInfo.Size))
	sb.WriteString(fmt.Sprintf("- **Has Existing README**: %v\n", analysis.RepoInfo.HasReadme))

	if len(analysis.RepoInfo.Topics) > 0 {
		sb.WriteString(fmt.Sprintf("- **Topics**: %s\n", strings.Join(analysis.RepoInfo.Topics, ", ")))
	}

	// Languages
	if len(analysis.RepoInfo.LanguagesPercent) > 0 {
		sb.WriteString("\n## Language Breakdown\n")
		for lang, pct := range analysis.RepoInfo.LanguagesPercent {
			sb.WriteString(fmt.Sprintf("- %s: %.1f%%\n", lang, pct))
		}
	}

	// Tech stack
	sb.WriteString("\n## Detected Tech Stack\n")
	sb.WriteString("```json\n")
	sb.Write(techJSON)
	sb.WriteString("\n```\n")

	// File structure
	if fileList != "" {
		sb.WriteString("\n## Repository Structure (top 2 levels)\n")
		sb.WriteString(fileList)
	}

	// Dependencies
	if depSummary != "" {
		sb.WriteString("\n## Key Dependencies\n")
		sb.WriteString(depSummary)
	}

	// Health score
	sb.WriteString("\n## Repository Health\n")
	sb.WriteString("```json\n")
	sb.Write(healthJSON)
	sb.WriteString("\n```\n")

	// Style instruction
	sb.WriteString(fmt.Sprintf("\n## Required Output Style: %s\n", string(style)))

	sb.WriteString(`
## Required README Sections
Generate a complete README.md with these sections (adapt based on available data):
1. Badges (shields.io) — language, license, stars, last commit
2. Project title and one-line tagline
3. Overview / About
4. Features list (infer from tech stack and description)
5. Tech Stack (with icons if possible using simple-icons or devicons CDN links)
6. Architecture (describe based on detected stack)
7. Installation & Setup
8. Environment Variables (table format if detected)
9. Usage / Quick Start
10. API Endpoints (if backend detected)
11. Folder Structure (use the provided file tree)
12. Screenshots (include placeholder section)
13. Deployment
14. Roadmap (3-5 realistic future features)
15. Contributing
16. License
17. Author / Contact

Generate the complete README.md now:`)

	return sb.String()
}

func buildFileList(nodes []*models.FileNode) string {
	if len(nodes) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("```\n")
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.Type == "dir" {
			sb.WriteString(fmt.Sprintf("├── %s/\n", node.Name))
			for _, child := range node.Children {
				if child == nil {
					continue
				}
				childType := ""
				if child.Type == "dir" {
					childType = "/"
				}
				sb.WriteString(fmt.Sprintf("│   ├── %s%s\n", child.Name, childType))
			}
		} else {
			sb.WriteString(fmt.Sprintf("├── %s\n", node.Name))
		}
	}
	sb.WriteString("```")
	return sb.String()
}


func buildDepSummary(deps []models.DependencyFile) string {
	if len(deps) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, df := range deps {
		sb.WriteString(fmt.Sprintf("### %s\n", df.FileName))
		shown := df.Dependencies
		if len(shown) > 20 {
			shown = shown[:20]
		}
		sb.WriteString(strings.Join(shown, ", "))
		if len(df.Dependencies) > 20 {
			sb.WriteString(fmt.Sprintf(" ... and %d more", len(df.Dependencies)-20))
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
