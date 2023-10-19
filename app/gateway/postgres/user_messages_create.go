package postgres

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/entity"
)

func (r *UserMessagesRepository) Create(ctx context.Context, message entity.UserMessage) error {
	const (
		operation = "Repository.UserMessagesRepository.Create"
		query     = `
			INSERT INTO user_messages (user_id, message)
				VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`
	)

	_, err := r.Client.Pool.Exec(
		ctx,
		query,
		message.UserID,
		message.Message,
	)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
