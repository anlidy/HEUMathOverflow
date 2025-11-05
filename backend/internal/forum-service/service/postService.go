package service

import (
	"MathOverflow/internal/common/config"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/model"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/repository"
	userpb "MathOverflow/proto/user"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService interface {
	UploadFile(ctx context.Context, file common.File) (string, syserror.Error)
}

type postService struct {
	cfg        config.Config
	postRepo   repository.PostRepo
	fileRepo   repository.FileRepo
	userClient userpb.UserServiceClient
	servName   string
}

func NewPostService(cfg config.Config, postRepo repository.PostRepo, fileRepo repository.FileRepo, userConn *grpc.ClientConn) PostService {
	return &postService{
		cfg:        cfg,
		postRepo:   postRepo,
		fileRepo:   fileRepo,
		userClient: userpb.NewUserServiceClient(userConn),
		servName:   "Post-Service",
	}
}

func (s *postService) UploadFile(ctx context.Context, file common.File) (string, syserror.Error) {
	// 将文件存入minio
	url, err := s.fileRepo.UploadFile(ctx, s.cfg.Minio.Bucket, file)
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return "", syserror.InternalError
	}
	return url, syserror.NoError
}

func (s *postService) CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error) {
	// 调用 UserService
	resp, err := s.userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		log.Println("调用 GetUserInfo 失败:", err)
		st, ok := status.FromError(err)
		if !ok {
			log.Println("非 gRPC 错误:", err)
			return -1, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return -1, syserror.InternalError
		case codes.NotFound:
			return -1, syserror.NotFoundError
		}
	}
	// 请求正常
	// 创建mongo文档
	
	// 创建帖子元数据
	var postID = utils.GenerateSnowflakeID()
	var post = model.Post{
		ID:       postID,
		AuthorID: userID,
		Title:    req.Title,
		Status:   model.Unanswered,
		Tags:     req.Tags,
	}
	fmt.Println(resp.Username, post)
	return postID, syserror.NoError
}
