package usecase

import (
	"context"
	"fmt"

	"github.com/chatbot-go/app/domain/types"
)

func (u *UseCase) CreateJobsControl(ctx context.Context, jobID types.Job) error {
	const operation = "UseCase.CreateJobsControl"

	err := u.JobsControlRepository.Create(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}

func (u *UseCase) UpdateJobsControl(ctx context.Context, jobID types.Job) error {
	const operation = "UseCase.UpdateJobsControl"

	err := u.JobsControlRepository.Update(ctx, jobID)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
