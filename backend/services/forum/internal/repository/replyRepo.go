package repository

import (
	"MathOverflow/services/forum/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrReplyStatusConflict = errors.New("reply status changed concurrently")

type ReplyRepo interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(replyID int64) (model.Reply, error)
	FindReplyDetailByID(replyID, userID int64) (model.ReplyDetail, error)
	FindReplyDetailsByPostID(postID, userID int64, offset, limit int) ([]model.ReplyDetail, int64, error)
	FindRepliesPageByPostID(postID int64, offset, limit int) ([]model.Reply, int64, error)
	FindRepliesByPostID(postID int64) ([]model.Reply, error)
	FindTopRepliesByLikes(postID int64, limit int) ([]model.Reply, error)
	FindRepliesByIDs(replyIDs []int64) ([]model.Reply, error)
	FindReplyLikeMap(userID int64, replyIDs []int64) (map[int64]bool, error)
	UpdateColumn(replyID int64, column string, value any) (bool, error)
	UpdateReply(reply *model.Reply) (bool, error)
	DeleteReply(replyID int64) error
	CreateReplyLike(rl *model.ReplyLike) (bool, error)
	DeleteReplyLike(userID, replyID int64) (bool, error)
	// redis
	IncreaseReplyStat(ctx context.Context, replyID int64, attr string) error
	DecreaseReplyStat(ctx context.Context, replyID int64, attr string) error
	MoveReplyStatToProcessing(ctx context.Context, key string) (string, bool, error)
	DeleteReplyStatKey(ctx context.Context, key string) error
	IncreasePostReplies(ctx context.Context, postID int64) error
	DecreasePostReplies(ctx context.Context, postID int64) error
	ChangeReplyAndPostStatus(ctx context.Context, replyID, postID int64, expectedReplyStatus int, replyStatus int, certifiedBy *int64, postStatus *int) error
}

type replyRepo struct {
	pg  *gorm.DB
	rdb *redis.Client
}

func NewReplyRepository(pg *gorm.DB, rdb *redis.Client) ReplyRepo {
	return &replyRepo{pg: pg, rdb: rdb}
}

// pg创建回帖记录
func (r *replyRepo) CreateReply(reply *model.Reply) error {
	return r.pg.Model(&model.Reply{}).Create(reply).Error
}

func (r *replyRepo) FindReplyByID(replyID int64) (model.Reply, error) {
	var result model.Reply
	err := r.pg.Model(&model.Reply{}).Where("id = ?", replyID).First(&result).Error
	return result, err
}

// pg根据replyID查询回帖
func (r *replyRepo) FindReplyDetailByID(replyID, userID int64) (model.ReplyDetail, error) {
	var result model.ReplyDetail
	err := r.pg.Raw(`
		SELECT 
			r.*,
			EXISTS (
				SELECT 1 FROM reply_likes rl
				WHERE rl.reply_id = r.id AND rl.user_id = ?
			) AS liked
		FROM replies r
		WHERE r.id = ?
	`, userID, replyID).Scan(&result).Error
	return result, err
}

