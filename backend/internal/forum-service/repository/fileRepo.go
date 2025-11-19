package repository

import (
	"MathOverflow/internal/common/client"
	common "MathOverflow/internal/common/model"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
)

type FileRepo interface {
	UploadFile(ctx context.Context, bucket string, file common.File) (string, error)
	PromoteFile(ctx context.Context, tmpURL, bucket string, hostID int64) (string, error)
	DownloadFile(ctx context.Context, filename string) (*common.File, error)
	DeleteFile(ctx context.Context, filename string) error
	FileExists(ctx context.Context, filename string) (bool, error)
}

type fileRepo struct {
	mc *client.MinioClinet
}

func NewFileRepository(mc *client.MinioClinet) FileRepo {
	return &fileRepo{mc: mc}
}

// 上传文件
// 上传文件并返回文件路径
func (r *fileRepo) UploadFile(ctx context.Context, bucket string, file common.File) (string, error) {
	err := r.mc.Upload(ctx, file.Filename, file.Data, file.Size, file.ContentType)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/api/v1/%s/file/%s", bucket, file.Filename) // /api/v1/forum/file/tmp/17302800000.jpg
	return url, nil
}

// 将临时文件转正存储
// hostID可代表postID和replyID
func (r *fileRepo) PromoteFile(ctx context.Context, tmpURL, bucket string, hostID int64) (string, error) {
	// 1. 提取对象名，得到 tmp/17302800000.jpg
	tmpName := strings.TrimPrefix(tmpURL, fmt.Sprintf("/api/v1/%s/file/", bucket))

	// 2. 新名称: {hostID}/{filename}
	newName := fmt.Sprintf("%d/%s", hostID, filepath.Base(tmpName))
	// 3. 复制对象
	src := minio.CopySrcOptions{
		Bucket: bucket,
		Object: tmpName,
	}
	dst := minio.CopyDestOptions{
		Bucket: bucket,
		Object: newName,
	}
	_, err := r.mc.Client.CopyObject(context.Background(), dst, src)
	if err != nil {
		return "", fmt.Errorf("复制失败: %v", err)
	}

	// 4. 删除旧的 tmp 对象
	err = r.DeleteFile(ctx, tmpName)
	if err != nil {
		log.Printf("删除临时文件失败: %v", err)
	}

	// 5. 返回新 URL
	newURL := fmt.Sprintf("/api/v1/%s/file/%s", bucket, newName)
	return newURL, nil
}

// 下载文件
func (r *fileRepo) DownloadFile(ctx context.Context, filename string) (*common.File, error) {
	file, err := r.mc.Download(ctx, filename)
	return file, err
}

// 删除文件
func (r *fileRepo) DeleteFile(ctx context.Context, filename string) error {
	err := r.mc.Delete(ctx, filename)
	return err
}

func (r *fileRepo) FileExists(ctx context.Context, filename string) (bool, error) {
	exists, err := r.mc.Exists(ctx, filename)
	return exists, err
}
