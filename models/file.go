package models

import "time"

type File struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint
	FileName     string `gorm:"column:filename"`
	OriginalName string `gorm:"column:original_name"`
	FilePath     string `gorm:"column:path"`
	Size         int64
	MimeType     string    `gorm:"column:mime_type"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}
