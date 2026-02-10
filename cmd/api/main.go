package main

import (
	"log"
	"os"

	"github.com/JAreyes98/healthconnect-storage-service/config"
	"github.com/JAreyes98/healthconnect-storage-service/internal/api/routes"
	"github.com/JAreyes98/healthconnect-storage-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Simplemente llama a InitDB sin pasarle nada
	db := config.InitDB()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	origins := os.Getenv("FRONTEND_ORIGINS")

	auditSvc, err := service.NewAuditService(rabbitURL)
	if err != nil {
		log.Fatalf("Critical: Could not connect to RabbitMQ: %v", err)
	}
	defer auditSvc.Close() // Ahora sí existe

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-API-Key, X-API-Secret",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	// Configurar rutas, etc.
	routes.SetupRoutes(app, db, auditSvc)

	log.Fatal(app.Listen(":8082"))
}
