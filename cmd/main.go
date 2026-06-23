package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"

	"readmeai/internal/handlers"
	"readmeai/internal/models"
	"readmeai/internal/services"
)

func main() {
	// Load .env file if present (development)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Warn if keys are missing or check fallback configuration
	openaiKey := os.Getenv("OPENAI_API_KEY")
	groqKey := os.Getenv("GROQ_API_KEY")
	if openaiKey == "" && groqKey == "" {
		log.Println("⚠️  WARNING: Neither OPENAI_API_KEY nor GROQ_API_KEY is set — README generation will fail")
	} else if openaiKey == "" && groqKey != "" {
		log.Println("ℹ️  INFO: OPENAI_API_KEY is not set — README generation will use Groq API exclusively")
	} else if openaiKey != "" && groqKey == "" {
		log.Println("ℹ️  INFO: GROQ_API_KEY is not set — OpenAI will be used without Groq fallback")
	} else {
		log.Println("✅ INFO: OpenAI active with Groq API fallback enabled")
	}

	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Println("⚠️  WARNING: GITHUB_TOKEN is not set — GitHub API rate limit is 60 req/hr")
	}

	// Create service and handlers
	svc := services.NewReadmeService()
	generateHandler := handlers.NewGenerateHandler(svc)
	repoHandler := handlers.NewRepositoryHandler(svc)

	// Setup Gin
	gin.SetMode(os.Getenv("GIN_MODE"))
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Static files
	r.Static("/static", "./public/static")
	r.StaticFile("/", "./public/index.html")

	// Rate limiter: 15 requests per minute per IP
	limiter := newIPRateLimiter(rate.Every(time.Minute/15), 5)



	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "ReadMeAI",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := r.Group("/api")
	api.Use(rateLimitMiddleware(limiter))
	{
		api.POST("/generate", generateHandler.Handle)
		api.GET("/repository", repoHandler.Handle)
	}

	log.Printf("🚀 ReadMeAI server starting on http://localhost:%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware adds CORS headers for development flexibility.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ipRateLimiter stores per-IP rate limiters.
type ipRateLimiter struct {
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (i *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	limiter, ok := i.limiters[ip]
	if !ok {
		limiter = rate.NewLimiter(i.r, i.b)
		i.limiters[ip] = limiter
	}
	return limiter
}

// rateLimitMiddleware enforces per-IP rate limiting.
func rateLimitMiddleware(limiter *ipRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		l := limiter.getLimiter(ip)
		if !l.Allow() {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Rate limit exceeded",
				Details: "You can generate up to 15 READMEs per minute. Please wait a moment.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
