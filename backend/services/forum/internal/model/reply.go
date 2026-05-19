package model

import (
	"time"

	"github.com/lib/pq"
)

type ReplyStatus int

const (
	NotSelected      ReplyStatus = iota + 1
	AuthorSelected               // 作者选定的答案
	TeacherCertified             // 教师精选的答案
)

// pg表
type Reply struct {
	ID            int64          `gorm:"primaryKey;autoIncrement:false" json:"reply_id,string"` // reply_id
	PostID        int64          `gorm:"not null" json:"post_id,string"`
	ReplierID     int64          `gorm:"not null" json:"-"`
	Content       string         `gorm:"type:text" json:"content"`
	Status        ReplyStatus    `gorm:"not null" json:"status"`
	Likes         uint           `json:"likes"`
	ImageURLs     pq.StringArray `gorm:"type:varchar(128)[]" json:"image_urls"`
	VoiceURL      string         `json:"voice_url"`
	VoiceText     string         `json:"voice_text"`
	AIAnswered    bool           `json:"ai_answered"`
	ParentReplyID *int64         `json:"parent_reply_id,string"`
	CertifiedBy   *int64         `json:"certified_by,string"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"-"`
	Post          Post           `gorm:"constraint:OnDelete:CASCADE;" json:"-"`                                         // 外键约束:级联删除
	ParentReply   *Reply         `gorm:"foreignKey:ParentReplyID;references:ID;constraint:OnDelete:SET NULL;" json:"-"` // 外键约束:级联置空
}

// 回帖点赞表
type ReplyLike struct {
	UserID    int64 `gorm:"primaryKey"`
	ReplyID   int64 `gorm:"primaryKey"`
	CreatedAt time.Time
	Reply     Reply `gorm:"foreignKey:ReplyID;references:ID;OnDelete:CASCADE;"`
}

// 存储reply数据及当前用户like状态
type ReplyDetail struct {
	Reply
	Liked bool
}

type ReplyStat struct {
	ReplyID int64
	Likes   int64
}
