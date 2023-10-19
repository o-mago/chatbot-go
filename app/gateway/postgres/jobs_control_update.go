package postgres

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/types"
)

const updateJobsControlQuery = `
UPDATE jobs_control SET 
	last_success_run = CURRENT_TIMESTAMP
WHERE job = $1
`

func (r *JobsControlRepository) Update(ctx context.Context, job types.Job) error {
	const operation = "Repository.JobsControl.Update"

	_, err := r.Client.Pool.Exec(
		ctx,
		updateJobsControlQuery,
		job,
	)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
