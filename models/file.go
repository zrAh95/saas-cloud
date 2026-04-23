package models

import "time"

type File struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	UserID       int       `gorm:"column:user_id"`
	ParentID     *uint     `gorm:"column:parent_id"`
	FileName     string    `gorm:"column:filename"`
	OriginalName string    `gorm:"column:original_name"`
	FilePath     string    `gorm:"column:path"`
	Size         int64     `gorm:"column:size"`
	MimeType     string    `gorm:"column:mime_type"`
	IsFolder     bool      `gorm:"column:is_folder"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (File) TableName() string {
	return "tb_files"
}
