package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name      string `json:"name" validate:"required,valid-name"`
	Email     string `json:"email" validate:"required,email,valid-email" gorm:"unique"`
	Password  string `json:"password" validate:"required,min=6"`
	ExpiresAt time.Time
	CreatedAt time.Time
}
