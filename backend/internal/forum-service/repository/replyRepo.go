package repository

import (
	"MathOverflow/internal/forum-service/model"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ReplyRepo interface {
	CreateReply(reply *model.Reply) error
	FindReplyDetailByID(replyID, userID int64) (model.ReplyDetail, error)
	FindReplyDetailsByPostID(postID, userID int64, offset, limit int) ([]model.ReplyDetail, int64, error)
	UpdateColumn(replyID int64, column string, value any) (bool, error)
	UpdateReply(reply *model.Reply) (bool, error)
	DeleteReply(replyID int64) error
	CreateReplyLike(rl *model.ReplyLike) error
	DeleteReplyLike(userID, replyID int64) error
	// redis
	IncreasePostReplies(ctx context.Context, postID int64) error
	DecreasePostReplies(ctx context.Context, postID int64) error
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
func (r *replyRepo) CreateReplyLike(rl *model.ReplyLike) error {
	return r.pg.Model(&model.ReplyLike{}).Create(rl).Error
}

// 删除点赞记录
func (r *replyRepo) DeleteReplyLike(userID, replyID int64) error {
	result := r.pg.Where("user_id = ? AND reply_id = ?", userID, replyID).Delete(&model.ReplyLike{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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
