package app

import (
	"github.com/chatbot-go/app/config"
	"github.com/chatbot-go/app/domain/usecase"
	"github.com/chatbot-go/app/gateway/client/twilio"
	"github.com/chatbot-go/app/gateway/postgres"
	"github.com/chatbot-go/app/gateway/redis"
	"github.com/chatbot-go/app/gateway/sqs"
)

type App struct {
	UseCase      *usecase.UseCase
	TwilioClient *twilio.Client
}

func New(config config.Config, db *postgres.Client, redisClient *redis.Client, sqsEnqueuer *sqs.Enqueuer) (*App, error) { //nolint: revive
	twilioClient := twilio.NewClient(config.Twilio)

	useCase := &usecase.UseCase{
		AppName:                config.App.Name,
		Enqueuer:               sqsEnqueuer,
		TwilioClient:           twilioClient,
		JobsControlRepository:  postgres.NewJobsControlRepository(db),
		UsersRepository:        postgres.NewUsersRepository(db),
		UserMessagesRepository: postgres.NewUserMessagesRepository(db),
	}

	return &App{
		UseCase:      useCase,
		TwilioClient: twilioClient,
	}, nil
}
