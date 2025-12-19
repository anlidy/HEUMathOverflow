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
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ReplyService interface {
	CreateNewReply(ctx context.Context, userID int64, req request.ReplyCreate) (int64, syserror.Error)
	GetOneReply(ctx context.Context, replyID, userID int64) (*response.UserInfo, *response.ReplyData, syserror.Error)
	GetManyReplies(ctx context.Context, postID, userID int64, page, limit int) ([]response.MultiReplyData, int64, syserror.Error)
	UpdateOneReply(ctx context.Context, userID int64, replyID int64, req request.ReplyUpdate) syserror.Error
	DeleteOneReply(ctx context.Context, replyID, userID int64, role int) syserror.Error
	LikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error
	CancelLikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error
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
	logger := utils.WithContext(ctx).WithField("service", s.servName)
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
			logger.WithError(err).Error("promote reply image file failed")
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
			logger.WithError(err).Error("promote reply voice file failed")
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
		logger.WithError(err).Error("create reply failed")
		return -1, syserror.InternalError
	}
	// 后台同步
	go func() {
		// redis的replies+1
		var ctx = context.Background()
		if err := s.replyRepo.IncreasePostReplies(ctx, req.PostID); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post replies failed")
		}
	}()
	return replyID, syserror.NoError
}

// 获取一条回帖
func (s *replyService) GetOneReply(ctx context.Context, replyID, userID int64) (*response.UserInfo, *response.ReplyData, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询回帖信息
	reply, err := s.replyRepo.FindReplyDetailByID(replyID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		logger.WithError(err).Error("find reply detail failed")
		return nil, nil, syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return nil, nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: reply.ReplierID})
	if err != nil {
		logger.WithError(err).Error("call GetUserInfo failed")
		st, ok := status.FromError(err)
		if !ok {
			logger.WithError(err).Error("non gRPC error when calling GetUserInfo")
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
		Reply: reply.Reply,
		Liked: reply.Liked,
	}

	return userInfo, replyData, syserror.NoError
}

// 获取分页帖子
func (s *replyService) GetManyReplies(ctx context.Context, postID, userID int64, page, limit int) ([]response.MultiReplyData, int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询回帖信息
	offset := (page - 1) * limit // 计算偏移量
	replies, total, err := s.replyRepo.FindReplyDetailsByPostID(postID, userID, offset, limit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, syserror.NotFoundError
		}
		logger.WithError(err).Error("find reply details by post id failed")
		return nil, 0, syserror.InternalError
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
		logger.WithError(err).Error("get user client failed")
		return nil, 0, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: userIDs})
	if err != nil {
		logger.WithError(err).Error("call BatchGetUserInfo failed")
		st, ok := status.FromError(err)
		if !ok {
			logger.WithError(err).Error("non gRPC error when calling BatchGetUserInfo")
			return nil, 0, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return nil, 0, syserror.InternalError
		case codes.Canceled:
			return nil, 0, syserror.NetworkError
		default:
			logger.WithField("grpc_message", st.Message()).Warn("grpc error when calling BatchGetUserInfo")
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
		replyDatas[i].ReplyData.Reply = replies[i].Reply
		replyDatas[i].ReplyData.Liked = replies[i].Liked
	}
	return replyDatas, total, syserror.NoError
}

// 更新一条回帖
func (s *replyService) UpdateOneReply(ctx context.Context, userID int64, replyID int64, req request.ReplyUpdate) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询该条回帖
	reply, err := s.replyRepo.FindReplyDetailByID(replyID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find reply detail failed")
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
			logger.WithError(err).Error("promote reply image file failed")
			return syserror.InternalError
		}
		// 完成添加
		reply.ImageURLs = append(reply.ImageURLs, newUrl)
	}

	// 保存回帖信息
	_, err = s.replyRepo.UpdateReply(&reply.Reply)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("update reply failed")
		return syserror.InternalError
	}

	// 后台删除图片
	go func() {
		var ctx = context.Background()
		for _, delUrl := range append(req.DeleteImageURLs, oldVoiceURL) {
			filename := strings.TrimPrefix(delUrl, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			if err := s.fileRepo.DeleteFile(ctx, filename); err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					utils.WithContext(ctx).WithField("service", s.servName).WithField("filename", filename).Info("minio file not found when deleting reply file")
					continue
				}
				utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("delete reply file failed")
				return
			}
		}
	}()
	return syserror.NoError
}

// 删除一条回帖
func (s *replyService) DeleteOneReply(ctx context.Context, replyID, userID int64, role int) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询回帖信息
	reply, err := s.replyRepo.FindReplyDetailByID(replyID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find reply detail failed")
		return syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: reply.ReplierID})
	if err != nil {
		logger.WithError(err).Error("call GetUserInfo failed")
		st, ok := status.FromError(err)
		if !ok {
			logger.WithError(err).Error("non gRPC error when calling GetUserInfo")
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
		logger.WithError(err).Error("delete reply failed")
		return syserror.InternalError
	}

	// 后台删除回帖包含的文件
	go func() {
		var ctx = context.Background()
		urls := append(reply.ImageURLs, reply.VoiceURL)
		for _, url := range urls {
			filename := strings.TrimPrefix(url, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			if err := s.fileRepo.DeleteFile(ctx, url); err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					utils.WithContext(ctx).WithField("service", s.servName).WithField("filename", filename).Info("minio file not found when deleting reply file")
					continue
				}
				utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("delete reply file failed")
				return
			}
		}
	}()
	// 后台同步
	go func() {
		// redis的replies-1
		var ctx = context.Background()
		if err := s.replyRepo.DecreasePostReplies(ctx, reply.PostID); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decrease post replies failed")
		}
	}()
	return syserror.NoError
}

// 给回帖点赞
func (s *replyService) LikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	var replyLike = model.ReplyLike{
		UserID:  userID,
		ReplyID: replyID,
	}
	err := s.replyRepo.CreateReplyLike(&replyLike)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return syserror.DuplicateError
		}
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create reply like failed")
		return syserror.InternalError
	}
	return syserror.NoError
}

// 取消点赞
func (s *replyService) CancelLikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	err := s.replyRepo.DeleteReplyLike(userID, replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("delete reply like failed")
		return syserror.InternalError
	}
	return syserror.NoError
}
