package models

import "time"

type FileAccsess struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	FileID      int       `gorm:"column:file_id"`
	OwnerID     int       `gorm:"column:owner_id"`
	TargetEmail string    `gorm:"column:target_email"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (FileAccsess) TableName() string {
	return "tb_file_access"
}
