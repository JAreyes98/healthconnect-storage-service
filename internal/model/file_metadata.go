package model

import (
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	AppID        uuid.UUID `gorm:"type:uuid;not null"` // Dueño del archivo
	BucketID     uuid.UUID `gorm:"type:uuid;not null"` // Dónde se guardó
	OriginalName string    `gorm:"not null"`
	PhysicalPath string    `gorm:"not null"` // Ruta en disco o Key en S3
	FileSize     int64
	ContentType  string
	CreatedAt    time.Time
}
