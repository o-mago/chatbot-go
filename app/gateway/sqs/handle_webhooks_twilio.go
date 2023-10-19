package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chatbot-go/app/domain/usecase"
)

func (h *Handler) WebhooksTwilio(ctx context.Context, data []byte, _ string) error {
	const operation = "SQS.Handler.WebhooksTwilio"

	var input usecase.ProcessTwilioWebhookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	err := h.useCase.ProcessTwilioWebhook(ctx, input)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
