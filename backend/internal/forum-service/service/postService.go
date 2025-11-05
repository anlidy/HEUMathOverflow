package service

import (
	"MathOverflow/internal/common/config"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/model"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/model/response"
	"MathOverflow/internal/forum-service/repository"
	userpb "MathOverflow/proto/user"
	"context"
	"log"
	"time"

	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PostService interface {
	UploadFile(ctx context.Context, file common.File) (string, syserror.Error)
	CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error)
	GetOnePost(ctx context.Context, userID, postID int64) (*response.UserInfo, *response.PostData, syserror.Error)
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
	// 生成帖子id
	var postID = utils.GenerateSnowflakeID()

	// 将临时url提升为正式url
	var images = []string{}
	for _, tmpUrl := range req.Images {
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, postID)
		if err != nil {
			continue
		}
		images = append(images, newUrl)
	}

	// 创建mongo文档
	var postContent = &model.PostContent{
		PostID:    postID,
		Content:   req.Content,
		Images:    images,
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
		Tags:     req.Tags,
		Status:   model.Unanswered,
		DocID:    utils.ObjectIDToString(docID),
	}
	err = s.postRepo.CreatePost(post)
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return -1, syserror.InternalError
	}
	return postID, syserror.NoError
}

// 查询一条帖子
func (s *postService) GetOnePost(ctx context.Context, userID, postID int64) (*response.UserInfo, *response.PostData, syserror.Error) {
	// 调用 UserService
	resp, err := s.userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: userID})
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
	// 聚合返回查询结果
	postData := &response.PostData{}
	copier.Copy(postData, &post)
	postData.Content = postContent.Content
	postData.Images = postContent.Images

	return userInfo, postData, syserror.NoError
}
