package repository

import (
	"MathOverflow/services/forum/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSnapshotRepo interface {
	Upsert(s model.UserSnapshot) error
	DeleteByID(userID int64) error
	FindByID(userID int64) (model.UserSnapshot, error)
	FindMapByIDs(userIDs []int64) (map[int64]model.UserSnapshot, error)
}

type userSnapshotRepo struct {
	pg *gorm.DB
}

func NewUserSnapshotRepository(pg *gorm.DB) UserSnapshotRepo {
	return &userSnapshotRepo{pg: pg}
}

func (r *userSnapshotRepo) Upsert(s model.UserSnapshot) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	return r.pg.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"username",
			"role",
			"avatar_url",
			"updated_at",
		}),
	}).Create(&s).Error
}

func (r *userSnapshotRepo) DeleteByID(userID int64) error {
	return r.pg.Where("user_id = ?", userID).Delete(&model.UserSnapshot{}).Error
}

func (r *userSnapshotRepo) FindByID(userID int64) (model.UserSnapshot, error) {
	var s model.UserSnapshot
	err := r.pg.Where("user_id = ?", userID).First(&s).Error
	return s, err
}

func (r *userSnapshotRepo) FindMapByIDs(userIDs []int64) (map[int64]model.UserSnapshot, error) {
	if len(userIDs) == 0 {
		return map[int64]model.UserSnapshot{}, nil
	}
	var list []model.UserSnapshot
	if err := r.pg.Where("user_id IN (?)", userIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	mp := make(map[int64]model.UserSnapshot, len(list))
	for _, s := range list {
		mp[s.UserID] = s
	}
	return mp, nil
}
