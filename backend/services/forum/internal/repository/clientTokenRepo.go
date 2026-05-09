package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"MathOverflow/common/utils"

	"github.com/redis/go-redis/v9"
)

type ClientTokenKind string

const (
	ClientTokenPost  ClientTokenKind = "post"
	ClientTokenReply ClientTokenKind = "reply"
)

type ClientTokenRepo interface {
	Issue(ctx context.Context, kind ClientTokenKind, userID int64) (string, error)
	IsIssued(ctx context.Context, kind ClientTokenKind, userID int64, token string) (bool, error)
	GetResult(ctx context.Context, kind ClientTokenKind, userID int64, token string) (int64, bool, error)
	SetResult(ctx context.Context, kind ClientTokenKind, userID int64, token string, id int64) error
	TryLock(ctx context.Context, kind ClientTokenKind, userID int64, token string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, kind ClientTokenKind, userID int64, token string) error
}

type clientTokenRepo struct {
	rdb *redis.Client
}

func NewClientTokenRepository(rdb *redis.Client) ClientTokenRepo {
	return &clientTokenRepo{rdb: rdb}
}

func (r *clientTokenRepo) issuedKey(kind ClientTokenKind, userID int64, token string) string {
	return fmt.Sprintf("client_token:%s:issued:%d:%s", kind, userID, token)
}

func (r *clientTokenRepo) lockKey(kind ClientTokenKind, userID int64, token string) string {
	return fmt.Sprintf("client_token:%s:lock:%d:%s", kind, userID, token)
}

func (r *clientTokenRepo) Issue(ctx context.Context, kind ClientTokenKind, userID int64) (string, error) {
	if r == nil || r.rdb == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	token := utils.GenerateUUID()
	// token 有效期：页面进入后的一次性 token，过期后前端需要重新获取
	if err := r.rdb.Set(ctx, r.issuedKey(kind, userID, token), "pending", 30*time.Minute).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (r *clientTokenRepo) IsIssued(ctx context.Context, kind ClientTokenKind, userID int64, token string) (bool, error) {
	if r == nil || r.rdb == nil {
		return false, fmt.Errorf("redis client is nil")
	}
	n, err := r.rdb.Exists(ctx, r.issuedKey(kind, userID, token)).Result()
	return n == 1, err
}

func (r *clientTokenRepo) GetResult(ctx context.Context, kind ClientTokenKind, userID int64, token string) (int64, bool, error) {
	if r == nil || r.rdb == nil {
		return 0, false, fmt.Errorf("redis client is nil")
	}
	val, err := r.rdb.Get(ctx, r.issuedKey(kind, userID, token)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	if val == "" || val == "pending" {
		return 0, false, nil
	}
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (r *clientTokenRepo) SetResult(ctx context.Context, kind ClientTokenKind, userID int64, token string, id int64) error {
	if r == nil || r.rdb == nil {
		return fmt.Errorf("redis client is nil")
	}
	key := r.issuedKey(kind, userID, token)
	ttl, err := r.rdb.TTL(ctx, key).Result()
	if err != nil {
		return err
	}
	// TTL <= 0 说明 token 已过期或未设置过期时间；按 token 过期处理
	if ttl <= 0 {
		return fmt.Errorf("client token expired before setting result")
	}
	return r.rdb.Set(ctx, key, strconv.FormatInt(id, 10), ttl).Err()
}

func (r *clientTokenRepo) TryLock(ctx context.Context, kind ClientTokenKind, userID int64, token string, ttl time.Duration) (bool, error) {
	if r == nil || r.rdb == nil {
		return false, fmt.Errorf("redis client is nil")
	}
	return r.rdb.SetNX(ctx, r.lockKey(kind, userID, token), "1", ttl).Result()
}

func (r *clientTokenRepo) Unlock(ctx context.Context, kind ClientTokenKind, userID int64, token string) error {
	if r == nil || r.rdb == nil {
		return fmt.Errorf("redis client is nil")
	}
	return r.rdb.Del(ctx, r.lockKey(kind, userID, token)).Err()
}
