package postgres

type JobsControlRepository struct {
	*Client
}

func NewJobsControlRepository(client *Client) *JobsControlRepository {
	return &JobsControlRepository{client}
}
