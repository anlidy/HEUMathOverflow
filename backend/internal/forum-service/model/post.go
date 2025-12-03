package model

import (
	"time"

	"github.com/lib/pq"
)

type PostStatus int

const (
	Unanswered PostStatus = iota + 1
	Answered
	Certified
)

// pg表
type Post struct {
	ID          int64          `gorm:"primaryKey;autoIncrement:false" json:"post_id,string"` // post_id
	AuthorID    int64          `gorm:"not null" json:"-"`
	Title       string         `gorm:"size:255" json:"title"` // size可被解释为varchar(size)
	Content     string         `gorm:"type:text" json:"content"`
	ImageURLs   pq.StringArray `gorm:"type:varchar(128)[]" json:"image_urls"`
	Tags        pq.StringArray `gorm:"type:varchar(32)[]" json:"tags"`
	Status      PostStatus     `gorm:"not null" json:"status"`
	Views       uint           `json:"views"`         // 浏览量
	Likes       uint           `json:"likes"`         // 赞同数
	Stars       uint           `json:"stars"`         // 收藏数
	Replies     uint           `json:"replies"`       // 评论数
	LastReplyAt *time.Time     `json:"last_reply_at"` // 指针类型便于判空
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

// 存储用户对帖子的点赞行为
type PostLike struct {
	PostID    int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Post      Post      `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE;"` // 级联删除
}

// 存储用户对帖子的收藏行为
type PostStar struct {
	PostID    int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"primaryKey"`
	CreatedAt time.Time
	Post      Post `gorm:"foreignKey:PostID;references:ID;"` // 不级联删除,显示帖子已删除
}
