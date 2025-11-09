package repository

import (
	"MathOverflow/internal/forum-service/model"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type ReplyRepo interface {
	CreateReply(reply *model.Reply) error
	FindReplyByID(replyID int64) (model.Reply, error)
	UpdateColumn(replyID int64, column string, value any) (bool, error)
	UpdateReply(reply *model.Reply) (bool, error)
	DeleteReply(replyID int64) error
	/// mongo
	MCreateReply(ctx context.Context, reply *model.ReplyContent) (primitive.ObjectID, error)
	MFindReply(ctx context.Context, filter bson.M) (*model.ReplyContent, error)
	MFindAllReply(ctx context.Context, filter bson.M) ([]model.ReplyContent, error)
	MUpdateReply(ctx context.Context, query bson.M, update bson.D) (bool, error)
	MDeleteReply(ctx context.Context, query bson.M) error
	MDeleteAllReplies(ctx context.Context, query bson.M) error
}

type replyRepo struct {
	pg      *gorm.DB
	replies *mongo.Collection
}

func NewReplyRepository(pg *gorm.DB, mdb *mongo.Database) ReplyRepo {
	replies := mdb.Collection("replies")
	return &replyRepo{pg: pg, replies: replies}
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

// 更新一列
func (r *replyRepo) UpdateColumn(replyID int64, column string, value any) (bool, error) {
	result := r.pg.Model(&model.Reply{}).Where("id = ?", replyID).Update(column, value)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 更新一个回帖
func (r *replyRepo) UpdateReply(reply *model.Reply) (bool, error) {
	result := r.pg.Model(&model.Reply{}).Where("id = ?", reply.ID).Updates(reply)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 删除一个回帖
func (r *replyRepo) DeleteReply(replyID int64) error {
	return r.pg.Where("id = ?", replyID).Delete(&model.Reply{}).Error
}

// 删除所有帖子的所有回复
func (r *replyRepo) DeleteAllReplies(postID int64) error {
	return r.pg.Where("post_id = ?", postID).Delete(&model.Reply{}).Error
}

// / mongo
// 创建一个回帖
func (r *replyRepo) MCreateReply(ctx context.Context, reply *model.ReplyContent) (primitive.ObjectID, error) {
	result, err := r.replies.InsertOne(ctx, reply)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// 查找一个回帖
func (r *replyRepo) MFindReply(ctx context.Context, filter bson.M) (*model.ReplyContent, error) {
	var reply *model.ReplyContent
	err := r.replies.FindOne(ctx, filter).Decode(reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// 查找所有回帖
func (r *replyRepo) MFindAllReply(ctx context.Context, filter bson.M) ([]model.ReplyContent, error) {
	// 执行查询
	cursor, err := r.replies.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 遍历结果
	var results []model.ReplyContent
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// 更新一个回帖
func (r *replyRepo) MUpdateReply(ctx context.Context, query bson.M, update bson.D) (bool, error) {
	result, err := r.replies.UpdateOne(ctx, query, update, options.Update().SetUpsert(false))
	if err != nil {
		return false, err
	}
	// 检查匹配情况
	if result.MatchedCount == 0 {
		return false, fmt.Errorf("未找到要更新的回帖内容")
	}
	return result.ModifiedCount > 0, nil
}

// 删除一个回帖
func (r *replyRepo) MDeleteReply(ctx context.Context, query bson.M) error {
	result, err := r.replies.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的回帖内容")
	}
	return nil
}

// 删除所有帖子的所有回复
func (r *replyRepo) MDeleteAllReplies(ctx context.Context, query bson.M) error {
	result, err := r.replies.DeleteMany(ctx, query)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的回帖内容")
	}
	return nil
}
