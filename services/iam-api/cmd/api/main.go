package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/services/iam-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/iam-api/internal/service"
)

func main() {
	port := getEnv("IAM_PORT", "4001")
	secret := getEnv("IAM_JWT_SECRET", "dev-secret-change-me")
	issuer := getEnv("IAM_JWT_ISSUER", "iam-api")
	ttl := getEnvDuration("IAM_JWT_TTL", time.Hour)

	authService := service.NewAuthService(secret, issuer, ttl)
	h := handler.NewHandler(authService)

	app := fiber.New(fiber.Config{
		AppName:      "IAM API",
		ServerHeader: "Fiber",
	})

	app.Use(logger.New(logger.Config{Format: "[${time}] ${status} - ${method} ${path}\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	h.RegisterRoutes(app)

	log.Printf("🚀 IAM service starting on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("IAM service failed to start: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil && duration > 0 {
			return duration
		}
	}
	return fallback
}

func init() {
	fmt.Print(`
╔════════════════════════════════════╗
║            IAM API                ║
║     Authentication Service        ║
╚════════════════════════════════════╝
`)
}
