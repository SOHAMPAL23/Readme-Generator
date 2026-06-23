package github

import (
	"encoding/json"
	"strings"

	"readmeai/internal/models"
)

// Analyzer detects technologies from repo file contents and dependency files.
type Analyzer struct {
	client *Client
}

// NewAnalyzer creates a new Analyzer with the given GitHub client.
func NewAnalyzer(client *Client) *Analyzer {
	return &Analyzer{client: client}
}

// Analyze fetches the repository contents and returns a full AnalysisResult.
func (a *Analyzer) Analyze(owner, repo, branch string) ([]string, []*models.FileNode, []models.DependencyFile, *models.TechStack) {
	rootContents, err := a.client.GetContents(owner, repo, "")
	if err != nil || rootContents == nil {
		return []string{}, []*models.FileNode{}, []models.DependencyFile{}, &models.TechStack{}
	}

	keyFiles := []string{}
	fileNodes := []*models.FileNode{}
	depFiles := []models.DependencyFile{}

	for _, item := range rootContents {
		name, _ := item["name"].(string)
		typ, _ := item["type"].(string)
		path, _ := item["path"].(string)

		keyFiles = append(keyFiles, strings.ToLower(name))
		node := &models.FileNode{
			Name: name,
			Type: typ,
			Path: path,
		}

		if typ == "dir" {
			subContents, err := a.client.GetContents(owner, repo, path)
			if err == nil && subContents != nil {
				for _, sub := range subContents {
					subName, _ := sub["name"].(string)
					subType, _ := sub["type"].(string)
					subPath, _ := sub["path"].(string)
					keyFiles = append(keyFiles, strings.ToLower(subPath))
					node.Children = append(node.Children, &models.FileNode{
						Name: subName,
						Type: subType,
						Path: subPath,
					})
				}
			}
		}
		fileNodes = append(fileNodes, node)
	}

	// Parse dependency files
	depFilenames := []string{
		"package.json", "go.mod", "requirements.txt",
		"pom.xml", "Cargo.toml", "composer.json", "Gemfile",
		"Pipfile", "pyproject.toml", "build.gradle",
	}

	for _, fn := range depFilenames {
		if containsFile(rootContents, fn) {
			content, err := a.client.GetRawFile(owner, repo, branch, fn)
			if err == nil && content != "" {
				deps := parseDependencies(fn, content)
				if len(deps) > 0 {
					depFiles = append(depFiles, models.DependencyFile{
						FileName:     fn,
						Dependencies: deps,
					})
				}
			}
		}
	}

	techStack := detectTechStack(keyFiles, depFiles)
	return keyFiles, fileNodes, depFiles, techStack
}

// containsFile checks if a filename exists in the contents listing.
func containsFile(contents []map[string]interface{}, name string) bool {
	for _, item := range contents {
		if n, ok := item["name"].(string); ok && strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

// parseDependencies extracts dependency names from common manifest files.
func parseDependencies(filename, content string) []string {
	switch strings.ToLower(filename) {
	case "package.json":
		return parsePackageJSON(content)
	case "go.mod":
		return parseGoMod(content)
	case "requirements.txt", "pipfile":
		return parseRequirementsTxt(content)
	case "cargo.toml":
		return parseCargoToml(content)
	}
	return []string{}
}

func parsePackageJSON(content string) []string {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal([]byte(content), &pkg); err != nil {
		return nil
	}
	deps := []string{}
	for k := range pkg.Dependencies {
		deps = append(deps, k)
	}
	for k := range pkg.DevDependencies {
		deps = append(deps, k)
	}
	return deps
}

func parseGoMod(content string) []string {
	deps := []string{}
	inRequire := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if inRequire || strings.HasPrefix(line, "require ") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 1 && !strings.HasPrefix(parts[0], "//") {
				deps = append(deps, parts[0])
			}
		}
	}
	return deps
}

func parseRequirementsTxt(content string) []string {
	deps := []string{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip version specifiers
		name := strings.FieldsFunc(line, func(r rune) bool {
			return r == '=' || r == '>' || r == '<' || r == '!' || r == '~' || r == '['
		})[0]
		deps = append(deps, strings.TrimSpace(name))
	}
	return deps
}

func parseCargoToml(content string) []string {
	deps := []string{}
	inDeps := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "[dependencies]" || line == "[dev-dependencies]" {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			deps = append(deps, strings.TrimSpace(parts[0]))
		}
	}
	return deps
}

