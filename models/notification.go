package models

import "time"

type Notification struct {
	ID        uint  `gorm:"primaryKey"`
	UserID    *uint `gorm:"index"`
	IsGlobal  bool  `gorm:"default:false"`
	Title     string
	Body      string
	Type      string `gorm:"index"`
	Data      string `gorm:"type:json"`
	IsRead    bool   `gorm:"default:false"`
	CreatedAt time.Time
}
