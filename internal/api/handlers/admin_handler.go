package handlers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{DB: db}
}

// --- APP CRUD ---

func (h *AdminHandler) CreateApp(c *fiber.Ctx) error {
	var app model.App
	if err := c.BodyParser(&app); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cuerpo inválido"})
	}
	app.ID = uuid.New()
	app.ApiKey = generateToken(16)
	app.ApiSecret = generateToken(32)

	h.DB.Create(&app)
	return c.Status(201).JSON(app)
}

func (h *AdminHandler) GetAllApps(c *fiber.Ctx) error {
	var apps []model.App
	h.DB.Preload("Buckets").Find(&apps)
	return c.JSON(apps)
}

func (h *AdminHandler) UpdateApp(c *fiber.Ctx) error {
	id := c.Params("id")
	var app model.App
	if err := h.DB.First(&app, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "App no encontrada"})
	}
	c.BodyParser(&app)
	h.DB.Save(&app)
	return c.JSON(app)
}

func (h *AdminHandler) DeleteApp(c *fiber.Ctx) error {
	h.DB.Delete(&model.App{}, "id = ?", c.Params("id"))
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
	return c.Status(201).JSON(bucket)
}

// GetBucketsByApp (GET /api/v1/admin/buckets/app/:appId)
func (h *AdminHandler) GetBucketsByApp(c *fiber.Ctx) error {
	appId := c.Params("appId")
	var buckets []model.Bucket

	if err := h.DB.Where("app_id = ?", appId).Find(&buckets).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
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
