package models

import "time"

type Notification struct {
	ID uint `gorm:"primaryKey"`

	UserID   *uint `gorm:"index"` // NULL = global
	IsGlobal bool  `gorm:"default:false"`

	Title string
	Body  string

	Type string `gorm:"index"`     // system, user, chat
	Data string `gorm:"type:json"` // JSON payload

	IsRead bool `gorm:"default:false"`

	CreatedAt time.Time
}
