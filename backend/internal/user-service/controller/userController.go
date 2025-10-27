package controller

import (
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/user-service/model"
	syserror "MathOverflow/internal/user-service/model/error"
	"MathOverflow/internal/user-service/model/request"
	"MathOverflow/internal/user-service/service"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) UserController {
	return UserController{userService: userService}
}

// 用户注册
func (uc *UserController) UserRegister(c *gin.Context) {
	var req request.UserRegister
	var ctx = c.Request.Context()
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "注册失败,数据格式有误", "code": http.StatusBadRequest, "data": nil})
		return
	}
	info, err := uc.userService.UserRegister(ctx, req)
	switch err {
	case syserror.EmailError:
		c.JSON(http.StatusBadRequest, gin.H{"message": "注册失败,邮箱格式有误", "code": http.StatusBadRequest, "data": nil})
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "注册失败", "code": http.StatusInternalServerError, "data": nil})
	case syserror.NoError:
		c.JSON(http.StatusOK, gin.H{"message": "注册失败", "code": http.StatusOK, "data": info})
	}
}

// 用户登录
func (uc *UserController) UserLogin(c *gin.Context) {
	var req request.UserLogin
	var ctx = c.Request.Context()
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "登录失败,数据格式有误", "code": http.StatusBadRequest, "data": nil})
		return
	}
	info, err := uc.userService.UserLogin(ctx, req)
	switch err {
	case syserror.EmailError:
		c.JSON(http.StatusBadRequest, gin.H{"message": "登录失败,邮箱格式有误", "code": http.StatusBadRequest, "data": nil})
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "登录失败", "code": http.StatusInternalServerError, "data": nil})
	case syserror.NoError:
		c.JSON(http.StatusOK, gin.H{"message": "登录成功", "code": http.StatusOK, "data": info})
	}
}

// 上传头像
func (uc *UserController) UserUploadAvatar(c *gin.Context) {
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")

	// 获取上传文件
	rawFile, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "文件上传失败", "data": nil})
		return
	}
	defer rawFile.Close()

	// 判断文件类型
	fileExt := filepath.Ext(fileHeader.Filename)
	if utils.IsValidImage(fileExt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "上传失败,非图片文件", "data": nil})
	}

	// 获取文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(fileExt)
	}

	fileName := fmt.Sprintf("avatar/%d%s", time.Now().Unix(), fileExt)
	var file = model.File{
		Filename:    fileName,
		Data:        rawFile,
		Size:        fileHeader.Size,
		ContentType: contentType,
	}

	avatarUrl, sysErr := uc.userService.UserUploadAvatar(ctx, userID, file)
	switch sysErr {
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "头像上传失败", "code": http.StatusInternalServerError, "data": nil})
	case syserror.NoError:
		c.JSON(http.StatusOK, gin.H{"message": "头像上传成功", "code": http.StatusOK, "data": gin.H{"avatar_url": avatarUrl}})
	}
}

// 下载头像
func (uc *UserController) UserDownloadAvatar(c *gin.Context) {
	var ctx = c.Request.Context()
	filename := c.Param("filename")

	file, err := uc.userService.UserDownloadAvatar(ctx, filename)
	defer file.Data.Close()
	switch err {
	case syserror.InternalError:
		c.String(http.StatusInternalServerError, "下载文件失败")
	case syserror.NotFoundError:
		c.String(http.StatusNotFound, "找不到该文件")
	}

	// 设置响应头
	c.Header("Content-Type", file.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.Filename))

	// 流式写出
	if _, err := io.Copy(c.Writer, file.Data); err != nil {
		c.String(http.StatusInternalServerError, "下载失败")
		return
	}
}
