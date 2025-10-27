package model

import "time"

type Role int

const (
	Student   Role = iota + 1 // 1
	Assistant                 // 2
	Teacher                   // 3
	Admin                     // 4
)

// 获取role的字符串名称
func GetRoleName(role Role) string {
	switch role {
	case Student:
		return "student"
	case Assistant:
		return "assitant"
	case Teacher:
		return "teacher"
	case Admin:
		return "admin"
	default:
		return ""
	}
}

// User 表示用户信息
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Username     string    `gorm:"size:50;unique;not null" json:"username"`
	Email        string    `gorm:"size:100;unique;not null" json:"email"`
	PasswordHash string    `gorm:"type:text;not null" json:"-"`
	Role         Role      `gorm:"not null" json:"role"`
	AvatarUrl    string    `gorm:"type:text" json:"avatar_url"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"-"`
}
