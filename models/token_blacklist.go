package models

import "time"

type TokenBlacklist struct {
	ID        uint   `gorm:"primaryKey"`
	Token     string `gorm:"unique"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
