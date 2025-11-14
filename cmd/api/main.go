package main

import (
	"fmt"
	"incident-backend/internal/handlers"
	"incident-backend/internal/middleware"
	"incident-backend/internal/services"
	"incident-backend/internal/storage"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get configuration from environment
	config := getConfig()

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Print banner
	printBanner()

	// Initialize storage
	jsonStore := storage.NewJSONStore(config.DataDir)
	orgStore := storage.NewOrgStore(jsonStore)
	userStore := storage.NewUserStore(jsonStore)
	incidentStore := storage.NewIncidentStore(jsonStore)
	teamStore := storage.NewTeamStore(jsonStore)

	// Initialize services
	authService := services.NewAuthService(userStore, orgStore, config.JWTSecret, config.JWTExpiry)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userStore, orgStore)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
		AppName:      "Incident Management API v1.0",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CORSMiddleware(config.AllowedOrigins))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Auth routes (protected)
	authProtected := auth.Group("")
	authProtected.Use(middleware.AuthMiddleware(authService))
	authProtected.Get("/me", authHandler.GetMe)
	authProtected.Post("/refresh", authHandler.RefreshToken)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(authService))

	// Incidents routes (will be implemented in Phase 2)
	// incidents := protected.Group("/incidents")

	// Teams routes (will be implemented in Phase 4)
	// teams := protected.Group("/teams")

	// Users routes (will be implemented later)
	// users := protected.Group("/users")

	// Organization routes (will be implemented later)
	// org := protected.Group("/organization")

	// Start server
	log.Printf("🚀 Server starting on port %s", config.Port)
	log.Printf("📂 Data directory: %s", config.DataDir)
	log.Printf("🌍 Allowed origins: %s", config.AllowedOrigins)
	log.Printf("\n✨ API Documentation:")
	log.Printf("   POST   /api/v1/auth/register - Register new user")
	log.Printf("   POST   /api/v1/auth/login - Login")
	log.Printf("   GET    /api/v1/auth/me - Get current user (auth required)")
	log.Printf("   POST   /api/v1/auth/refresh - Refresh token (auth required)")
	log.Printf("   GET    /health - Health check\n")

	if err := app.Listen(":" + config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Config holds application configuration
type Config struct {
	Port                string
	JWTSecret           string
	JWTExpiry           time.Duration
	CloudflareAPIKey    string
	CloudflareAccountID string
	AllowedOrigins      string
	DataDir             string
}

// getConfig loads configuration from environment variables
func getConfig() *Config {
	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	return &Config{
		Port:                getEnv("PORT", "3000"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTExpiry:           jwtExpiry,
		CloudflareAPIKey:    getEnv("CLOUDFLARE_API_KEY", ""),
		CloudflareAccountID: getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
		AllowedOrigins:      getEnv("ALLOWED_ORIGINS", "http://localhost:3001"),
		DataDir:             getEnv("DATA_DIR", "./data"),
	}
}

// validateConfig validates required configuration
func validateConfig(config *Config) error {
	if config.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(config.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// customErrorHandler handles Fiber errors
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║        🔥 Incident Management Platform - API Server              ║
║                                                                   ║
║        Multi-Tenant • AI-Powered • Real-Time Collaboration       ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
