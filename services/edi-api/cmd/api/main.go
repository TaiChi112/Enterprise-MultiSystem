package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/user/pos-wms-mvp/services/edi-api/internal/handler"
	"github.com/user/pos-wms-mvp/services/edi-api/internal/service"
)

func main() {
	port := os.Getenv("EDI_PORT")
	if port == "" {
		port = "4005"
	}

	svc := service.NewService()
	h := handler.NewHandler(svc)

	app := fiber.New(fiber.Config{
		AppName:      "EDI API - Electronic Data Interchange Simulator",
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

	log.Printf("🚀 Server starting on :%s\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func init() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║       EDI API - Electronic Data Interchange Simulator         ║
║                      Microservice                             ║
║                                                               ║
║              Backend: Go | Stateless Service                 ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
	`)
}
