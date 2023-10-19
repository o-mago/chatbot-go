package twilio

import (
	"context"
	"fmt"

	api "github.com/twilio/twilio-go/rest/api/v2010"

	"github.com/chatbot-go/app/domain/dto"
	"github.com/chatbot-go/app/domain/erring"
)

//nolint:revive
func (c *Client) SendMessage(ctx context.Context, input dto.SendMessageInput) error {
	const operation = "Client.Twilio.SendMessage"

	params := &api.CreateMessageParams{}
	params.SetFrom(string(input.Provider) + ":" + c.originNumber)
	params.SetBody(input.Message)
	params.SetTo(string(input.Provider) + ":" + input.DestinationNumber)

	resp, err := c.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("%s -> %w", operation, err)
	}

	// TODO: try to validate a case where SID is empty
	if resp.Sid == nil {
		return fmt.Errorf("%s -> %w", operation, erring.ErrMissingTwilioSid)
	}

	return nil
}
