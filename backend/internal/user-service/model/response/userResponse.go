package response

import "MathOverflow/internal/user-service/model"

type UserInfo struct {
	ID        int64      `json:"id,string"` // 便于前端存储
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      model.Role `json:"role"`
	AvatarUrl string     `json:"avatar_url"`
}

type RPCUserInfo struct {
	ID        int64
	Username  string
	Role      model.Role
	AvatarUrl string
}
