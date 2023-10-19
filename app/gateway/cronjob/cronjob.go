package cronjob

import (
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/chatbot-go/app/domain/types"
)

func New(useCase useCase) *cli.App {
	handler := NewHandler(useCase)

	return &cli.App{
		Commands: []*cli.Command{
			{
				Name:  string(types.SendMessage),
				Usage: "Send mgm referrals payment requests",
				Action: runJobAction(func(ctx *cli.Context) error {
					return handler.SendMessage(ctx.Context)
				}, handler, types.SendMessage),
			},
		},
	}
}

func runJobAction(action cli.ActionFunc, handler *Handler, jobID types.Job) cli.ActionFunc {
	return func(cliCtx *cli.Context) error {
		err := handler.useCase.CreateJobsControl(cliCtx.Context, jobID)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = action(cliCtx)
		if err != nil {
			slog.ErrorContext(cliCtx.Context, err.Error())

			return fmt.Errorf("%w", err)
		}

		err = handler.useCase.UpdateJobsControl(cliCtx.Context, jobID)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return nil
	}
}
