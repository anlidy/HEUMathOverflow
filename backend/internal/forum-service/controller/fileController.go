package controller

import (
	"MathOverflow/internal/common/api"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/service"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type ForumController struct {
	forumServ service.ForumService
	name      string
}

func NewForumController(forumServ service.ForumService) ForumController {
	return ForumController{forumServ: forumServ, name: "Forum-Controller"}
}

// 上传临时图片文件
func (pc *ForumController) UploadTempFile(c *gin.Context) {
	var ctx = c.Request.Context()
	// 获取上传文件
	rawFile, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		utils.WithContext(ctx).WithField("controller", pc.name).WithError(err).Error("parse upload file failed")
		api.JSON(c).Code(http.StatusBadRequest).Message("上传失败").Send()
		return
	}
	defer rawFile.Close()

	// 判断文件类型
	fileExt := filepath.Ext(fileHeader.Filename)
	if !utils.IsAllowedFile(fileExt) {
		api.JSON(c).Code(http.StatusBadRequest).Message("上传失败,不支持此文件格式").Send()
		return
	}

	// 获取文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(fileExt)
	}

	// 构造临时文件名
	var fileID = utils.GenerateSnowflakeID()
	tempName := fmt.Sprintf("tmp/%d%s", fileID, fileExt) // tmp/17302800000.jpg
	var file = common.File{
		Filename:    tempName,
		Data:        rawFile,
		Size:        fileHeader.Size,
		ContentType: contentType,
	}
	tempUrl, sysErr := pc.forumServ.UploadFile(ctx, file)
	switch sysErr {
	case syserror.InternalError:
		api.JSON(c).Code(http.StatusInternalServerError).Message("上传失败").Send()
	case syserror.NoError:
		api.JSON(c).Code(http.StatusOK).Message("上传成功").Data(gin.H{"image_url": tempUrl}).Send()
	}
}

// 下载文件
func (pc *ForumController) DownloadFile(c *gin.Context) {
	var ctx = c.Request.Context()
	filename := c.Param("filename")
	file, err := pc.forumServ.DownloadFile(ctx, filename)
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
