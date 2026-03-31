package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/services/ecm-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/ecm-api/internal/service"
)

func main() {
	port := os.Getenv("ECM_PORT")
	if port == "" {
		port = "4010"
	}

	uploadDir := os.Getenv("ECM_UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	svc := service.NewService(uploadDir)
	h := handler.NewHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "ECM API - Enterprise Content Management",
		ServerHeader: "Fiber",
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	h.RegisterRoutes(app)

	log.Printf("Server starting on :%s (upload dir: %s)\n", port, uploadDir)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func init() {
	fmt.Println(`
+---------------------------------------------------------------+
|                   ECM API Microservice                        |
|          Enterprise Content Management (Upload)               |
+---------------------------------------------------------------+
	`)
}
