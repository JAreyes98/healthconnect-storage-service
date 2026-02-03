package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/JAreyes98/healthconnect-storage-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB    *gorm.DB
	Audit *service.AuditService
}

func NewAdminHandler(db *gorm.DB, audit *service.AuditService) *AdminHandler {
	return &AdminHandler{
		DB:    db,
		Audit: audit,
	}
}

// --- APP CRUD ---

func (h *AdminHandler) CreateApp(c *fiber.Ctx) error {
	var app model.App
	if err := c.BodyParser(&app); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":     "Error de parseo: " + err.Error(),
			"sent_data": string(c.Body()),
		})
	}
	app.ID = uuid.New()
	app.ApiKey = generateToken(16)
	app.ApiSecret = generateToken(32)

	h.DB.Create(&app)
	h.Audit.LogEvent("ADMIN_APP_CREATE", fmt.Sprintf("New App created: %s (ID: %s)", app.AppName, app.ID), "INFO")
	return c.Status(201).JSON(app)
}

func (h *AdminHandler) GetAllApps(c *fiber.Ctx) error {
	var apps []model.App

	if err := h.DB.Preload("Buckets").Find(&apps).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch apps"})
	}

	for i := range apps {
		for j := range apps[i].Buckets {
			var total int64

			h.DB.Model(&model.FileMetadata{}).
				Where("bucket_id = ?", apps[i].Buckets[j].ID).
				Select("COALESCE(SUM(file_size), 0)").
				Scan(&total)

			apps[i].Buckets[j].TotalSize = total
		}
	}

	h.Audit.LogEvent("ADMIN_APP_ALL", "Fetched all apps with storage metrics", "INFO")
	return c.JSON(apps)
}

func (h *AdminHandler) UpdateApp(c *fiber.Ctx) error {
	id := c.Params("id")
	var app model.App
	if err := h.DB.First(&app, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App not found"})
	}
	c.BodyParser(&app)
	h.DB.Save(&app)

	h.Audit.LogEvent("ADMIN_APP_UPDATE", fmt.Sprintf("New App created: %s (ID: %s)", app.AppName, app.ID), "INFO")

	return c.JSON(app)
}

func (h *AdminHandler) DeleteApp(c *fiber.Ctx) error {
	h.DB.Delete(&model.App{}, "id = ?", c.Params("id"))

	h.Audit.LogEvent("ADMIN_APP_DELETE", fmt.Sprintf("App deleted ID: %s", c.Params("id")), "INFO")

	return c.SendStatus(204)
}

// --- BUCKET CRUD ---

// RegisterBucket (POST /api/v1/admin/buckets)
func (h *AdminHandler) RegisterBucket(c *fiber.Ctx) error {
	var bucket model.Bucket
	if err := c.BodyParser(&bucket); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cuerpo inválido"})
	}

	bucket.ID = uuid.New()
	if err := h.DB.Create(&bucket).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "No se pudo crear el bucket: " + err.Error()})
	}

	h.Audit.LogEvent("ADMIN_BUCKET_REGISTER", fmt.Sprintf("Bucket %s registered for App %s", bucket.Name, bucket.AppID), "INFO")

	return c.Status(201).JSON(bucket)
}

// GetBucketsByApp (GET /api/v1/admin/buckets/app/:appId)
func (h *AdminHandler) GetAllBuckets(c *fiber.Ctx) error {
	var buckets []model.Bucket
	h.DB.Find(&buckets)

	h.DB.Preload("App").Find(&buckets)
	for i := range buckets {
		var total int64
		h.DB.Model(&model.FileMetadata{}).
			Where("bucket_id = ?", buckets[i].ID).
			Select("COALESCE(SUM(file_size), 0)").
			Scan(&total)

		buckets[i].TotalSize = total
	}

	return c.JSON(buckets)
}

func (h *AdminHandler) GetBucketById(c *fiber.Ctx) error {
	bucketId := c.Params("id")
	var bucket model.Bucket
	var queryErr error
	queryErr = h.DB.Where("id = ?", bucketId).Preload("App").Find(&bucket).Error

	if queryErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": queryErr.Error()})
	}

	return c.JSON(bucket)
}

func (h *AdminHandler) GetBucketsByApp(c *fiber.Ctx) error {
	appId := c.Params("appId")
	var buckets []model.Bucket
	var queryErr error

	if appId == "all" {
		queryErr = h.DB.Find(&buckets).Error
	} else {
		queryErr = h.DB.Where("app_id = ?", appId).Find(&buckets).Error
	}

	if queryErr != nil {
		return c.Status(500).JSON(fiber.Map{"error": queryErr.Error()})
	}

	h.Audit.LogEvent("ADMIN_BUCKET_SEARCH", fmt.Sprintf("Searching for: %s", appId), "INFO")

	return c.JSON(buckets)
}

// --- UTILS ---

func generateToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
