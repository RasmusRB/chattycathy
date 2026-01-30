package models

import (
	"time"
)

// Ping represents a ping record in the database
type Ping struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Message   string    `gorm:"type:varchar(255);not null" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Ping) TableName() string {
	return "pings"
}
