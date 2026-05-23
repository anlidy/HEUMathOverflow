package main

import (
	"MathOverflow/common/client"
	"MathOverflow/common/config"
	common "MathOverflow/common/model"
	"MathOverflow/common/outbox"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/cache"
	"MathOverflow/services/forum/internal/consumer"
	handler "MathOverflow/services/forum/internal/handler"
	"MathOverflow/services/forum/internal/model"
	"MathOverflow/services/forum/internal/repository"
	router "MathOverflow/services/forum/internal/router"
	"MathOverflow/services/forum/internal/service"
	"MathOverflow/services/forum/internal/worker"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

func main() {
	utils.InitLogger("forum-service")

	dir, _ := os.Getwd()
	utils.Logger().WithField("working_dir", dir).Info("forum-service starting")
	// 加载配置

	cfg, err := config.LoadConfig("config/forum.yaml")
	if err != nil {
		utils.Logger().WithError(err).Fatal("failed to load forum.yaml")
	}

	// 初始化RabbitMQ
	rabbit, err := client.InitRabbitMQ(cfg.RabbitMQ)
	if err != nil {
		panic(err)
	}
	defer rabbit.Close()

	// 初始化redis数据库
	rdb, err := client.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}

	// 初始化pgsql数据库
	pg, err := client.InitPostgres(cfg.Postgres)
	if err != nil {
		panic(err)
	}
	if err := migrateForumLegacyData(pg); err != nil {
		panic(err)
	}
	// 迁移表
	if err := pg.AutoMigrate(&model.Post{}, &model.PostLike{}, &model.PostStar{}); err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.Reply{}, &model.ReplyLike{}); err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.UserSnapshot{}, &common.OutboxMessage{}); err != nil {
		panic(err)
	}
	if err := pg.AutoMigrate(&model.ChatSession{}, &model.ChatMessage{}); err != nil {
		panic(err)
	}
	// 添加triggers
	InitTriggers(pg)

	// 初始化minio数据库
	mc, err := client.InitMinIO(cfg.Minio)
	if err != nil {
		panic(err)
	}

	es, err := client.InitESClient(cfg)
	if err != nil {
		utils.Logger().WithError(err).Fatal("init es failed")
	}

	// 初始化服务
	fileRepo := repository.NewFileRepository(mc)
	postRepo := repository.NewPostRepository(pg, rdb)
	replyRepo := repository.NewReplyRepository(pg, rdb)
	userSnapshotRepo := repository.NewUserSnapshotRepository(pg)
	tokenRepo := repository.NewClientTokenRepository(rdb)
	chatRepo := repository.NewChatRepository(pg)
	postCache := cache.NewPostCache(rdb, cache.PostCacheConfig{})
	commentCache := cache.NewCommentCache(rdb, cache.CommentCacheConfig{})

	forumServ := service.NewForumService(cfg, fileRepo)
	postServ := service.NewPostService(cfg, rabbit, postRepo, replyRepo, fileRepo, userSnapshotRepo, tokenRepo, postCache)
	replyServ := service.NewReplyService(cfg, postRepo, replyRepo, fileRepo, userSnapshotRepo, tokenRepo, commentCache, postCache, postCache)
	searchServ := service.NewSearchService(cfg, es, postRepo, userSnapshotRepo)
	chatServ := service.NewChatService(cfg, chatRepo)

	forumHandler := handler.NewForumHandler(forumServ)
	postHandler := handler.NewPostHandler(postServ)
	replyHandler := handler.NewReplyHandler(replyServ)
	searchHandler := handler.NewSearchHandler(searchServ)
	tokenHandler := handler.NewTokenHandler(tokenRepo)
	chatHandler := handler.NewChatHandler(chatServ)

	// 初始化worker
	postWorker := worker.NewPostWorker(pg, rdb, rabbit)
	replyWorker := worker.NewReplyWorker(pg, rdb, rabbit)
	hotWarmWorker := worker.NewHotPostWarmWorker(rdb, postRepo, postCache)
	outboxMessageWorker := outbox.NewMessageWorker(pg, rabbit, "forum-post-events", "Forum-Outbox-Worker", func(message common.OutboxMessage, mq *client.RabbitMQClient) string {
		workerCount := mq.PostWorkerCount
		if workerCount <= 0 {
			workerCount = 1
		}
		return fmt.Sprintf("%s.%d", message.Topic, message.ResourceID%workerCount)
	})
	ragOutboxMessageWorker := outbox.NewMessageWorker(pg, rabbit, "rag-events", "Rag-Outbox-Worker", func(message common.OutboxMessage, mq *client.RabbitMQClient) string {
		workerCount := mq.RagWorkerCount
		if workerCount <= 0 {
			workerCount = 1
		}
		return fmt.Sprintf("%s.%d", message.Topic, message.ResourceID%workerCount)
	})

	// 并发启动
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		interval := 5 * time.Second
		var count int64 = 500
		go postWorker.WriteBackPostWorker(ctx, interval, count)
		return nil
	})
	eg.Go(func() error {
		interval := 5 * time.Second
		var count int64 = 500
		go replyWorker.WriteBackReplyWorker(ctx, interval, count)
		return nil
	})
	eg.Go(func() error {
		interval := 30 * time.Second
		var topN int64 = 100
		go hotWarmWorker.Run(ctx, interval, topN)
		return nil
	})
	// 定时转发 outbox_messages 中待发送的帖子事件
	eg.Go(func() error {
		interval := 3 * time.Second
		go outboxMessageWorker.Run(ctx, interval, 100)
		return nil
	})
	eg.Go(func() error {
		interval := 3 * time.Second
		go ragOutboxMessageWorker.Run(ctx, interval, 100)
		return nil
	})

	eg.Go(func() error {
		return consumer.StartUserConsumer(ctx, rabbit, userSnapshotRepo)
	})

	eg.Go(func() error {
		r := router.SetupRouter(rdb, forumHandler, postHandler, replyHandler, searchHandler, tokenHandler, chatHandler)
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
			Handler: r, // Gin 实现了 http.Handler 接口
		}
		go func() {
			<-ctx.Done()
			srv.Shutdown(context.Background())
			utils.Logger().Info("Forum Service Exited.")
		}()
		utils.Logger().Info("Forum Service is running...")
		return srv.ListenAndServe()
	})

	if err := eg.Wait(); err != nil {
		utils.Logger().WithError(err).Error("Forum服务异常退出")
	}

}

