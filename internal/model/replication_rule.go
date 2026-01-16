package model

import "github.com/google/uuid"

type ReplicationRule struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	AppID          uuid.UUID `gorm:"type:uuid;not null"`
	SourceBucketID uuid.UUID `gorm:"type:uuid;not null"`
	TargetBucketID uuid.UUID `gorm:"type:uuid;not null"`
	Active         bool      `gorm:"default:true"`
}
