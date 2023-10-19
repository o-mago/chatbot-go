package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/chatbot-go/app/domain/entity"
	"github.com/chatbot-go/app/domain/erring"
	"github.com/chatbot-go/app/domain/types"
)

const getJobsControlSelectClause = `
SELECT
	job,
	last_success_run
FROM jobs_control
WHERE job = $1
`

func (r *JobsControlRepository) GetByJob(ctx context.Context, job types.Job) (entity.JobControl, error) {
	const operation = "Repository.JobsControl.GetByJob"

	var jobControl entity.JobControl

	err := r.Client.Pool.QueryRow(
		ctx,
		getJobsControlSelectClause,
		job,
	).Scan(
		&jobControl.Job,
		&jobControl.LastSuccessRun,
	)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return entity.JobControl{}, fmt.Errorf("%s -> %w", operation, erring.ErrJobNotFound)
		}

		return entity.JobControl{}, fmt.Errorf("%s -> %w", operation, err)
	}

	return jobControl, nil
}
