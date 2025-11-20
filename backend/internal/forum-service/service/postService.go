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
	"time"

	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PostService interface {
	CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error)
	GetOnePost(ctx context.Context, postID int64) (*response.UserInfo, *response.PostData, syserror.Error)
}

type postService struct {
	cfg            config.Config
	postRepo       repository.PostRepo
	fileRepo       repository.FileRepo
	servName       string
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
}

func (s *postService) getUserClient() (userpb.UserServiceClient, error) {
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

func NewPostService(cfg config.Config, postRepo repository.PostRepo, fileRepo repository.FileRepo) PostService {
	return &postService{
		cfg:      cfg,
		postRepo: postRepo,
		fileRepo: fileRepo,
		servName: "Post-Service",
	}
}

// 创建新帖子
func (s *postService) CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error) {
	// 生成帖子id
	var postID = utils.GenerateSnowflakeID()

	// 将临时url提升为正式url
	var images = []string{}
	for _, tmpUrl := range req.ImageURL {
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, postID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return -1, syserror.ResourceExpiredError
			}
			log.Printf("[%s] %v\n", s.servName, err)
			return -1, syserror.InternalError
		}
		images = append(images, newUrl)
	}

	// 创建mongo文档
	var postContent = &model.PostContent{
		PostID:    postID,
		Content:   req.Content,
		ImageURLs: images,
		CreatedAt: time.Now(),
	}
	docID, err := s.postRepo.CreateOnePost(ctx, postContent)
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}

	// 创建帖子元数据
	var post = &model.Post{
		ID:       postID,
		AuthorID: userID,
		Title:    req.Title,
		Tags:     pq.StringArray(req.Tags),
		Status:   model.Unanswered,
		DocID:    utils.ObjectIDToString(docID),
	}
	err = s.postRepo.CreatePost(post)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			return -1, syserror.DuplicateError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}
	return postID, syserror.NoError
}

// 获取一条帖子
func (s *postService) GetOnePost(ctx context.Context, postID int64) (*response.UserInfo, *response.PostData, syserror.Error) {
	// 查询帖子元信息
	post, err := s.postRepo.FindPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.InternalError
	}
	// 查询帖子正文内容
	postContent, err := s.postRepo.FindOnePost(ctx, bson.M{"post_id": postID})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, nil, syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		return nil, nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: post.AuthorID})
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
		default:
			log.Printf("GRPC请求失败! code:%v err:%v\n", st.Code(), err)
			return nil, nil, syserror.InternalError
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
	postData := &response.PostData{
		Post:        post,
		PostContent: postContent,
	}

	return userInfo, postData, syserror.NoError
}
