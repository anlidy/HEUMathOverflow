package controller

import (
	"MathOverflow/internal/common/api"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	searchServ service.SearchService
	name       string
}

func NewSearchController(serv service.SearchService) SearchController {
	return SearchController{searchServ: serv, name: "Search-Controller"}
}

// 搜索帖子
func (sc *SearchController) SearchPosts(c *gin.Context) {
	ctx := c.Request.Context()
	var req request.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("搜索参数不合法").Send()
		return
	}
	// 计算分页
	page := max(req.Page, 1)
	size := req.PageSize
	if size < 1 || size > 100 {
		size = 20
	}
	req.From = (page - 1) * size

	result, total, syserr := sc.searchServ.SearchPosts(ctx, req)
	switch syserr {
	case syserror.NetworkError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("服务器网络异常,请稍后重试").Send()
		return
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusOK).Message("搜索完成").Data(result).Pagination(page, size, total).Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("搜索失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("搜索完成").Data(result).Pagination(page, size, total).Send()
}
