package cache

import (
	"context"
	"encoding/json"
	"fmt"
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

func (c *CommentCache) listKey(postID int64) string {
	return fmt.Sprintf("comments:post:%d", postID)
}

func (c *CommentCache) totalKey(postID int64) string {
	return fmt.Sprintf("comments:post:%d:total", postID)
}

func (c *CommentCache) commentKey(commentID int64) string {
	return fmt.Sprintf("comment:%d", commentID)
}

func (c *CommentCache) lockKey(postID int64) string {
	return fmt.Sprintf("lock:comments:post:%d", postID)
}

func (c *CommentCache) GetTotal(ctx context.Context, postID int64) (int64, bool, error) {
	if c == nil || c.rdb == nil {
		return 0, false, nil
	}
	n, err := c.rdb.Get(ctx, c.totalKey(postID)).Int64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return n, true, nil
}

func (c *CommentCache) SetTotal(ctx context.Context, postID int64, total int64, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	ttl = ClampMin(Jitter(ttl, 0.2), 5*time.Second)
	return c.rdb.Set(ctx, c.totalKey(postID), strconv.FormatInt(total, 10), ttl).Err()
}

func (c *CommentCache) GetList(ctx context.Context, postID int64) ([]int64, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}

	raw, err := c.rdb.Get(ctx, c.listKey(postID)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var ids []int64
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil, false, err
	}
	return ids, true, nil
}

func (c *CommentCache) SetList(ctx context.Context, postID int64, ids []int64, ttl time.Duration) error {
	if c == nil || c.rdb == nil {
		return nil
	}

	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	ttl = ClampMin(Jitter(ttl, 0.2), 5*time.Second)
	return c.rdb.Set(ctx, c.listKey(postID), raw, ttl).Err()
}

func (c *CommentCache) GetComments(ctx context.Context, commentIDs []int64) (map[int64]model.Reply, []int64, error) {
	hits := make(map[int64]model.Reply, len(commentIDs))
	missing := make([]int64, 0)
	if c == nil || c.rdb == nil || len(commentIDs) == 0 {
		return hits, commentIDs, nil
	}
	keys := make([]string, 0, len(commentIDs))
	for _, id := range commentIDs {
		keys = append(keys, c.commentKey(id))
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
		pipe.Set(ctx, c.commentKey(r.ID), raw, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *CommentCache) InvalidatePost(ctx context.Context, postID int64) {
	if c == nil || c.rdb == nil {
		return
	}
	_ = c.rdb.Del(ctx,
		c.listKey(postID),
		c.totalKey(postID),
	).Err()
}

func (c *CommentCache) InvalidateComment(ctx context.Context, commentID int64) {
	if c == nil || c.rdb == nil {
		return
	}
	_ = c.rdb.Del(ctx, c.commentKey(commentID)).Err()
}

func (c *CommentCache) GetOrLoad(
	ctx context.Context,
	postID int64,
	loader func(context.Context) (ids []int64, total int64, replies []model.Reply, err error),
	listTTL time.Duration,
	detailTTL time.Duration,
) ([]int64, int64, error) {

	// 1️⃣ 先查缓存
	if ids, ok, err := c.GetList(ctx, postID); err == nil && ok {
		total, _, _ := c.GetTotal(ctx, postID)
		return ids, total, nil
	}

	key := strconv.FormatInt(postID, 10)

	v, err, _ := c.sf.Do(key, func() (any, error) {

		// double check
		if ids, ok, _ := c.GetList(ctx, postID); ok {
			total, _, _ := c.GetTotal(ctx, postID)
			return pageLoadResult{IDs: ids, Total: total}, nil
		}

		// 分布式锁
		token, locked, e := c.mutex.TryLock(ctx, c.lockKey(postID), c.cfg.LockTTL)
		if e == nil && locked {
			defer func() { _ = c.mutex.Unlock(ctx, c.lockKey(postID), token) }()

			ids, total, replies, err := loader(ctx)
			if err != nil {
				return nil, err
			}

			_ = c.SetList(ctx, postID, ids, listTTL)
			_ = c.SetTotal(ctx, postID, total, listTTL)
			_ = c.SetComments(ctx, replies, detailTTL)

			return pageLoadResult{IDs: ids, Total: total}, nil
		}

		// fallback
		ids, total, replies, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		_ = c.SetList(ctx, postID, ids, listTTL)
		_ = c.SetTotal(ctx, postID, total, listTTL)
		_ = c.SetComments(ctx, replies, detailTTL)

		return pageLoadResult{IDs: ids, Total: total}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	out := v.(pageLoadResult)
	return out.IDs, out.Total, nil
}
