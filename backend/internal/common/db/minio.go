package db

import (
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/user-service/model"
	"context"
	"fmt"
	"io"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClinet interface {
	Upload(ctx context.Context, objectName string, data io.ReadCloser, size int64, contentType string) error
	Download(ctx context.Context, objectName string) (*model.File, error)
	Delete(ctx context.Context, objectName string) error
	Exists(ctx context.Context, objectName string) (bool, error)
}

type minioClient struct {
	client *minio.Client
	bucket string
}

// 初始化minio
func InitMinIO(cfg config.MinioConfig) (MinioClinet, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	// 创建 bucket（如果不存在）
	ctx := context.Background()
	exists, errBucket := client.BucketExists(ctx, cfg.Bucket)
	if errBucket != nil {
		return nil, errBucket
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	fmt.Println("MinIO connected, bucket:", cfg.Bucket)
	return &minioClient{client: client, bucket: cfg.Bucket}, nil
}

// 流式上传文件
func (c *minioClient) Upload(ctx context.Context, objectName string, data io.ReadCloser, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// 流式下载文件
func (c *minioClient) Download(ctx context.Context, objectName string) (*model.File, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object failed: %w", err)
	}

	// 获取对象元数据
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, fmt.Errorf("get object stat failed: %w", err)
	}

	return &model.File{
		Data:        obj,
		ContentType: stat.ContentType,
		Size:        stat.Size,
		Filename:    path.Base(objectName),
	}, nil
}

// 删除文件
func (c *minioClient) Delete(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

// Exists 检查对象是否存在
func (c *minioClient) Exists(ctx context.Context, objectName string) (bool, error) {
	// 调用 StatObject 检查对象元信息
	_, err := c.client.StatObject(ctx, c.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		// 如果是 NotFound 错误，则返回 false
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		// 其他错误（网络、权限等）则直接返回错误
		return false, fmt.Errorf("stat object failed: %w", err)
	}
	return true, nil
}
