package client

import (
	"MathOverflow/internal/common/config"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
)

type MinioClinet struct {
	Client *minio.Client
	Bucket string
}

// 初始化minio
func InitMinIO(cfg config.MinioConfig) (*MinioClinet, error) {
	maxRetries := 5
	retryInterval := 3 * time.Second

	var client *minio.Client
	var err error

	for i := 1; i <= maxRetries; i++ {
		client, err = minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: cfg.UseSSL,
		})
		if err == nil {
			fmt.Println("MinIO connected:", cfg.Endpoint)
			break
		}
		utils.Logger().WithField("retry", i).WithField("max_retries", maxRetries).WithError(err).Error("连接 MinIO 失败")
		if i < maxRetries {
			time.Sleep(retryInterval)
			utils.Logger().Info("正在重试连接 MinIO...")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client after %d retries: %w", maxRetries, err)
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

	// 创建自动清理临时文件规则
	if cfg.CleanTmp {
		config := lifecycle.NewConfiguration()
		rule := lifecycle.Rule{
			ID:     "DeleteTempUploads",
			Status: "Enabled",
			RuleFilter: lifecycle.Filter{
				Prefix: "tmp/",
			},
			Expiration: lifecycle.Expiration{
				Days: lifecycle.ExpirationDays(cfg.CleanCycleDays), // 自动清理1天前的对象
			},
		}
		config.Rules = []lifecycle.Rule{rule}
		err = client.SetBucketLifecycle(ctx, cfg.Bucket, config)
		if err != nil {
			utils.Logger().WithError(err).Fatal("设置 MinIO 生命周期失败")
		}
		fmt.Printf("已为 bucket: %s 启用临时文件自动清理,清理周期为 %d天.\n", cfg.Bucket, cfg.CleanCycleDays)
	}

	return &MinioClinet{Client: client, Bucket: cfg.Bucket}, nil
}

// 流式上传文件
func (c *MinioClinet) Upload(ctx context.Context, objectName string, data io.ReadCloser, size int64, contentType string) error {
	_, err := c.Client.PutObject(ctx, c.Bucket, objectName, data, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// 流式下载文件
func (c *MinioClinet) Download(ctx context.Context, objectName string) (*common.File, error) {
	obj, err := c.Client.GetObject(ctx, c.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object failed: %w", err)
	}

	// 获取对象元数据
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, fmt.Errorf("get object stat failed: %w", err)
	}

	return &common.File{
		Data:        obj,
		ContentType: stat.ContentType,
		Size:        stat.Size,
		Filename:    path.Base(objectName),
	}, nil
}

// 删除文件
func (c *MinioClinet) Delete(ctx context.Context, objectName string) error {
	return c.Client.RemoveObject(ctx, c.Bucket, objectName, minio.RemoveObjectOptions{})
}

// Exists 检查对象是否存在
func (c *MinioClinet) Exists(ctx context.Context, objectName string) (bool, error) {
	// 调用 StatObject 检查对象元信息
	_, err := c.Client.StatObject(ctx, c.Bucket, objectName, minio.StatObjectOptions{})
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
