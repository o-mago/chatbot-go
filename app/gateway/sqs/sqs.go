package sqs

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/chatbot-go/app/config"
)

const (
	groupIDWebhooksTwilio = "webhooks-twilio"
)

var queues = &Queues{}

type Queues struct {
	WebhooksTwilio Queue
}

type Queue struct {
	Name    string
	URL     *string
	Handler func(context.Context, []byte, string) error
}

type Client struct {
	client *awssqs.Client
	cfg    config.SQS
}

func New(ctx context.Context, cfg config.SQS, development bool) (*Client, *Enqueuer, error) {
	const operation = "SQS.New"

	loadOptions := []func(*awsconfig.LoadOptions) error{awsconfig.WithRetryMaxAttempts(cfg.SessionMaxRetries)}

	if development {
		loadOptions = append(loadOptions,
			awsconfig.WithRegion("us-east-1"),
			awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
			awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...any) (aws.Endpoint, error) {
					return aws.Endpoint{URL: "http://localhost:4566"}, nil
				},
			)),
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s -> %w", operation, err)
	}

	otelaws.AppendMiddlewares(&awsCfg.APIOptions)

	client := &Client{awssqs.NewFromConfig(awsCfg), cfg}

	err = client.setupQueues(ctx, development)
	if err != nil {
		return nil, nil, fmt.Errorf("%s -> %w", operation, err)
	}

	enqueuer := &Enqueuer{client}

	return client, enqueuer, nil
}

func (c *Client) setupQueues(ctx context.Context, development bool) error {
	const operation = "SQS.Client.setupQueues"

	var err error

	if development {
		for _, queue := range []string{
			c.cfg.WebhooksTwilioQueue,
		} {
			_, err = c.client.CreateQueue(ctx, &awssqs.CreateQueueInput{
				QueueName: aws.String(queue),
				Attributes: map[string]string{
					"FifoQueue": strconv.FormatBool(strings.HasSuffix(queue, ".fifo")),
				},
			})
			if err != nil {
				return fmt.Errorf("%s -> %w", operation, err)
			}
		}
	}

	queues.WebhooksTwilio = Queue{Name: c.cfg.WebhooksTwilioQueue}

	queues.WebhooksTwilio.URL, err = c.getQueueURL(ctx, queues.WebhooksTwilio.Name)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}

func (c *Client) getQueueURL(ctx context.Context, name string) (*string, error) {
	const operation = "SQS.Client.getQueueURL"

	output, err := c.client.GetQueueUrl(ctx, &awssqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("%s (%s) -> %w", operation, name, err)
	}

	return output.QueueUrl, nil
}
