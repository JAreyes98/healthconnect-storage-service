package handlers

import (
	"bytes"
	"fmt"
	"io"
	"mime"
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

	// 1. Get Source Bucket
	var sourceBucket model.Bucket
	if err := h.DB.Where("app_id = ? AND name = ?", appID, bucketName).First(&sourceBucket).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Bucket not found or access denied"})
	}

	// 2. Read file content once to allow multiple uploads (replication)
	var fileData []byte
	var err error
	stream := c.Context().RequestBodyStream()
	if stream != nil {
		fileData, err = io.ReadAll(stream)
	} else {
		fileData = c.Body()
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read upload stream"})
	}

	// 3. Prepare File IDs
	fileID := uuid.New()
	ext := filepath.Ext(originalName)
	physicalName := fileID.String() + ext

	// 4. Find Replication Rules
	var rules []model.ReplicationRule
	h.DB.Preload("TargetBucket").Where("source_bucket_id = ?", sourceBucket.ID).Find(&rules)

	// 5. Upload to Primary Bucket and Replicas
	// We'll collect all target buckets (Source + Replicas)
	type uploadTarget struct {
		Bucket    model.Bucket
		IsPrimary bool
	}

	targets := []uploadTarget{{Bucket: sourceBucket, IsPrimary: true}}
	for _, r := range rules {
		targets = append(targets, uploadTarget{Bucket: r.TargetBucket, IsPrimary: false})
	}

	for _, target := range targets {
		strat, ok := storage.GetStrategy(target.Bucket.ProviderType)

		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "Invalid storage provider"})
		}
		// Note: each target might have its own Cipher setting
		path, err := strat.Upload(bytes.NewReader(fileData), physicalName, target.Bucket.Config, target.Bucket.Cipher)

		if err != nil {
			fmt.Sprintf("Failed to upload to bucket %s: %v", target.Bucket.Name, err.Error())
			h.Audit.LogEvent("REPLICATION_ERROR",
				fmt.Sprintf("Failed to upload to bucket %s: %v", target.Bucket.Name, err), "ERROR")
			if target.IsPrimary {
				return c.Status(500).JSON(fiber.Map{"error": "Primary upload failed", "details": err.Error()})
			}
			continue // Skip failed replica, but keep going
		}
		fmt.Sprintf("Generating filemetadata on bucket %s", target.Bucket.Name)

		// 6. Save Metadata for each successful upload
		fileMeta := model.FileMetadata{
			ID:           uuid.New(), // Each entry gets its own ID if you want to track them separately
			AppID:        appID,
			BucketID:     target.Bucket.ID,
			OriginalName: originalName,
			PhysicalPath: path,
			FileSize:     int64(len(fileData)),
		}
		h.DB.Create(&fileMeta)
	}

	h.Audit.LogEvent("FILE_UPLOAD", fmt.Sprintf("File %s processed. Replicas created: %d", originalName, len(rules)), "INFO")

	return c.Status(201).JSON(fiber.Map{
		"message":            "Upload successful",
		"file_id":            fileID,
		"replicas_processed": len(targets),
	})
}

func (h *StorageHandler) ViewFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	var file model.FileMetadata

	if err := h.DB.Preload("Bucket").First(&file, "id = ?", fileID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	strategy, ok := storage.GetStrategy(file.Bucket.ProviderType)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid storage provider"})
	}

	reader, err := strategy.Download(file.PhysicalPath, file.Bucket.Config)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer reader.Close()

	// Read the entire content to avoid socket hang up during streaming
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error reading file content"})
	}

	contentType := mime.TypeByExtension(filepath.Ext(file.OriginalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Set("Content-Disposition", "inline; filename=\""+file.OriginalName+"\"")
	c.Set("Content-Type", contentType)

	return c.Send(fileBytes)
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

func (h *StorageHandler) DownloadFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	appID := c.Locals("app_id").(uuid.UUID)

	h.Audit.LogEvent("FILE_DOWNLOAD", fmt.Sprintf("Downloading file ID: %s", fileID), "INFO")

	var meta model.FileMetadata
	if err := h.DB.Preload("Bucket").Where("id = ? AND app_id = ?", fileID, appID).First(&meta).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "File not found or access denied"})
	}

	strategy, ok := storage.GetStrategy(meta.Bucket.ProviderType)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid storage provider"})
	}

	reader, err := strategy.Download(meta.PhysicalPath, meta.Bucket.Config)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not retrieve file from storage"})
	}
	defer reader.Close()

	contentType := mime.TypeByExtension(filepath.Ext(meta.OriginalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", meta.OriginalName))
	c.Set("Content-Type", contentType)

	if meta.Bucket.Cipher {
		encryptedData, err := io.ReadAll(reader)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Error reading encrypted stream"})
		}

		decryptedData, err := crypto.Decrypt(encryptedData)
		if err != nil {
			h.Audit.LogEvent("DECRYPTION_FAILED", fmt.Sprintf("Critical: Failed to decrypt file %s", fileID), "ERROR")
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decrypt file"})
		}

		return c.Send(decryptedData)
	}

	return c.SendStream(reader)
}

func (h *AdminHandler) GetBucketFiles(c *fiber.Ctx) error {
	bucketID := c.Params("id")
	var files []model.FileMetadata
	var bucket model.Bucket

	if err := h.DB.Where("bucket_id = ?", bucketID).Find(&files).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not fetch files"})
	}

	if err := h.DB.Where("id = ?", bucketID).Find(&bucket).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not find bucket"})
	}

	for i := range files {
		var total int64
		h.DB.Model(&model.FileMetadata{}).
			Where("bucket_id = ?", files[i].ID).
			Select("COALESCE(SUM(file_size), 0)").
			Scan(&total)

		files[i].IsCiphered = bucket.Cipher
	}

	return c.JSON(files)
}
