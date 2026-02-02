// internal/model/app.go
package model

import "github.com/google/uuid"

type App struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AppName   string    `gorm:"unique;not null" json:"app_name"`
	ApiKey    string    `gorm:"unique;not null" json:"api_key"`
	ApiSecret string    `gorm:"not null" json:"api_secret"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	Buckets   []Bucket  `gorm:"foreignKey:AppID" json:"buckets"`
}
