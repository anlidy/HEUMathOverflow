package router

import (
	"MathOverflow/common/middleware"
	handler "MathOverflow/services/forum/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client,
	forumHandler handler.ForumHandler,
	postHandler handler.PostHandler,
	replyHandler handler.ReplyHandler,
	searchHandler handler.SearchHandler,
	tokenHandler handler.TokenHandler) *gin.Engine {

	r := gin.Default()
	r.Use(
		middleware.RequestIDMiddleware(),
		middleware.MetricsMiddleware("forum-service"),
		middleware.LoggingMiddleware(),
	)

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	forum := r.Group("/forum")

	// 论坛
	forum.GET("/file/*filename", forumHandler.DownloadFile) // 无需鉴权
	authForum := forum.Group("")
	authForum.POST("/upload", forumHandler.UploadTempFile) // 临时文件上传

	// 帖子相关
	authPost := authForum.Group("/posts")
	authPost.POST("", postHandler.CreateNewPost)                       // 创建帖子
	authPost.GET("", postHandler.GetManyPostData)                      // 获取多条帖子
	authPost.GET("/client-token", tokenHandler.IssuePostToken)         // 获取发帖token
	authPost.GET("/:postID", postHandler.GetPostData)                  // 获取单条帖子
	authPost.GET("/:postID/replies", replyHandler.BatchGetReply)       // 根据条件获取所有回复
	authPost.PATCH("/:postID", postHandler.UpdateOnePost)              // 更新帖子
	authPost.PATCH("/:postID/certified", postHandler.SetPostCertified) // 设置帖子精选
	authPost.DELETE("/:postID", postHandler.DeleteOnePost)             // 删除帖子
	authPost.POST("/like/:postID", postHandler.LikeOnePost)            // 点赞帖子
	authPost.DELETE("/like/:postID", postHandler.CancelLikeOnePost)    // 取消点赞
	authPost.POST("/star/:postID", postHandler.StarOnePost)            // 收藏帖子
	authPost.DELETE("/star/:postID", postHandler.CancelStarOnePost)    // 取消收藏
	authPost.GET("/starred", postHandler.GetUserStarredPosts)          // 查询当前用户的收藏列表

	// 回帖相关
	authReply := authForum.Group("/replies")
	authReply.POST("", replyHandler.CreateNewReply)                     // 创建回复
	authReply.GET("/client-token", tokenHandler.IssueReplyToken)        // 获取回帖token
	authReply.GET("/:replyID", replyHandler.GetOneReply)                // 获取单条回复
	authReply.PATCH("/:replyID", replyHandler.UpdateOneReply)           // 更新单条回复
	authReply.DELETE("/:replyID", replyHandler.DeleteOneReply)          // 删除单条回复
	authReply.POST("/like/:replyID", replyHandler.LikeOneReply)         // 点赞回复
	authReply.DELETE("/like/:replyID", replyHandler.CancelLikeOneReply) // 取消点赞
	authReply.PATCH("/status", replyHandler.ChangeReplyStatus)          // 修改回复的状态

	// 搜索
	search := authForum.Group("/search")
	search.POST("", searchHandler.SearchPosts) // 搜索帖子:tags和关键词共用
	return r
}
