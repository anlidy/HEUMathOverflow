package controller

import (
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/service"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postServ service.PostService
	name     string
}

func NewPostController(postServ service.PostService) PostController {
	return PostController{postServ: postServ, name: "Post-Controller"}
}

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

func (pc *PostController) CreateNewPost(c *gin.Context) {

}
