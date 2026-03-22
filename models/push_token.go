package models

import "time"

type PushToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	Token     string    `gorm:"unique"`
	Device    string    `gorm:"type:varchar(20)"`
	CreatedAt time.Time `json:"created_at"`
}
