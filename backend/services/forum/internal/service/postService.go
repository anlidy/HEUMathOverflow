package service

import (
	userpb "MathOverflow/api/user"
	"MathOverflow/common/client"
	"MathOverflow/common/config"
	"MathOverflow/common/event"
	common "MathOverflow/common/model"
	"MathOverflow/common/outbox"
	"MathOverflow/common/utils"
	"MathOverflow/services/forum/internal/cache"
	"MathOverflow/services/forum/internal/model"
	syserror "MathOverflow/services/forum/internal/model/error"
	"MathOverflow/services/forum/internal/model/request"
	"MathOverflow/services/forum/internal/model/response"
	"MathOverflow/services/forum/internal/repository"
	"context"
	"encoding/json"
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
	replyRepo      repository.ReplyRepo
	userSnapshot   repository.UserSnapshotRepo
	tokenRepo      repository.ClientTokenRepo
	postCache      *cache.PostCache
	servName       string
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
	userResolver   *userResolver
	statWriter     *statWriter
	fileCleanup    *fileCleanup
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

func NewPostService(cfg config.Config, _ *client.RabbitMQClient, postRepo repository.PostRepo, replyRepo repository.ReplyRepo, fileRepo repository.FileRepo, userSnapshot repository.UserSnapshotRepo, tokenRepo repository.ClientTokenRepo, postCache *cache.PostCache) PostService {
	svc := &postService{
		cfg:          cfg,
		postRepo:     postRepo,
		replyRepo:    replyRepo,
		fileRepo:     fileRepo,
		userSnapshot: userSnapshot,
		tokenRepo:    tokenRepo,
		postCache:    postCache,
		servName:     "Post-Service",
		statWriter:   newStatWriter("Post-Service"),
		fileCleanup:  newFileCleanup("Post-Service", fileRepo),
	}
	svc.userResolver = newUserResolver(userSnapshot, svc.getUserClient)
	return svc
}

func (s *postService) selectRagAnswers(ctx context.Context, postID int64, replies []model.Reply) ([]event.PostAnswer, error) {
	selected := make([]model.Reply, 0)
	selectedIDs := make(map[int64]struct{})
	for _, reply := range replies {
		if reply.Status == model.AuthorSelected || reply.Status == model.TeacherCertified {
			selected = append(selected, reply)
			selectedIDs[reply.ID] = struct{}{}
		}
	}
	if len(selected) < 3 {
		topReplies, err := s.replyRepo.FindTopRepliesByLikes(postID, 3)
		if err != nil {
			return nil, err
		}
		for _, reply := range topReplies {
			if len(selected) >= 3 {
				break
			}
			if _, ok := selectedIDs[reply.ID]; ok {
				continue
			}
			selected = append(selected, reply)
			selectedIDs[reply.ID] = struct{}{}
		}
	}
	if len(selected) > 3 {
		selected = selected[:3]
	}
	answers := make([]event.PostAnswer, 0, len(selected))
	for _, reply := range selected {
		answers = append(answers, event.PostAnswer{
			ReplyID:   reply.ID,
			ReplierID: reply.ReplierID,
			Content:   reply.Content,
		})
	}
	return answers, nil
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
	err = s.postRepo.RunInTx(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&model.Post{}).Create(&post).Error; err != nil {
			return err
		}
		payload := event.ForumPostPayload{
			PostID:     postID,
			AuthorID:   post.AuthorID,
			AuthorName: authorName,
			Title:      post.Title,
			Content:    post.Content,
			Tags:       post.Tags,
			Status:     int(post.Status),
			CreatedAt:  post.CreatedAt,
		}
		eventBody, err := json.Marshal(event.ForumPostEvent{
			Type:      event.ForumPostCreated,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		return s.postRepo.AddOutboxMessageInTx(ctx, tx, string(event.ForumPostCreated), postID, eventBody)
	})
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

	userInfo := s.userResolver.ResolveOne(ctx, post.AuthorID, logger)

	// 聚合返回查询结果
	postData := &response.PostData{
		Post:    post,
		Liked:   liked,
		Starred: starred,
	}

	s.statWriter.Submit("increase post views", func(ctx context.Context) error {
		return s.postRepo.IncreasePostStat(ctx, postID, "views")
	})

	return &userInfo, postData, syserror.NoError
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

	userInfoMap := s.userResolver.ResolveMany(ctx, userIDs, logger)

	// 聚合返回查询结果
	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		postDatas[i].UserInfo = userInfoMap[posts[i].AuthorID]
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
	err = s.postRepo.RunInTx(ctx, func(tx *gorm.DB) error {
		result := tx.Model(&model.Post{}).Where("id = ?", newPost.ID).Updates(newPost)
		if result.Error != nil {
			return result.Error
		}
		payload := event.ForumPostPayload{
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
		eventBody, err := json.Marshal(event.ForumPostEvent{
			Type:      event.ForumPostUpdated,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		return s.postRepo.AddOutboxMessageInTx(ctx, tx, string(event.ForumPostUpdated), postID, eventBody)
	})
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

	s.fileCleanup.SubmitObjectURLs("delete post images", s.cfg.Minio.Bucket, req.DeleteImageURLs)
	return syserror.NoError
}

func (s *postService) SetPostCertified(ctx context.Context, postID int64, req request.PostCertifiedUpdate, role int) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	if role < int(common.Teacher) {
		return syserror.PermissionDeniedError
	}

	post, err := s.postRepo.FindPostByID(postID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find post failed before set certified")
		return syserror.InternalError
	}
	replies, err := s.replyRepo.FindRepliesByPostID(postID)
	if err != nil {
		logger.WithError(err).Error("find post replies failed before set certified")
		return syserror.InternalError
	}
	answers, err := s.selectRagAnswers(ctx, postID, replies)
	if err != nil {
		logger.WithError(err).Error("select rag answers failed before set certified")
		return syserror.InternalError
	}

	err = s.postRepo.RunInTx(ctx, func(tx *gorm.DB) error {
		result := tx.Model(&model.Post{}).Where("id = ?", postID).Update("is_certified", req.IsCertified)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var ragType event.RagEventType = event.RagDataDeleted
		if req.IsCertified {
			ragType = event.RagDataCreated
		}
		eventBody, err := json.Marshal(event.RagDataEvent{
			Type: ragType,
			Payload: event.RagDataPayload{
				PostID:   post.ID,
				AuthorID: post.AuthorID,
				Title:    post.Title,
				Content:  post.Content,
				Tags:     post.Tags,
				Answers:  answers,
			},
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		return outbox.NewStore(tx).AddMessageInTx(ctx, tx, "post", post.ID, string(ragType), "rag-events", json.RawMessage(eventBody))
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("set post certified failed")
		return syserror.InternalError
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
	authorRole := s.userResolver.ResolveOne(ctx, post.AuthorID, logger).Role

	// 验证操作者权限是否低于帖子作者
	if role <= authorRole {
		// 验证当前登录的用户是否为该帖子的作者
		if userID != post.AuthorID {
			return syserror.PermissionDeniedError
		}
	}

	// 删除帖子
	// 附属的回贴会一并删除
	replies, err := s.replyRepo.FindRepliesByPostID(postID)
	if err != nil {
		logger.WithError(err).Error("find post replies failed before delete post")
		return syserror.InternalError
	}
	ragAnswers, err := s.selectRagAnswers(ctx, postID, replies)
	if err != nil {
		logger.WithError(err).Error("select rag answers failed before delete post")
		return syserror.InternalError
	}
	err = s.postRepo.RunInTx(ctx, func(tx *gorm.DB) error {
		result := tx.Where("id = ?", postID).Delete(&model.Post{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		payload := event.ForumPostPayload{PostID: postID}
		eventBody, err := json.Marshal(event.ForumPostEvent{
			Type:      event.ForumPostDeleted,
			Payload:   payload,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		if err := s.postRepo.AddOutboxMessageInTx(ctx, tx, string(event.ForumPostDeleted), postID, eventBody); err != nil {
			return err
		}
		if !post.IsCertified {
			return nil
		}
		ragEventBody, err := json.Marshal(event.RagDataEvent{
			Type: event.RagDataDeleted,
			Payload: event.RagDataPayload{
				PostID:   post.ID,
				AuthorID: post.AuthorID,
				Title:    post.Title,
				Content:  post.Content,
				Tags:     post.Tags,
				Answers:  ragAnswers,
			},
			CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		return outbox.NewStore(tx).AddMessageInTx(ctx, tx, "post", post.ID, string(event.RagDataDeleted), "rag-events", json.RawMessage(ragEventBody))
	})
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

	s.fileCleanup.SubmitObjectURLs("delete post files", s.cfg.Minio.Bucket, post.ImageURLs)

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
	s.statWriter.Submit("increase post likes", func(ctx context.Context) error {
		return s.postRepo.IncreasePostStat(ctx, postID, "likes")
	})
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
	s.statWriter.Submit("decrease post likes", func(ctx context.Context) error {
		return s.postRepo.DecreasePostStat(ctx, postID, "likes")
	})
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
	s.statWriter.Submit("increase post stars", func(ctx context.Context) error {
		return s.postRepo.IncreasePostStat(ctx, postID, "stars")
	})
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
	s.statWriter.Submit("decrease post stars", func(ctx context.Context) error {
		return s.postRepo.DecreasePostStat(ctx, postID, "stars")
	})
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

	userInfoMap := s.userResolver.ResolveMany(ctx, userIDs, logger)

	postDatas := make([]response.MultiPostData, len(posts))
	for i := range posts {
		postDatas[i].UserInfo = userInfoMap[posts[i].AuthorID]
		postDatas[i].PostData.Post = posts[i]
	}
	return postDatas, total, syserror.NoError
}