// detectTechStack infers the technology stack from file names and dependencies.
func detectTechStack(keyFiles []string, depFiles []models.DependencyFile) *models.TechStack {
	ts := &models.TechStack{}

	// Flatten all dependency names for easy lookup
	allDeps := map[string]bool{}
	for _, df := range depFiles {
		for _, d := range df.Dependencies {
			allDeps[strings.ToLower(d)] = true
		}
	}

	fileSet := map[string]bool{}
	for _, f := range keyFiles {
		fileSet[f] = true
	}

	// Frontend
	if allDeps["react"] || allDeps["react-dom"] {
		ts.Frontend = append(ts.Frontend, "React")
	}
	if allDeps["next"] || fileSet["next.config.js"] || fileSet["next.config.ts"] || fileSet["next.config.mjs"] {
		ts.Frontend = append(ts.Frontend, "Next.js")
	}
	if allDeps["vue"] || allDeps["@vue/core"] {
		ts.Frontend = append(ts.Frontend, "Vue.js")
	}
	if allDeps["@angular/core"] {
		ts.Frontend = append(ts.Frontend, "Angular")
	}
	if allDeps["svelte"] {
		ts.Frontend = append(ts.Frontend, "Svelte")
	}
	if allDeps["vite"] {
		ts.Frontend = append(ts.Frontend, "Vite")
	}
	if allDeps["tailwindcss"] {
		ts.Frontend = append(ts.Frontend, "Tailwind CSS")
	}

	// Backend
	if fileSet["go.mod"] || fileSet["main.go"] {
		ts.Backend = append(ts.Backend, "Go")
	}
	if fileSet["package.json"] && !containsStr(ts.Frontend, "React") && !containsStr(ts.Frontend, "Next.js") {
		ts.Backend = append(ts.Backend, "Node.js")
	}
	if allDeps["express"] {
		ts.Backend = append(ts.Backend, "Express.js")
	}
	if allDeps["fastapi"] || allDeps["uvicorn"] {
		ts.Backend = append(ts.Backend, "FastAPI")
	}
	if allDeps["flask"] {
		ts.Backend = append(ts.Backend, "Flask")
	}
	if allDeps["django"] {
		ts.Backend = append(ts.Backend, "Django")
	}
	if fileSet["pom.xml"] || fileSet["build.gradle"] {
		ts.Backend = append(ts.Backend, "Spring Boot")
	}

	// Databases
	if allDeps["pg"] || allDeps["postgres"] || allDeps["psycopg2"] || allDeps["psycopg2-binary"] || allDeps["github.com/lib/pq"] || allDeps["gorm.io/driver/postgres"] {
		ts.Databases = append(ts.Databases, "PostgreSQL")
	}
	if allDeps["mysql2"] || allDeps["mysql"] || allDeps["gorm.io/driver/mysql"] {
		ts.Databases = append(ts.Databases, "MySQL")
	}
	if allDeps["mongoose"] || allDeps["pymongo"] || allDeps["motor"] {
		ts.Databases = append(ts.Databases, "MongoDB")
	}
	if allDeps["redis"] || allDeps["ioredis"] || allDeps["go-redis/redis"] || allDeps["github.com/go-redis/redis/v8"] {
		ts.Databases = append(ts.Databases, "Redis")
	}
	if allDeps["better-sqlite3"] || allDeps["sqlite3"] || allDeps["gorm.io/driver/sqlite"] || fileSet["*.db"] {
		ts.Databases = append(ts.Databases, "SQLite")
	}

	// DevOps
	if fileSet["dockerfile"] || fileSet["dockerfile.dev"] {
		ts.DevOps = append(ts.DevOps, "Docker")
	}
	if fileSet["docker-compose.yml"] || fileSet["docker-compose.yaml"] {
		ts.DevOps = append(ts.DevOps, "Docker Compose")
	}
	if fileSet[".github/workflows"] || fileSet["github/workflows"] {
		ts.DevOps = append(ts.DevOps, "GitHub Actions")
	}
	if fileSet["k8s"] || fileSet["kubernetes"] || fileSet["helm"] {
		ts.DevOps = append(ts.DevOps, "Kubernetes")
	}
	if fileSet["terraform"] || fileSet["main.tf"] {
		ts.DevOps = append(ts.DevOps, "Terraform")
	}

	// Cloud
	if fileSet[".ebextensions"] || allDeps["aws-sdk"] || allDeps["@aws-sdk/client-s3"] {
		ts.Cloud = append(ts.Cloud, "AWS")
	}
	if fileSet["app.yaml"] || fileSet[".gcloudignore"] {
		ts.Cloud = append(ts.Cloud, "GCP")
	}
	if fileSet["azure-pipelines.yml"] {
		ts.Cloud = append(ts.Cloud, "Azure")
	}
	if fileSet["vercel.json"] || fileSet[".vercel"] {
		ts.Cloud = append(ts.Cloud, "Vercel")
	}
	if fileSet["netlify.toml"] {
		ts.Cloud = append(ts.Cloud, "Netlify")
	}
	if fileSet["render.yaml"] {
		ts.Cloud = append(ts.Cloud, "Render")
	}
	if fileSet["fly.toml"] {
		ts.Cloud = append(ts.Cloud, "Fly.io")
	}
	if fileSet["railway.json"] || fileSet["railway.toml"] {
		ts.Cloud = append(ts.Cloud, "Railway")
	}

	// Testing
	if allDeps["jest"] || allDeps["vitest"] {
		ts.Testing = append(ts.Testing, "Jest/Vitest")
	}
	if allDeps["pytest"] || allDeps["unittest"] {
		ts.Testing = append(ts.Testing, "pytest")
	}
	if allDeps["cypress"] || allDeps["playwright"] {
		ts.Testing = append(ts.Testing, "E2E Testing")
	}
	if fileSet["_test.go"] {
		ts.Testing = append(ts.Testing, "Go testing")
	}

	// Package managers
	if fileSet["package-lock.json"] {
		ts.PackageManager = "npm"
	} else if fileSet["yarn.lock"] {
		ts.PackageManager = "Yarn"
	} else if fileSet["pnpm-lock.yaml"] {
		ts.PackageManager = "pnpm"
	} else if fileSet["go.mod"] {
		ts.PackageManager = "Go modules"
	} else if fileSet["requirements.txt"] {
		ts.PackageManager = "pip"
	} else if fileSet["gemfile"] {
		ts.PackageManager = "Bundler"
	}

	return ts
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
