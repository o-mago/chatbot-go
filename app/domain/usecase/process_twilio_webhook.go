package usecase

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/entity"
)

type ProcessTwilioWebhookInput struct {
	PhoneNumber string `json:"phone_number"`
	MessageBody string `json:"message_body"`
}

func (u *UseCase) ProcessTwilioWebhook(ctx context.Context, input ProcessTwilioWebhookInput) error {
	const operation = "UseCase.ProcessTwilioWebhook"

	user, err := u.UsersRepository.GetByPhoneNumber(ctx, input.PhoneNumber)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	err = u.UserMessagesRepository.Create(ctx, entity.UserMessage{
		UserID:  user.ID,
		Message: input.MessageBody,
	})
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
