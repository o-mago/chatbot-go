package sqs

type Enqueuer struct {
	client *Client
}

func NewEnqueuer(client *Client) *Enqueuer {
	return &Enqueuer{
		client: client,
	}
}
