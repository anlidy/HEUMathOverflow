package repository

import (
	"MathOverflow/services/forum/internal/model"
	"context"

	"gorm.io/gorm"
)

type ChatRepo interface {
	CreateSession(session *model.ChatSession) error
	ListSessionsByUser(userID int64, offset, limit int) ([]model.ChatSession, int64, error)
	GetSessionByIDAndUser(sessionID string, userID int64) (model.ChatSession, error)
	ListMessages(sessionID string, offset, limit int) ([]model.ChatMessage, int64, error)
	ListMessagesAfter(sessionID string, afterMessageID *int64) ([]model.ChatMessage, error)
	RunInTx(ctx context.Context, fn func(tx *gorm.DB) error) error
	CreateMessagesTx(tx *gorm.DB, messages []model.ChatMessage) error
	UpdateSessionTx(tx *gorm.DB, session *model.ChatSession) error
}

type chatRepo struct {
	pg *gorm.DB
}

func NewChatRepository(pg *gorm.DB) ChatRepo {
	return &chatRepo{pg: pg}
}

func (r *chatRepo) CreateSession(session *model.ChatSession) error {
	return r.pg.Create(session).Error
}

func (r *chatRepo) ListSessionsByUser(userID int64, offset, limit int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64
	if err := r.pg.Model(&model.ChatSession{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.pg.Where("user_id = ?", userID).Order("updated_at DESC").Offset(offset).Limit(limit).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (r *chatRepo) GetSessionByIDAndUser(sessionID string, userID int64) (model.ChatSession, error) {
	var session model.ChatSession
	err := r.pg.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	return session, err
}

func (r *chatRepo) ListMessages(sessionID string, offset, limit int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64
	if err := r.pg.Model(&model.ChatMessage{}).Where("session_id = ?", sessionID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.pg.Where("session_id = ?", sessionID).Order("id ASC").Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

func (r *chatRepo) ListMessagesAfter(sessionID string, afterMessageID *int64) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	query := r.pg.Where("session_id = ?", sessionID)
	if afterMessageID != nil {
		query = query.Where("id > ?", *afterMessageID)
	}
	err := query.Order("id ASC").Find(&messages).Error
	return messages, err
}

func (r *chatRepo) RunInTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.pg.WithContext(ctx).Transaction(fn)
}

func (r *chatRepo) CreateMessagesTx(tx *gorm.DB, messages []model.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}
	return tx.Create(&messages).Error
}

func (r *chatRepo) UpdateSessionTx(tx *gorm.DB, session *model.ChatSession) error {
	return tx.Save(session).Error
}
