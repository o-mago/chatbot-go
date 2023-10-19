package cronjob

import (
	"context"
	"fmt"
)

func (h *Handler) SendMessage(ctx context.Context) error {
	const operation = "Cronjob.Handler.SendMessage"

	err := h.useCase.SendMessage(ctx)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	return nil
}
