package cache

import (
	"context"
	"time"

	"MathOverflow/internal/common/utils"

	"github.com/redis/go-redis/v9"
)

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
`)

type RedisMutex struct {
	rdb *redis.Client
}

func NewRedisMutex(rdb *redis.Client) *RedisMutex {
	return &RedisMutex{rdb: rdb}
}

// TryLock returns (token, true, nil) on success.
func (m *RedisMutex) TryLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	if m == nil || m.rdb == nil {
		return "", false, nil
	}
	if ttl <= 0 {
		ttl = 3 * time.Second
	}
	token := utils.GenerateUUID()
	ok, err := m.rdb.SetNX(ctx, key, token, ttl).Result()
	return token, ok, err
}

func (m *RedisMutex) Unlock(ctx context.Context, key, token string) error {
	if m == nil || m.rdb == nil {
		return nil
	}
	_, err := unlockScript.Run(ctx, m.rdb, []string{key}, token).Result()
	return err
}
