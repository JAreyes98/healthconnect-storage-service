package middleware

import (
	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// internal/api/middleware/auth.go
func StorageAuth(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		apiSecret := c.Get("X-API-Secret")

		var app model.App
		if err := db.Where("api_key = ? AND api_secret = ?", apiKey, apiSecret).First(&app).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}

		c.Locals("app_id", app.ID) // Guardamos el ID de la app para filtrar queries
		return c.Next()
	}
}
