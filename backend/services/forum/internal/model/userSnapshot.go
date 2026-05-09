package model

import "time"

// UserSnapshot stores a redundant copy of user info in forum DB to
// decouple read paths from user-service availability.
type UserSnapshot struct {
	UserID    int64     `gorm:"primaryKey;autoIncrement:false" json:"user_id,string"`
	Username  string    `gorm:"size:50;not null" json:"username"`
	Role      int       `gorm:"not null" json:"role"`
	AvatarURL string    `gorm:"type:text" json:"avatar_url"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
