package response

type UserInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"` // 返回前端时需转换为字符串
	AvatarUrl string `json:"avatar_url"`
}
