package repository

import (
	"MathOverflow/services/user/internal/model"

	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(user *model.User) error
	FindUserByID(userID int64) (model.User, error)
	BatchFindUserByID(userIDs []int64) ([]model.User, error)
	FindUserByEmail(email string) (model.User, error)
	FindUserByUsername(username string) (model.User, error)
	UpdateColumn(userID int64, column string, value any) (bool, error)
	UpdateUser(user *model.User) (bool, error)
	DeleteUser(userID int64) (bool, error)
}

type userRepo struct {
	pg *gorm.DB // 持有一个pg数据库实例
}

func NewUserRepository(pg *gorm.DB) UserRepo {
	return &userRepo{pg: pg}
}

// 创建新用户
func (r *userRepo) CreateUser(user *model.User) error {
	err := r.pg.Model(&model.User{}).Create(user).Error
	return err
}

// 根据用户id查询User
func (r *userRepo) FindUserByID(userID int64) (model.User, error) {
	var result model.User
	err := r.pg.Where("id = ?", userID).First(&result).Error
	return result, err
}

// 根据用户id批量查询Users
func (r *userRepo) BatchFindUserByID(userIDs []int64) ([]model.User, error) {
	var users []model.User
	err := r.pg.Where("id IN ?", userIDs).Find(&users).Error
	return users, err
}

// 根据用户email查询User
func (r *userRepo) FindUserByEmail(email string) (model.User, error) {
	var result model.User
	err := r.pg.Where("email = ?", email).First(&result).Error
	return result, err
}

// 根据用户username查询User
func (r *userRepo) FindUserByUsername(username string) (model.User, error) {
	var result model.User
	err := r.pg.Where("username = ?", username).First(&result).Error
	return result, err
}

// 更新一列
func (r *userRepo) UpdateColumn(userID int64, column string, value any) (bool, error) {
	result := r.pg.Model(&model.User{}).Where("id = ?", userID).Update(column, value)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// 更新一个用户
func (r *userRepo) UpdateUser(user *model.User) (bool, error) {
	err := r.pg.Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// 删除一个用户
func (r *userRepo) DeleteUser(userID int64) (bool, error) {
	err := r.pg.Where("id = ?", userID).Delete(&model.User{}).Error
	if err != nil {
		return false, err
	}
	return true, nil
}
