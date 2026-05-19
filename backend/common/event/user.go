package event

import "time"

type UserEventType string

const (
	UserUpserted UserEventType = "user.upserted"
	UserDeleted  UserEventType = "user.deleted"
)

type UserEvent struct {
	Type      UserEventType `json:"type"`
	Payload   UserPayload   `json:"payload"`
	CreatedAt time.Time     `json:"created_at"`
}

type UserPayload struct {
	UserID    int64  `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Role      int64  `json:"role,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
