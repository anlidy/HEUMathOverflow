package handler

import (
	"net/http"

	"MathOverflow/common/api"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/repository"

	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	tokenRepo repository.ClientTokenRepo
	name      string
}

func NewTokenHandler(tokenRepo repository.ClientTokenRepo) TokenHandler {
	return TokenHandler{tokenRepo: tokenRepo, name: "Token-Handler"}
}

func (tc *TokenHandler) IssuePostToken(c *gin.Context) {
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

func (tc *TokenHandler) IssueReplyToken(c *gin.Context) {
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
