package controller

import (
	"MathOverflow/internal/forum-service/model"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/service"
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
	userInfo, postData, syserr := pc.postServ.GetOnePost(ctx, postID)
	switch syserr {
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
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

// 批量获取帖子信息
func (pc *PostController) GetManyPostData(c *gin.Context) {
	// 获取偏移量
	var offsetStr, limitStr, orderStr = c.Query("offset"), c.Query("limit"), c.Query("order")
	offset, _ := strconv.Atoi(offsetStr) // 默认取0
	limit, _ := strconv.Atoi(limitStr)
	order, _ := strconv.Atoi(orderStr) // 默认取0,推荐排序
	if limit == 0 {
		limit = 20 // 默认取20条
	}
	var ctx = c.Request.Context()
	multiData, syserr := pc.postServ.GetManyPosts(ctx, offset, limit, model.OrderBy(order))
	switch syserr {
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求成功", "code": http.StatusOK, "data": multiData})

}

// 更新帖子
func (pc *PostController) UpdateOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var req request.PostUpdate
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "更新失败,数据格式有误", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.UpdateOnePost(ctx, userID, postID, req)
	switch syserr {
	case syserror.PermissionDeniedError:
		c.JSON(http.StatusUnauthorized, gin.H{"message": "您无权修改他人的帖子", "code": http.StatusUnauthorized})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要修改的帖子", "code": http.StatusNotFound})
		return
	case syserror.ResourceExpiredError:
		c.JSON(http.StatusGone, gin.H{"message": "附件资源已过期,请重新上传", "code": http.StatusGone})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新失败", "code": http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帖子更新成功", "code": http.StatusOK})
}

// 删除帖子及所有回帖
func (pc *PostController) DeleteOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	var role = c.GetInt("role")
	syserr := pc.postServ.DeleteOnePost(ctx, postID, userID, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		c.JSON(http.StatusUnauthorized, gin.H{"message": "您无权删除他人的帖子", "code": http.StatusUnauthorized})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要删除的帖子", "code": http.StatusNotFound})
		return
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "code": http.StatusOK})
}

// / 点赞接口
// 点赞帖子
func (pc *PostController) LikeOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.LikeOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "您已点过赞", "code": http.StatusConflict})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要点赞的帖子", "code": http.StatusNotFound})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "点赞成功", "code": http.StatusOK})
}

// 取消点赞
func (pc *PostController) CancelLikeOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.CancelLikeOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "未点赞过该帖子", "code": http.StatusNotFound})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取消点赞成功", "code": http.StatusOK})
}

// 查询当前用户是否点赞过帖子
func (pc *PostController) GetPostLikeStatus(c *gin.Context) {
	postIDStr := c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	liked, syserr := pc.postServ.HasLikedPost(ctx, postID, userID)
	if syserr == syserror.InternalError {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "查询成功", "code": http.StatusOK, "data": gin.H{"liked": liked}})
}

// / 收藏接口
// 收藏帖子
func (pc *PostController) StarOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.StarOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "您已收藏", "code": http.StatusConflict})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到要收藏的帖子", "code": http.StatusNotFound})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "收藏成功", "code": http.StatusOK})
}

// 取消收藏
func (pc *PostController) CancelStarOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.CancelStarOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "未收藏过该帖子", "code": http.StatusNotFound})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "操作失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "取消收藏成功", "code": http.StatusOK})
}

// 查询当前用户是否收藏过帖子
func (pc *PostController) GetPostStarStatus(c *gin.Context) {
	postIDStr := c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	starred, syserr := pc.postServ.HasStarredPost(ctx, postID, userID)
	if syserr == syserror.InternalError {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询失败", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "查询成功", "code": http.StatusOK, "data": gin.H{"starred": starred}})
}

// 查询当前用户收藏的帖子（分页）
func (pc *PostController) GetUserStarredPosts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.Query("offset"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit == 0 {
		limit = 20
	}
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	posts, syserr := pc.postServ.GetUserStarredPosts(ctx, userID, offset, limit)
	switch syserr {
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "查询成功", "code": http.StatusOK, "data": posts})
}
