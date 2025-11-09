package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReplyStatus int

const (
	NotSelected      ReplyStatus = iota + 1
	AuthorSelected               // 作者选定的答案
	TeacherCertified             // 教师精选的答案
)

// pg表
type Reply struct {
	ID            int64       `gorm:"primaryKey;autoIncrement:false" json:"reply_id"` // reply_id
	PostID        int64       `gorm:"not null" json:"post_id"`
	ReplierID     int64       `gorm:"not null" json:"replier_id"`
	ParentReplyID *int64      `json:"parent_reply_id"`
	DocID         string      `gorm:"type:char(24);not null" json:"-"`
	Status        ReplyStatus `gorm:"not null" json:"status"`
	CertifiedBy   *int64      `json:"certified_by"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"-"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"-"`
}

// mongo集合
type ReplyContent struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ReplyID    int64              `bson:"reply_id"`
	Content    string             `bson:"content"`
	Voice      string             `bson:"voice"` //存储语音url
	VoiceText  string             `bson:"voice_text,omitempty"`
	Images     []string           `bson:"images"`
	AIAnswered bool               `bson:"ai_answered"`
}
