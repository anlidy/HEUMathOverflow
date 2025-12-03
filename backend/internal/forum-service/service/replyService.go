package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/model"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/model/response"
	"MathOverflow/internal/forum-service/repository"
	userpb "MathOverflow/proto/user"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ReplyService interface {
	CreateNewReply(ctx context.Context, userID int64, req request.ReplyCreate) (int64, syserror.Error)
	GetOneReply(ctx context.Context, replyID int64) (*response.UserInfo, *response.ReplyData, syserror.Error)
	GetManyReplies(ctx context.Context, replyID int64, offset, limit int) ([]response.MultiReplyData, syserror.Error)
	UpdateOneReply(ctx context.Context, userID int64, replyID int64, req request.ReplyUpdate) syserror.Error
	DeleteOneReply(ctx context.Context, replyID, userID int64, role int) syserror.Error
	LikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error
	CancelLikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error
	HasLikedReply(ctx context.Context, replyID, userID int64) (bool, syserror.Error)
}

type replyService struct {
	cfg            config.Config
	replyRepo      repository.ReplyRepo
	fileRepo       repository.FileRepo
	servName       string
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
}

func (s *replyService) getUserClient() (userpb.UserServiceClient, error) {
	var initErr error

	s.userClientOnce.Do(func() {
		// 使用局部变量接收
		conn, err := client.GetGRPCConn(s.cfg.GRPC.UserServiceAddr)
		if err != nil {
			initErr = err
			return
		}
		s.userClient = userpb.NewUserServiceClient(conn)
	})

	// 如果初始化失败，下次允许重试
	if initErr != nil {
		s.userClientOnce = sync.Once{}
		return nil, initErr
	}

	return s.userClient, nil
}

func NewReplyService(cfg config.Config, replyRepo repository.ReplyRepo, fileRepo repository.FileRepo) ReplyService {
	return &replyService{
		cfg:       cfg,
		replyRepo: replyRepo,
		fileRepo:  fileRepo,
		servName:  "Reply-Service",
	}
}

// 创建新回贴
func (s *replyService) CreateNewReply(ctx context.Context, userID int64, req request.ReplyCreate) (int64, syserror.Error) {
	// 生成回帖id
	var replyID = utils.GenerateSnowflakeID()
	// 将临时url提升为正式url
	var images = []string{}
	for _, tmpUrl := range req.ImageURLs {
		if strings.TrimSpace(tmpUrl) == "" {
			continue
		}
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, replyID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return -1, syserror.ResourceExpiredError
			}
			log.Printf("[%s] %v\n", s.servName, err)
			return -1, syserror.InternalError
		}
		images = append(images, newUrl)
	}
	var voice string
	if req.VoiceURL != "" {
		var err error
		voice, err = s.fileRepo.PromoteFile(ctx, req.VoiceURL, s.cfg.Minio.Bucket, replyID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return -1, syserror.ResourceExpiredError
			}
			log.Printf("[%s] %v\n", s.servName, err)
			return -1, syserror.InternalError
		}
	}

	// 创建回帖元数据
	var reply = &model.Reply{
		ID:            replyID,
		PostID:        req.PostID,
		ReplierID:     userID,
		Content:       req.Content,
		VoiceURL:      voice,
		ImageURLs:     images,
		AIAnswered:    false,
		Status:        model.NotSelected,
		ParentReplyID: req.ParentReplyID,
	}
	// 创建回帖
	err := s.replyRepo.CreateReply(reply)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return -1, syserror.DuplicateError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}
	// 更新原帖的状态
	return replyID, syserror.NoError
}

// 获取一条回帖
func (s *replyService) GetOneReply(ctx context.Context, replyID int64) (*response.UserInfo, *response.ReplyData, syserror.Error) {
	// 查询回帖信息
	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: reply.ReplierID})
	if err != nil {
		log.Printf("[%s] 调用 GetUserInfo 失败: %v\n", s.servName, err)
		st, ok := status.FromError(err)
		if !ok {
			log.Println("非 gRPC 错误:", err)
			return nil, nil, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return nil, nil, syserror.InternalError
		case codes.NotFound:
			return nil, nil, syserror.NotFoundError
		case codes.Canceled:
			return nil, nil, syserror.NetworkError
		}
	}
	// 请求成功
	var userInfo = &response.UserInfo{
		ID:        resp.UserId,
		Username:  resp.Username,
		Role:      int(resp.Role),
		AvatarUrl: resp.AvatarUrl,
	}
	// 聚合返回查询结果
	replyData := &response.ReplyData{
		Reply: reply,
	}

	return userInfo, replyData, syserror.NoError
}

