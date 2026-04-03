package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/user/pos-wms-mvp/pkg/config"
	"github.com/user/pos-wms-mvp/pkg/observability"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/repository"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/service"
)

func main() {
	// Load configuration
	dsn := config.GetDatabaseURL()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize database connection
	db, err := repository.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database connection established")

	// Initialize service layer
	svc := service.NewService(db)
	h := handler.NewHandler(svc)

	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:      "POS & WMS MVP API",
		ServerHeader: "Fiber",
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))
	app.Use(observability.PrometheusMiddleware())

	rateLimitMax := getEnvInt("RATE_LIMIT_MAX", 50)
	rateLimitWindow := getEnvDuration("RATE_LIMIT_WINDOW", time.Second)

	// Protect sales endpoint from request floods.
	app.Use("/api/sales", limiter.New(limiter.Config{
		Max:        rateLimitMax,
		Expiration: rateLimitWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "rate limit exceeded",
				"code":    "TOO_MANY_REQUESTS",
			})
		},
	}))
	log.Printf("✓ Rate limiter enabled for /api/sales (max=%d, window=%s)", rateLimitMax, rateLimitWindow)

	// Observability endpoint for Prometheus scraping
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/metrics/rate-limit", observability.RateLimitMetricsHandler)

	// Register routes
	h.RegisterRoutes(app)

	// Start server
	log.Printf("🚀 Server starting on :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func init() {
	// Print banner
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║          POS & WMS MVP - Point of Sale & Warehouse          ║
║                   Management System                          ║
║                                                               ║
║              Backend: Go | Database: PostgreSQL              ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
	`)
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultValue
	}
	return n
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return defaultValue
	}
	return d
}
