package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/config"
	"github.com/user/pos-wms-mvp/internal/handler"
	"github.com/user/pos-wms-mvp/internal/repository"
	"github.com/user/pos-wms-mvp/internal/service"
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
