package repository

import (
	"MathOverflow/internal/forum-service/model"

	"gorm.io/gorm"
)

type PostRepo interface {
	CreatePost(post *model.Post) error
	FindPostByID(postID int64) (model.Post, error)
	FindManyPosts(offset, limit int, order model.OrderBy) ([]model.Post, error)
	UpdateColumn(postID int64, column string, value any) (bool, error)
	UpdatePost(post *model.Post) (bool, error)
	DeletePost(postID int64) error
	CreatePostLike(pl *model.PostLike) error
	DeletePostLike(userID, postID int64) error
	HasPostLike(userID, postID int64) (bool, error)
	CreatePostStar(pl *model.PostStar) error
	DeletePostStar(userID, postID int64) error
	HasPostStar(userID, postID int64) (bool, error)
	FindUserStarredPosts(userID int64, offset, limit int) ([]model.Post, error)
}

type postRepo struct {
	pg *gorm.DB
}

func NewPostRepository(pg *gorm.DB) PostRepo {
	return &postRepo{pg: pg}
}

// pg创建帖子记录
func (r *postRepo) CreatePost(post *model.Post) error {
	return r.pg.Model(&model.Post{}).Create(post).Error
}

// 根据postID查询帖子(view++)
func (r *postRepo) FindPostByID(postID int64) (model.Post, error) {
	var result model.Post
	// 事务保证一致性
	err := r.pg.Transaction(func(tx *gorm.DB) error {
		// 1. 查询完整帖子
		if err := tx.First(&result, postID).Error; err != nil {
			return err
		}

		// 2. views + 1 原子自增
		err := tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("views", gorm.Expr("views + 1")).Error
		if err != nil {
			return err
		}
		return nil
	})
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
func (r *postRepo) FindUserStarredPosts(userID int64, offset, limit int) ([]model.Post, error) {
	var results []model.Post
	err := r.pg.Table("post_stars").
		Select("posts.*").
		Joins("JOIN posts ON posts.id = post_stars.post_id").
		Where("post_stars.user_id = ?", userID).
		Order("post_stars.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&results).Error
	return results, err
}
