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
	ID            int64       `gorm:"primaryKey;autoIncrement:false" json:"reply_id,string"` // reply_id
	PostID        int64       `gorm:"not null" json:"post_id,string"`
	ReplierID     int64       `gorm:"not null" json:"-"`
	ParentReplyID *int64      `json:"parent_reply_id,string"`
	DocID         string      `gorm:"type:char(24);not null" json:"-"`
	Status        ReplyStatus `gorm:"not null" json:"status"`
	CertifiedBy   *int64      `json:"certified_by"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"-"`
}

// mongo集合
type ReplyContent struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	ReplyID    int64              `bson:"reply_id" json:"-"`
	Content    string             `bson:"content" json:"content"`
	VoiceURL   string             `bson:"voice_url" json:"voice_url"`
	VoiceText  string             `bson:"voice_text,omitempty" json:"voice_text"`
	ImageURLs  []string           `bson:"image_urls" json:"image_urls"`
	AIAnswered bool               `bson:"ai_answered" json:"ai_answered"`
}
