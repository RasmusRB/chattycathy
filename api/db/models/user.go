package models

import (
	"time"
)

// User represents a user in the database
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GoogleID    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"google_id"`
	Email       string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Picture     string    `gorm:"type:varchar(512)" json:"picture"`
	Role        string    `gorm:"type:varchar(50);default:'user'" json:"role"`
	LastLoginAt time.Time `gorm:"autoUpdateTime" json:"last_login_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
