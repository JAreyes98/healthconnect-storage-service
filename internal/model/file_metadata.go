package model

import (
	"time"

	"github.com/google/uuid"
)

type FileMetadata struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AppID        uuid.UUID `gorm:"type:uuid;not null" json:"appId"`
	BucketID     uuid.UUID `gorm:"type:uuid;not null" json:"bucketId"`
	OriginalName string    `gorm:"not null" json:"originalName"`
	PhysicalPath string    `gorm:"not null" json:"physicalPath"`
	FileSize     int64     `json:"fileSize"`
	ContentType  string    `json:"contentType"`
	CreatedAt    time.Time `json:"createdAt"`
	IsCiphered   bool      `gorm:"-" json:"is_ciphered"`
	Bucket       Bucket    `gorm:"foreignKey:BucketID" json:"bucket"`
}
