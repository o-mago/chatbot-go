package postgres

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/types"
)

const createJobsControlQuery = `
INSERT INTO jobs_control (job)
VALUES ($1)
ON CONFLICT DO NOTHING
`

func (r *JobsControlRepository) Create(ctx context.Context, job types.Job) error {
	const operation = "Repository.JobsControl.Create"

	_, err := r.Client.Pool.Exec(
		ctx,
		createJobsControlQuery,
		job,
	)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
