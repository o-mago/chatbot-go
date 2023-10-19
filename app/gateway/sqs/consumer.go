package sqs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"github.com/chatbot-go/app/domain/erring"
	"github.com/chatbot-go/app/telemetry"
)

var (
	ErrConsumerClosed = errors.New("sqs: consumer closed")

	consumerCtx    context.Context
	consumerCancel context.CancelFunc
)

func (c *Client) Shutdown() {
	consumerCancel()
}

func (c *Client) ListenAndConsume(ctx context.Context, useCase useCase) error {
	const operation = "SQS.Client.ListenAndConsume"

	handler := NewHandler(useCase)

	consumerCtx, consumerCancel = context.WithCancel(ctx)
	group, groupCtx := errgroup.WithContext(consumerCtx)
	group.Go(func() error {
		return c.startConsumers(consumerCtx, queues.WebhooksTwilio, handler.WebhooksTwilio)
	})
	group.Go(func() error {
		<-groupCtx.Done()

		consumerCancel()

		return fmt.Errorf("%s -> %w", operation, ErrConsumerClosed)
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}

func (c *Client) startConsumers(ctx context.Context, queue Queue, handler func(context.Context, []byte, string) error) error {
	const operation = "SQS.Client.startConsumers"

	queue.Handler = handler

	logAttrs := []any{slog.String("sqs_queue_name", queue.Name)}

	slog.DebugContext(ctx, "sqs consumer started", logAttrs...)

	wg := &sync.WaitGroup{}
	wg.Add(c.cfg.MaxWorkers)

	for i := 1; i <= c.cfg.MaxWorkers; i++ {
		go c.consumeMessages(ctx, queue, wg, i)
	}

	wg.Wait()

	slog.DebugContext(ctx, "sqs consumer stopped", logAttrs...)

	return fmt.Errorf("%s -> %w", operation, ErrConsumerClosed)
}

func (c *Client) consumeMessages(ctx context.Context, queue Queue, wg *sync.WaitGroup, id int) {
	const operation = "SQS.Client.consumeMessages"

	logAttrs := []any{
		slog.Int("sqs_worker_id", id),
		slog.String("sqs_queue_name", queue.Name),
	}

	for {
		select {
		case <-ctx.Done():
			wg.Done()

			return
		default:
		}

		workerCtx, workerSpan := telemetry.StartConsumerSpan(context.WithoutCancel(ctx), operation)
		workerSpan.SetAttributes(
			attribute.Int("sqs_worker_id", id),
			attribute.String("sqs_queue_name", queue.Name),
		)

		output, err := c.client.ReceiveMessage(workerCtx, &sqs.ReceiveMessageInput{
			AttributeNames:        []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameAll},
			MessageAttributeNames: []string{string(sqstypes.QueueAttributeNameAll)},
			MaxNumberOfMessages:   int32(c.cfg.MaxMessages),
			QueueUrl:              queue.URL,
			VisibilityTimeout:     int32(c.cfg.VisibilityTimeout.Seconds()),
		})
		if err != nil {
			slog.ErrorContext(
				workerCtx,
				fmt.Errorf("%s -> receive message: %w", operation, err).Error(),
				logAttrs...,
			)

			workerSpan.RecordError(err)
			workerSpan.SetStatus(codes.Error, err.Error())
		}

		if output == nil || len(output.Messages) == 0 {
			workerSpan.End()
			time.Sleep(c.cfg.PollInterval)

			continue
		}

		for _, msg := range output.Messages {
			msgCtx, msgSpan := telemetry.StartInternalSpan(workerCtx, operation+".message")
			msgSpan.SetAttributes(attribute.String("sqs_message_id", *msg.MessageId))

			if err := c.handleMessage(msgCtx, queue, msg); err != nil {
				slog.ErrorContext(
					msgCtx,
					fmt.Errorf("%s -> handle message: %w", operation, err).Error(),
					append(logAttrs, slog.Any("sqs_message", msg))...,
				)

				msgSpan.RecordError(err)
				msgSpan.SetStatus(codes.Error, err.Error())
			}

			msgSpan.End()
		}

		workerSpan.End()
	}
}

func (c *Client) handleMessage(ctx context.Context, queue Queue, msg sqstypes.Message) error {
	const operation = "SQS.Client.handleMessage"

	groupID := msg.Attributes[string(sqstypes.MessageSystemAttributeNameMessageGroupId)]

	err := queue.Handler(ctx, []byte(*msg.Body), groupID)
	if err != nil {
		sqsEventErr := new(erring.SQSEventError)
		if errors.As(err, &sqsEventErr) {
			return c.changeMessageVisibility(ctx, queue, msg, sqsEventErr.NewVisibilityTimeout)
		}

		return fmt.Errorf("%s -> %w", operation, err)
	}

	_, err = c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      queue.URL,
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}

func (c *Client) changeMessageVisibility(ctx context.Context, queue Queue, msg sqstypes.Message, newVisibilityTimeout int32) error {
	const operation = "SQS.Client.changeMessageVisibility"

	_, err := c.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          queue.URL,
		ReceiptHandle:     msg.ReceiptHandle,
		VisibilityTimeout: newVisibilityTimeout,
	})
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
