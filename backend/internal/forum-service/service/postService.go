package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/event"
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
	"time"

	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PostService interface {
	CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error)
	GetOnePost(ctx context.Context, postID, userID int64) (*response.UserInfo, *response.PostData, syserror.Error)
	GetManyPosts(ctx context.Context, page, limit int, order model.OrderBy) ([]response.MultiPostData, syserror.Error)
	UpdateOnePost(ctx context.Context, userID int64, postID int64, req request.PostUpdate) syserror.Error
	DeleteOnePost(ctx context.Context, postID, userID int64, role int) syserror.Error
	LikeOnePost(ctx context.Context, postID, userID int64) syserror.Error
	CancelLikeOnePost(ctx context.Context, postID, userID int64) syserror.Error
	StarOnePost(ctx context.Context, postID, userID int64) syserror.Error
	CancelStarOnePost(ctx context.Context, postID, userID int64) syserror.Error
	GetUserStarredPosts(ctx context.Context, userID int64, page, limit int) ([]response.MultiPostData, int64, syserror.Error)
}

type postService struct {
	cfg            config.Config
	postRepo       repository.PostRepo
	fileRepo       repository.FileRepo
	servName       string
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
	mq             *client.RabbitMQClient
}

func (s *postService) getUserClient() (userpb.UserServiceClient, error) {
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

func NewPostService(cfg config.Config, rabbit *client.RabbitMQClient, postRepo repository.PostRepo, fileRepo repository.FileRepo) PostService {
	return &postService{
		cfg:      cfg,
		mq:       rabbit,
		postRepo: postRepo,
		fileRepo: fileRepo,
		servName: "Post-Service",
	}
}

// 创建新帖子
func (s *postService) CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return 0, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		logger.WithError(err).Error("call GetUserInfo failed")
		st, ok := status.FromError(err)
		if !ok {
			logger.WithError(err).Error("non gRPC error when calling GetUserInfo")
			return 0, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return 0, syserror.InternalError
		case codes.NotFound:
			return 0, syserror.NotFoundError
		case codes.Canceled:
			return 0, syserror.NetworkError
		default:
			logger.WithField("code", st.Code()).WithError(err).Error("grpc request failed when calling GetUserInfo")
			return 0, syserror.InternalError
		}
	}
	// 获取作者名称
	authorName := resp.Username

	// 生成帖子id
	var postID = utils.GenerateSnowflakeID()

	// 将临时url提升为正式url
	var images = []string{}
	for _, tmpUrl := range req.ImageURLs {
		if strings.TrimSpace(tmpUrl) == "" {
			continue
		}
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, postID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return -1, syserror.ResourceExpiredError
			}
			logger.WithError(err).Error("promote post image file failed")
			return -1, syserror.InternalError
		}
		images = append(images, newUrl)
	}

	// 创建帖子数据
	var post = model.Post{
		ID:        postID,
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		ImageURLs: images,
		Tags:      pq.StringArray(req.Tags),
		Status:    model.Unanswered,
		CreatedAt: time.Now(),
	}
	err = s.postRepo.CreatePost(&post)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return -1, syserror.DuplicateError
		}
		logger.WithError(err).Error("create post failed")
		return -1, syserror.InternalError
	}

	// 发布帖子创建消息
	var payload = event.ForumPostPayload{
		PostID:     postID,
		AuthorID:   post.AuthorID,
		AuthorName: authorName, // 创建时需要存一次用户名
		Title:      post.Title,
		Content:    post.Content,
		Tags:       post.Tags,
		Status:     int(post.Status),
		CreatedAt:  post.CreatedAt,
	}
	event.PublishPostEvent(s.mq, event.ForumPostCreated, payload)
	return postID, syserror.NoError
}

// 获取一条帖子
func (s *postService) GetOnePost(ctx context.Context, postID, userID int64) (*response.UserInfo, *response.PostData, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询帖子信息
	post, err := s.postRepo.FindPostDetail(postID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		logger.WithError(err).Error("find post detail failed")
		return nil, nil, syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return nil, nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: post.AuthorID})
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
		default:
			logger.WithField("code", st.Code()).WithError(err).Error("grpc request failed when calling GetUserInfo")
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
		Post:    post.Post,
		Liked:   post.Liked,
		Starred: post.Starred,
	}

	// 后台完成同步
	go func() {
		// redis的views+1
		var ctx = context.Background()
		if err := s.postRepo.IncreasePostStat(ctx, postID, "views"); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post views failed")
		}
	}()

	return userInfo, postData, syserror.NoError
}

