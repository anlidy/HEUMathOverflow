package model

import (
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostStatus int

const (
	Unanswered PostStatus = iota + 1
	Answered
	Certified
)

// pg表
type Post struct {
	ID          int64          `gorm:"primaryKey;autoIncrement:false" json:"post_id"` // post_id
	AuthorID    int64          `gorm:"unique;not null" json:"author_id"`
	Title       string         `gorm:"size:255" json:"title"` // size可被解释为varchar(size)
	Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`
	Status      PostStatus     `gorm:"not null" json:"status"`
	DocID       string         `gorm:"type:char(24);not null" json:"-"` // objectID 为24字节
	LastReplyAt *time.Time     `json:"last_reply_at"`                   // 指针类型便于判空
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// mongo集合
type PostContent struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // omitempty表示忽略空字段
	PostID    int64              `bson:"post_id"`
	Content   string             `bson:"content"`
	Images    []string           `bson:"images"`
	CreatedAt time.Time          `bson:"created_at"`
}
