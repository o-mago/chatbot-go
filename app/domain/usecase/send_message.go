package usecase

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/dto"
	"github.com/chatbot-go/app/domain/types"
)

func (u *UseCase) SendMessage(ctx context.Context) error {
	const operation = "UseCase.SendMessage"

	users, err := u.UsersRepository.List(ctx)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	for _, user := range users {
		err = u.TwilioClient.SendMessageTemplate(ctx, dto.SendMessageTemplateInput{
			Provider:          dto.WhatsappProvider,
			DestinationNumber: user.PhoneNumber,
			TemplateID:        types.ListTemplate,
			Variables:         map[string]string{"1": user.Name},
		})
		if err != nil {
			return fmt.Errorf("%s -> %w", operation, err)
		}
	}

	return nil
}
