package model

import (
	"time"

	"github.com/lib/pq"
)

// 帖子状态
type AnswerStatus int

const (
	Unanswered AnswerStatus = iota + 1 // 未回答
	Answered                           // 已回答
)

// 排序类型
type OrderBy int

const (
	Recommend = iota // 推荐排序
	Hottest          // 热门排序
	Latest           // 最新排序
)

// pg表
type Post struct {
	ID          int64          `gorm:"primaryKey;autoIncrement:false" json:"post_id,string"` // post_id
	AuthorID    int64          `gorm:"not null" json:"-"`
	Title       string         `gorm:"size:255" json:"title"` // size可被解释为varchar(size)
	Content     string         `gorm:"type:text" json:"content"`
	ImageURLs   pq.StringArray `gorm:"type:varchar(128)[]" json:"image_urls"`
	Tags        pq.StringArray `gorm:"type:varchar(32)[]" json:"tags"`
	Status      AnswerStatus   `gorm:"not null" json:"status"`
	IsCertified bool           `gorm:"not null" json:"is_certified"` //是否精选
	Views       int64          `json:"views"`                        // 浏览量
	Likes       int64          `json:"likes"`                        // 赞同数
	Stars       int64          `json:"stars"`                        // 收藏数
	Replies     int64          `json:"replies"`                      // 评论数
	LastReplyAt *time.Time     `json:"last_reply_at"`                // 指针类型便于判空
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
	ID        int64     `gorm:"primaryKey"`
	PostID    *int64    `gorm:"index"` // 可以为 NULL
	UserID    int64     `gorm:"index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Post      Post      `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:SET NULL;"`
}

// redis缓存帖子的高频数据
type PostStat struct {
	PostID  int64
	Views   int64 // 可能累计为复数
	Likes   int64
	Stars   int64
	Replies int64
}

// 存储post表及当前用户的like&star情况
type PostDetail struct {
	Post
	Liked   bool
	Starred bool
}
