package router

import (
	"MathOverflow/internal/common/middleware"
	"MathOverflow/internal/forum-service/controller"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client,
	forumContrller controller.ForumController,
	postController controller.PostController,
	replyController controller.ReplyController,
	searchController controller.SearchController,
	tokenController controller.TokenController) *gin.Engine {

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
	forum.GET("/file/*filename", forumContrller.DownloadFile) // 无需鉴权
	authForum := forum.Group("")
	authForum.POST("/upload", forumContrller.UploadTempFile) // 临时文件上传

	// 帖子相关
	authPost := authForum.Group("/posts")
	authPost.POST("", postController.CreateNewPost)                       // 创建帖子
	authPost.GET("", postController.GetManyPostData)                      // 获取多条帖子
	authPost.GET("/client-token", tokenController.IssuePostToken)         // 获取发帖token
	authPost.GET("/:postID", postController.GetPostData)                  // 获取单条帖子
	authPost.GET("/:postID/replies", replyController.BatchGetReply)       // 根据条件获取所有回复
	authPost.PATCH("/:postID", postController.UpdateOnePost)              // 更新帖子
	authPost.PATCH("/:postID/certified", postController.SetPostCertified) // 设置帖子精选
	authPost.DELETE("/:postID", postController.DeleteOnePost)             // 删除帖子
	authPost.POST("/like/:postID", postController.LikeOnePost)            // 点赞帖子
	authPost.DELETE("/like/:postID", postController.CancelLikeOnePost)    // 取消点赞
	authPost.POST("/star/:postID", postController.StarOnePost)            // 收藏帖子
	authPost.DELETE("/star/:postID", postController.CancelStarOnePost)    // 取消收藏
	authPost.GET("/starred", postController.GetUserStarredPosts)          // 查询当前用户的收藏列表

	// 回帖相关
	authReply := authForum.Group("/replies")
	authReply.POST("", replyController.CreateNewReply)                     // 创建回复
	authReply.GET("/client-token", tokenController.IssueReplyToken)        // 获取回帖token
	authReply.GET("/:replyID", replyController.GetOneReply)                // 获取单条回复
	authReply.PATCH("/:replyID", replyController.UpdateOneReply)           // 更新单条回复
	authReply.DELETE("/:replyID", replyController.DeleteOneReply)          // 删除单条回复
	authReply.POST("/like/:replyID", replyController.LikeOneReply)         // 点赞回复
	authReply.DELETE("/like/:replyID", replyController.CancelLikeOneReply) // 取消点赞
	authReply.PATCH("/status", replyController.ChangeReplyStatus)          // 修改回复的状态

	// 搜索
	search := authForum.Group("/search")
	search.POST("", searchController.SearchPosts) // 搜索帖子:tags和关键词共用
	return r
}
