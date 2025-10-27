package repository

import (
	"MathOverflow/internal/user-service/model"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepo interface {
	GetSession(ctx context.Context, sessionID string) (model.Session, error)
	SetSession(ctx context.Context, sessionID string, session model.Session, ttl time.Duration) error
	ResetTTL(ctx context.Context, sessionID string, ttl time.Duration) error
	DeleteSession(ctx context.Context, sessionID string) error
}

type sessionRepo struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepo {
	return &sessionRepo{rdb: rdb}
}

func (s *sessionRepo) GetSession(ctx context.Context, sessionID string) (model.Session, error) {
	key := "session:" + sessionID

	// 读取所有字段
	data, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return model.Session{}, err
	}

	// 如果session不存在
	if len(data) == 0 {
		return model.Session{}, fmt.Errorf("session %s not found", sessionID)
	}

	// 解析 userID
	userID, _ := strconv.ParseInt(data["userID"], 10, 64)
	// 解析 role
	role, _ := strconv.ParseInt(data["role"], 10, 0)
	// 解析 remember
	remember, _ := strconv.ParseBool(data["remember"])

	return model.Session{
		UserID:   userID,
		Role:     model.Role(role),
		Remember: remember,
	}, nil
}

func (s *sessionRepo) SetSession(ctx context.Context, sessionID string, session model.Session, ttl time.Duration) error {
	key := "session:" + sessionID

	pipe := s.rdb.TxPipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"userID":   session.UserID,
		"role":     session.Role,
		"remember": session.Remember,
	})
	pipe.Expire(ctx, key, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *sessionRepo) ResetTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := "session:" + sessionID
	ok, err := s.rdb.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("session %s not found", sessionID)
	}
	return nil
}

func (s *sessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	return s.rdb.Del(ctx, "session:"+sessionID).Err()
}
