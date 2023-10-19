package postgres

type UserMessagesRepository struct {
	*Client
}

func NewUserMessagesRepository(client *Client) *UserMessagesRepository {
	return &UserMessagesRepository{client}
}
