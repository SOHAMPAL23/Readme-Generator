package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"readmeai/core/handlers"
	"readmeai/core/models"
	"readmeai/core/services"
)

var app *gin.Engine

func init() {
	// Initialize Gin
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	// Add CORS middleware
	app.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Add Rate Limiter Middleware
	limiter := newIPRateLimiter(rate.Every(time.Minute/15), 5)
	
	// Create services and handlers
	svc := services.NewReadmeService()
	generateHandler := handlers.NewGenerateHandler(svc)
	repoHandler := handlers.NewRepositoryHandler(svc)

	// API Routes
	apiGroup := app.Group("/api")
	apiGroup.Use(rateLimitMiddleware(limiter))
	{
		apiGroup.POST("/generate", generateHandler.Handle)
		apiGroup.GET("/repository", repoHandler.Handle)
	}

	// Health Check
	app.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "ReadMeAI-Vercel",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	
	// Print a log for initialization (Vercel logs this on cold start)
	log.Println("✅ ReadMeAI Serverless Function Initialized")
	if os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("GROQ_API_KEY") == "" {
		log.Println("⚠️  WARNING: API keys are not set!")
	}
}

// Handler is the Vercel Serverless Function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}

// Rate Limiter logic
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
