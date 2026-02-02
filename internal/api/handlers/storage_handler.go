package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/JAreyes98/healthconnect-storage-service/internal/crypto"
	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/JAreyes98/healthconnect-storage-service/internal/service"
	"github.com/JAreyes98/healthconnect-storage-service/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StorageHandler struct {
	DB    *gorm.DB
	Audit *service.AuditService
}

func NewStorageHandler(db *gorm.DB, audit *service.AuditService) *StorageHandler {
	return &StorageHandler{
		DB:    db,
		Audit: audit,
	}
}
func (h *StorageHandler) UploadFile(c *fiber.Ctx) error {
	appID := c.Locals("app_id").(uuid.UUID)
	bucketName := c.Get("X-Bucket-Name")
	originalName := c.Get("X-Original-Filename")

	var bucket model.Bucket
	if err := h.DB.Where("app_id = ? AND name = ?", appID, bucketName).First(&bucket).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Bucket not found or access denied"})
	}

	// 1. Generate unique ID for the physical file
	fileID := uuid.New()

	// Use the UUID as the filename, preserving extension if you wish
	// or just the UUID for maximum obfuscation
	ext := filepath.Ext(originalName)
	physicalName := fileID.String() + ext

	var bodyReader io.Reader
	stream := c.Context().RequestBodyStream()
	if stream != nil {
		bodyReader = stream
	} else {
		bodyReader = bytes.NewReader(c.Body())
	}

	// 2. Pass the NEW physicalName to the strategy
	strat := storage.GetStrategy(bucket.ProviderType)
	physicalPath, err := strat.Upload(bodyReader, physicalName, bucket.Config, bucket.Cipher)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Upload failed", "details": err.Error()})
	}

	// 3. Save Metadata (Original Name vs Physical Path)
	fileMeta := model.FileMetadata{
		ID:           fileID, // Use the same generated ID
		AppID:        appID,
		BucketID:     bucket.ID,
		OriginalName: originalName, // Still saved for the user
		PhysicalPath: physicalPath, // Path contains the UUID, not the name
		FileSize:     int64(c.Request().Header.ContentLength()),
	}

	h.DB.Create(&fileMeta)
	h.Audit.LogEvent("FILE_UPLOAD", fmt.Sprintf("File %s stored with ID %s", originalName, fileID), "INFO")
	return c.Status(201).JSON(fileMeta)
}

func (h *StorageHandler) GetMetadata(c *fiber.Ctx) error {
	id := c.Params("id")
	appID := c.Locals("app_id").(uuid.UUID)

	var meta model.FileMetadata
	if err := h.DB.Where("id = ? AND app_id = ?", id, appID).First(&meta).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Metadata no encontrada"})
	}
	return c.JSON(meta)
}

// DownloadFile (GET /api/v1/storage/download/:id)
func (h *StorageHandler) DownloadFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	appID := c.Locals("app_id").(uuid.UUID)

	h.Audit.LogEvent("FILE_DOWNLOAD", fmt.Sprintf("Downloading file ID: %s", fileID), "INFO")

	// 1. Buscar metadata y validar que pertenece a la App
	var meta model.FileMetadata
	if err := h.DB.Where("id = ? AND app_id = ?", fileID, appID).First(&meta).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found or access denied"})
	}

	// 2. Obtener información del bucket para saber si está cifrado
	var bucket model.Bucket
	if err := h.DB.First(&bucket, "id = ?", meta.BucketID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error retrieving bucket info"})
	}

	// 3. Configurar el nombre original para la descarga
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", meta.OriginalName))

	// 4. Lógica de entrega (Descifrado vs Directo)
	if bucket.Cipher {
		// Leer el archivo cifrado desde el disco
		encryptedData, err := os.ReadFile(meta.PhysicalPath)
		if err != nil {
			h.Audit.LogEvent("READ_FILE_FAILED", fmt.Sprintf("Critical: Failed to read file %s", fileID), "ERROR")
			return c.Status(500).JSON(fiber.Map{"error": "Could not read file from storage"})
		}

		// Descifrar usando la llave del .env
		decryptedData, err := crypto.Decrypt(encryptedData)
		if err != nil {
			log.Printf("DECRYPTION ERROR: %v", err)
			h.Audit.LogEvent("DECRYPTION_FAILED", fmt.Sprintf("Critical: Failed to decrypt file %s", fileID), "ERROR")
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decrypt file"})
		}

		// Enviar los bytes descifrados
		return c.Send(decryptedData)
	}

	// Si no está cifrado, usar SendFile (más eficiente para archivos planos)
	return c.SendFile(meta.PhysicalPath)
}
