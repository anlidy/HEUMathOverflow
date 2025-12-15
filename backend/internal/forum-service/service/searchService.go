package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/utils"
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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SearchService interface {
	SearchPosts(ctx context.Context, req request.SearchRequest) ([]response.MultiPostData, int, syserror.Error)
}

type searchService struct {
	cfg            config.Config
	es             *client.ESClient
	postRepo       repository.PostRepo
	userClient     userpb.UserServiceClient
	userClientOnce sync.Once
	servName       string
}

func NewSearchService(cfg config.Config, es *client.ESClient, postRepo repository.PostRepo) SearchService {
	return &searchService{
		cfg:      cfg,
		es:       es,
		postRepo: postRepo,
		servName: "Search-Service",
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
		}
	}

	// 请求成功
	var userMap = resp.Users

	// 聚合返回查询结果
	postDatas := make([]response.MultiPostData, len(postMap))
	var idx = 0
	for _, pid := range postIDs {
		if post, ok := postMap[pid]; ok {
			var user = userMap[post.AuthorID] // 不存在的用户查询得到空值
			postDatas[idx].UserInfo = response.UserInfo{
				ID:        user.UserId,
				Username:  user.Username,
				Role:      int(user.Role),
				AvatarUrl: user.AvatarUrl,
			}
			postDatas[idx].PostData.Post = post
			idx++
		}
	}
	return postDatas, esResp.Total, syserror.NoError
}
