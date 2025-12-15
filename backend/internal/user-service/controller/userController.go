package controller

import (
	"MathOverflow/internal/common/api"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	syserror "MathOverflow/internal/user-service/model/error"
	"MathOverflow/internal/user-service/model/request"
	"MathOverflow/internal/user-service/service"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
	name        string
}

func NewUserController(userService service.UserService) UserController {
	return UserController{userService: userService, name: "User-Controller"}
}

// 用户注册
func (uc *UserController) UserRegister(c *gin.Context) {
	var req request.UserRegister
	var ctx = c.Request.Context()
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("注册失败,数据格式有误").Send()
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		api.JSON(c).Code(http.StatusBadRequest).Message("注册失败,用户名不能为空").Send()
	}
	session, info, err := uc.userService.UserRegister(ctx, req)
	switch err {
	case syserror.EmailError:
		api.JSON(c).Code(http.StatusBadRequest).Message("注册失败,邮箱格式有误").Send()
	case syserror.EmailExistsError:
		api.JSON(c).Code(http.StatusConflict).Message("注册失败,用户已存在").Send()
	case syserror.NameExistsError:
		api.JSON(c).Code(http.StatusConflict).Message("注册失败,该用户名已被注册").Send()
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("注册失败").Send()
	case syserror.NoError:
		// 设置 Cookie
		c.SetCookie("session-id", session.SessionID, int(session.TTL.Seconds()), "/", "localhost", false, true) // 生产环境改成 math-overflow.edu
		api.JSON(c).Code(http.StatusOK).Message("注册成功").Data(gin.H{"user_info": info}).Send()
	}
}

// 用户登录
func (uc *UserController) UserLogin(c *gin.Context) {
	var ctx = c.Request.Context()
	var req request.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("登录失败,数据格式有误").Send()
		return
	}
	session, info, syserr := uc.userService.UserLogin(ctx, req)
	switch syserr {
	case syserror.EmailError:
		api.JSON(c).Code(http.StatusBadRequest).Message("登录失败,邮箱格式有误").Send()
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("登录失败,找不到该用户").Send()
	case syserror.PasswordError:
		api.JSON(c).Code(http.StatusUnauthorized).Message("登录失败,账户或密码错误").Send()
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("登录失败").Send()
	case syserror.NoError:
		// 设置 Cookie
		c.SetCookie("session-id", session.SessionID, int(session.TTL.Seconds()), "/", "localhost", false, true) // 生产环境改成 math-overflow.edu
		api.JSON(c).Code(http.StatusOK).Message("登录成功").Data(gin.H{"user_info": info}).Send()
	}
}

// 上传头像
func (uc *UserController) UserUploadAvatar(c *gin.Context) {
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	// fmt.Println(c.ContentType())
	// 获取上传文件
	rawFile, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		utils.WithContext(ctx).WithField("controller", uc.name).WithError(err).Error("parse avatar upload file failed")
		api.JSON(c).Code(http.StatusBadRequest).Message("文件上传失败").Send()
		return
	}
	defer rawFile.Close()

	// 判断文件类型
	fileExt := filepath.Ext(fileHeader.Filename)
	if !utils.IsValidImage(fileExt) {
		api.JSON(c).Code(http.StatusBadRequest).Message("上传失败,非图片文件").Send()
		return
	}

	// 获取文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(fileExt)
	}

	fileName := fmt.Sprintf("%d%s", userID, fileExt) // userID.jpg
	var file = common.File{
		Filename:    fileName,
		Data:        rawFile,
		Size:        fileHeader.Size,
		ContentType: contentType,
	}

	avatarUrl, sysErr := uc.userService.UserUploadAvatar(ctx, userID, file)
	switch sysErr {
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("头像上传失败").Send()
	case syserror.NoError:
		api.JSON(c).Code(http.StatusOK).Message("头像上传成功").Data(gin.H{"avatar_url": avatarUrl}).Send()
	}
}

// 下载头像
func (uc *UserController) UserDownloadAvatar(c *gin.Context) {
	var ctx = c.Request.Context()
	filename := c.Param("filename")
	file, err := uc.userService.UserDownloadAvatar(ctx, filename)
	switch err {
	case syserror.InternalError:
		c.String(http.StatusInternalServerError, "下载文件失败")
		return
	case syserror.NotFoundError:
		c.String(http.StatusNotFound, "找不到该文件")
		return
	}
	// 若无报错, 结束时关闭指针
	defer file.Data.Close()
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

// 修改用户信息
func (uc *UserController) UserUploadProfie(c *gin.Context) {
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	var req request.UserProfie
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("数据格式校验不通过").Send()
		return
	}
	err := uc.userService.UserUpdateProfie(ctx, userID, req)
	switch err {
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到该用户").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("更新失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("用户信息更新成功").Send()
}

// 更新用户密码
func (uc *UserController) UserUpdatePassword(c *gin.Context) {
	var ctx = c.Request.Context()
	var userID = c.GetInt64("userID")
	var req request.UserPassword
	if err := c.ShouldBindJSON(&req); err != nil {
		api.JSON(c).Code(http.StatusBadRequest).Message("数据格式校验不通过").Send()
		return
	}
	err := uc.userService.UserUpdatePassword(ctx, userID, req)
	switch err {
	case syserror.NotFoundError:
		api.JSON(c).Code(http.StatusNotFound).Message("找不到该用户").Send()
		return
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("密码修改失败").Send()
		return
	}
	api.JSON(c).Code(http.StatusOK).Message("密码修改成功").Send()
}
