package controller

import (
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "搜索参数不合法", "code": http.StatusBadRequest})
		return
	}

	result, total, syserr := sc.searchServ.SearchPosts(ctx, req)
	switch syserr {
	case syserror.NetworkError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器网络异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "搜索失败", "data": nil, "code": http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "搜索完成", "data": result, "total": total, "code": http.StatusOK})
}