// 批量获取帖子
func (s *postService) GetManyPosts(ctx context.Context, page, limit int, order model.OrderBy) ([]response.MultiPostData, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询帖子信息
	offset := (page - 1) * limit // 计算偏移量
	posts, err := s.postRepo.FindManyPosts(offset, limit, order)
	if err != nil {
		logger.WithError(err).Error("find many posts failed")
		return nil, syserror.InternalError
	}
	// 构建查询的userID数组
	userIDs := make([]int64, len(posts))
	for i, post := range posts {
		userIDs[i] = post.AuthorID
	}
	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return nil, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: userIDs})
	if err != nil {
		logger.WithError(err).Error("call BatchGetUserInfo failed")
		st, ok := status.FromError(err)
		if !ok {
			logger.WithError(err).Error("non gRPC error when calling BatchGetUserInfo")
			return nil, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return nil, syserror.InternalError
		case codes.Canceled:
			return nil, syserror.NetworkError
		default:
			logger.WithField("grpc_message", st.Message()).Warn("grpc error when calling BatchGetUserInfo")
		}
	}
	// 请求成功
	var userMap = resp.Users
	// 聚合返回查询结果
	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		var user = userMap[posts[i].AuthorID] // 不存在的用户查询得到空值
		postDatas[i].UserInfo = response.UserInfo{
			ID:        user.UserId,
			Username:  user.Username,
			Role:      int(user.Role),
			AvatarUrl: user.AvatarUrl,
		}
		postDatas[i].PostData.Post = posts[i]
	}
	return postDatas, syserror.NoError
}

// 更新一条帖子
func (s *postService) UpdateOnePost(ctx context.Context, userID int64, postID int64, req request.PostUpdate) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询该条帖子
	post, err := s.postRepo.FindPostDetail(postID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find post detail failed")
		return syserror.InternalError
	}
	// 验证当前登录的用户是否为该帖子的作者
	if userID != post.AuthorID {
		return syserror.PermissionDeniedError
	}

	// 完成内容修改
	post.Title = req.Title
	post.Tags = req.Tags
	post.Content = req.Content
	// 删除指定的url
	post.ImageURLs = utils.SliceFilter(post.ImageURLs, req.DeleteImageURLs)
	// 将临时url提升为正式url
	for _, tmpUrl := range req.AddImageURLs {
		if strings.TrimSpace(tmpUrl) == "" {
			continue
		}
		newUrl, err := s.fileRepo.PromoteFile(ctx, tmpUrl, s.cfg.Minio.Bucket, postID)
		if err != nil {
			if minio.ToErrorResponse(err).Code == "NoSuchKey" {
				return syserror.ResourceExpiredError
			}
			logger.WithError(err).Error("promote post image file failed")
			return syserror.InternalError
		}
		// 完成添加
		fmt.Println(newUrl)
		post.ImageURLs = append(post.ImageURLs, newUrl)
	}

	// 保存帖子信息
	_, err = s.postRepo.UpdatePost(&post.Post)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("update post failed")
		return syserror.InternalError
	}

	// 发布事件
	var payload = event.ForumPostPayload{
		PostID:    postID,
		Title:     post.Title,
		Content:   post.Content,
		Tags:      post.Tags,
		Status:    int(post.Status),
		CreatedAt: post.CreatedAt,
		Views:     post.Views,
		Likes:     post.Likes,
		Stars:     post.Stars,
		Replies:   post.Replies,
	}
	event.PublishPostEvent(s.mq, event.ForumPostUpdated, payload)

	// 后台删除图片
	go func(ctx context.Context) {
		for _, delUrl := range req.DeleteImageURLs {
			filename := strings.TrimPrefix(delUrl, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			if err := s.fileRepo.DeleteFile(ctx, filename); err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					utils.WithContext(ctx).WithField("service", s.servName).WithField("filename", filename).Info("minio file not found when deleting post file")
					continue
				}
				utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("delete post file failed")
				return
			}
		}
	}(ctx)
	return syserror.NoError
}

