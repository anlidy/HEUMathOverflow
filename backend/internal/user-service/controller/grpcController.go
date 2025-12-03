package controller

import (
	syserror "MathOverflow/internal/user-service/model/error"
	"MathOverflow/internal/user-service/service"
	userpb "MathOverflow/proto/user"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userService service.UserService
	name        string
}

func NewUserServer(userService service.UserService) UserServer {
	return UserServer{userService: userService, name: "User-gRPC"}
}

// / gRPC 调用接口
// 查询用户信息
func (us *UserServer) GetUserInfo(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	info, err := us.userService.RPCGetUserInfo(ctx, req.UserId)
	switch err {
	case syserror.NotFoundError:
		return nil, status.Error(codes.NotFound, "user not found")
	case syserror.InternalError:
		return nil, status.Error(codes.Internal, "internal error")
	}
	// 正常返回
	return info, nil
}

func (us *UserServer) BatchGetUserInfo(ctx context.Context, req *userpb.BatchGetUserRequest) (*userpb.BatchGetUserResponse, error) {
	infoMap, err := us.userService.RPCBatchGetUserInfo(ctx, req.UserIds)
	switch err {
	case syserror.InternalError:
		return nil, status.Error(codes.Internal, "internal error")
	}
	// 正常返回
	var resp = &userpb.BatchGetUserResponse{
		Users: infoMap,
	}
	return resp, nil
}
