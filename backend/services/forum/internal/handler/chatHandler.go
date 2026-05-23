package handler

import (
	"MathOverflow/common/api"
	requestmodel "MathOverflow/services/forum/internal/model/request"
	responsemodel "MathOverflow/services/forum/internal/model/response"
	syserror "MathOverflow/services/forum/internal/model/error"
	"MathOverflow/services/forum/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatServ service.ChatService
	name     string
}

func NewChatHandler(chatServ service.ChatService) ChatHandler {
	return ChatHandler{chatServ: chatServ, name: "Chat-Handler"}
}

func (hc *ChatHandler) CreateSession(c *gin.Context) {
	var req requestmodel.ChatSessionCreate
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		api.JSON(c).Code(http.StatusBadRequest).Message("会话参数不合法").Send()
		return
	}
	session, syserr := hc.chatServ.CreateSession(c.Request.Context(), c.GetInt64("userID"), req)
	if syserr != syserror.NoError {
		api.JSON(c).Code(http.StatusInternalServerError).Message("创建会话失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("创建会话成功").Data(responsemodel.NewChatSessionItem(session)).Send()
}

func (hc *ChatHandler) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	sessions, total, syserr := hc.chatServ.ListSessions(c.Request.Context(), c.GetInt64("userID"), page, pageSize)
	if syserr != syserror.NoError {
		api.JSON(c).Code(http.StatusInternalServerError).Message("获取会话列表失败").Send()
		return
	}
	items := make([]responsemodel.ChatSessionItem, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, responsemodel.NewChatSessionItem(session))
	}
	api.JSON(c).Code(http.StatusOK).Message("获取会话列表成功").Data(items).Pagination(page, pageSize, int(total)).Send()
}

func (hc *ChatHandler) ListMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	sessionID := c.Param("sessionID")
	session, messages, total, syserr := hc.chatServ.ListMessages(c.Request.Context(), c.GetInt64("userID"), sessionID, page, pageSize)
	if syserr == syserror.NotFoundError {
		api.JSON(c).Code(http.StatusNotFound).Message("找不到会话").Send()
		return
	}
	if syserr != syserror.NoError {
		api.JSON(c).Code(http.StatusInternalServerError).Message("获取消息失败").Send()
		return
	}
	items := make([]responsemodel.ChatMessageItem, 0, len(messages))
	for _, message := range messages {
		items = append(items, responsemodel.NewChatMessageItem(message))
	}
	api.JSON(c).Code(http.StatusOK).Message("获取消息成功").Data(gin.H{
		"session":  responsemodel.NewChatSessionItem(session),
		"messages": items,
	}).Pagination(page, pageSize, int(total)).Send()
}

func (hc *ChatHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("sessionID")
	var req requestmodel.ChatMessageCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("消息参数不合法").Send()
		return
	}
	data, syserr := hc.chatServ.SendMessage(c.Request.Context(), c.GetInt64("userID"), sessionID, req)
	if syserr == syserror.NotFoundError {
		api.JSON(c).Code(http.StatusNotFound).Message("找不到会话").Send()
		return
	}
	if syserr == syserror.ConflictError {
		api.JSON(c).Code(http.StatusBadRequest).Message("消息不能为空").Send()
		return
	}
	if syserr == syserror.NetworkError {
		api.JSON(c).Code(http.StatusBadGateway).Message("AI服务不可用").Send()
		return
	}
	if syserr != syserror.NoError {
		api.JSON(c).Code(http.StatusInternalServerError).Message("发送消息失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("发送消息成功").Data(data).Send()
}
