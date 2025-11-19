package controller

import (
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postServ service.PostService
	name     string
}

func NewPostController(postServ service.PostService) PostController {
	return PostController{postServ: postServ, name: "Post-Controller"}
}

// 创建帖子
func (pc *PostController) CreateNewPost(c *gin.Context) {
	var req request.PostCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "发布失败,数据格式有误", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	postID, err := pc.postServ.CreateNewPost(ctx, userID, req)
	switch err {
	case syserror.ResourceExpiredError:
		c.JSON(http.StatusGone, gin.H{"message": "图片附件已过期,请重新上传", "code": http.StatusGone, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "帖子发布失败", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "帖子已经发布", "code": http.StatusConflict, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "code": http.StatusOK, "data": gin.H{"post_id": strconv.FormatInt(postID, 10)}})
}

// 获取帖子信息
func (pc *PostController) GetPostData(c *gin.Context) {
	var ctx = c.Request.Context()
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var userID = c.GetInt64("userID")
	userInfo, postData, syserr := pc.postServ.GetOnePost(ctx, userID, postID)
	switch syserr {
	case syserror.NetworkError:
		log.Printf("[%s] gRPC网络异常: %v\n", pc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		log.Printf("[%s] %v\n", pc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到该用户的帖子", "code": http.StatusNotFound, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求成功", "code": http.StatusOK, "data": gin.H{
		"user_info": userInfo,
		"post_data": postData,
	}})
}
