package model

// redis sessionID -> Session
type Session struct {
	UserID   int64 `json:"userID"`
	Role     Role  `json:"role"`
	Remember bool  `json:"remember"`
}
