package models

import "time"

type OTP struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	UserID    int       `gorm:"column:user_id"`
	OTPCode   string    `gorm:"column:otp_code"`
	ExpiredAt time.Time `gorm:"column:expired_at"`
	IsUsed    int       `gorm:"column:is_used"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (OTP) TableName() string {
	return "tb_otps"
}
