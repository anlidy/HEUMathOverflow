package request

type UserRegister struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type UserProfie struct {
	Username string `json:"username"`
}

type UserPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