// 删除一条帖子
func (s *postService) DeleteOnePost(ctx context.Context, postID, userID int64, role int) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 查询该条帖子
	post, err := s.postRepo.FindPostDetail(postID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find post detail failed")
		return syserror.InternalError
	}
	// 获取帖子作者信息
	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		logger.WithError(err).Error("get user client failed")
		return syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: post.AuthorID})
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
		default:
			logger.WithField("code", st.Code()).WithError(err).Error("grpc request failed when calling GetUserInfo")
			return syserror.InternalError
		}
	}
	// 验证操作者权限是否低于帖子作者
	if role <= int(resp.Role) {
		// 验证当前登录的用户是否为该帖子的作者
		if userID != post.AuthorID {
			return syserror.PermissionDeniedError
		}
	}

	// 删除帖子
	// 附属的回贴会一并删除
	err = s.postRepo.DeletePost(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("delete post failed")
		return syserror.InternalError
	}

	// 发布事件
	var payload = event.ForumPostPayload{PostID: postID}
	event.PublishPostEvent(s.mq, event.ForumPostDeleted, payload)

	// 后台删除帖子包含的文件
	go func(ctx context.Context) {
		urls := post.ImageURLs
		for _, url := range urls {
			filename := strings.TrimPrefix(url, fmt.Sprintf("/api/v1/%s/file/", s.cfg.Minio.Bucket))
			if err := s.fileRepo.DeleteFile(ctx, url); err != nil {
				if minio.ToErrorResponse(err).Code == "NoSuchKey" {
					utils.WithContext(ctx).WithField("service", s.servName).WithField("filename", filename).Info("minio file not found when deleting post file")
					continue
				}
				utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("delete post file failed")
				return
			}
		}
	}(ctx)

	return syserror.NoError
}

// 给帖子点赞
func (s *postService) LikeOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	var postLike = model.PostLike{
		UserID: userID,
		PostID: postID,
	}
	err := s.postRepo.CreatePostLike(&postLike)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return syserror.DuplicateError
		}
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create post like failed")
		return syserror.InternalError
	}
	// 后台完成同步
	go func() {
		// redis的likes+1
		var ctx = context.Background()
		if err := s.postRepo.IncreasePostStat(ctx, postID, "likes"); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post likes failed")
		}
	}()
	return syserror.NoError
}

// 取消点赞
func (s *postService) CancelLikeOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	err := s.postRepo.DeletePostLike(userID, postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("delete post like failed")
		return syserror.InternalError
	}
	// 后台完成同步
	go func() {
		// redis的likes-1
		var ctx = context.Background()
		if err := s.postRepo.DecreasePostStat(ctx, postID, "likes"); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decrease post likes failed")
		}
	}()
	return syserror.NoError
}

// / 收藏接口
// 收藏帖子
func (s *postService) StarOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	var postStar = model.PostStar{
		ID:     utils.GenerateSnowflakeID(),
		UserID: userID,
		PostID: &postID,
	}
	err := s.postRepo.CreatePostStar(&postStar)
	if err != nil {
		if utils.IsPgDuplicateKey(err) {
			return syserror.DuplicateError
		}
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create post star failed")
		return syserror.InternalError
	}
	// 后台完成同步
	go func() {
		// redis的stars+1
		var ctx = context.Background()
		if err := s.postRepo.IncreasePostStat(ctx, postID, "stars"); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post stars failed")
		}
	}()
	return syserror.NoError
}

// 取消收藏帖子
func (s *postService) CancelStarOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	err := s.postRepo.DeletePostStar(userID, postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("delete post star failed")
		return syserror.InternalError
	}
	// 后台完成同步
	go func() {
		// redis的stars-1
		var ctx = context.Background()
		if err := s.postRepo.DecreasePostStat(ctx, postID, "stars"); err != nil {
			utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decrease post stars failed")
		}
	}()
	return syserror.NoError
}

// 查询用户收藏的帖子（分页）
func (s *postService) GetUserStarredPosts(ctx context.Context, userID int64, page, limit int) ([]response.MultiPostData, int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	offset := (page - 1) * limit // 计算偏移量
	posts, total, err := s.postRepo.FindUserStarredPosts(userID, offset, limit)
	if err != nil {
		logger.WithError(err).Error("find user starred posts failed")
		return nil, 0, syserror.InternalError
	}

	// 如果没有收藏，直接返回空列表
	if len(posts) == 0 {
		return []response.MultiPostData{}, 0, syserror.NoError
	}

	// 构建查询的userID数组
	userIDs := make([]int64, len(posts))
	for i, post := range posts {
		userIDs[i] = post.AuthorID
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
			return nil, 0, syserror.InternalError
		}
	}

	// 聚合返回查询结果
	var userMap = resp.Users
	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		var user = userMap[posts[i].AuthorID]
		postDatas[i].UserInfo = response.UserInfo{
			ID:        user.UserId,
			Username:  user.Username,
			Role:      int(user.Role),
			AvatarUrl: user.AvatarUrl,
		}
		postDatas[i].PostData.Post = posts[i]
	}
	return postDatas, total, syserror.NoError
}
