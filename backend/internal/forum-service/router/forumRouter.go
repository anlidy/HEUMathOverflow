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
	// 无需鉴权
	forum.GET("/file/*filename", forumContrller.DownloadFile)

	// 鉴权--论坛
	authForum := forum.Group("")
	authForum.Use(middleware.AuthMiddleware(rdb))
	authForum.POST("/upload", forumContrller.UploadTempFile)

	// 鉴权--帖子相关
	authPost := authForum.Group("/post")
	authPost.GET("/:postID", postController.GetPostData)
	authPost.POST("/createPost", postController.CreateNewPost)

	// 鉴权--回帖相关
	authReply := authForum.Group("/reply")
	authReply.GET("/:replyID", replyController.GetOneReply)
	authReply.POST("/createReply", replyController.CreateNewReply)
	return r
}
