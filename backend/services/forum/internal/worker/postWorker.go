package worker

import (
	"MathOverflow/common/client"
	"MathOverflow/common/event"
	common "MathOverflow/common/model"
	"MathOverflow/common/outbox"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/model"
	"MathOverflow/services/forum/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostWorker struct {
	pg       *gorm.DB
	rdb      *redis.Client
	postRepo repository.PostRepo
	outbox   outbox.Store
	name     string
}

func NewPostWorker(pg *gorm.DB, rdb *redis.Client, mq *client.RabbitMQClient) PostWorker {
	_ = mq
	return PostWorker{pg: pg, rdb: rdb, postRepo: repository.NewPostRepository(pg, rdb), outbox: outbox.NewStore(pg), name: "Post-Worker"}
}

// redis读取postStat
func (w *PostWorker) getPostStat(ctx context.Context, key string) (*model.PostStat, error) {
	var postID int64
	_, err := fmt.Sscanf(key, "post:%d:stat", &postID)
	if err != nil {
		_, err = fmt.Sscanf(key, "post:%d:stat:processing", &postID)
	}
	if err != nil {
		return nil, fmt.Errorf("parse error:%v", err)
	}

	m, err := w.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis get failed: %w", err)
	}

	// 如果 key 不存在，HGETALL 返回空 map
	pc := &model.PostStat{
		PostID: postID, Views: 0, Likes: 0, Stars: 0,
	}

	// 解析字段（不存在的字段保持 0）
	if v, ok := m["views"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			pc.Views = n
		}
	}
	if v, ok := m["likes"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			pc.Likes = n
		}
	}
	if v, ok := m["stars"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			pc.Stars = n
		}
	}
	if v, ok := m["replies"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			pc.Replies = n
		}
	}
	return pc, nil
}

// 从redis周期性回写数据库
func (w *PostWorker) WriteBackPost(ctx context.Context, scanCursor uint64, count int64) uint64 {
	var keys []string
	var err error
	keys, scanCursor, err = w.rdb.Scan(ctx, scanCursor, "post:*:stat", count).Result()
	if err != nil {
		return scanCursor
	}
	wroteAny := false
	for _, key := range keys {
		processingKey, moved, err := w.postRepo.MovePostStatToProcessing(ctx, key)
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", key).WithError(err).Error("move post stat to processing failed")
			continue
		}
		if !moved {
			continue
		}

		// 从redis读取冻结后的增量快照
		stat, err := w.getPostStat(ctx, processingKey)
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", processingKey).WithError(err).Error("get post stat from redis failed")
			continue
		}
		if stat == nil || stat.PostID == 0 || stat.Views == 0 && stat.Likes == 0 && stat.Stars == 0 && stat.Replies == 0 {
			_ = w.postRepo.DeletePostStatKey(ctx, processingKey)
			continue
		}
		// 批量写回数据库
		err = w.pg.Exec(
			`UPDATE posts SET views = views + ?, likes = likes + ?, stars = stars + ?, replies = replies + ? WHERE id = ?`,
			stat.Views, stat.Likes, stat.Stars, stat.Replies, stat.PostID,
		).Error
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithError(err).Error("write back post stat to postgres failed")
			continue
		}
		wroteAny = true

		post, err := w.postRepo.FindPostByID(stat.PostID)
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("post_id", stat.PostID).WithError(err).Error("load latest post stat snapshot failed")
			continue
		}

		favors, totalScore := utils.CalcPostScores(&common.PostStat{
			Status:    int(post.Status),
			CreatedAt: post.CreatedAt,
			Views:     post.Views,
			Likes:     post.Likes,
			Stars:     post.Stars,
			Replies:   post.Replies,
		})

		// 发布PostStat绝对值快照事件
		var evt = event.ForumPostPayload{
			PostID:     post.ID,
			Status:     int(post.Status),
			CreatedAt:  post.CreatedAt,
			Views:      post.Views,
			Likes:      post.Likes,
			Stars:      post.Stars,
			Replies:    post.Replies,
			Favors:     favors,
			TotalScore: totalScore,
		}
		eventBody, err := json.Marshal(event.ForumPostEvent{
			Type:      event.ForumPostStatUpdated,
			Payload:   evt,
			CreatedAt: time.Now(),
		})
		if err != nil {
			utils.Logger().WithField("worker", w.name).WithField("post_id", post.ID).WithError(err).Error("marshal post stat event failed")
			continue
		}
		if err := w.outbox.AddMessageInTx(ctx, w.pg, "post", post.ID, string(event.ForumPostStatUpdated), "forum-post-events", json.RawMessage(eventBody)); err != nil {
			utils.Logger().WithField("worker", w.name).WithField("post_id", post.ID).WithError(err).Error("save post stat event to outbox failed")
			continue
		}

		// 删除已处理的冻结快照
		if err := w.postRepo.DeletePostStatKey(ctx, processingKey); err != nil {
			utils.Logger().WithField("worker", w.name).WithField("key", processingKey).WithError(err).Error("delete post stat from redis failed")
			continue
		}
	}
	if wroteAny {
		_ = w.rdb.Incr(ctx, repository.HottestPostsCacheVersionKey).Err()
	}
	return scanCursor
}

// 定期从redis回写post
func (w *PostWorker) WriteBackPostWorker(ctx context.Context, interval time.Duration, count int64) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	utils.Logger().WithField("worker", w.name).WithField("interval", interval.String()).Info("post worker started")
	var scanCursor uint64 = 0
	for {
		select {
		case <-ticker.C:
			scanCursor = w.WriteBackPost(ctx, scanCursor, count)
		case <-ctx.Done():
			utils.Logger().WithField("worker", w.name).Info("post worker stopped")
			return
		}
	}
}
