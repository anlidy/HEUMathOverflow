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

type PostRepo interface {
	CreatePost(post *model.Post) error
	FindPostByID(postID int64) (model.Post, error)
	UpdateColumn(postID int64, column string, value any) (bool, error)
	UpdatePost(post *model.Post) (bool, error)
	DeletePost(postID int64) error
	// mongo
	CreateOnePost(ctx context.Context, post *model.PostContent) (primitive.ObjectID, error)
	FindOnePost(ctx context.Context, filter bson.M) (*model.PostContent, error)
	UpdateOnePost(ctx context.Context, query bson.M, update bson.D) (bool, error)
	DeleteOnePost(ctx context.Context, query bson.M) error
}

type postRepo struct {
	pg    *gorm.DB
	posts *mongo.Collection
}

func NewPostRepository(pg *gorm.DB, mdb *mongo.Database) PostRepo {
	posts := mdb.Collection("posts")
	return &postRepo{pg: pg, posts: posts}
}

// pg创建帖子记录
func (r *postRepo) CreatePost(post *model.Post) error {
	return r.pg.Model(&model.Post{}).Create(post).Error
}

// 根据postID查询帖子
func (r *postRepo) FindPostByID(postID int64) (model.Post, error) {
	var result model.Post
	err := r.pg.Where("id = ?", postID).First(&result).Error
	return result, err
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
	return r.pg.Where("id = ?", postID).Delete(&model.Post{}).Error
}

// / Mongo接口
// 创建一个帖子
func (r *postRepo) CreateOnePost(ctx context.Context, post *model.PostContent) (primitive.ObjectID, error) {
	result, err := r.posts.InsertOne(ctx, post)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// 查找一个帖子
func (r *postRepo) FindOnePost(ctx context.Context, filter bson.M) (*model.PostContent, error) {
	var post *model.PostContent
	err := r.posts.FindOne(ctx, filter).Decode(post)
	if err != nil {
		return nil, err
	}
	return post, nil
}

// 更新一个帖子
func (r *postRepo) UpdateOnePost(ctx context.Context, query bson.M, update bson.D) (bool, error) {
	result, err := r.posts.UpdateOne(ctx, query, update, options.Update().SetUpsert(false))
	if err != nil {
		return false, err
	}
	// 检查匹配情况
	if result.MatchedCount == 0 {
		return false, fmt.Errorf("未找到要更新的帖子内容")
	}
	return result.ModifiedCount > 0, nil
}

// 删除一个帖子
func (r *postRepo) DeleteOnePost(ctx context.Context, query bson.M) error {
	result, err := r.posts.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("未找到要删除的帖子内容")
	}
	return nil
}
