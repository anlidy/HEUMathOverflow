package controller

import (
	"MathOverflow/internal/common/api"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/model"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/service"
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
		api.JSON(c).Code(http.StatusBadRequest).Message("发布失败,数据格式有误").Send()
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	replyID, syserr := rc.replyServ.CreateNewReply(ctx, userID, req)
	switch syserr {
	case syserror.InProgressError:
		api.JSON(c).Code(http.StatusAccepted).Message("请求处理中,请稍后重试").Send()
		return
	case syserror.TokenExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("页面已过期,请刷新后重试").Send()
		return
	case syserror.ResourceExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("附件已失效,请重新上传").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("回帖发布失败").Send()
		return
	case syserror.DuplicateError:
		api.JSON(c).Code(http.StatusConflict).Message("回帖已经发布").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("回帖发布成功").Data(gin.H{"reply_id": strconv.FormatInt(replyID, 10)}).Send()
}

// 获取回帖
func (rc *ReplyController) GetOneReply(c *gin.Context) {
	ctx := c.Request.Context()
	replyIDStr := c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的回帖id").Send()
		return
	}
	var userID = c.GetInt64("userID")
	userInfo, replyData, syserr := rc.replyServ.GetOneReply(ctx, replyID, userID)
	switch syserr {
	case syserror.NetworkError:
		utils.WithContext(ctx).WithField("controller", rc.name).WithError(err).Error("gRPC network error in GetOneReply")
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		utils.WithContext(ctx).WithField("controller", rc.name).WithError(err).Error("internal error in GetOneReply")
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到该用户的回帖").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(gin.H{
		"user_info": userInfo,
		"post_data": replyData,
	}).Send()
}

// 获取分页回帖
func (rc *ReplyController) BatchGetReply(c *gin.Context) {
	var replyIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(replyIDStr, 10, 64)
	var userID = c.GetInt64("userID")
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("获取失败,无效的帖子id").Send()
		return
	}
	// 获取偏移量
	var pageStr, limitStr = c.Query("page"), c.Query("page_size")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	page = max(page, 1) // 从1开始
	// 强制分页规范：page_size 固定 20
	if limitStr != "" && limit != 20 {
		api.JSON(c).Code(http.StatusBadRequest).Message("page_size 仅支持 20").Send()
		return
	}
	limit = 20
	var ctx = c.Request.Context()
	multiData, total, syserr := rc.replyServ.GetManyReplies(ctx, postID, userID, page, limit)
	switch syserr {
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到指定帖子的回帖").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(multiData).Pagination(page, limit, int(total)).Send()
}

// 更新回帖
func (rc *ReplyController) UpdateOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的回帖id").Send()
		return
	}
	var req request.ReplyUpdate
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("更新失败,数据格式有误").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := rc.replyServ.UpdateOneReply(ctx, userID, replyID, req)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("您无权修改他人的回帖").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要修改的回帖").Send()
		return
	case syserror.ResourceExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("附件资源已过期,请重新上传").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("更新失败").Send()
		return
	}

	api.JSON(c).Code(http.StatusOK).Message("回帖更新成功").Send()
}

// 删除一条回帖
func (rc *ReplyController) DeleteOneReply(c *gin.Context) {
	ctx := c.Request.Context()
	replyIDStr := c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的回帖id").Send()
		return
	}
	var userID = c.GetInt64("userID")
	var role = c.GetInt("role")
	syserr := rc.replyServ.DeleteOneReply(ctx, replyID, userID, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("您无权删除他人的回帖").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要删除的回帖").Send()
		return
	case syserror.NetworkError:
		utils.WithContext(ctx).WithField("controller", rc.name).WithError(err).Error("network error in DeleteOneReply")
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("删除失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("删除成功").Send()
}

// 点赞接口
// 点赞帖子
func (pc *ReplyController) LikeOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的回帖id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.replyServ.LikeOneReply(ctx, replyID, userID)
	switch syserr {
	case syserror.DuplicateError:
		api.JSON(c).Code(http.StatusConflict).Message("您已点过赞").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要点赞的回帖").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("点赞成功").Send()
}

// 取消点赞
func (pc *ReplyController) CancelLikeOneReply(c *gin.Context) {
	var replyIDStr = c.Param("replyID")
	replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的回帖id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.replyServ.CancelLikeOneReply(ctx, replyID, userID)
	switch syserr {
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("未点赞过该回帖").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("取消点赞成功").Send()
}

// 更改回帖状态
// 发帖作者 -> 作者认可
// 教师 -> 教师认可
func (pc *ReplyController) ChangeReplyStatus(c *gin.Context) {
	var req request.ReplyStatusUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("状态修改失败,数据格式有误").Send()
		return
	}
	if req.Status < int(requestedReplyStatusMin()) || req.Status > int(requestedReplyStatusMax()) {
		api.JSON(c).Code(http.StatusBadRequest).Message("状态值无效").Send()
		return
	}
	var userID = c.GetInt64("userID")
	var role = c.GetInt("role")
	var ctx = c.Request.Context()
	syserr := pc.replyServ.ChangeReplyStatus(ctx, req, userID, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("您无权修改该回帖状态").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要修改状态的回贴").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("回帖状态修改成功").Send()
}

func requestedReplyStatusMin() model.ReplyStatus {
	return model.NotSelected
}

func requestedReplyStatusMax() model.ReplyStatus {
	return model.TeacherCertified
}
