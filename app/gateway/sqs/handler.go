package sqs

import (
	"context"

	"github.com/chatbot-go/app/domain/usecase"
)

type Handler struct {
	useCase useCase
}

func NewHandler(uc useCase) *Handler {
	return &Handler{
		useCase: uc,
	}
}

//go:generate moq -rm -out handler_mocks.gen.go . useCase

type useCase interface {
	ProcessTwilioWebhook(ctx context.Context, input usecase.ProcessTwilioWebhookInput) error
}
