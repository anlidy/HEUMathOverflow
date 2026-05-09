package worker

import (
	"context"
	"strconv"
	"time"

	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/cache"
	"MathOverflow/services/forum/internal/repository"

	"github.com/redis/go-redis/v9"
)

type HotPostWarmWorker struct {
	rdb       *redis.Client
	postRepo  repository.PostRepo
	postCache *cache.PostCache
	name      string
}

func NewHotPostWarmWorker(rdb *redis.Client, postRepo repository.PostRepo, postCache *cache.PostCache) *HotPostWarmWorker {
	return &HotPostWarmWorker{
		rdb:       rdb,
		postRepo:  postRepo,
		postCache: postCache,
		name:      "HotPostWarm-Worker",
	}
}

func (w *HotPostWarmWorker) Run(ctx context.Context, interval time.Duration, topN int64) {
	if w == nil || w.rdb == nil || w.postRepo == nil || w.postCache == nil {
		return
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	if topN <= 0 {
		topN = 100
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	utils.Logger().WithField("worker", w.name).WithField("interval", interval.String()).WithField("topN", topN).Info("hot post warm worker started")

	for {
		select {
		case <-ticker.C:
			// zset支持分数排序/topN获取,知道哪些是热点
			members, err := w.postCache.GetHot24h(ctx, topN)
			if err != nil && err != redis.Nil {
				utils.Logger().WithField("worker", w.name).WithError(err).Warn("load hot zset failed")
				continue
			}
			if len(members) == 0 {
				continue
			}

			hotIDs := make([]int64, 0, len(members))
			for _, m := range members {
				id, err := strconv.ParseInt(m, 10, 64)
				if err != nil || id <= 0 {
					continue
				}
				// visit只能判断是不是热点,而无法找出哪些是热点,故需要zset先进行排序获取热点
				hot, _, err := w.postCache.IsHot(ctx, id)
				if err == nil && hot {
					hotIDs = append(hotIDs, id)
				}
			}

			// 预热热点帖子详情缓存
			for _, id := range hotIDs {
				if _, ok, _ := w.postCache.Get(ctx, id); ok { //如果缓存已经存在则跳过
					continue
				}
				post, err := w.postRepo.FindPostByID(id)
				if err != nil {
					continue
				}
				_ = w.postCache.SetHot(ctx, id, post)
			}

		case <-ctx.Done():
			utils.Logger().WithField("worker", w.name).Info("hot post warm worker stopped")
			return
		}
	}
}
