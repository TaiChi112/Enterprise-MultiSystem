package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/services/idp-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/idp-api/internal/service"
)

func main() {
	port := os.Getenv("IDP_PORT")
	if port == "" {
		port = "4011"
	}

	svc := service.NewService()
	h := handler.NewHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "IDP API - Intelligent Document Processing",
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

	log.Printf("Server starting on :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func init() {
	fmt.Println(`
+---------------------------------------------------------------+
|                   IDP API Microservice                        |
|     Intelligent Document Processing (AI Simulation)           |
+---------------------------------------------------------------+
	`)
}
