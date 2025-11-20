package router

import (
	"MathOverflow/internal/common/middleware"
	"MathOverflow/internal/forum-service/controller"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRouter(rdb *redis.Client,
	forumContrller controller.ForumController,
	postController controller.PostController,
	replyController controller.ReplyController) *gin.Engine {

	r := gin.Default()
	r.Use(middleware.CorsMiddleware([]string{"http://localhost:3000", "https://math-overflow.edu"}))
	api := r.Group("/api")
	v1 := api.Group("/v1")
	forum := v1.Group("/forum")

	// 论坛
	authForum := forum.Group("")
	authForum.Use(middleware.AuthMiddleware(rdb))
	authForum.POST("/upload", forumContrller.UploadTempFile)
	forum.GET("/file/*filename", forumContrller.DownloadFile) // 无需鉴权

	// 帖子相关
	authPost := authForum.Group("/posts")
	authPost.POST("", postController.CreateNewPost)
	authPost.GET("/:postID", postController.GetPostData)
	authPost.GET("/:postID/replies", replyController.BatchGetReply) // 根据条件获取所有回帖

	// 回帖相关
	authReply := authForum.Group("/replies")
	authReply.POST("", replyController.CreateNewReply)
	authReply.GET("/:replyID", replyController.GetOneReply)
	return r
}
