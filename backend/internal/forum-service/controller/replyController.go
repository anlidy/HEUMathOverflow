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

type ReplyController struct {
	replyServ service.ReplyService
	name      string
}

func NewReplyController(replyServ service.ReplyService) ReplyController {
	return ReplyController{replyServ: replyServ, name: "Reply-Controller"}
}

// 创建回帖
func (rc *ReplyController) CreateNewReply(c *gin.Context) {
	var req request.ReplyCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "发布失败,数据格式有误", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	replyID, syserr := rc.replyServ.CreateNewReply(ctx, userID, req)
	switch syserr {
	case syserror.ResourceExpiredError:
		c.JSON(http.StatusGone, gin.H{"message": "附件已失效,请重新上传", "code": http.StatusGone, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "回帖发布失败", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "回帖已经发布", "code": http.StatusConflict, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "回帖发布成功", "code": http.StatusOK, "data": gin.H{"reply_id": strconv.FormatInt(replyID, 10)}})
}

// 获取回帖
func (rc *ReplyController) GetOneReply(c *gin.Context) {
	replyIDStr := c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	userInfo, replyData, syserr := rc.replyServ.GetOneReply(ctx, userID, replyID)
	switch syserr {
	case syserror.NetworkError:
		log.Printf("[%s] gRPC网络异常: %v\n", rc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		log.Printf("[%s] %v\n", rc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到该用户的回帖", "code": http.StatusNotFound, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求成功", "code": http.StatusOK, "data": gin.H{
		"user_info": userInfo,
		"post_data": replyData,
	}})
}
