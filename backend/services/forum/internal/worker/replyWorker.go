package worker

import (
	"MathOverflow/common/client"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/model"
	"MathOverflow/services/forum/internal/repository"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ReplyWorker struct {
	pg        *gorm.DB
	rdb       *redis.Client
	replyRepo repository.ReplyRepo
	name      string
}

func NewReplyWorker(pg *gorm.DB, rdb *redis.Client, mq *client.RabbitMQClient) ReplyWorker {
	_ = mq
	return ReplyWorker{pg: pg, rdb: rdb, replyRepo: repository.NewReplyRepository(pg, rdb), name: "Reply-Worker"}
}

func (w *ReplyWorker) getReplyStat(ctx context.Context, key string) (*model.ReplyStat, error) {
	var replyID int64
	_, err := fmt.Sscanf(key, "reply:%d:stat", &replyID)
	if err != nil {
		_, err = fmt.Sscanf(key, "reply:%d:stat:processing", &replyID)
	}
	if err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	m, err := w.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}
	rc := &model.ReplyStat{ReplyID: replyID, Likes: 0}
	if v, ok := m["likes"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			rc.Likes = n
		}
	}
	return rc, nil
}

func (w *ReplyWorker) WriteBackReply(ctx context.Context, scanCursor uint64, count int64) uint64 {
	keys, nextCursor, err := w.rdb.Scan(ctx, scanCursor, "reply:*:stat", count).Result()
	if err != nil {
		return scanCursor
	}
	for _, key := range keys {
		processingKey, moved, err := w.replyRepo.MoveReplyStatToProcessing(ctx, key)
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", key).WithError(err).Error("move reply stat to processing failed")
			continue
		}
		if !moved {
			continue
		}
		stat, err := w.getReplyStat(ctx, processingKey)
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", processingKey).WithError(err).Error("get reply stat from redis failed")
			continue
		}
		if stat == nil || stat.ReplyID == 0 || stat.Likes == 0 {
			_ = w.replyRepo.DeleteReplyStatKey(ctx, processingKey)
			continue
		}
		err = w.pg.Exec(`UPDATE replies SET likes = likes + ? WHERE id = ?`, stat.Likes, stat.ReplyID).Error
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithError(err).Error("write back reply stat to postgres failed")
			continue
		}
		if err := w.replyRepo.DeleteReplyStatKey(ctx, processingKey); err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", processingKey).WithError(err).Error("delete reply stat from redis failed")
			continue
		}
	}
	return nextCursor
}

func (w *ReplyWorker) WriteBackReplyWorker(ctx context.Context, interval time.Duration, count int64) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	utils.Logger().WithField("worker", w.name).WithField("interval", interval.String()).Info("reply worker started")
	var scanCursor uint64
	for {
		select {
		case <-ticker.C:
			scanCursor = w.WriteBackReply(ctx, scanCursor, count)
		case <-ctx.Done():
			utils.Logger().WithField("worker", w.name).Info("reply worker stopped")
			return
		}
	}
}