// 查询帖子的回帖数据及用户对这些回帖的点赞情况
func (r *replyRepo) FindReplyDetailsByPostID(postID, userID int64, offset, limit int) ([]model.ReplyDetail, int64, error) {
	var results []model.ReplyDetail
	var total int64

	// 总数
	err := r.pg.Model(&model.Reply{}).
		Where("post_id = ?", postID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询回复 + 当前用户是否点赞
	err = r.pg.Raw(`
        SELECT 
            r.*,
            EXISTS (
                SELECT 1 FROM reply_likes rl
                WHERE rl.reply_id = r.id AND rl.user_id = ?
            ) AS liked
        FROM replies r
        WHERE r.post_id = ?
        ORDER BY r.status DESC
        OFFSET ? LIMIT ?
    `, userID, postID, offset, limit).Scan(&results).Error

	return results, total, err
}

// FindRepliesPageByPostID loads replies without user-specific liked flag (for caching).
func (r *replyRepo) FindRepliesPageByPostID(postID int64, offset, limit int) ([]model.Reply, int64, error) {
	var results []model.Reply
	var total int64
	err := r.pg.Model(&model.Reply{}).
		Where("post_id = ?", postID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.pg.Model(&model.Reply{}).
		Where("post_id = ?", postID).
		Order("status DESC").
		Offset(offset).Limit(limit).
		Find(&results).Error
	return results, total, err
}

func (r *replyRepo) FindRepliesByPostID(postID int64) ([]model.Reply, error) {
	var results []model.Reply
	err := r.pg.Model(&model.Reply{}).
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&results).Error
	return results, err
}

func (r *replyRepo) FindTopRepliesByLikes(postID int64, limit int) ([]model.Reply, error) {
	if limit <= 0 {
		limit = 3
	}
	var results []model.Reply
	err := r.pg.Model(&model.Reply{}).
		Where("post_id = ?", postID).
		Order("likes DESC, created_at ASC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

func (r *replyRepo) FindRepliesByIDs(replyIDs []int64) ([]model.Reply, error) {
	if len(replyIDs) == 0 {
		return nil, nil
	}
	var results []model.Reply
	err := r.pg.Model(&model.Reply{}).Where("id IN (?)", replyIDs).Find(&results).Error
	return results, err
}

func (r *replyRepo) FindReplyLikeMap(userID int64, replyIDs []int64) (map[int64]bool, error) {
	mp := make(map[int64]bool, len(replyIDs))
	if len(replyIDs) == 0 {
		return mp, nil
	}
	type row struct {
		ReplyID int64 `gorm:"column:reply_id"`
	}
	var rows []row
	err := r.pg.Model(&model.ReplyLike{}).
		Select("reply_id").
		Where("user_id = ? AND reply_id IN (?)", userID, replyIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		mp[r.ReplyID] = true
	}
	return mp, nil
}

// pg更新一列
func (r *replyRepo) UpdateColumn(replyID int64, column string, value any) (bool, error) {
	result := r.pg.Model(&model.Reply{}).Where("id = ?", replyID).Update(column, value)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// pg更新一个回帖
func (r *replyRepo) UpdateReply(reply *model.Reply) (bool, error) {
	result := r.pg.Model(&model.Reply{}).Where("id = ?", reply.ID).Updates(reply)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// pg删除一个回帖
func (r *replyRepo) DeleteReply(replyID int64) error {
	result := r.pg.Where("id = ?", replyID).Delete(&model.Reply{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// 删除所有帖子的所有回复
func (r *replyRepo) DeleteAllReplies(postID int64) error {
	result := r.pg.Where("post_id = ?", postID).Delete(&model.Reply{})
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
func (r *replyRepo) CreateReplyLike(rl *model.ReplyLike) (bool, error) {
	result := r.pg.Model(&model.ReplyLike{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(rl)
	return result.RowsAffected > 0, result.Error
}

// 删除点赞记录
func (r *replyRepo) DeleteReplyLike(userID, replyID int64) (bool, error) {
	result := r.pg.Where("user_id = ? AND reply_id = ?", userID, replyID).Delete(&model.ReplyLike{})
	return result.RowsAffected > 0, result.Error
}

func (r *replyRepo) IncreaseReplyStat(ctx context.Context, replyID int64, attr string) error {
	key := fmt.Sprintf("reply:%d:stat", replyID)
	_, err := r.rdb.HIncrBy(ctx, key, attr, 1).Result()
	if err != nil {
		return fmt.Errorf("increase reply %s failed: %w", attr, err)
	}
	return nil
}

func (r *replyRepo) DecreaseReplyStat(ctx context.Context, replyID int64, attr string) error {
	key := fmt.Sprintf("reply:%d:stat", replyID)
	_, err := r.rdb.HIncrBy(ctx, key, attr, -1).Result()
	if err != nil {
		return fmt.Errorf("decrease reply %s failed: %w", attr, err)
	}
	return nil
}

func (r *replyRepo) MoveReplyStatToProcessing(ctx context.Context, key string) (string, bool, error) {
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

func (r *replyRepo) DeleteReplyStatKey(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

// redis增加 post replies 次数
func (r *replyRepo) IncreasePostReplies(ctx context.Context, postID int64) error {
	key := fmt.Sprintf("post:%d:stat", postID)

	// 原子 +1，如果 key 或字段不存在，Redis 会自动创建
	_, err := r.rdb.HIncrBy(ctx, key, "replies", 1).Result()
	if err != nil {
		return fmt.Errorf("increase replies failed: %w", err)
	}

	return nil
}

// redis减少 post replies 次数
func (r *replyRepo) DecreasePostReplies(ctx context.Context, postID int64) error {
	key := fmt.Sprintf("post:%d:stat", postID)

	// 原子 -1，如果 key 或字段不存在，Redis 会自动创建
	_, err := r.rdb.HIncrBy(ctx, key, "replies", -1).Result()
	if err != nil {
		return fmt.Errorf("decrease replies failed: %w", err)
	}

	return nil
}

// 同时修改回帖和帖子的状态
func (r *replyRepo) ChangeReplyAndPostStatus(ctx context.Context, replyID, postID int64, expectedReplyStatus int, replyStatus int, certifiedBy *int64, postStatus *int) error {
	return r.pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"status":       replyStatus,
			"certified_by": certifiedBy,
		}
		result := tx.Model(&model.Reply{}).
			Where("id = ? AND post_id = ? AND status = ?", replyID, postID, expectedReplyStatus).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrReplyStatusConflict
		}

		if postStatus != nil {
			if err := tx.Model(&model.Post{}).Where("id = ?", postID).Update("status", *postStatus).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
