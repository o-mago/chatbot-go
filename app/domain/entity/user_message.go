package entity

import (
	"time"
)

type UserMessage struct {
	ID      string
	UserID  string
	Message string

	CreatedAt time.Time
}
