package models

import "time"

type File struct {
    ID           uint      `gorm:"primaryKey;column:id"`
    UserID       int       `gorm:"column:user_id"`
    FileName     string    `gorm:"column:filename"`        // FIX
    OriginalName string    `gorm:"column:original_name"`
    FilePath     string    `gorm:"column:path"`            // FIX
    Size         int64     `gorm:"column:size"`
    MimeType     string    `gorm:"column:mime_type"`
    CreatedAt    time.Time `gorm:"column:created_at"`
}

func (File) TableName() string {
    return "tb_files"
}
