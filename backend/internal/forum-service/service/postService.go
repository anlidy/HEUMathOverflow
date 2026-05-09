package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/event"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/forum-service/cache"
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
	SetPostCertified(ctx context.Context, postID int64, req request.PostCertifiedUpdate, role int) syserror.Error
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
	userSnapshot   repository.UserSnapshotRepo
	tokenRepo      repository.ClientTokenRepo
	postCache      *cache.PostCache
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

func NewPostService(cfg config.Config, rabbit *client.RabbitMQClient, postRepo repository.PostRepo, fileRepo repository.FileRepo, userSnapshot repository.UserSnapshotRepo, tokenRepo repository.ClientTokenRepo, postCache *cache.PostCache) PostService {
	return &postService{
		cfg:          cfg,
		mq:           rabbit,
		postRepo:     postRepo,
		fileRepo:     fileRepo,
		userSnapshot: userSnapshot,
		tokenRepo:    tokenRepo,
		postCache:    postCache,
		servName:     "Post-Service",
	}
}

func (s *postService) snapshotToUserInfo(snap model.UserSnapshot) response.UserInfo {
	return response.UserInfo{
		ID:        snap.UserID,
		Username:  snap.Username,
		Role:      snap.Role,
		AvatarUrl: snap.AvatarURL,
	}
}

func (s *postService) getUserInfoFromSnapshot(ctx context.Context, userID int64) (*response.UserInfo, bool) {
	if s.userSnapshot == nil {
		return nil, false
	}
	snap, err := s.userSnapshot.FindByID(userID)
	if err != nil {
		return nil, false
	}
	ui := s.snapshotToUserInfo(snap)
	return &ui, true
}

