package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"MathOverflow/services/forum/internal/model"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type PostCacheConfig struct {
	L1Size            int
	VisitWindow       time.Duration
	HotVisitThreshold int64
	HotTTLMin         time.Duration
	HotTTLMax         time.Duration
	LockTTL           time.Duration
}

type PostCache struct {
	rdb   *redis.Client
	l1    *LRU[int64, model.Post]
	mutex *RedisMutex
	sf    singleflight.Group
	cfg   PostCacheConfig
}

type Hotness interface {
	IsHot(ctx context.Context, postID int64) (bool, int64, error)
}

func NewPostCache(rdb *redis.Client, cfg PostCacheConfig) *PostCache {
	if cfg.L1Size <= 0 {
		cfg.L1Size = 1000
	}
	if cfg.VisitWindow <= 0 {
		cfg.VisitWindow = 60 * time.Second
	}
	if cfg.HotVisitThreshold <= 0 {
		cfg.HotVisitThreshold = 10
	}
	if cfg.HotTTLMin <= 0 {
		cfg.HotTTLMin = 5 * time.Minute
	}
	if cfg.HotTTLMax <= 0 {
		cfg.HotTTLMax = 10 * time.Minute
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 5 * time.Second
	}
	return &PostCache{
		rdb:   rdb,
		l1:    NewLRU[int64, model.Post](cfg.L1Size),
		mutex: NewRedisMutex(rdb),
		cfg:   cfg,
	}
}

func (c *PostCache) visitKey(postID int64) string { return fmt.Sprintf("post:visit:%d", postID) }
func (c *PostCache) postKey(postID int64) string  { return fmt.Sprintf("post:%d", postID) }
func (c *PostCache) lockKey(postID int64) string  { return fmt.Sprintf("lock:post:%d", postID) }
func (c *PostCache) hotBucketKey(t time.Time) string {
	return fmt.Sprintf("post:hot:%s", t.Format("2006010215")) // 小时桶
}

// RecordVisit implements: INCR + EXPIRE 60s.
func (c *PostCache) RecordVisit(ctx context.Context, postID int64) (int64, bool, error) {
	if c == nil || c.rdb == nil {
		return 0, false, nil
	}

	now := time.Now()
	bucketKey := c.hotBucketKey(now)
	pipe := c.rdb.Pipeline()
	// 写入当前小时桶
	pipe.ZIncrBy(ctx, bucketKey, 1, strconv.FormatInt(postID, 10))
	// 桶TTL（略大于24h，避免边界问题）
	pipe.Expire(ctx, bucketKey, 25*time.Hour)
	// 短期热点计数（60s窗口）
	incrCmd := pipe.Incr(ctx, c.visitKey(postID))
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, false, err
	}

	n := incrCmd.Val()
	// 只在第一次设置 visitKey TTL（固定窗口）
	if n == 1 {
		_ = c.rdb.Expire(ctx, c.visitKey(postID), c.cfg.VisitWindow).Err()
	}

	return n, n >= c.cfg.HotVisitThreshold, nil
}

// 聚合24h桶内数据
func (c *PostCache) GetHot24h(ctx context.Context, topN int64) ([]string, error) {
	now := time.Now()
	keys := make([]string, 0, 24)

	// 合并过去24个桶
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(i) * time.Hour)
		keys = append(keys, c.hotBucketKey(t))
	}

	// 聚合
	store := redis.ZStore{Keys: keys}
	zs, err := c.rdb.ZUnionWithScores(ctx, store).Result()
	if err != nil {
		return nil, err
	}

	// 按 score 排序（降序）
	sort.Slice(zs, func(i, j int) bool {
		return zs[i].Score > zs[j].Score
	})

	// 取 TopN
	n := min(int(topN), len(zs))
	res := make([]string, 0, n)
	for i := 0; i < n; i++ {
		res = append(res, zs[i].Member.(string))
	}
	return res, nil
}

