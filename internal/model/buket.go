// internal/model/bucket.go
package model

import "github.com/google/uuid"

type Bucket struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AppID        uuid.UUID `gorm:"type:uuid;not null" json:"app_id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"`
	Config       string    `json:"config"`
	IsDefault    bool      `json:"is_default"`
	Cipher       bool      `json:"cipher" gorm:"default:false"`
}
