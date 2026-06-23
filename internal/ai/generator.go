package ai

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"readmeai/internal/models"
)

const openAIURL = "https://api.openai.com/v1/chat/completions"
const groqURL = "https://api.groq.com/openai/v1/chat/completions"

// Generator calls OpenAI and fallback AI services to generate README content.
type Generator struct {
	client      *resty.Client
	model       string
	groqClient  *resty.Client
	groqModel   string
	hasGroq     bool
}

// NewGenerator creates a new generator with OpenAI and Groq clients, including retry logic.
func NewGenerator() *Generator {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	r := resty.New().
		SetAuthToken(openaiKey).
		SetHeader("Content-Type", "application/json").
		SetTimeout(120 * time.Second).
		// Retry up to 3 times on 429 (rate limit) with exponential backoff
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(30 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r != nil && r.StatusCode() == http.StatusTooManyRequests
		})

	groqKey := os.Getenv("GROQ_API_KEY")
	groqModel := os.Getenv("GROQ_MODEL")
	if groqModel == "" {
		groqModel = "llama-3.3-70b-versatile"
	}

	var groqClient *resty.Client
	hasGroq := false

	if groqKey != "" {
		hasGroq = true
		groqClient = resty.New().
			SetAuthToken(groqKey).
			SetHeader("Content-Type", "application/json").
			SetTimeout(120 * time.Second).
			// Retry up to 3 times on 429 (rate limit) with exponential backoff
			SetRetryCount(3).
			SetRetryWaitTime(5 * time.Second).
			SetRetryMaxWaitTime(30 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return r != nil && r.StatusCode() == http.StatusTooManyRequests
			})
	}

	return &Generator{
		client:     r,
		model:      model,
		groqClient: groqClient,
		groqModel:  groqModel,
		hasGroq:    hasGroq,
	}
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Generate calls OpenAI first, and if it fails or is not configured, falls back to Groq API.
func (g *Generator) Generate(analysis *models.AnalysisResult, style models.ReadmeStyle) (string, error) {
	systemPrompt, userPrompt := BuildPrompt(analysis, style)

	var openAIError error
	openaiKey := os.Getenv("OPENAI_API_KEY")

	if openaiKey != "" {
		log.Printf("Attempting README generation using OpenAI (%s)...", g.model)
		readme, err := g.callLLM(g.client, openAIURL, g.model, systemPrompt, userPrompt, "OpenAI")
		if err == nil {
			return readme, nil
		}
		openAIError = err
		log.Printf("OpenAI generation failed: %v", err)
	} else {
		openAIError = fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Fallback to Groq if configured
	if g.hasGroq {
		log.Printf("Falling back to Groq API (%s)...", g.groqModel)
		readme, err := g.callLLM(g.groqClient, groqURL, g.groqModel, systemPrompt, userPrompt, "Groq")
		if err == nil {
			log.Println("Successfully generated README using Groq fallback!")
			return readme, nil
		}
		return "", fmt.Errorf("both OpenAI and Groq generation failed. OpenAI error: %v; Groq error: %v", openAIError, err)
	}

	return "", fmt.Errorf("OpenAI generation failed and no Groq fallback is configured. OpenAI error: %w", openAIError)
}

// callLLM is a helper to perform OpenAI-compatible chat completion requests.
func (g *Generator) callLLM(client *resty.Client, endpoint, model, systemPrompt, userPrompt, provider string) (string, error) {
	reqBody := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   4096,
		Temperature: 0.4,
	}

	var result openAIResponse
	resp, err := client.R().
		SetBody(reqBody).
		SetResult(&result).
		Post(endpoint)

	if err != nil {
		return "", fmt.Errorf("%s request error: %w", provider, err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("%s API error (%s): %s", provider, result.Error.Type, result.Error.Message)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return "", fmt.Errorf("invalid %s API key — check your credentials", provider)
	}
	if resp.StatusCode() == http.StatusTooManyRequests {
		return "", fmt.Errorf("%s rate limit exceeded — please try again in a moment", provider)
	}
	if resp.IsError() {
		return "", fmt.Errorf("%s API returned status %d: %s", provider, resp.StatusCode(), resp.String())
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("%s response returned no choices", provider)
	}

	readme := result.Choices[0].Message.Content
	if readme == "" {
		return "", fmt.Errorf("%s response returned an empty content body", provider)
	}

	return readme, nil
}
