package event

import (
	"MathOverflow/internal/common/client"
	"MathOverflow/internal/common/utils"
	"time"
)

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

func PublishUserEvent(mq *client.RabbitMQClient, t UserEventType, payload UserPayload) {
	if mq == nil {
		utils.Logger().Warn("mq is nil, cannot publish user event")
		return
	}

	evt := UserEvent{
		Type:      t,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	if err := mq.PublishEvent(string(t), evt); err != nil {
		utils.Logger().WithError(err).Error("publish user event failed")
	}
}
