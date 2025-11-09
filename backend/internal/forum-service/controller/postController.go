package controller

import (
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/service"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
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

// 上传临时图片文件
func (pc *PostController) UploadTempFile(c *gin.Context) {
	var ctx = c.Request.Context()
	// 获取上传文件
	rawFile, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("[%s] %v\n", pc.name, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "文件上传失败", "data": nil})
		return
	}
	defer rawFile.Close()

	// 判断文件类型
	fileExt := filepath.Ext(fileHeader.Filename)
	if !utils.IsValidImage(fileExt) {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "上传失败,非图片文件", "data": nil})
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
	tempUrl, sysErr := pc.postServ.UploadFile(ctx, file)
	switch sysErr {
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "图片上传失败", "code": http.StatusInternalServerError, "data": nil})
	case syserror.NoError:
		c.JSON(http.StatusOK, gin.H{"message": "图片上传成功", "code": http.StatusOK, "data": gin.H{"image_url": tempUrl}})
	}
}

// 下载文件
func (pc *PostController) UserDownloadAvatar(c *gin.Context) {
	var ctx = c.Request.Context()
	filename := c.Param("filename")
	file, err := pc.postServ.DownloadFile(ctx, filename)
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

// 创建帖子
func (pc *PostController) CreateNewPost(c *gin.Context) {
	var req request.PostCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "发布失败,数据格式有误", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var ctx = c.Request.Context()
	userID := c.GetInt64("userID")
	postID, err := pc.postServ.CreateNewPost(ctx, userID, req)
	switch err {
	case syserror.ResourceExpiredError:
		c.JSON(http.StatusGone, gin.H{"message": "图片附件已过期,请重新上传", "code": http.StatusGone, "data": nil})
		return
	case syserror.InternalError:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "帖子发布失败", "code": http.StatusBadRequest, "data": nil})
		return
	case syserror.DuplicateError:
		c.JSON(http.StatusConflict, gin.H{"message": "帖子已经发布", "code": http.StatusConflict, "data": nil})
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "code": http.StatusOK, "data": gin.H{"post_id": postID}})
}

// 获取帖子信息
func (pc *PostController) GetPostData(c *gin.Context) {
	var ctx = c.Request.Context()
	var postIDStr = c.Param("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的帖子id", "code": http.StatusBadRequest, "data": nil})
		return
	}
	var userID = c.GetInt64("userID")
	userInfo, postData, syserr := pc.postServ.GetOnePost(ctx, userID, postID)
	switch syserr {
	case syserror.NetworkError:
		log.Printf("[%s] gRPC网络异常: %v\n", pc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后重试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.InternalError:
		log.Printf("[%s] %v\n", pc.name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器异常,请稍后再试", "code": http.StatusInternalServerError, "data": nil})
		return
	case syserror.NotFoundError:
		c.JSON(http.StatusNotFound, gin.H{"message": "找不到该用户的帖子", "code": http.StatusNotFound, "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求成功", "code": http.StatusOK, "data": gin.H{
		"user_info": userInfo,
		"post_data": postData,
	}})
}
