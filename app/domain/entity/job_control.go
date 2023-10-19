package entity

import (
	"time"

	"github.com/chatbot-go/app/domain/types"
)

type JobControl struct {
	Job            types.Job
	LastSuccessRun *time.Time
}
