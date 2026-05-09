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
	"MathOverflow/internal/forum-service/search"
	userpb "MathOverflow/proto/user"
	"bytes"
	"context"
	"encoding/json"
	"sync"
)

type SearchService interface {
	SearchPosts(ctx context.Context, req request.SearchRequest) ([]response.MultiPostData, int, syserror.Error)
}

type searchService struct {
	cfg            config.Config
	es             *client.ESClient
	postRepo       repository.PostRepo
	userSnapshot   repository.UserSnapshotRepo
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
	servName       string
}

func NewSearchService(cfg config.Config, es *client.ESClient, postRepo repository.PostRepo, userSnapshot repository.UserSnapshotRepo) SearchService {
	return &searchService{
		cfg:          cfg,
		es:           es,
		postRepo:     postRepo,
		userSnapshot: userSnapshot,
		servName:     "Search-Service",
	}
}

func (s *searchService) getUserClient() (userpb.UserServiceClient, error) {
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

func (s *searchService) SearchPosts(ctx context.Context, req request.SearchRequest) ([]response.MultiPostData, int, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	body := search.BuildQuery(req) // 构建请求体
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		logger.WithError(err).Error("encode es query failed")
		return nil, 0, syserror.InternalError
	}

	res, err := s.es.Client.Search(
		s.es.Client.Search.WithContext(ctx),
		s.es.Client.Search.WithIndex(s.es.Index),
		s.es.Client.Search.WithBody(&buf),
		s.es.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		logger.WithError(err).Error("es search failed")
		return nil, 0, syserror.InternalError
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			// 索引不存在 → 当成“无数据”
			logger.WithField("status", res.Status()).Warn("es index does not build yet")
			return []response.MultiPostData{}, 0, syserror.NotFoundError
		}
		logger.WithField("status", res.Status()).Error("es search response error")
		return nil, 0, syserror.InternalError
	}

	esResp, err := search.ParseSearchResponse(res)
	if err != nil {
		logger.WithError(err).Error("parse es search response failed")
		return nil, 0, syserror.InternalError
	}

	// 解析id数组
	var (
		postIDs = make([]int64, len(esResp.PostMetas))
		userIDs = make([]int64, len(esResp.PostMetas))
	)
	for i, val := range esResp.PostMetas {
		postIDs[i] = val.PostID
		userIDs[i] = val.AuthorID
	}

	// 获取帖子map
	postMap, err := s.postRepo.FindPostMapByIDs(postIDs)
	if err != nil {
		logger.WithError(err).Error("find post map by ids failed")
		return nil, 0, syserror.InternalError
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
		if err != nil {
			logger.WithError(err).Warn("get user client failed when filling missing snapshots")
		} else {
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
	postDatas := make([]response.MultiPostData, len(postMap))
	var idx = 0
	for _, pid := range postIDs {
		if post, ok := postMap[pid]; ok {
			if snap, ok := snapMap[post.AuthorID]; ok {
				postDatas[idx].UserInfo = response.UserInfo{
					ID:        snap.UserID,
					Username:  snap.Username,
					Role:      snap.Role,
					AvatarUrl: snap.AvatarURL,
				}
			} else if user, ok := grpcMap[post.AuthorID]; ok {
				postDatas[idx].UserInfo = response.UserInfo{
					ID:        user.UserId,
					Username:  user.Username,
					Role:      int(user.Role),
					AvatarUrl: user.AvatarUrl,
				}
			} else {
				postDatas[idx].UserInfo = response.UserInfo{
					Username: "用户已注销",
				}
			}

			postDatas[idx].PostData.Post = post
			idx++
		}
	}
	return postDatas, esResp.Total, syserror.NoError
}
