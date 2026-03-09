package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetOTP struct {
	gorm.Model
	Email     string    `gorm:"index"`
	OTP       string    `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
	Attempts  int
	IsUsed    bool `gorm:"default:false"`
	CreatedAt time.Time
}

func (p *PasswordResetOTP) TableName() string {
	return "password_reset_otps"
}
