package service

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/config"
	syserror "MathOverflow/internal/forum-service/model/error"
	"MathOverflow/internal/forum-service/model/request"
	"MathOverflow/internal/forum-service/model/response"
	"MathOverflow/internal/forum-service/repository"
	"MathOverflow/internal/forum-service/search"
	userpb "MathOverflow/proto/user"
	"bytes"
	"context"
	"encoding/json"
	"log"
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
	body := search.BuildQuery(req) // 构建请求体
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		log.Printf("[%s]encode es query err: %v\n", s.servName, err)
		return nil, 0, syserror.InternalError
	}

	res, err := s.es.Client.Search(
		s.es.Client.Search.WithContext(ctx),
		s.es.Client.Search.WithIndex(s.es.Index),
		s.es.Client.Search.WithBody(&buf),
		s.es.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Printf("[%s]es search err: %v", s.servName, err)
		return nil, 0, syserror.InternalError
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s]es search response error: %s", s.servName, res.String())
		return nil, 0, syserror.InternalError
	}

	esResp, err := search.ParseSearchResponse(res)
	if err != nil {
		log.Printf("[%s]%v\n", s.servName, err)
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
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, 0, syserror.InternalError
	}

	// 获取grpc client
	userClient, err := s.getUserClient()
	if err != nil {
		log.Printf("[%s] %v\n", s.servName, err)
		return nil, 0, syserror.NetworkError
	}
	// 调用 UserService
	resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: userIDs})
	if err != nil {
		log.Printf("[%s] 调用 BatchGetUserInfo 失败: %v\n", s.servName, err)
		st, ok := status.FromError(err)
		if !ok {
			log.Println("非 gRPC 错误:", err)
			return nil, 0, syserror.NetworkError
		}
		switch st.Code() {
		case codes.Internal:
			return nil, 0, syserror.InternalError
		case codes.Canceled:
			return nil, 0, syserror.NetworkError
		default:
			log.Println(st.Message())
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
