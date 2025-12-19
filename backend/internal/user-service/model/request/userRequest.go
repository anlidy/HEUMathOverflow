package request

import "MathOverflow/internal/user-service/model"

type UserRegister struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLogin struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"` // 不能加binding, 若传false通不过ShouldBindJSON
}

type UserProfie struct {
	Username string `json:"username" `
}

type UserPassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type UserRole struct {
	ID      int64      `json:"id,string" binding:"required"`
	NewRole model.Role `json:"new_role,string" binding:"required"`
}

type UserDelete struct {
	Password string `json:"password" binding:"required"`
}