func InitTriggers(db *gorm.DB) {
	// 插入回帖时更新原帖的回复时间
	sql := `
        CREATE OR REPLACE FUNCTION update_post_after_insert_reply_func()
        RETURNS TRIGGER AS $$
        BEGIN
            UPDATE posts
            SET last_reply_at = NEW.created_at
            WHERE id = NEW.post_id;
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;

        DROP TRIGGER IF EXISTS update_post_after_insert_reply ON replies;

        CREATE TRIGGER update_post_after_insert_reply
        AFTER INSERT ON replies
        FOR EACH ROW
        EXECUTE FUNCTION update_post_after_insert_reply_func();
        `

	if err := db.Exec(sql).Error; err != nil {
		panic(err)
	}

	// 保证 post_stars 幂等（去重 + 建唯一约束），避免重复收藏导致统计漂移。
	// 说明：Postgres 的 UNIQUE 对 NULL 不敏感（允许多条 NULL），因此加 partial unique index 仅约束有效 post_id。
	sql = `
		-- 删除重复收藏（保留 created_at 最早的一条）
		WITH ranked AS (
			SELECT
				id,
				ROW_NUMBER() OVER (PARTITION BY post_id, user_id ORDER BY created_at ASC) AS rn
			FROM post_stars
			WHERE post_id IS NOT NULL
		)
		DELETE FROM post_stars
		WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

		-- 建立唯一索引：同一用户同一帖子最多一条收藏记录
		CREATE UNIQUE INDEX IF NOT EXISTS idx_post_stars_post_user
		ON post_stars (post_id, user_id)
		WHERE post_id IS NOT NULL;
	`
	if err := db.Exec(sql).Error; err != nil {
		panic(err)
	}
}

func migrateForumLegacyData(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE posts ADD COLUMN IF NOT EXISTS is_certified boolean`,
		`UPDATE posts SET is_certified = false WHERE is_certified IS NULL`,
		`ALTER TABLE posts ALTER COLUMN is_certified SET DEFAULT false`,
		`ALTER TABLE posts ALTER COLUMN is_certified SET NOT NULL`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}
