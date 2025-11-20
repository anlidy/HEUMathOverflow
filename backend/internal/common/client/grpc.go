package client

import (
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	connMap   = make(map[string]*grpc.ClientConn)
	connMutex sync.Mutex
)

func GetGRPCConn(target string) (*grpc.ClientConn, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// 如果已存在，直接复用
	if c, ok := connMap[target]; ok {
		return c, nil
	}

	// 不存在则创建新的连接
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	connMap[target] = conn
	return conn, nil
}