// 创建新帖子
func (s *postService) CreateNewPost(ctx context.Context, userID int64, req request.PostCreate) (int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	token := strings.TrimSpace(req.ClientToken)
	if token == "" {
		return 0, syserror.TokenExpiredError
	}

	// 幂等：优先查 redis 结果缓存（重复请求直接返回首次创建的 id）
	if s.tokenRepo != nil {
		if id, ok, err := s.tokenRepo.GetResult(ctx, repository.ClientTokenPost, userID, token); err == nil && ok {
			return id, syserror.NoError
		} else if err != nil {
			logger.WithError(err).Warn("get post token result from redis failed")
		}
	}

	// 首次创建：必须是后端签发且未过期的 token（redis 中存在 issued key）
	if s.tokenRepo != nil {
		issued, err := s.tokenRepo.IsIssued(ctx, repository.ClientTokenPost, userID, token)
		if err != nil {
			logger.WithError(err).Error("check post token issued failed")
			return 0, syserror.InternalError
		}
		if !issued {
			return 0, syserror.TokenExpiredError
		}
	}

	// token TTL 内防重复：用 redis 锁避免并发重复创建；重复请求优先等待 result 再返回 id
	if s.tokenRepo == nil {
		return 0, syserror.InternalError
	}
	locked, err := s.tokenRepo.TryLock(ctx, repository.ClientTokenPost, userID, token, 5*time.Minute)
	if err != nil {
		logger.WithError(err).Error("lock post token failed")
		return 0, syserror.InternalError
	}
	if !locked {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if id, ok, e := s.tokenRepo.GetResult(ctx, repository.ClientTokenPost, userID, token); e == nil && ok {
				return id, syserror.NoError
			}
			time.Sleep(200 * time.Millisecond)
		}
		return 0, syserror.InProgressError
	}
	defer func() {
		_ = s.tokenRepo.Unlock(ctx, repository.ClientTokenPost, userID, token)
	}()

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
	// 冗余用户信息到 forum DB（异步更新也会覆盖）
	if s.userSnapshot != nil {
		_ = s.userSnapshot.Upsert(model.UserSnapshot{
			UserID:    resp.UserId,
			Username:  resp.Username,
			Role:      int(resp.Role),
			AvatarURL: resp.AvatarUrl,
		})
	}

	// 将临时url提升为正式url
	var postID = utils.GenerateSnowflakeID()
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
			logger.WithError(err).Warn("create post duplicate key")
			return -1, syserror.InternalError
		}
		logger.WithError(err).Error("create post failed")
		return -1, syserror.InternalError
	}
	_ = s.postRepo.BumpHottestCacheVersion(ctx)
	if s.postCache != nil {
		s.postCache.Invalidate(ctx, postID)
	}

	// 缓存 token -> postID 映射（供重试幂等返回）
	if s.tokenRepo != nil {
		if err := s.tokenRepo.SetResult(ctx, repository.ClientTokenPost, userID, token, postID); err != nil {
			logger.WithError(err).Warn("set post token result failed")
		}
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
	hot := false
	if s.postCache != nil {
		if _, h, err := s.postCache.RecordVisit(ctx, postID); err == nil {
			hot = h
		} else {
			logger.WithError(err).Warn("record post visit failed")
		}
	}

	// 查询帖子信息：L1（本地LRU）-> L2（Redis）-> DB（单飞 + 互斥锁防击穿）
	var post model.Post
	var err error
	if s.postCache != nil {
		post, err = s.postCache.GetOrLoad(ctx, postID, hot, func(ctx context.Context) (model.Post, error) {
			return s.postRepo.FindPostByID(postID)
		})
	} else {
		post, err = s.postRepo.FindPostByID(postID)
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		logger.WithError(err).Error("find post failed")
		return nil, nil, syserror.InternalError
	}

	liked, starred := false, false
	if userID != 0 {
		l, s2, err := s.postRepo.FindPostFlags(postID, userID)
		if err != nil {
			logger.WithError(err).Warn("find post flags failed")
		} else {
			liked, starred = l, s2
		}
	}

	// 优先从 forum DB 的用户快照读取，避免强依赖 user-service
	userInfo, ok := s.getUserInfoFromSnapshot(ctx, post.AuthorID)
	if !ok {
		// 缺失时再回退到 gRPC（并回写快照），避免历史数据在首次请求时展示异常
		var userDeleted = false
		userClient, err := s.getUserClient()
		if err == nil {
			resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: post.AuthorID})
			if err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.NotFound {
					userDeleted = true
				} else {
					logger.WithError(err).Warn("get user info via grpc failed, fallback to deleted user")
				}
			} else {
				userInfo = &response.UserInfo{
					ID:        resp.UserId,
					Username:  resp.Username,
					Role:      int(resp.Role),
					AvatarUrl: resp.AvatarUrl,
				}
				if s.userSnapshot != nil {
					_ = s.userSnapshot.Upsert(model.UserSnapshot{
						UserID:    resp.UserId,
						Username:  resp.Username,
						Role:      int(resp.Role),
						AvatarURL: resp.AvatarUrl,
					})
				}
			}
		}
		if userInfo == nil {
			userInfo = &response.UserInfo{Username: "用户已注销"}
		}
		if userDeleted {
			userInfo = &response.UserInfo{Username: "用户已注销"}
		}
	}

	// 聚合返回查询结果
	postData := &response.PostData{
		Post:    post,
		Liked:   liked,
		Starred: starred,
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
	posts, err := s.postRepo.FindManyPosts(ctx, offset, limit, order)
	if err != nil {
		logger.WithError(err).Error("find many posts failed")
		return nil, syserror.InternalError
	}
	// 构建查询的userID数组
	userIDs := make([]int64, len(posts))
	for i, post := range posts {
		userIDs[i] = post.AuthorID
	}

	// 1) 快照优先
	dedup := make(map[int64]struct{}, len(userIDs))
	uniqIDs := make([]int64, 0, len(userIDs))
	for _, uid := range userIDs {
		if _, ok := dedup[uid]; ok {
			continue
		}
		dedup[uid] = struct{}{}
		uniqIDs = append(uniqIDs, uid)
	}

	snapMap := map[int64]model.UserSnapshot{}
	if s.userSnapshot != nil {
		if m, err := s.userSnapshot.FindMapByIDs(uniqIDs); err == nil {
			snapMap = m
		} else {
			logger.WithError(err).Warn("load user snapshots failed")
		}
	}

	// 2) 缺失快照时回退 gRPC 并回写
	grpcMap := map[int64]*userpb.GetUserResponse{}
	missing := make([]int64, 0)
	for _, uid := range uniqIDs {
		if _, ok := snapMap[uid]; ok {
			continue
		}
		missing = append(missing, uid)
	}
	if len(missing) > 0 {
		userClient, err := s.getUserClient()
		if err == nil {
			resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: missing})
			if err != nil {
				logger.WithError(err).Warn("batch get user info via grpc failed")
			} else {
				grpcMap = resp.Users
				if s.userSnapshot != nil {
					for _, u := range grpcMap {
						_ = s.userSnapshot.Upsert(model.UserSnapshot{
							UserID:    u.UserId,
							Username:  u.Username,
							Role:      int(u.Role),
							AvatarURL: u.AvatarUrl,
						})
					}
				}
			}
		}
	}

	// 聚合返回查询结果
	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		if snap, ok := snapMap[posts[i].AuthorID]; ok {
			postDatas[i].UserInfo = s.snapshotToUserInfo(snap)
		} else if u, ok := grpcMap[posts[i].AuthorID]; ok {
			postDatas[i].UserInfo = response.UserInfo{
				ID:        u.UserId,
				Username:  u.Username,
				Role:      int(u.Role),
				AvatarUrl: u.AvatarUrl,
			}
		} else {
			postDatas[i].UserInfo = response.UserInfo{
				Username: "用户已注销",
			}
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
		// fmt.Println(newUrl)
		post.ImageURLs = append(post.ImageURLs, newUrl)
	}

	// 保存帖子信息----需要创建一个新结构体,防止其他字段被覆盖
	newPost := &model.Post{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		ImageURLs: post.ImageURLs,
		Tags:      post.Tags,
	}
	_, err = s.postRepo.UpdatePost(newPost)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("update post failed")
		return syserror.InternalError
	}
	_ = s.postRepo.BumpHottestCacheVersion(ctx)
	if s.postCache != nil {
		s.postCache.Invalidate(ctx, postID)
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

func (s *postService) SetPostCertified(ctx context.Context, postID int64, req request.PostCertifiedUpdate, role int) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	if role < int(common.Teacher) {
		return syserror.PermissionDeniedError
	}

	ok, err := s.postRepo.UpdateColumn(postID, "is_certified", req.IsCertified)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("set post certified failed")
		return syserror.InternalError
	}
	if !ok {
		return syserror.NotFoundError
	}

	_ = s.postRepo.BumpHottestCacheVersion(ctx)
	if s.postCache != nil {
		s.postCache.Invalidate(ctx, postID)
	}
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
	// 获取帖子作者角色（快照优先，缺失时回退 gRPC；NotFound 视为已注销）
	authorRole := 0
	if snap, ok := s.getUserInfoFromSnapshot(ctx, post.AuthorID); ok {
		authorRole = snap.Role
	} else {
		userClient, err := s.getUserClient()
		if err == nil {
			resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: post.AuthorID})
			if err != nil {
				st, ok := status.FromError(err)
				if ok && st.Code() == codes.NotFound {
					authorRole = 0
				} else {
					logger.WithError(err).Warn("get author role via grpc failed, treat as deleted user")
					authorRole = 0
				}
			} else {
				authorRole = int(resp.Role)
				if s.userSnapshot != nil {
					_ = s.userSnapshot.Upsert(model.UserSnapshot{
						UserID:    resp.UserId,
						Username:  resp.Username,
						Role:      int(resp.Role),
						AvatarURL: resp.AvatarUrl,
					})
				}
			}
		}
	}

	// 验证操作者权限是否低于帖子作者
	if role <= authorRole {
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
	_ = s.postRepo.BumpHottestCacheVersion(ctx)
	if s.postCache != nil {
		s.postCache.Invalidate(ctx, postID)
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
	created, err := s.postRepo.CreatePostLike(&postLike)
	if err != nil {
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create post like failed")
		return syserror.InternalError
	}
	if !created {
		// already liked -> idempotent success
		return syserror.NoError
	}
	// redis的likes+1
	if err := s.postRepo.IncreasePostStat(ctx, postID, "likes"); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post likes failed")
	}
	return syserror.NoError
}

