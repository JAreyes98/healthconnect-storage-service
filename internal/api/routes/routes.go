package routes

import (
	"github.com/JAreyes98/healthconnect-storage-service/internal/api/handlers"
	"github.com/JAreyes98/healthconnect-storage-service/internal/api/middleware"
	"github.com/JAreyes98/healthconnect-storage-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB, auditSvc *service.AuditService) {
	admin := handlers.NewAdminHandler(db, auditSvc)
	replicate := handlers.NewReplicationHandler(db)
	storageHandler := handlers.NewStorageHandler(db, auditSvc)

	// 1. Definimos el grupo base VERSIONADO
	v1 := app.Group("/api/v1/storage")

	// 2. Definimos el grupo ADMIN (hijo de v1) -> /api/v1/admin
	adminGroup := v1.Group("/admin")

	// Apps (Rutas finales: /api/v1/admin/apps...)
	adminGroup.Post("/apps", admin.CreateApp)
	adminGroup.Get("/apps", admin.GetAllApps)
	adminGroup.Put("/apps/:id", admin.UpdateApp)
	adminGroup.Delete("/apps/:id", admin.DeleteApp)

	// Buckets
	adminGroup.Get("/buckets", admin.GetAllBuckets)
	adminGroup.Post("/buckets", admin.RegisterBucket)
	adminGroup.Get("/buckets/app/:appId", admin.GetBucketsByApp)

	// Replication
	adminGroup.Post("/replication", replicate.CreateRule)
	adminGroup.Get("/replication/app/:appId", replicate.GetRulesByApp)
	adminGroup.Delete("/replication/:id", replicate.DeleteRule)

	// 3. Definimos el grupo STORAGE (hijo de v1) -> /api/v1/storage
	// NOTA: Aquí usamos 'v1.Group', NO 'adminGroup.Group'
	storageGroup := v1.Group("/storage", middleware.StorageAuth(db))
	storageGroup.Post("/upload", storageHandler.UploadFile)
	storageGroup.Get("/download/:id", storageHandler.DownloadFile)
}
