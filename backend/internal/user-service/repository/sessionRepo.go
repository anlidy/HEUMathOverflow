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
	SetSession(ctx context.Context, session model.Session) error
	UpdateUserRole(ctx context.Context, userID int64, newRole model.Role) error
	ResetTTL(ctx context.Context, session model.Session, ttl time.Duration) error
	DeleteSession(ctx context.Context, session model.Session) error
}

type sessionRepo struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepo {
	return &sessionRepo{rdb: rdb}
}

// 获取session
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
	// 解析 role
	roleStr, err := s.rdb.Get(ctx, data["roleKey"]).Result()
	if err != nil {
		return model.Session{}, err
	}
	role, _ := strconv.ParseInt(roleStr, 10, 0)

	// 解析 userID
	userID, _ := strconv.ParseInt(data["userID"], 10, 64)

	// 解析 remember
	remember, _ := strconv.ParseBool(data["remember"])

	return model.Session{
		UserID:   userID,
		Role:     model.Role(role),
		Remember: remember,
	}, nil
}

// 创建session
func (s *sessionRepo) SetSession(ctx context.Context, session model.Session) error {
	key := "session:" + session.SessionID
	roleKey := "role:" + strconv.FormatInt(session.UserID, 10)
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, roleKey, int(session.Role), session.TTL) // 存储role:userID -> role,便于后续修改权限
	pipe.HSet(ctx, key, map[string]any{
		"userID":   session.UserID,
		"roleKey":  roleKey,
		"remember": session.Remember,
	})
	pipe.Expire(ctx, key, session.TTL)

	_, err := pipe.Exec(ctx)
	return err
}

// 更新用户role
func (s *sessionRepo) UpdateUserRole(ctx context.Context, userID int64, newRole model.Role) error {
	roleKey := "role:" + strconv.FormatInt(userID, 10)
	return s.rdb.SetArgs(ctx, roleKey, int(newRole), redis.SetArgs{
		KeepTTL: true, // redis >= 7.0
	}).Err()
}

// 重置ttl为指定值
func (s *sessionRepo) ResetTTL(ctx context.Context, session model.Session, ttl time.Duration) error {
	key := "session:" + session.SessionID
	ok, err := s.rdb.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("session %s not found", session.SessionID)
	}
	return nil
}

// 删除session
func (s *sessionRepo) DeleteSession(ctx context.Context, session model.Session) error {
	roleKey := "role:" + strconv.FormatInt(session.UserID, 10)
	pipe := s.rdb.TxPipeline()
	pipe.Del(ctx, "session:"+session.SessionID)
	pipe.Del(ctx, roleKey)

	_, err := pipe.Exec(ctx)
	return err
}