// 取消点赞
func (s *postService) CancelLikeOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	deleted, err := s.postRepo.DeletePostLike(userID, postID)
	if err != nil {
		logger.WithError(err).Error("delete post like failed")
		return syserror.InternalError
	}
	if !deleted {
		// already unliked -> idempotent success
		return syserror.NoError
	}
	// redis的likes-1
	if err := s.postRepo.DecreasePostStat(ctx, postID, "likes"); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decrease post likes failed")
	}
	return syserror.NoError
}

// / 收藏接口
// 收藏帖子
func (s *postService) StarOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// Fast-path: already starred -> idempotent success.
	if exists, err := s.postRepo.HasPostStar(userID, postID); err == nil && exists {
		return syserror.NoError
	}
	var postStar = model.PostStar{
		ID:     utils.GenerateSnowflakeID(),
		UserID: userID,
		PostID: &postID,
	}
	created, err := s.postRepo.CreatePostStar(&postStar)
	if err != nil {
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create post star failed")
		return syserror.InternalError
	}
	if !created {
		return syserror.NoError
	}
	// redis的stars+1
	if err := s.postRepo.IncreasePostStat(ctx, postID, "stars"); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("increase post stars failed")
	}
	return syserror.NoError
}

// 取消收藏帖子
func (s *postService) CancelStarOnePost(ctx context.Context, postID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	deleted, err := s.postRepo.DeletePostStar(userID, postID)
	if err != nil {
		logger.WithError(err).Error("delete post star failed")
		return syserror.InternalError
	}
	if !deleted {
		// already unstarred -> idempotent success
		return syserror.NoError
	}
	// redis的stars-1
	if err := s.postRepo.DecreasePostStat(ctx, postID, "stars"); err != nil {
		utils.WithContext(ctx).WithField("service", s.servName).WithError(err).Error("decrease post stars failed")
	}
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

	// 快照优先，缺失时回退 gRPC
	dedup := make(map[int64]struct{}, len(userIDs))
	uniqIDs := make([]int64, 0, len(userIDs))
	for _, uid := range userIDs {
		if _, ok := dedup[uid]; ok {
			continue
		}
		dedup[uid] = struct{}{}
		uniqIDs = append(uniqIDs, uid)
	}

	snapMap := map[int64]model.UserSnapshot{}
	if s.userSnapshot != nil {
		if m, err := s.userSnapshot.FindMapByIDs(uniqIDs); err == nil {
			snapMap = m
		} else {
			logger.WithError(err).Warn("load user snapshots failed")
		}
	}

	grpcMap := map[int64]*userpb.GetUserResponse{}
	missing := make([]int64, 0)
	for _, uid := range uniqIDs {
		if _, ok := snapMap[uid]; ok {
			continue
		}
		missing = append(missing, uid)
	}
	if len(missing) > 0 {
		userClient, err := s.getUserClient()
		if err == nil {
			resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: missing})
			if err != nil {
				logger.WithError(err).Warn("batch get user info via grpc failed")
			} else {
				grpcMap = resp.Users
				if s.userSnapshot != nil {
					for _, u := range grpcMap {
						_ = s.userSnapshot.Upsert(model.UserSnapshot{
							UserID:    u.UserId,
							Username:  u.Username,
							Role:      int(u.Role),
							AvatarURL: u.AvatarUrl,
						})
					}
				}
			}
		}
	}

	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		if snap, ok := snapMap[posts[i].AuthorID]; ok {
			postDatas[i].UserInfo = s.snapshotToUserInfo(snap)
		} else if u, ok := grpcMap[posts[i].AuthorID]; ok {
			postDatas[i].UserInfo = response.UserInfo{
				ID:        u.UserId,
				Username:  u.Username,
				Role:      int(u.Role),
				AvatarUrl: u.AvatarUrl,
			}
		} else {
			postDatas[i].UserInfo = response.UserInfo{
				Username: "用户已注销",
			}
		}
		postDatas[i].PostData.Post = posts[i]
	}
	return postDatas, total, syserror.NoError
}
