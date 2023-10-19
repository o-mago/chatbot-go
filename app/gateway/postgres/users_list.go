package postgres

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/entity"
)

func (r *UsersRepository) List(ctx context.Context) ([]entity.User, error) {
	const (
		operation = "Repository.UsersRepository.List"
		query     = `
			SELECT
				id,
				name,
				phone_number,
				created_at
			FROM users
		`
	)

	rows, err := r.Client.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s -> %w", operation, err)
	}
	defer rows.Close()

	var users []entity.User

	for rows.Next() {
		var user entity.User

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.PhoneNumber,
			&user.CreatedAt,
		); err != nil {
			return []entity.User{}, fmt.Errorf("%s -> %w", operation, err)
		}

		users = append(users, user)
	}

	return users, nil
}
