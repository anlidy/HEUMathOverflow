package controller

import (
	"MathOverflow/internal/common/api"
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
		api.JSON(c).Code(http.StatusBadRequest).Message("发布失败,数据格式有误").Send()
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	postID, err := pc.postServ.CreateNewPost(ctx, userID, req)
	switch err {
	case syserror.InProgressError:
		api.JSON(c).Code(http.StatusAccepted).Message("请求处理中,请稍后重试").Send()
		return
	case syserror.TokenExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("页面已过期,请刷新后重试").Send()
		return
	case syserror.ResourceExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("图片附件已过期,请重新上传").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("帖子发布失败").Send()
		return
	case syserror.DuplicateError:
		api.JSON(c).Code(http.StatusConflict).Message("帖子已经发布").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("发布成功").Data(gin.H{"post_id": strconv.FormatInt(postID, 10)}).Send()
}

// 获取帖子信息
func (pc *PostController) GetPostData(c *gin.Context) {
	var ctx = c.Request.Context()
	var postIDStr = c.Param("postID")
	var userID = c.GetInt64("userID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	userInfo, postData, syserr := pc.postServ.GetOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到该用户的帖子").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(gin.H{
		"user_info": userInfo,
		"post_data": postData,
	}).Send()
}

// 批量获取帖子信息
func (pc *PostController) GetManyPostData(c *gin.Context) {
	// 获取偏移量
	var pageStr, limitStr, orderStr = c.Query("page"), c.Query("page_size"), c.Query("order")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	order, _ := strconv.Atoi(orderStr) // 默认取0,推荐排序
	page = max(page, 1)                // 从1开始
	if limit == 0 {
		limit = 20 // 默认取20条
	}
	var ctx = c.Request.Context()
	multiData, syserr := pc.postServ.GetManyPosts(ctx, page, limit, model.OrderBy(order))
	switch syserr {
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(multiData).Pagination(page, limit, 0).Send()

}

// 更新帖子
func (pc *PostController) UpdateOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var req request.PostUpdate
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("更新失败,数据格式有误").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.UpdateOnePost(ctx, userID, postID, req)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("您无权修改他人的帖子").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要修改的帖子").Send()
		return
	case syserror.ResourceExpiredError:
		api.JSON(c).Code(http.StatusGone).Message("附件资源已过期,请重新上传").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("更新失败").Send()
		return
	}

	api.JSON(c).Code(http.StatusOK).Message("帖子更新成功").Send()
}

func (pc *PostController) SetPostCertified(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var req request.PostCertifiedUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("精选设置失败,数据格式有误").Send()
		return
	}
	ctx := c.Request.Context()
	role := c.GetInt("role")
	syserr := pc.postServ.SetPostCertified(ctx, postID, req, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("只有教师可以设置精选贴").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要设置的帖子").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("精选设置失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("帖子精选状态修改成功").Send()
}

// 删除帖子及所有回帖
func (pc *PostController) DeleteOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	var role = c.GetInt("role")
	syserr := pc.postServ.DeleteOnePost(ctx, postID, userID, role)
	switch syserr {
	case syserror.PermissionDeniedError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("您无权删除他人的帖子").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要删除的帖子").Send()
		return
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("删除失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("删除成功").Send()
}

// / 点赞接口
// 点赞帖子
func (pc *PostController) LikeOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.LikeOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.DuplicateError:
		api.JSON(c).Code(http.StatusConflict).Message("您已点过赞").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要点赞的帖子").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("点赞成功").Send()
}

// 取消点赞
func (pc *PostController) CancelLikeOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.CancelLikeOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("未点赞过该帖子").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("取消点赞成功").Send()
}

// / 收藏接口
// 收藏帖子
func (pc *PostController) StarOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.StarOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.DuplicateError:
		api.JSON(c).Code(http.StatusConflict).Message("您已收藏").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到要收藏的帖子").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("收藏成功").Send()
}

// 取消收藏
func (pc *PostController) CancelStarOnePost(c *gin.Context) {
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("无效的帖子id").Send()
		return
	}
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	syserr := pc.postServ.CancelStarOnePost(ctx, postID, userID)
	switch syserr {
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("未收藏过该帖子").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("操作失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("取消收藏成功").Send()
}

// 查询当前用户收藏的帖子（分页）
func (pc *PostController) GetUserStarredPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("page_size"))
	page = max(page, 1) // 从1开始
	if limit == 0 {
		limit = 20
	}
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	posts, total, syserr := pc.postServ.GetUserStarredPosts(ctx, userID, page, limit)
	switch syserr {
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器异常,请稍后再试").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("查询成功").Data(posts).Pagination(page, limit, int(total)).Send()
}
