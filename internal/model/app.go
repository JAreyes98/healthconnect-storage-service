// internal/model/app.go
package model

import "github.com/google/uuid"

type App struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	AppName   string    `gorm:"unique;not null"` // 'patient-service', 'billing-service'
	ApiKey    string    `gorm:"unique;not null"`
	ApiSecret string    `gorm:"not null"` // Almacenado de forma segura
	IsActive  bool      `gorm:"default:true"`
	Buckets   []Bucket  `gorm:"foreignKey:AppID"`
}
