package repository

import (
	"MathOverflow/internal/forum-service/model"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostRepo interface {
	CreatePost(post *model.Post) error
	FindPostByID(postID int64) (model.Post, error)
	FindManyPosts(offset, limit int, order model.OrderBy) ([]model.Post, error)
	FindPostMapByIDs(postIDs []int64) (map[int64]model.Post, error)
	UpdateColumn(postID int64, column string, value any) (bool, error)
	UpdatePost(post *model.Post) (bool, error)
	DeletePost(postID int64) error
	CreatePostLike(pl *model.PostLike) error
	DeletePostLike(userID, postID int64) error
	HasPostLike(userID, postID int64) (bool, error)
	CreatePostStar(pl *model.PostStar) error
	DeletePostStar(userID, postID int64) error
	HasPostStar(userID, postID int64) (bool, error)
	FindUserStarredPosts(userID int64, offset, limit int) ([]model.Post, int64, error)
	// redis
	IncreasePostStat(ctx context.Context, postID int64, attr string) error
	DecreasePostStat(ctx context.Context, postID int64, attr string) error
}

type postRepo struct {
	pg  *gorm.DB
	rdb *redis.Client
}

func NewPostRepository(pg *gorm.DB, rdb *redis.Client) PostRepo {
	return &postRepo{pg: pg, rdb: rdb}
}

// pg创建帖子记录
func (r *postRepo) CreatePost(post *model.Post) error {
	return r.pg.Model(&model.Post{}).Create(post).Error
}

// 根据postID查询帖子
func (r *postRepo) FindPostByID(postID int64) (model.Post, error) {
	var result model.Post
	err := r.pg.Model(&model.Post{}).Where("id = ?", postID).First(&result).Error
	return result, err
}

// 查询多条帖子
func (r *postRepo) FindManyPosts(offset, limit int, order model.OrderBy) ([]model.Post, error) {
	var results []model.Post
	var orderStr string
	switch order {
	case model.Hottest: // 最热排序
		orderStr = "views * 0.1 + likes * 3 + replies * 4 + stars * 5 + status * 10 DESC"
	case model.Latest: // 最新排序
		orderStr = "created_at DESC"
	default: // 推荐排序(时间衰减)
		orderStr = `(views * 0.1 + likes * 3 + replies * 4 + stars * 5 + status * 10) 
		/ pow(EXTRACT(EPOCH FROM (now() - created_at)) / 3600 + 2, 1.5) DESC`
	}
	err := r.pg.Offset(offset).Limit(limit).Order(orderStr).Find(&results).Error
	return results, err
}

// 根据postID查询多条帖子
func (r *postRepo) FindPostMapByIDs(postIDs []int64) (map[int64]model.Post, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	// 查询
	var list []model.Post
	err := r.pg.Where("id IN (?)", postIDs).Find(&list).Error
	if err != nil {
		return nil, err
	}

	// 构建 map，方便按原顺序重排
	mp := make(map[int64]model.Post, len(list))
	for _, post := range list {
		mp[post.ID] = post
	}

	return mp, nil
}

// 更新一列
func (r *postRepo) UpdateColumn(postID int64, column string, value any) (bool, error) {
	result := r.pg.Model(&model.Post{}).Where("id = ?", postID).Update(column, value)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 更新一个帖子
func (r *postRepo) UpdatePost(post *model.Post) (bool, error) {
	result := r.pg.Model(&model.Post{}).Where("id = ?", post.ID).Updates(post)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 删除一个帖子
func (r *postRepo) DeletePost(postID int64) error {
	result := r.pg.Where("id = ?", postID).Delete(&model.Post{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// / 点赞接口
// 创建点赞记录
func (r *postRepo) CreatePostLike(pl *model.PostLike) error {
	return r.pg.Model(&model.PostLike{}).Create(pl).Error
}

// 删除点赞记录
func (r *postRepo) DeletePostLike(userID, postID int64) error {
	result := r.pg.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.PostLike{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 查询点赞记录是否存在
func (r *postRepo) HasPostLike(userID, postID int64) (bool, error) {
	var count int64
	err := r.pg.Model(&model.PostLike{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// / 收藏接口
// 创建收藏记录
func (r *postRepo) CreatePostStar(pl *model.PostStar) error {
	return r.pg.Model(&model.PostStar{}).Create(pl).Error
}

// 删除点赞记录
func (r *postRepo) DeletePostStar(userID, postID int64) error {
	result := r.pg.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.PostStar{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 查询收藏记录是否存在
func (r *postRepo) HasPostStar(userID, postID int64) (bool, error) {
	var count int64
	err := r.pg.Model(&model.PostStar{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 查询用户收藏的帖子（分页，按收藏时间倒序）
func (r *postRepo) FindUserStarredPosts(userID int64, offset, limit int) ([]model.Post, int64, error) {
	var results []model.Post
	var count int64
	err := r.pg.Model(&model.PostStar{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return results, 0, err
	}
	err = r.pg.Table("post_stars").
		Select("posts.*").
		Joins("JOIN posts ON posts.id = post_stars.post_id").
		Where("post_stars.user_id = ?", userID).
		Order("post_stars.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&results).Error
	return results, count, err
}

// redis增加{attr}次数
func (r *postRepo) IncreasePostStat(ctx context.Context, postID int64, attr string) error {
	key := fmt.Sprintf("post:%d:stat", postID)

	// 原子 +1，如果 key 或字段不存在，Redis 会自动创建
	_, err := r.rdb.HIncrBy(ctx, key, attr, 1).Result()
	if err != nil {
		return fmt.Errorf("increase %s failed: %w", attr, err)
	}

	return nil
}

// redis减少{attr}次数
func (r *postRepo) DecreasePostStat(ctx context.Context, postID int64, attr string) error {
	key := fmt.Sprintf("post:%d:stat", postID)

	// 原子 -1，如果 key 或字段不存在，Redis 会自动创建
	_, err := r.rdb.HIncrBy(ctx, key, attr, -1).Result()
	if err != nil {
		return fmt.Errorf("decrease %s failed: %w", attr, err)
	}

	return nil
}
