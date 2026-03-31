package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/services/erp-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/erp-api/internal/service"
)

func main() {
	port := os.Getenv("ERP_PORT")
	if port == "" {
		port = "4007"
	}

	// Initialize service layer (ERP does not need a database repository)
	svc := service.NewService()
	h := handler.NewHandler(svc)

	// Create Fiber app with configuration
	app := fiber.New(fiber.Config{
		AppName:      "ERP API - Enterprise Resource Planning",
		ServerHeader: "Fiber",
	})

	// Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

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
║         ERP API - Enterprise Resource Planning               ║
║            Financial Aggregator Service                       ║
║                                                               ║
║              Backend: Go | Architecture: Aggregator           ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
	`)
}
