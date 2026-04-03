package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/pkg/config"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/repository"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/service"
)

func main() {
	dsn := config.GetDatabaseURL()
	port := os.Getenv("SCM_PORT")
	if port == "" {
		port = "4004"
	}

	db, err := repository.NewDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database connection established")

	svc := service.NewService(db)
	h := handler.NewHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "SCM API - Supply Chain Management",
		ServerHeader: "Fiber",
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	h.RegisterRoutes(app)

	log.Printf("🚀 Server starting on :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func init() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║         SCM API - Supply Chain Management                     ║
║                   Microservice                                ║
║                                                               ║
║              Backend: Go | Database: PostgreSQL              ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
	`)
}
