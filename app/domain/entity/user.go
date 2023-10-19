package entity

import (
	"time"
)

type User struct {
	ID          string
	Name        string
	PhoneNumber string

	CreatedAt time.Time
}
