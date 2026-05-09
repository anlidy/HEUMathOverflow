package service

import (
	"MathOverflow/common/config"
	common "MathOverflow/common/model"
	"MathOverflow/common/utils"
	syserror "MathOverflow/services/forum/internal/model/error"
	"MathOverflow/services/forum/internal/repository"
	"context"
)

type ForumService interface {
	UploadFile(ctx context.Context, file common.File) (string, syserror.Error)
	DownloadFile(ctx context.Context, filename string) (*common.File, syserror.Error)
}

type forumService struct {
	cfg      config.Config
	fileRepo repository.FileRepo
	servName string
}

func NewForumService(cfg config.Config, fileRepo repository.FileRepo) ForumService {
	return &forumService{
		cfg:      cfg,
		fileRepo: fileRepo,
		servName: "Forum-Service",
	}
}

// 上传文件,返回临时链接
func (s *forumService) UploadFile(ctx context.Context, file common.File) (string, syserror.Error) {
	// 将文件存入minio
	url, err := s.fileRepo.UploadFile(ctx, s.cfg.Minio.Bucket, file)
	if err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("upload file failed")
		return "", syserror.InternalError
	}
	return url, syserror.NoError
}

// 下载文件
func (s *forumService) DownloadFile(ctx context.Context, filename string) (*common.File, syserror.Error) {
	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, filename)
	if err != nil {
		return nil, syserror.InternalError
	} else if !exists {
		return nil, syserror.NotFoundError
	}
	// 从minio流式读取文件
	file, err := s.fileRepo.DownloadFile(ctx, filename)
	if err != nil {
		return nil, syserror.InternalError
	}
	return file, syserror.NoError
}
