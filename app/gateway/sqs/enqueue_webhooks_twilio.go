package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/chatbot-go/app/domain/dto"
)

func (e *Enqueuer) WebhooksTwilio(ctx context.Context, webhook dto.WebhookTwilio) error {
	const operation = "SQS.Enqueuer.WebhooksTwilio"

	body, err := json.Marshal(webhook)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	_, err = e.client.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    queues.WebhooksTwilio.URL,
		MessageBody: aws.String(string(body)),

		// required for FIFO queues
		MessageGroupId:         aws.String(groupIDWebhooksTwilio),
		MessageDeduplicationId: aws.String(webhook.MessageSid),
	})
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
