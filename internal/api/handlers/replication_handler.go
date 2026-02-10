package handlers

import (
	"path/filepath"

	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/JAreyes98/healthconnect-storage-service/storage"
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

func (h *ReplicationHandler) CreateRule(c *fiber.Ctx) error {
	var rule model.ReplicationRule
	if err := c.BodyParser(&rule); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Basic validation: Source and Target cannot be the same
	if rule.SourceBucketID == rule.TargetBucketID {
		return c.Status(400).JSON(fiber.Map{"error": "Source and Target buckets must be different"})
	}

	// 2. Security validation: Verify both buckets exist and belong to the same App
	var count int64
	h.DB.Model(&model.Bucket{}).Where("id IN ? AND app_id = ?",
		[]uuid.UUID{rule.SourceBucketID, rule.TargetBucketID}, rule.AppID).Count(&count)

	if count != 2 {
		return c.Status(404).JSON(fiber.Map{"error": "One or both buckets not found or do not belong to this application"})
	}

	// 3. Prevent duplicate rules
	var existing model.ReplicationRule
	err := h.DB.Where("source_bucket_id = ? AND target_bucket_id = ?",
		rule.SourceBucketID, rule.TargetBucketID).First(&existing).Error
	if err == nil {
		return c.Status(409).JSON(fiber.Map{"error": "Replication rule already exists"})
	}

	rule.ID = uuid.New()
	rule.Active = true

	if err := h.DB.Create(&rule).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create replication rule"})
	}

	return c.Status(201).JSON(rule)
}

func (h *ReplicationHandler) GetRulesByApp(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("appId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid App ID format"})
	}

	var rules []model.ReplicationRule
	// Preload TargetBucket to show provider details in the frontend
	h.DB.Preload("TargetBucket").Where("app_id = ?", appID).Find(&rules)

	return c.JSON(rules)
}

func (h *ReplicationHandler) GetRules(c *fiber.Ctx) error {
	// appID, err := uuid.Parse(c.Params("appId"))
	// if err != nil {
	// 	return c.Status(400).JSON(fiber.Map{"error": "Invalid App ID format"})
	// }

	var rules []model.ReplicationRule
	// Preload TargetBucket to show provider details in the frontend
	h.DB.Preload("TargetBucket").Preload("ReplicationOnApp").Preload("SourceBucket").Find(&rules)
	return c.JSON(rules)
}

func (h *ReplicationHandler) DeleteRule(c *fiber.Ctx) error {
	ruleID := c.Params("id")
	result := h.DB.Delete(&model.ReplicationRule{}, "id = ?", ruleID)

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Rule not found"})
	}

	return c.SendStatus(204)
}

// ToggleRule (PATCH /api/v1/admin/replication/:id/toggle)
func (h *ReplicationHandler) ToggleRule(c *fiber.Ctx) error {
	id := c.Params("id")
	var rule model.ReplicationRule

	if err := h.DB.Preload("TargetBucket").First(&rule, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Regla no encontrada"})
	}

	rule.Active = !rule.Active
	h.DB.Save(&rule)

	// Si se activa, disparamos el escaneo de archivos faltantes
	if rule.Active {
		// Tip: En producción esto debería ser una Goroutine o un Job de RabbitMQ
		go h.SyncBuckets(rule)
	}

	return c.JSON(fiber.Map{"status": "updated", "active": rule.Active})
}

// ToggleRule (PATCH /api/v1/admin/replication/:id/toggle)
func (h *ReplicationHandler) SyncBuckets(rule model.ReplicationRule) {
	// 1. Buscar archivos en el bucket origen
	var sourceFiles []model.FileMetadata
	h.DB.Where("bucket_id = ?", rule.SourceBucketID).Find(&sourceFiles)

	for _, file := range sourceFiles {
		// 2. Verificar si ya existe en el bucket destino (por OriginalName o un Hash si lo tuvieras)
		var exists int64
		h.DB.Model(&model.FileMetadata{}).
			Where("bucket_id = ? AND original_name = ?", rule.TargetBucketID, file.OriginalName).
			Count(&exists)

		if exists == 0 {
			// 3. El archivo falta en el destino -> Replicar
			h.replicateMissingFile(file, rule)
		}
	}
}

func (h *ReplicationHandler) replicateMissingFile(file model.FileMetadata, rule model.ReplicationRule) {
	// Obtenemos el bucket origen para saber cómo descargar
	var sourceBucket model.Bucket
	h.DB.First(&sourceBucket, "id = ?", rule.SourceBucketID)

	sourceStrat, ok1 := storage.GetStrategy(sourceBucket.ProviderType)
	targetStrat, ok2 := storage.GetStrategy(rule.TargetBucket.ProviderType)

	if !ok1 || !ok2 {
		return
	}
	// A. Descargar del origen
	reader, err := sourceStrat.Download(file.PhysicalPath, sourceBucket.Config)
	if err != nil {
		return
	}
	defer reader.Close()

	// B. Subir al destino (El nombre físico se mantiene para consistencia)
	newPath, err := targetStrat.Upload(reader, filepath.Base(file.PhysicalPath), rule.TargetBucket.Config, rule.TargetBucket.Cipher)
	if err != nil {
		return
	}

	// C. Registrar metadata del nuevo archivo replicado
	newMeta := model.FileMetadata{
		ID:           uuid.New(),
		AppID:        file.AppID,
		BucketID:     rule.TargetBucketID,
		OriginalName: file.OriginalName,
		PhysicalPath: newPath,
		FileSize:     file.FileSize,
	}
	h.DB.Create(&newMeta)
}
