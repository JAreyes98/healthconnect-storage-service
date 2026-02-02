package main

import (
	"log"
	"os"

	"github.com/JAreyes98/healthconnect-storage-service/config"
	"github.com/JAreyes98/healthconnect-storage-service/internal/api/routes"
	"github.com/JAreyes98/healthconnect-storage-service/internal/service"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Simplemente llama a InitDB sin pasarle nada
	db := config.InitDB()

	rabbitURL := os.Getenv("RABBITMQ_URL")

	auditSvc, err := service.NewAuditService(rabbitURL)
	if err != nil {
		log.Fatalf("Critical: Could not connect to RabbitMQ: %v", err)
	}
	defer auditSvc.Close() // Ahora sí existe

	app := fiber.New()

	// Configurar rutas, etc.
	routes.SetupRoutes(app, db, auditSvc)

	log.Fatal(app.Listen(":8082"))
}
