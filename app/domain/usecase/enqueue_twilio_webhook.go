package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/chatbot-go/app/domain/dto"
)

type EnqueueTwilioWebhookInput struct {
	MessageBody string
	MessageSid  string
	PhoneNumber string
}

func (u *UseCase) EnqueueTwilioWebhook(ctx context.Context, input EnqueueTwilioWebhookInput) error {
	const operation = "UseCase.EnqueueTwilioWebhook"

	webhook := dto.WebhookTwilio{
		MessageSid:  input.MessageSid,
		MessageBody: input.MessageBody,
		PhoneNumber: strings.Split(input.PhoneNumber, ":")[1],
	}

	err := u.Enqueuer.WebhooksTwilio(ctx, webhook)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
