package repository

import (
	"MathOverflow/internal/forum-service/model"

	"gorm.io/gorm"
)

type ReplyRepo interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(replyID int64) (model.Reply, error)
	FindRepliesByPostID(postID int64, offset, limit int) ([]model.Reply, error)
	UpdateColumn(replyID int64, column string, value any) (bool, error)
	UpdateReply(reply *model.Reply) (bool, error)
	DeleteReply(replyID int64) error
	CreateReplyLike(rl *model.ReplyLike) error
	DeleteReplyLike(userID, replyID int64) error
	HasReplyLike(userID, replyID int64) (bool, error)
}

type replyRepo struct {
	pg *gorm.DB
}

func NewReplyRepository(pg *gorm.DB) ReplyRepo {
	return &replyRepo{pg: pg}
}

// pg创建回帖记录
func (r *replyRepo) CreateReply(reply *model.Reply) error {
	return r.pg.Model(&model.Reply{}).Create(reply).Error
}

// pg根据replyID查询回帖
func (r *replyRepo) FindReplyByID(replyID int64) (model.Reply, error) {
	var result model.Reply
	err := r.pg.Where("id = ?", replyID).First(&result).Error
	return result, err
}

// pg根据postID,offset,limit查询多个回帖
func (r *replyRepo) FindRepliesByPostID(postID int64, offset, limit int) ([]model.Reply, error) {
	var results []model.Reply
	// 确保被认证的答案置顶
	err := r.pg.Where("post_id = ?", postID).Offset(offset).Limit(limit).Order("status DESC").Find(&results).Error
	return results, err
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

// 查询点赞记录是否存在
func (r *replyRepo) HasReplyLike(userID, replyID int64) (bool, error) {
	var count int64
	err := r.pg.Model(&model.ReplyLike{}).Where("user_id = ? AND reply_id = ?", userID, replyID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
