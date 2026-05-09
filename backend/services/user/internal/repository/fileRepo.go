package repository

import (
	"MathOverflow/common/client"
	common "MathOverflow/common/model"
	"context"
	"fmt"
)

type FileRepo interface {
	UploadFile(ctx context.Context, bucket string, file common.File) (string, error)
	DownloadFile(ctx context.Context, filename string) (*common.File, error)
	DeleteFile(ctx context.Context, filename string) error
	FileExists(ctx context.Context, filename string) (bool, error)
}

type fileRepo struct {
	client *client.MinioClinet
}

func NewFileRepository(client *client.MinioClinet) FileRepo {
	return &fileRepo{client: client}
}

// 上传文件并返回文件路径
func (r *fileRepo) UploadFile(ctx context.Context, bucket string, file common.File) (string, error) {

	err := r.client.Upload(ctx, file.Filename, file.Data, file.Size, file.ContentType)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/api/v1/%s/avatar/%s", bucket, file.Filename) // /api/v1/user/avatar/20251027.jpg
	return url, nil
}

// 下载文件
func (r *fileRepo) DownloadFile(ctx context.Context, filename string) (*common.File, error) {
	file, err := r.client.Download(ctx, filename)
	return file, err
}

// 删除文件
func (r *fileRepo) DeleteFile(ctx context.Context, filename string) error {
	err := r.client.Delete(ctx, filename)
	return err
}

func (r *fileRepo) FileExists(ctx context.Context, filename string) (bool, error) {
	exists, err := r.client.Exists(ctx, filename)
	return exists, err
}
