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
	"log"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ReplyService interface {
	CreateNewReply(ctx context.Context, userID int64, req request.ReplyCreate) (int64, syserror.Error)
	GetOneReply(ctx context.Context, userID, replyID int64) (*response.UserInfo, *response.ReplyData, syserror.Error)
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
	var err error

	s.userClientOnce.Do(func() {
		var userConn *grpc.ClientConn
		userConn, err = client.GetGRPCConn(s.cfg.GRPC.UserServiceAddr)
		if err != nil {
			return
		}
		s.userClient = userpb.NewUserServiceClient(userConn)
	})

	// 允许再次初始化
	if err != nil {
		s.userClientOnce = sync.Once{}
	}

	return s.userClient, err
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
	for _, tmpUrl := range req.Images {
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

	voice, err := s.fileRepo.PromoteFile(ctx, req.Voice, s.cfg.Minio.Bucket, replyID)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return -1, syserror.ResourceExpiredError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}

	// 创建mongo文档
	var replyContent = &model.ReplyContent{
		ReplyID:    replyID,
		Content:    req.Content,
		Voice:      voice,
		Images:     images,
		AIAnswered: false,
	}
	docID, err := s.replyRepo.MCreateReply(ctx, replyContent)
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}

	// 创建回帖元数据
	var reply = &model.Reply{
		ID:            replyID,
		PostID:        req.PostID,
		ReplierID:     userID,
		ParentReplyID: req.ParentReplyID,
		DocID:         utils.ObjectIDToString(docID),
		Status:        model.NotSelected,
	}

	err = s.replyRepo.CreateReply(reply)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			return -1, syserror.DuplicateError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}
	return replyID, syserror.NoError
}

// 获取一条回帖
func (s *replyService) GetOneReply(ctx context.Context, userID, replyID int64) (*response.UserInfo, *response.ReplyData, syserror.Error) {
	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		return nil, nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: userID})
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
		}
	}
	// 请求成功
	var userInfo = &response.UserInfo{
		ID:        userID,
		Username:  resp.Username,
		Role:      int(resp.Role),
		AvatarUrl: resp.AvatarUrl,
	}
	// 查询回帖元信息
	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.InternalError
	}
	// 查询回帖正文内容
	replyContent, err := s.replyRepo.MFindReply(ctx, bson.M{"reply_id": replyID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.InternalError
	}
	// 聚合返回查询结果
	replyData := &response.ReplyData{}
	copier.Copy(replyData, &reply)
	copier.Copy(replyData, replyContent)

	return userInfo, replyData, syserror.NoError
}
