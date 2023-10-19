package cronjob

import (
	"context"

	"github.com/chatbot-go/app/domain/types"
)

//go:generate moq -fmt goimports -out handler_mocks.gen.go . useCase

type useCase interface {
	SendMessage(ctx context.Context) error
	CreateJobsControl(ctx context.Context, jobID types.Job) error
	UpdateJobsControl(ctx context.Context, jobID types.Job) error
}

type Handler struct {
	useCase useCase
}

func NewHandler(uc useCase) *Handler {
	handler := Handler{
		useCase: uc,
	}

	return &handler
}
