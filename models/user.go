package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey;column:id"`
	Name       string    `gorm:"column:name"`
	Email      string    `gorm:"column:email"`
	Phone      string    `gorm:"column:phone"`
	AvatarPath string    `gorm:"column:avatar_path"`
	Password   string    `gorm:"column:password"`
	IsVerified int       `gorm:"column:is_verified"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (User) TableName() string {
	return "tb_users"
}
