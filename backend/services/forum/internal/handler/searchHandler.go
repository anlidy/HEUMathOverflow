package handler

import (
	"MathOverflow/common/api"
	syserror "MathOverflow/services/forum/internal/model/error"
	"MathOverflow/services/forum/internal/model/request"
	"MathOverflow/services/forum/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchServ service.SearchService
	name       string
}

func NewSearchHandler(serv service.SearchService) SearchHandler {
	return SearchHandler{searchServ: serv, name: "Search-Handler"}
}

// 搜索帖子
func (sc *SearchHandler) SearchPosts(c *gin.Context) {
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
