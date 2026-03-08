package models

import "time"

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"unique"`
	TokenID   string `gorm:"unique"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
