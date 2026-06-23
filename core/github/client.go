package github

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

const baseURL = "https://api.github.com"

// Client wraps the Resty HTTP client for GitHub API calls.
type Client struct {
	resty *resty.Client
}

// NewClient creates a new GitHub API client with optional auth token.
func NewClient() *Client {
	r := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		SetTimeout(15 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second)

	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		r.SetAuthToken(token)
	}

	return &Client{resty: r}
}

// GetRepo fetches core repository metadata.
func (c *Client) GetRepo(owner, repo string) (map[string]interface{}, error) {
	var result map[string]interface{}
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/repos/%s/%s", owner, repo))

	if err != nil {
		return nil, fmt.Errorf("github API request failed: %w", err)
	}
	if resp.StatusCode() == 404 {
		return nil, fmt.Errorf("repository %s/%s not found — make sure it's public", owner, repo)
	}
	if resp.StatusCode() == 403 || resp.StatusCode() == 429 {
		return nil, fmt.Errorf("GitHub API rate limit exceeded — set GITHUB_TOKEN to increase the limit")
	}
	if resp.IsError() {
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode(), resp.String())
	}
	return result, nil
}

// GetLanguages fetches the bytes-per-language breakdown.
func (c *Client) GetLanguages(owner, repo string) (map[string]interface{}, error) {
	var result map[string]interface{}
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/repos/%s/%s/languages", owner, repo))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch languages: %w", err)
	}
	if resp.IsError() {
		return nil, nil // Non-fatal — return empty
	}
	return result, nil
}

// GetTopics fetches repository topics.
func (c *Client) GetTopics(owner, repo string) ([]string, error) {
	var result struct {
		Names []string `json:"names"`
	}
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/repos/%s/%s/topics", owner, repo))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch topics: %w", err)
	}
	if resp.IsError() {
		return []string{}, nil // Non-fatal
	}
	return result.Names, nil
}

// GetContents fetches the file/directory listing at a given path.
func (c *Client) GetContents(owner, repo, path string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path))

	if err != nil {
		return nil, fmt.Errorf("failed to fetch contents at %q: %w", path, err)
	}
	if resp.StatusCode() == 404 {
		return []map[string]interface{}{}, nil // Path doesn't exist
	}
	if resp.IsError() {
		return nil, nil // Non-fatal for sub-paths
	}
	return result, nil
}

// GetFileContent fetches the raw encoded content of a file (≤1MB).
// The caller should use DecodeBase64Content from parser.go to decode the result.
func (c *Client) GetFileContent(owner, repo, path string) (string, error) {
	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	resp, err := c.resty.R().
		SetResult(&result).
		Get(fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path))

	if err != nil || resp.IsError() {
		return "", nil // Non-fatal
	}

	return result.Content, nil
}

// GetRawFile fetches raw file content directly.
func (c *Client) GetRawFile(owner, repo, branch, path string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, path)
	resp, err := c.resty.R().Get(url)
	if err != nil || resp.IsError() {
		return "", nil
	}
	return resp.String(), nil
}
