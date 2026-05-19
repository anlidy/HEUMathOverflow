package service

import (
	userpb "MathOverflow/api/user"
	"MathOverflow/services/forum/internal/model"
	"MathOverflow/services/forum/internal/model/response"
	"MathOverflow/services/forum/internal/repository"
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userClientGetter func() (userpb.UserServiceClient, error)

type userResolver struct {
	userSnapshot  repository.UserSnapshotRepo
	getUserClient userClientGetter
}

func newUserResolver(userSnapshot repository.UserSnapshotRepo, getUserClient userClientGetter) *userResolver {
	return &userResolver{userSnapshot: userSnapshot, getUserClient: getUserClient}
}

func (r *userResolver) ResolveOne(ctx context.Context, userID int64, logger *logrus.Entry) response.UserInfo {
	if r != nil && r.userSnapshot != nil {
		if snap, err := r.userSnapshot.FindByID(userID); err == nil {
			return snapshotToUserInfo(snap)
		}
	}

	if r == nil || r.getUserClient == nil {
		return deletedUserInfo(userID)
	}

	userClient, err := r.getUserClient()
	if err != nil {
		if logger != nil {
			logger.WithError(err).Warn("get user client failed when resolving user info")
		}
		return deletedUserInfo(userID)
	}

	resp, err := userClient.GetUserInfo(ctx, &userpb.GetUserRequest{UserId: userID})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return deletedUserInfo(userID)
		}
		if logger != nil {
			logger.WithError(err).Warn("get user info via grpc failed, fallback to deleted user")
		}
		return deletedUserInfo(userID)
	}

	r.upsertSnapshot(resp)
	return grpcUserToUserInfo(resp)
}

func (r *userResolver) ResolveMany(ctx context.Context, userIDs []int64, logger *logrus.Entry) map[int64]response.UserInfo {
	result := make(map[int64]response.UserInfo)
	uniqIDs := dedupUserIDs(userIDs)
	if len(uniqIDs) == 0 {
		return result
	}

	missing := make([]int64, 0, len(uniqIDs))
	if r != nil && r.userSnapshot != nil {
		if snapMap, err := r.userSnapshot.FindMapByIDs(uniqIDs); err == nil {
			for _, uid := range uniqIDs {
				if snap, ok := snapMap[uid]; ok {
					result[uid] = snapshotToUserInfo(snap)
					continue
				}
				missing = append(missing, uid)
			}
		} else {
			if logger != nil {
				logger.WithError(err).Warn("load user snapshots failed")
			}
			missing = append(missing, uniqIDs...)
		}
	} else {
		missing = append(missing, uniqIDs...)
	}

	if len(missing) == 0 {
		return result
	}
	if r == nil || r.getUserClient == nil {
		for _, uid := range missing {
			result[uid] = deletedUserInfo(uid)
		}
		return result
	}

	userClient, err := r.getUserClient()
	if err != nil {
		if logger != nil {
			logger.WithError(err).Warn("get user client failed when filling missing snapshots")
		}
		for _, uid := range missing {
			result[uid] = deletedUserInfo(uid)
		}
		return result
	}

	resp, err := userClient.BatchGetUserInfo(ctx, &userpb.BatchGetUserRequest{UserIds: missing})
	if err != nil {
		if logger != nil {
			logger.WithError(err).Warn("batch get user info via grpc failed")
		}
		for _, uid := range missing {
			result[uid] = deletedUserInfo(uid)
		}
		return result
	}

	for _, uid := range missing {
		user, ok := resp.Users[uid]
		if !ok || user == nil {
			result[uid] = deletedUserInfo(uid)
			continue
		}
		r.upsertSnapshot(user)
		result[uid] = grpcUserToUserInfo(user)
	}

	return result
}

func (r *userResolver) upsertSnapshot(user *userpb.GetUserResponse) {
	if r == nil || r.userSnapshot == nil || user == nil {
		return
	}
	_ = r.userSnapshot.Upsert(model.UserSnapshot{
		UserID:    user.UserId,
		Username:  user.Username,
		Role:      int(user.Role),
		AvatarURL: user.AvatarUrl,
	})
}

func snapshotToUserInfo(snap model.UserSnapshot) response.UserInfo {
	return response.UserInfo{
		ID:        snap.UserID,
		Username:  snap.Username,
		Role:      snap.Role,
		AvatarUrl: snap.AvatarURL,
	}
}

func grpcUserToUserInfo(user *userpb.GetUserResponse) response.UserInfo {
	if user == nil {
		return response.UserInfo{Username: "用户已注销"}
	}
	return response.UserInfo{
		ID:        user.UserId,
		Username:  user.Username,
		Role:      int(user.Role),
		AvatarUrl: user.AvatarUrl,
	}
}

func deletedUserInfo(userID int64) response.UserInfo {
	return response.UserInfo{ID: userID, Username: "用户已注销"}
}

func dedupUserIDs(userIDs []int64) []int64 {
	if len(userIDs) == 0 {
		return nil
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
	return uniqIDs
}
