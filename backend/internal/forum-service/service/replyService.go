package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
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
	ChangeReplyStatus(ctx context.Context, req request.ReplyStatusUpdate, userID int64, role int) syserror.Error
}

type replyService struct {
	cfg            config.Config
	postRepo       repository.PostRepo
	replyRepo      repository.ReplyRepo
	fileRepo       repository.FileRepo
	userSnapshot   repository.UserSnapshotRepo
	tokenRepo      repository.ClientTokenRepo
	commentCache   *cache.CommentCache
	postCache      *cache.PostCache
	hotness        cache.Hotness
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

func NewReplyService(cfg config.Config, postRepo repository.PostRepo, replyRepo repository.ReplyRepo, fileRepo repository.FileRepo, userSnapshot repository.UserSnapshotRepo, tokenRepo repository.ClientTokenRepo, commentCache *cache.CommentCache, postCache *cache.PostCache, hotness cache.Hotness) ReplyService {
	return &replyService{
		cfg:          cfg,
		postRepo:     postRepo,
		replyRepo:    replyRepo,
		fileRepo:     fileRepo,
		userSnapshot: userSnapshot,
		tokenRepo:    tokenRepo,
		commentCache: commentCache,
		postCache:    postCache,
		hotness:      hotness,
		servName:     "Reply-Service",
	}
}

func (s *replyService) snapshotToUserInfo(snap model.UserSnapshot) response.UserInfo {
	return response.UserInfo{
		ID:        snap.UserID,
		Username:  snap.Username,
		Role:      snap.Role,
		AvatarUrl: snap.AvatarURL,
	}
}

func (s *replyService) getUserInfoFromSnapshot(userID int64) (*response.UserInfo, bool) {
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

// 创建新回贴
func (s *replyService) CreateNewReply(ctx context.Context, userID int64, req request.ReplyCreate) (int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	token := strings.TrimSpace(req.ClientToken)
	if token == "" {
		return 0, syserror.TokenExpiredError
	}

	if s.tokenRepo == nil {
		return 0, syserror.InternalError
	}

	// 幂等：优先查 redis 结果缓存（重复请求直接返回首次创建的 id）
	if id, ok, err := s.tokenRepo.GetResult(ctx, repository.ClientTokenReply, userID, token); err == nil && ok {
		return id, syserror.NoError
	} else if err != nil {
		logger.WithError(err).Warn("get reply token result from redis failed")
	}

	// 首次创建：必须是后端签发且未过期的 token
	issued, err := s.tokenRepo.IsIssued(ctx, repository.ClientTokenReply, userID, token)
	if err != nil {
		logger.WithError(err).Error("check reply token issued failed")
		return 0, syserror.InternalError
	}
	if !issued {
		return 0, syserror.TokenExpiredError
	}

	// token TTL 内防重复：用 redis 锁避免并发重复创建；重复请求优先等待 result 再返回 id
	locked, err := s.tokenRepo.TryLock(ctx, repository.ClientTokenReply, userID, token, 5*time.Minute)
	if err != nil {
		logger.WithError(err).Error("lock reply token failed")
		return 0, syserror.InternalError
	}
	if !locked {
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if id, ok, e := s.tokenRepo.GetResult(ctx, repository.ClientTokenReply, userID, token); e == nil && ok {
				return id, syserror.NoError
			}
			time.Sleep(200 * time.Millisecond)
		}
		return 0, syserror.InProgressError
	}
	defer func() {
		_ = s.tokenRepo.Unlock(ctx, repository.ClientTokenReply, userID, token)
	}()

	replyID := utils.GenerateSnowflakeID()
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
	err = s.replyRepo.CreateReply(reply)
	if err != nil {
		logger.WithError(err).Error("create reply failed")
		return -1, syserror.InternalError
	}

	// 缓存 token -> replyID 映射（供重试幂等返回）
	if err := s.tokenRepo.SetResult(ctx, repository.ClientTokenReply, userID, token, replyID); err != nil {
		logger.WithError(err).Warn("set reply token result failed")
	}
	if s.commentCache != nil {
		s.commentCache.InvalidatePost(ctx, req.PostID)
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

	// 优先从快照读取用户信息
	userInfo, ok := s.getUserInfoFromSnapshot(reply.ReplierID)
	if !ok {
		// 回退 gRPC 并回写快照
		var userDeleted = false
		userClient, err := s.getUserClient()
		if err == nil {
			resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: reply.ReplierID})
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
	replyData := &response.ReplyData{
		Reply: reply.Reply,
		Liked: reply.Liked,
	}

	return userInfo, replyData, syserror.NoError
}

// 获取分页帖子
func (s *replyService) GetManyReplies(ctx context.Context, postID, userID int64, page, limit int) ([]response.MultiReplyData, int64, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)

	// 评论分页缓存：仅缓存热点帖子的前 1 页（page_size 固定 20）
	if s.commentCache != nil && page == 1 && limit == 20 {
		hot := false
		if s.hotness != nil {
			if h, _, err := s.hotness.IsHot(ctx, postID); err == nil {
				hot = h
			} else {
				logger.WithError(err).Warn("check post hotness failed")
			}
		}
		if hot {
			offset := (page - 1) * limit
			listTTL := 60 * time.Second
			detailTTL := 5 * time.Minute

			ids, total, err := s.commentCache.GetOrLoad(
				ctx,
				postID,
				func(ctx context.Context) ([]int64, int64, []model.Reply, error) {
					replies, total, err := s.replyRepo.FindRepliesPageByPostID(postID, offset, limit)
					if err != nil {
						return nil, 0, nil, err
					}
					ids := make([]int64, 0, len(replies))
					for _, r := range replies {
						ids = append(ids, r.ID)
					}
					return ids, total, replies, nil
				},
				listTTL,
				detailTTL,
			)
			if err == nil && len(ids) == 0 {
				return []response.MultiReplyData{}, total, syserror.NoError
			}
			if err == nil && len(ids) > 0 {
				hits, missing, err := s.commentCache.GetComments(ctx, ids)
				if err != nil {
					logger.WithError(err).Warn("get comment detail cache failed")
				}
				if len(missing) > 0 {
					dbReplies, err := s.replyRepo.FindRepliesByIDs(missing)
					if err != nil {
						logger.WithError(err).Warn("find missing replies by ids failed")
					} else {
						_ = s.commentCache.SetComments(ctx, dbReplies, detailTTL)
						for _, r := range dbReplies {
							hits[r.ID] = r
						}
					}
				}

				ordered := make([]model.Reply, 0, len(ids))
				for _, id := range ids {
					if r, ok := hits[id]; ok {
						ordered = append(ordered, r)
					}
				}

				if total == 0 && len(ordered) > 0 {
					_, t, e := s.replyRepo.FindRepliesPageByPostID(postID, 0, 1)
					if e == nil {
						total = t
						_ = s.commentCache.SetTotal(ctx, postID, total, listTTL)
					}
				}

				likeMap := map[int64]bool{}
				if m, e := s.replyRepo.FindReplyLikeMap(userID, ids); e == nil {
					likeMap = m
				} else {
					logger.WithError(e).Warn("find reply like map failed")
				}

				userIDs := make([]int64, len(ordered))
				for i, r := range ordered {
					userIDs[i] = r.ReplierID
				}
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
				missingUsers := make([]int64, 0)
				for _, uid := range uniqIDs {
					if _, ok := snapMap[uid]; ok {
						continue
					}
					missingUsers = append(missingUsers, uid)
				}
				if len(missingUsers) > 0 {
					userClient, err := s.getUserClient()
					if err == nil {
						resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: missingUsers})
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

				replyDatas := make([]response.MultiReplyData, len(ordered))
				for i := range ordered {
					if snap, ok := snapMap[ordered[i].ReplierID]; ok {
						replyDatas[i].UserInfo = s.snapshotToUserInfo(snap)
					} else if u, ok := grpcMap[ordered[i].ReplierID]; ok {
						replyDatas[i].UserInfo = response.UserInfo{
							ID:        u.UserId,
							Username:  u.Username,
							Role:      int(u.Role),
							AvatarUrl: u.AvatarUrl,
						}
					} else {
						replyDatas[i].UserInfo = response.UserInfo{Username: "用户已注销"}
					}
					replyDatas[i].ReplyData.Reply = ordered[i]
					replyDatas[i].ReplyData.Liked = likeMap[ordered[i].ID]
				}
				return replyDatas, total, syserror.NoError
			}
		}
	}

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
	// 快照优先，缺失回退 gRPC
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
	// 聚合返回查询结果
	replyDatas := make([]response.MultiReplyData, len(replies))
	for i := range replies {
		if snap, ok := snapMap[replies[i].ReplierID]; ok {
			replyDatas[i].UserInfo = s.snapshotToUserInfo(snap)
		} else if u, ok := grpcMap[replies[i].ReplierID]; ok {
			replyDatas[i].UserInfo = response.UserInfo{
				ID:        u.UserId,
				Username:  u.Username,
				Role:      int(u.Role),
				AvatarUrl: u.AvatarUrl,
			}
		} else {
			replyDatas[i].UserInfo = response.UserInfo{
				Username: "用户已注销",
			}
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

	// 保存回帖信息---需要创建一个新结构体,防止其他字段被覆盖
	newReply := &model.Reply{
		ID:        reply.ID,
		Content:   reply.Content,
		VoiceURL:  reply.VoiceURL,
		VoiceText: reply.VoiceText,
		ImageURLs: reply.ImageURLs,
	}
	_, err = s.replyRepo.UpdateReply(newReply)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("update reply failed")
		return syserror.InternalError
	}
	if s.commentCache != nil {
		s.commentCache.InvalidateComment(ctx, replyID)
		s.commentCache.InvalidatePost(ctx, reply.PostID)
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
	if s.commentCache != nil {
		s.commentCache.InvalidateComment(ctx, replyID)
		s.commentCache.InvalidatePost(ctx, reply.PostID)
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
	created, err := s.replyRepo.CreateReplyLike(&replyLike)
	if err != nil {
		if utils.IsPgViolateForeignKey(err) {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("create reply like failed")
		return syserror.InternalError
	}
	if !created {
		// already liked -> idempotent success
		return syserror.NoError
	}
	return syserror.NoError
}

// 取消点赞
func (s *replyService) CancelLikeOneReply(ctx context.Context, replyID, userID int64) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	deleted, err := s.replyRepo.DeleteReplyLike(userID, replyID)
	if err != nil {
		logger.WithError(err).Error("delete reply like failed")
		return syserror.InternalError
	}
	if !deleted {
		// already unliked -> idempotent success
		return syserror.NoError
	}
	return syserror.NoError
}

// 更改回帖状态
func (s *replyService) ChangeReplyStatus(ctx context.Context, req request.ReplyStatusUpdate, userID int64, role int) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	reply, err := s.replyRepo.FindReplyByID(req.ReplyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find reply failed")
		return syserror.InternalError
	}

	post, err := s.postRepo.FindPostByID(reply.PostID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("find post failed")
		return syserror.InternalError
	}

	targetStatus := common.ReplyStatus(req.Status)
	switch targetStatus {
	case common.NotSelected, common.AuthorSelected, common.TeacherCertified:
	default:
		return syserror.PermissionDeniedError
	}

	isTeacher := role >= int(common.Teacher)
	isAuthor := userID == post.AuthorID
	if !isTeacher && !isAuthor {
		return syserror.PermissionDeniedError
	}
	if isAuthor {
		if reply.Status != model.NotSelected && reply.Status != model.AuthorSelected {
			return syserror.PermissionDeniedError
		}
		if targetStatus != common.NotSelected && targetStatus != common.AuthorSelected {
			return syserror.PermissionDeniedError
		}
	}

	var certifiedBy *int64
	if targetStatus != common.NotSelected {
		certifiedBy = &userID
	}

	var postStatus *int
	if reply.Status == model.NotSelected && (targetStatus == common.AuthorSelected || targetStatus == common.TeacherCertified) && post.Status != model.Answered {
		answered := int(model.Answered)
		postStatus = &answered
	}

	err = s.replyRepo.ChangeReplyAndPostStatus(ctx, req.ReplyID, reply.PostID, int(targetStatus), certifiedBy, postStatus)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return syserror.NotFoundError
		}
		logger.WithError(err).Error("change reply status failed")
		return syserror.InternalError
	}

	if s.commentCache != nil {
		s.commentCache.InvalidatePost(ctx, reply.PostID)
	}
	if s.postCache != nil {
		s.postCache.Invalidate(ctx, reply.PostID)
	}
	return syserror.NoError
}
