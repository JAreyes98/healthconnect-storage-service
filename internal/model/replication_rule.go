package model

import "github.com/google/uuid"

type ReplicationRule struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AppID            uuid.UUID `gorm:"type:uuid;not null" json:"appId"`
	SourceBucketID   uuid.UUID `gorm:"type:uuid;not null" json:"sourceBucketId"`
	TargetBucketID   uuid.UUID `gorm:"type:uuid;not null" json:"targetBucketId"`
	Active           bool      `gorm:"default:true" json:"active"`
	ReplicationOnApp App       `gorm:"foreignKey:AppID" json:"replicationOnApp"`
	SourceBucket     Bucket    `gorm:"foreignKey:SourceBucketID" json:"sourceBucket"`
	TargetBucket     Bucket    `gorm:"foreignKey:TargetBucketID" json:"targetBucket"`
}
