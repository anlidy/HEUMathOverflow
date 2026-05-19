package repository

import (
	common "MathOverflow/common/model"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostRepo interface {
	CreatePost(post *model.Post) error
	FindPostByID(postID int64) (model.Post, error)
	// FindPostByID(postID int64) (model.Post, error)
	FindPostDetail(postID, userID int64) (model.PostDetail, error)
	FindPostFlags(postID, userID int64) (bool, bool, error)
	FindManyPosts(ctx context.Context, offset, limit int, order model.OrderBy) ([]model.Post, error)
	FindPostMapByIDs(postIDs []int64) (map[int64]model.Post, error)
	UpdateColumn(postID int64, column string, value any) (bool, error)
	UpdatePost(post *model.Post) (bool, error)
	DeletePost(postID int64) error
	CreatePostLike(pl *model.PostLike) (bool, error)
	DeletePostLike(userID, postID int64) (bool, error)
	CreatePostStar(pl *model.PostStar) (bool, error)
	DeletePostStar(userID, postID int64) (bool, error)
	HasPostStar(userID, postID int64) (bool, error)
	FindUserStarredPosts(userID int64, offset, limit int) ([]model.Post, int64, error)
	RunInTx(ctx context.Context, fn func(tx *gorm.DB) error) error
	AddOutboxMessageInTx(ctx context.Context, tx *gorm.DB, topic string, postID int64, payload []byte) error
	// redis
	IncreasePostStat(ctx context.Context, postID int64, attr string) error
	DecreasePostStat(ctx context.Context, postID int64, attr string) error
	MovePostStatToProcessing(ctx context.Context, key string) (string, bool, error)
	DeletePostStatKey(ctx context.Context, key string) error
	BumpHottestCacheVersion(ctx context.Context) error
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

func (r *postRepo) FindPostByID(postID int64) (model.Post, error) {
	var result model.Post
	err := r.pg.Model(&model.Post{}).Where("id = ?", postID).First(&result).Error
	return result, err
}

// // 根据postID查询帖子
// func (r *postRepo) FindPostByID(postID int64) (model.Post, error) {
// 	var result model.Post
// 	err := r.pg.Model(&model.Post{}).Where("id = ?", postID).First(&result).Error
// 	return result, err
// }

// 查询帖子数据及当前用户是否like&star
func (r *postRepo) FindPostDetail(postID, userID int64) (model.PostDetail, error) {
	var result model.PostDetail

	err := r.pg.Raw(`
        SELECT 
            p.*,
            EXISTS(SELECT 1 FROM post_likes pl 
                   WHERE pl.post_id = p.id AND pl.user_id = ?) AS liked,
            EXISTS(SELECT 1 FROM post_stars ps 
                   WHERE ps.post_id = p.id AND ps.user_id = ?) AS starred
        FROM posts p
        WHERE p.id = ?
    `, userID, userID, postID).Scan(&result).Error

	return result, err
}

func (r *postRepo) FindPostFlags(postID, userID int64) (bool, bool, error) {
	type flags struct {
		Liked   bool `gorm:"column:liked"`
		Starred bool `gorm:"column:starred"`
	}
	var f flags
	err := r.pg.Raw(`
        SELECT
            EXISTS(SELECT 1 FROM post_likes pl WHERE pl.post_id = ? AND pl.user_id = ?) AS liked,
            EXISTS(SELECT 1 FROM post_stars ps WHERE ps.post_id = ? AND ps.user_id = ?) AS starred
    `, postID, userID, postID, userID).Scan(&f).Error
	return f.Liked, f.Starred, err
}

// 查询多条帖子
func (r *postRepo) FindManyPosts(ctx context.Context, offset, limit int, order model.OrderBy) ([]model.Post, error) {
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

	// 热门帖子（热点接口）走 redis 缓存，减少高频排序查询压力
	if order == model.Hottest && r.rdb != nil && ctx != nil && offset >= 0 && offset <= 2000 && limit > 0 {
		ver, err := r.rdb.Get(ctx, HottestPostsCacheVersionKey).Int64()
		if err == redis.Nil {
			ver = 1
			_ = r.rdb.Set(ctx, HottestPostsCacheVersionKey, "1", 0).Err()
		} else if err != nil {
			ver = 0
		}
		if ver > 0 {
			cacheKey := fmt.Sprintf("forum:posts:hottest:v%d:o%d:l%d", ver, offset, limit)
			if raw, err := r.rdb.Get(ctx, cacheKey).Bytes(); err == nil && len(raw) > 0 {
				if e := json.Unmarshal(raw, &results); e == nil {
					return results, nil
				}
			}
			err = r.pg.Offset(offset).Limit(limit).Order(orderStr).Find(&results).Error
			if err != nil {
				return results, err
			}
			if raw, e := json.Marshal(results); e == nil {
				_ = r.rdb.Set(ctx, cacheKey, raw, 30*time.Second).Err()
			}
			return results, nil
		}
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
func (r *postRepo) CreatePostLike(pl *model.PostLike) (bool, error) {
	result := r.pg.Model(&model.PostLike{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(pl)
	return result.RowsAffected > 0, result.Error
}

// 删除点赞记录
func (r *postRepo) DeletePostLike(userID, postID int64) (bool, error) {
	result := r.pg.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.PostLike{})
	return result.RowsAffected > 0, result.Error
}

// / 收藏接口
// 创建收藏记录
func (r *postRepo) CreatePostStar(pl *model.PostStar) (bool, error) {
	result := r.pg.Model(&model.PostStar{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(pl)
	return result.RowsAffected > 0, result.Error
}

// 删除点赞记录
func (r *postRepo) DeletePostStar(userID, postID int64) (bool, error) {
	result := r.pg.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.PostStar{})
	return result.RowsAffected > 0, result.Error
}

func (r *postRepo) HasPostStar(userID, postID int64) (bool, error) {
	var count int64
	err := r.pg.Model(&model.PostStar{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Limit(1).
		Count(&count).Error
	return count > 0, err
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

func (r *postRepo) RunInTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.pg.WithContext(ctx).Transaction(fn)
}

func (r *postRepo) AddOutboxMessageInTx(ctx context.Context, tx *gorm.DB, topic string, postID int64, payload []byte) error {
	message := common.OutboxMessage{
		ID:           utils.GenerateSnowflakeID(),
		ResourceType: "post",
		ResourceID:   postID,
		Topic:        topic,
		Target:       "forum-post-events",
		Payload:      payload,
		Status:       common.OutboxMessagePending,
		RetryAt:      time.Now(),
	}
	return tx.WithContext(ctx).Create(&message).Error
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

func (r *postRepo) MovePostStatToProcessing(ctx context.Context, key string) (string, bool, error) {
	processingKey := fmt.Sprintf("%s:processing", key)
	moved, err := r.rdb.Eval(ctx, `
local src = KEYS[1]
local dst = KEYS[2]
if redis.call("EXISTS", src) == 0 then
    return 0
end
if redis.call("EXISTS", dst) == 1 then
    return -1
end
redis.call("RENAME", src, dst)
return 1
`, []string{key, processingKey}).Int()
	if err != nil {
		return "", false, err
	}
	switch moved {
	case 1:
		return processingKey, true, nil
	case 0:
		return processingKey, false, nil
	default:
		return processingKey, false, nil
	}
}

func (r *postRepo) DeletePostStatKey(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *postRepo) BumpHottestCacheVersion(ctx context.Context) error {
	if r == nil || r.rdb == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return r.rdb.Incr(ctx, HottestPostsCacheVersionKey).Err()
}
