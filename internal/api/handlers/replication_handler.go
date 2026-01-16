package handlers

import (
	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReplicationHandler struct {
	DB *gorm.DB
}

func NewReplicationHandler(db *gorm.DB) *ReplicationHandler {
	return &ReplicationHandler{DB: db}
}

// CreateRule (POST /api/v1/admin/replication)
func (h *ReplicationHandler) CreateRule(c *fiber.Ctx) error {
	var rule model.ReplicationRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cuerpo inválido"})
	}
	rule.ID = uuid.New()
	h.DB.Create(&rule)
	return c.Status(201).JSON(rule)
}

// GetRulesByApp (GET /api/v1/admin/replication/app/:appId)
func (h *ReplicationHandler) GetRulesByApp(c *fiber.Ctx) error {
	var rules []model.ReplicationRule
	h.DB.Where("app_id = ?", c.Params("appId")).Find(&rules)
	return c.JSON(rules)
}

// DeleteRule (DELETE /api/v1/admin/replication/:id)
func (h *ReplicationHandler) DeleteRule(c *fiber.Ctx) error {
	h.DB.Delete(&model.ReplicationRule{}, "id = ?", c.Params("id"))
	return c.SendStatus(204)
}
