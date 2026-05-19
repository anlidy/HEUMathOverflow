package service

import (
	userpb "MathOverflow/api/user"
	"MathOverflow/common/client"
	"MathOverflow/common/config"
	"MathOverflow/common/utils"
	syserror "MathOverflow/services/forum/internal/model/error"
	"MathOverflow/services/forum/internal/model/request"
	"MathOverflow/services/forum/internal/model/response"
	"MathOverflow/services/forum/internal/repository"
	"MathOverflow/services/forum/internal/search"
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
	userResolver   *userResolver
}

func NewSearchService(cfg config.Config, es *client.ESClient, postRepo repository.PostRepo, userSnapshot repository.UserSnapshotRepo) SearchService {
	svc := &searchService{
		cfg:          cfg,
		es:           es,
		postRepo:     postRepo,
		userSnapshot: userSnapshot,
		servName:     "Search-Service",
	}
	svc.userResolver = newUserResolver(userSnapshot, svc.getUserClient)
	return svc
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

	userInfoMap := s.userResolver.ResolveMany(ctx, userIDs, logger)

	// 聚合返回查询结果
	postDatas := make([]response.MultiPostData, len(postMap))
	var idx = 0
	for _, pid := range postIDs {
		if post, ok := postMap[pid]; ok {
			postDatas[idx].UserInfo = userInfoMap[post.AuthorID]

			postDatas[idx].PostData.Post = post
			idx++
		}
	}
	return postDatas, esResp.Total, syserror.NoError
}
