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
	userInfo, replyData, syserr := rc.replyServ.GetOneReply(ctx, replyID)
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

// 获取分页回帖
func (rc *ReplyController) BatchGetReply(c *gin.Context) {
	var replyIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "获取失败,无效的帖子id", "code": http.StatusBadRequest, "data": nil})
		return
	}
	// 获取偏移量
	var offsetStr, limitStr = c.Query("offset"), c.Query("limit")
	offset, _ := strconv.Atoi(offsetStr) // 默认取0
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 20 // 默认取20条
	}
	var ctx = c.Request.Context()
	multiData, syserr := rc.replyServ.GetManyReplies(ctx, postID, offset, limit)
	switch syserr {
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到指定帖子的回帖", "code": http.StatusNotFound, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求成功", "code": http.StatusOK, "data": multiData})
}

// 更新回帖
func (rc *ReplyController) UpdateOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest})
		return
	}
	var req request.ReplyUpdate
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新失败,数据格式有误", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := rc.replyServ.UpdateOneReply(ctx, userID, replyID, req)
	switch syserr {
	case syserror.PermissionDeniedError:
		c.JSON(http.StatusUnauthorized, gin.H{"message": "您无权修改他人的回帖", "code": http.StatusUnauthorized})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要修改的回帖", "code": http.StatusNotFound})
		return
	case syserror.ResourceExpiredError:
		c.JSON(http.StatusGone, gin.H{"message": "附件资源已过期,请重新上传", "code": http.StatusGone})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "code": http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "回帖更新成功", "code": http.StatusOK})
}

// 删除一条回帖
func (rc *ReplyController) DeleteOneReply(c *gin.Context) {
	replyIDStr := c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	var role = c.GetInt("role")
	syserr := rc.replyServ.DeleteOneReply(ctx, replyID, userID, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		c.JSON(http.StatusUnauthorized, gin.H{"message": "您无权删除他人的回帖", "code": http.StatusUnauthorized})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要删除的回帖", "code": http.StatusNotFound})
		return
	case syserror.NetworkError:
		log.Printf("[%s] %v\n", rc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "code": http.StatusOK})
}

// 点赞接口
// 点赞帖子
func (pc *ReplyController) LikeOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.replyServ.LikeOneReply(ctx, replyID, userID)
	switch syserr {
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "您已点过赞", "code": http.StatusConflict})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "点赞成功", "code": http.StatusOK})
}

// 取消点赞
func (pc *ReplyController) CancelLikeOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.replyServ.CancelLikeOneReply(ctx, replyID, userID)
	switch syserr {
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "未点赞过该回帖", "code": http.StatusNotFound})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功", "code": http.StatusOK})
}

// 查询当前用户是否点赞过回复
func (pc *ReplyController) GetReplyLikeStatus(c *gin.Context) {
	replyIDStr := c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的回帖id", "code": http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	liked, syserr := pc.replyServ.HasLikedReply(ctx, replyID, userID)
	if syserr == syserror.InternalError {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "查询成功", "code": http.StatusOK, "data": gin.H{"liked": liked}})
}
