package controller

import (
	"net/http"

	"MathOverflow/internal/common/api"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/repository"

	"github.com/gin-gonic/gin"
)

type TokenController struct {
	tokenRepo repository.ClientTokenRepo
	name      string
}

func NewTokenController(tokenRepo repository.ClientTokenRepo) TokenController {
	return TokenController{tokenRepo: tokenRepo, name: "Token-Controller"}
}

func (tc *TokenController) IssuePostToken(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	token, err := tc.tokenRepo.Issue(ctx, repository.ClientTokenPost, userID)
	if err != nil {
		utils.WithContext(ctx).WithField("controller", tc.name).WithError(err).Error("issue post token failed")
		api.JSON(c).Code(http.StatusInternalServerError).Message("获取token失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(gin.H{"client_token": token}).Send()
}

func (tc *TokenController) IssueReplyToken(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("userID")
	token, err := tc.tokenRepo.Issue(ctx, repository.ClientTokenReply, userID)
	if err != nil {
		utils.WithContext(ctx).WithField("controller", tc.name).WithError(err).Error("issue reply token failed")
		api.JSON(c).Code(http.StatusInternalServerError).Message("获取token失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("请求成功").Data(gin.H{"client_token": token}).Send()
}
