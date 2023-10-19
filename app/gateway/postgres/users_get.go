package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/chatbot-go/app/domain/entity"
)

func (r *UsersRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (entity.User, error) {
	const (
		operation = "Repository.UsersRepository.GetByPhoneNumber"
		query     = `
			SELECT
				id,
				name,
				created_at
			FROM users
			WHERE phone_number = $1
		`
	)

	var user entity.User

	err := r.Client.Pool.QueryRow(
		ctx,
		query,
		phoneNumber,
	).Scan(
		&user.ID,
		&user.Name,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return entity.User{}, fmt.Errorf("%s -> %w", operation, err)
		}

		return entity.User{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return user, nil
}
