package usecase

import (
	"context"

	"github.com/chatbot-go/app/domain/dto"
	"github.com/chatbot-go/app/domain/entity"
	"github.com/chatbot-go/app/domain/types"
)

type UseCase struct {
	AppName string

	// Messaging
	Enqueuer enqueuer

	// Clients
	TwilioClient twilioClient

	// Repos
	JobsControlRepository  jobsControlRepository
	UsersRepository        usersRepository
	UserMessagesRepository userMessagesRepository
}

type enqueuer interface {
	WebhooksTwilio(ctx context.Context, webhook dto.WebhookTwilio) error
}

type jobsControlRepository interface {
	Create(ctx context.Context, job types.Job) error
	Update(ctx context.Context, job types.Job) error
	GetByJob(ctx context.Context, job types.Job) (entity.JobControl, error)
}

type usersRepository interface {
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (entity.User, error)
	List(ctx context.Context) ([]entity.User, error)
}

type userMessagesRepository interface {
	Create(ctx context.Context, message entity.UserMessage) error
}

type twilioClient interface {
	SendMessage(ctx context.Context, input dto.SendMessageInput) error
	SendMessageTemplate(ctx context.Context, input dto.SendMessageTemplateInput) error
}