// 获取分页帖子
func (s *replyService) GetManyReplies(ctx context.Context, replyID int64, offset, limit int) ([]response.MultiReplyData, syserror.Error) {
	// 查询回帖信息
	replies, err := s.replyRepo.FindRepliesByPostID(replyID, offset, limit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, syserror.InternalError
	}
	// 查询回帖正文内容
	// 构建查询的docID和userID数组
	userIDs := make([]int64, len(replies))
	for i, reply := range replies {
		userIDs[i] = reply.ReplierID
	}
	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: userIDs})
	if err != nil {
		log.Printf("[%s] 调用 BatchGetUserInfo 失败: %v\n", s.servName, err)
		st, ok := status.FromError(err)
		if !ok {
			log.Println("非 gRPC 错误:", err)
			return nil, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return nil, syserror.InternalError
		case codes.Canceled:
			return nil, syserror.NetworkError
		default:
			log.Println(st.Message())
		}
	}
	// 请求成功
	var userMap = resp.Users
	// 聚合返回查询结果
	replyDatas := make([]response.MultiReplyData, len(replies))
	for i := range replies {
		var user = userMap[replies[i].ReplierID] // 不存在的用户查询得到空值
		replyDatas[i].UserInfo = response.UserInfo{
			ID:        user.UserId,
			Username:  user.Username,
			Role:      int(user.Role),
			AvatarUrl: user.AvatarUrl,
		}
		replyDatas[i].ReplyData.Reply = replies[i]
	}
	return replyDatas, syserror.NoError
}

// 更新一条回帖
func (s *replyService) UpdateOneReply(ctx context.Context, userID int64, replyID int64, req request.ReplyUpdate) syserror.Error {
	// 查询该条回帖
	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}
	// 验证当前登录的用户是否为该回帖的作者
	if userID != reply.ReplierID {
		return syserror.PermissionDeniedError
	}

	// 完成内容修改
	reply.Content = req.Content
	var oldVoiceURL string
	if reply.VoiceURL != req.VoiceURL {
		oldVoiceURL = reply.VoiceURL
		reply.VoiceURL = req.VoiceURL
		reply.VoiceText = ""
	}
	// 删除指定的url
	reply.ImageURLs = utils.SliceFilter(reply.ImageURLs, req.DeleteImageURLs)
	// 将临时url提升为正式url
	for _, tmpUrl := range req.AddImageURLs {
		if strings.TrimSpace(tmpUrl) == "" {
			continue
		}
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, replyID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return syserror.ResourceExpiredError
			}
			log.Printf("[%s] %v\n", s.servName, err)
			return syserror.InternalError
		}
		// 完成添加
		reply.ImageURLs = append(reply.ImageURLs, newUrl)
	}

	// 保存回帖信息
	_, err = s.replyRepo.UpdateReply(&reply)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}

	// 后台删除图片
	go func() {
		for _, delUrl := range append(req.DeleteImageURLs, oldVoiceURL) {
			filename := strings.TrimPrefix(delUrl, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			err := s.fileRepo.DeleteFile(ctx, filename)
			if err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					log.Printf("[Minio] 找不到文件: %s\n", filename)
					continue
				}
				log.Printf("[%s] %v\n", s.servName, err)
				return
			}
		}
	}()
	return syserror.NoError
}

// 删除一条回帖
func (s *replyService) DeleteOneReply(ctx context.Context, replyID, userID int64, role int) syserror.Error {
	// 查询回帖信息
	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: reply.ReplierID})
	if err != nil {
		log.Printf("[%s] 调用 GetUserInfo 失败: %v\n", s.servName, err)
		st, ok := status.FromError(err)
		if !ok {
			log.Println("非 gRPC 错误:", err)
			return syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return syserror.InternalError
		case codes.NotFound:
			return syserror.NotFoundError
		case codes.Canceled:
			return syserror.NetworkError
		}
	}
	// 验证操作者权限是否低于回帖人
	if role <= int(resp.Role) {
		// 验证当前登录的用户是否为回帖人
		if userID != reply.ReplierID {
			return syserror.PermissionDeniedError
		}
	}

	// 删除回帖
	err = s.replyRepo.DeleteReply(replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}

	// 后台删除回帖包含的文件
	go func() {
		urls := append(reply.ImageURLs, reply.VoiceURL)
		for _, url := range urls {
			filename := strings.TrimPrefix(url, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			err := s.fileRepo.DeleteFile(ctx, url)
			if err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					log.Printf("[Minio] 找不到文件: %s\n", filename)
					continue
				}
				log.Printf("[%s] %v\n", s.servName, err)
				return
			}
		}
	}()
	return syserror.NoError
}

// 给回帖点赞
func (s *replyService) LikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error {
	var replyLike = model.ReplyLike{
		UserID:  userID,
		ReplyID: replyID,
	}
	err := s.replyRepo.CreateReplyLike(&replyLike)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return syserror.DuplicateError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}
	return syserror.NoError
}

// 取消点赞
func (s *replyService) CancelLikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error {
	err := s.replyRepo.DeleteReplyLike(userID, replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return syserror.InternalError
	}
	return syserror.NoError
}

// 查询用户是否点赞过指定回复
func (s *replyService) HasLikedReply(ctx context.Context, replyID, userID int64) (bool, syserror.Error) {
	liked, err := s.replyRepo.HasReplyLike(userID, replyID)
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return false, syserror.InternalError
	}
	return liked, syserror.NoError
}
