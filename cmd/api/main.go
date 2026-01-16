package main

import (
	"log"

	"github.com/JAreyes98/healthconnect-storage-service/config"
	"github.com/JAreyes98/healthconnect-storage-service/internal/api/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Simplemente llama a InitDB sin pasarle nada
	db := config.InitDB()

	app := fiber.New()

	// Configurar rutas, etc.
	routes.SetupRoutes(app, db)

	log.Fatal(app.Listen(":8082"))
}
