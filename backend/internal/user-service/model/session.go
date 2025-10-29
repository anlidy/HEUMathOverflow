package model

import "time"

// redis sessionID -> Session
type Session struct {
	SessionID string        `json:"-"`
	UserID    int64         `json:"userID"`
	Role      Role          `json:"role"`
	Remember  bool          `json:"remember"`
	TTL       time.Duration `json:"-"`
}
