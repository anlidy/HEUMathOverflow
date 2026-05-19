package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"MathOverflow/services/forum/internal/model"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type CommentCacheConfig struct {
	CachePagesUpTo int
	ListTTLMin     time.Duration
	ListTTLMax     time.Duration
	DetailTTLMin   time.Duration
	DetailTTLMax   time.Duration
	LockTTL        time.Duration
}

type CommentCache struct {
	rdb   *redis.Client
	mutex *RedisMutex
	sf    singleflight.Group
	cfg   CommentCacheConfig
}

type pageLoadResult struct {
	IDs   []int64
	Total int64
}

type commentPageCache struct {
	IDs   []int64 `json:"ids"`
	Total int64   `json:"total"`
}

func NewCommentCache(rdb *redis.Client, cfg CommentCacheConfig) *CommentCache {
	if cfg.CachePagesUpTo <= 0 {
		cfg.CachePagesUpTo = 3
	}
	if cfg.ListTTLMin <= 0 {
		cfg.ListTTLMin = 30 * time.Second
	}
	if cfg.ListTTLMax <= 0 {
		cfg.ListTTLMax = 60 * time.Second
	}
	if cfg.DetailTTLMin <= 0 {
		cfg.DetailTTLMin = 2 * time.Minute
	}
	if cfg.DetailTTLMax <= 0 {
		cfg.DetailTTLMax = 5 * time.Minute
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 3 * time.Second
	}
	return &CommentCache{rdb: rdb, mutex: NewRedisMutex(rdb), cfg: cfg}
}

func (c *CommentCache) GetList(ctx context.Context, postID int64) ([]int64, int64, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, 0, false, nil
	}

	raw, err := c.rdb.Get(ctx, commentListKey(postID)).Bytes()
	if err == redis.Nil {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, err
	}

	var page commentPageCache
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, 0, false, err
	}
	return page.IDs, page.Total, true, nil
}

func (c *CommentCache) SetList(ctx context.Context, postID int64, ids []int64, total int64, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return nil
	}

	raw, err := json.Marshal(commentPageCache{IDs: ids, Total: total})
	if err != nil {
		return err
	}

	ttl = ClampMin(Jitter(ttl, 0.2), 5*time.Second)
	return c.rdb.Set(ctx, commentListKey(postID), raw, ttl).Err()
}

func (c *CommentCache) GetComments(ctx context.Context, commentIDs []int64) (map[int64]model.Reply, []int64, error) {
	hits := make(map[int64]model.Reply, len(commentIDs))
	missing := make([]int64, 0)
	if c == nil || c.rdb == nil || len(commentIDs) == 0 {
		return hits, commentIDs, nil
	}
	keys := make([]string, 0, len(commentIDs))
	for _, id := range commentIDs {
		keys = append(keys, commentDetailKey(id))
	}
	vals, err := c.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return hits, commentIDs, err
	}
	for i, v := range vals {
		id := commentIDs[i]
		if v == nil {
			missing = append(missing, id)
			continue
		}
		var raw []byte
		switch vv := v.(type) {
		case string:
			if vv == "" {
				missing = append(missing, id)
				continue
			}
			raw = []byte(vv)
		case []byte:
			if len(vv) == 0 {
				missing = append(missing, id)
				continue
			}
			raw = vv
		default:
			missing = append(missing, id)
			continue
		}
		var r model.Reply
		if err := json.Unmarshal(raw, &r); err != nil {
			missing = append(missing, id)
			continue
		}
		hits[id] = r
	}
	return hits, missing, nil
}

func (c *CommentCache) SetComments(ctx context.Context, replies []model.Reply, ttl time.Duration) error {
	if c == nil || c.rdb == nil || len(replies) == 0 {
		return nil
	}
	ttl = ClampMin(Jitter(ttl, 0.2), 10*time.Second)
	pipe := c.rdb.Pipeline()
	for _, r := range replies {
		raw, err := json.Marshal(r)
		if err != nil {
			continue
		}
		pipe.Set(ctx, commentDetailKey(r.ID), raw, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CommentCache) InvalidatePost(ctx context.Context, postID int64) {
	if c == nil || c.rdb == nil {
		return
	}
	_ = c.rdb.Del(ctx, commentListKey(postID)).Err()
}

func (c *CommentCache) InvalidateComment(ctx context.Context, commentID int64) {
	if c == nil || c.rdb == nil {
		return
	}
	_ = c.rdb.Del(ctx, commentDetailKey(commentID)).Err()
}

func (c *CommentCache) GetOrLoad(
	ctx context.Context,
	postID int64,
	loader func(context.Context) (ids []int64, total int64, replies []model.Reply, err error),
	listTTL time.Duration,
	detailTTL time.Duration,
) ([]int64, int64, error) {

	// 1️⃣ 先查缓存
	if ids, total, ok, err := c.GetList(ctx, postID); err == nil && ok {
		return ids, total, nil
	}

	key := strconv.FormatInt(postID, 10)

	v, err, _ := c.sf.Do(key, func() (any, error) {

		// double check
		if ids, total, ok, _ := c.GetList(ctx, postID); ok {
			return pageLoadResult{IDs: ids, Total: total}, nil
		}

		// 分布式锁
		lockKey := commentListLockKey(postID)
		token, locked, e := c.mutex.TryLock(ctx, lockKey, c.cfg.LockTTL)
		if e == nil && locked {
			defer func() { _ = c.mutex.Unlock(ctx, lockKey, token) }()

			ids, total, replies, err := loader(ctx)
			if err != nil {
				return nil, err
			}

			_ = c.SetList(ctx, postID, ids, total, listTTL)
			_ = c.SetComments(ctx, replies, detailTTL)

			return pageLoadResult{IDs: ids, Total: total}, nil
		}

		if page, ok := waitForCacheFill(ctx, 300*time.Millisecond, 20*time.Millisecond, 80*time.Millisecond, func(ctx context.Context) (pageLoadResult, bool, error) {
			ids, total, ok, err := c.GetList(ctx, postID)
			if err != nil || !ok {
				return pageLoadResult{}, ok, err
			}
			return pageLoadResult{IDs: ids, Total: total}, true, nil
		}); ok {
			return page, nil
		}

		// fallback
		ids, total, replies, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		_ = c.SetList(ctx, postID, ids, total, listTTL)
		_ = c.SetComments(ctx, replies, detailTTL)

		return pageLoadResult{IDs: ids, Total: total}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	out := v.(pageLoadResult)
	return out.IDs, out.Total, nil
}