// IsHot checks visit counter without increasing it.
func (c *PostCache) IsHot(ctx context.Context, postID int64) (bool, int64, error) {
	if c == nil || c.rdb == nil {
		return false, 0, nil
	}
	n, err := c.rdb.Get(ctx, c.visitKey(postID)).Int64()
	if err == redis.Nil {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return n >= c.cfg.HotVisitThreshold, n, nil
}

func (c *PostCache) Get(ctx context.Context, postID int64) (model.Post, bool, error) {
	var zero model.Post
	if c == nil {
		return zero, false, nil
	}
	if c.l1 != nil {
		if v, ok := c.l1.Get(postID); ok {
			return v, true, nil
		}
	}
	if c.rdb == nil {
		return zero, false, nil
	}

	key := c.postKey(postID)
	pipe := c.rdb.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		if err == redis.Nil {
			return zero, false, nil
		}
		return zero, false, err
	}

	raw, err := getCmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			return zero, false, nil
		}
		return zero, false, err
	}
	var post model.Post
	if err := json.Unmarshal(raw, &post); err != nil {
		return zero, false, err
	}
	ttl := ttlCmd.Val()
	if ttl > 0 && c.l1 != nil {
		l1TTL := min(ttl, 1*time.Minute) // 防止ttl过长导致不新鲜
		c.l1.Add(postID, post, l1TTL)
	}
	return post, true, nil
}

func (c *PostCache) SetHot(ctx context.Context, postID int64, post model.Post) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	ttl := c.cfg.HotTTLMin
	if c.cfg.HotTTLMax > c.cfg.HotTTLMin {
		span := c.cfg.HotTTLMax - c.cfg.HotTTLMin
		off := time.Duration(time.Now().UnixNano() % int64(span))
		ttl = c.cfg.HotTTLMin + off
	}
	ttl = ClampMin(Jitter(ttl, 0.2), 30*time.Second)

	raw, err := json.Marshal(post)
	if err != nil {
		return err
	}
	if err := c.rdb.Set(ctx, c.postKey(postID), raw, ttl).Err(); err != nil {
		return err
	}
	if c.l1 != nil { // redis -> l1双写
		c.l1.Add(postID, post, ttl)
	}
	return nil
}

func (c *PostCache) Invalidate(ctx context.Context, postID int64) {
	if c == nil {
		return
	}
	if c.l1 != nil {
		c.l1.Remove(postID)
	}
	if c.rdb != nil {
		_ = c.rdb.Del(ctx, c.postKey(postID)).Err()
	}
}

// GetOrLoad loads from cache or uses loader with stampede protection.
// Only when hot==true the loaded post will be cached.
func (c *PostCache) GetOrLoad(ctx context.Context, postID int64, hot bool, loader func(context.Context) (model.Post, error)) (model.Post, error) {
	if c == nil {
		return loader(ctx)
	}
	if p, ok, err := c.Get(ctx, postID); err == nil && ok {
		return p, nil
	}
	v, err, _ := c.sf.Do(strconv.FormatInt(postID, 10), func() (any, error) {
		if p, ok, e := c.Get(ctx, postID); e == nil && ok {
			return p, nil
		}

		// cross-instance mutex
		token, locked, e := c.mutex.TryLock(ctx, c.lockKey(postID), c.cfg.LockTTL)
		if e == nil && locked {
			defer func() { _ = c.mutex.Unlock(ctx, c.lockKey(postID), token) }()
			p, e := loader(ctx)
			if e != nil {
				return model.Post{}, e
			}
			if hot { // 热点缓存下来
				_ = c.SetHot(ctx, postID, p)
			}
			return p, nil
		}

		// wait for others to fill
		deadline := time.Now().Add(300 * time.Millisecond)
		for time.Now().Before(deadline) {
			if p, ok, e2 := c.Get(ctx, postID); e2 == nil && ok {
				return p, nil
			}
			time.Sleep(30 * time.Millisecond)
		}

		p, e2 := loader(ctx)
		if e2 != nil {
			return model.Post{}, e2
		}
		if hot {
			_ = c.SetHot(ctx, postID, p)
		}
		return p, nil
	})
	if err != nil {
		return model.Post{}, err
	}
	return v.(model.Post), nil
}
